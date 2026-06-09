package main

import (
	"crypto/rand"
	"crypto/tls"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Morolis/cb/server/handlers"
	"github.com/Morolis/cb/server/internal/tlsutil"
	"github.com/Morolis/cb/server/middleware"
	"github.com/Morolis/cb/server/store"
	"github.com/Morolis/cb/server/ws"
	"github.com/gin-gonic/gin"
)

//go:embed all:web/dist
var webFS embed.FS

func main() {
	jwtSecret := os.Getenv("CB_JWT_SECRET")

	dbPath := os.Getenv("CB_DB_PATH")
	if dbPath == "" {
		dbPath = "cb.db"
	}

	addr := os.Getenv("CB_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	tlsCert := os.Getenv("CB_TLS_CERT")
	tlsKey := os.Getenv("CB_TLS_KEY")
	tlsAuto := os.Getenv("CB_TLS_AUTO")

	// Auto-generate self-signed cert if requested
	if tlsAuto == "true" && (tlsCert == "" || tlsKey == "") {
		certDir := os.Getenv("CB_TLS_DIR")
		if certDir == "" {
			certDir = "."
		}
		c, k, err := tlsutil.GenerateSelfSignedCert(certDir)
		if err != nil {
			log.Fatalf("Failed to generate TLS cert: %v", err)
		}
		tlsCert = c
		tlsKey = k
		fmt.Printf("Auto-generated TLS cert: %s\n", certDir)
	}

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer s.Close()

	if jwtSecret == "" {
		jwtSecret, _ = s.GetSetting("jwt_secret")
		if jwtSecret == "" {
			jwtSecret = generateRandomSecret()
			s.SetSetting("jwt_secret", jwtSecret)
			fmt.Println("Auto-generated JWT secret (saved to database). Set CB_JWT_SECRET env to override.")
		}
	}

	hub := ws.NewHub()
	go hub.Run()

	// TLS Manager (supports hot-reload via GetCertificate)
	var tlsManager *tlsutil.Manager
	if tlsCert != "" && tlsKey != "" {
		tlsManager, err = tlsutil.NewManager(tlsCert, tlsKey)
		if err != nil {
			log.Printf("Warning: failed to load TLS cert: %v (running without TLS)", err)
			tlsManager = nil
		}
	}

	authHandler := handlers.NewAuthHandler(s, jwtSecret)
	webhookHandler := handlers.NewWebhookHandler(s)
	snippetHandler := handlers.NewSnippetHandler(s, webhookHandler, hub)
	deviceHandler := handlers.NewDeviceHandler(s)
	wsHandler := handlers.NewWSHandler(hub)
	settingsHandler := handlers.NewSettingsHandler(s, tlsManager)
	adminHandler := handlers.NewAdminHandler(s)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/auth/register", middleware.RateLimit(5, 5), authHandler.Register)
		v1.POST("/auth/login", middleware.RateLimit(10, 10), authHandler.Login)
	}

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtSecret, s))
	{
		protected.POST("/snippets", snippetHandler.Create)
		protected.GET("/snippets", snippetHandler.List)
		protected.GET("/snippets/alias/:alias", snippetHandler.GetByAlias)
		protected.GET("/snippets/prefix/:prefix", snippetHandler.GetByPrefix)
		protected.GET("/snippets/:id", snippetHandler.Get)
		protected.PUT("/snippets/:id", snippetHandler.Update)
		protected.DELETE("/snippets/:id", snippetHandler.Delete)
		protected.GET("/snippets/:id/versions", snippetHandler.ListVersions)
		protected.POST("/snippets/:id/rollback", snippetHandler.Rollback)
		protected.POST("/devices/heartbeat", deviceHandler.Heartbeat)
		protected.GET("/devices", deviceHandler.ListOnline)

		// Webhooks
		protected.POST("/webhooks", webhookHandler.Create)
		protected.GET("/webhooks", webhookHandler.List)
		protected.DELETE("/webhooks/:id", webhookHandler.Delete)
		protected.PUT("/webhooks/:id/toggle", webhookHandler.Toggle)
		protected.GET("/webhooks/:id/logs", webhookHandler.ListLogs)

		protected.PUT("/account/password", adminHandler.ChangePassword)
	}

	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(jwtSecret, s), middleware.AdminMiddleware())
	{
		admin.GET("/system", adminHandler.SystemInfo)
		admin.GET("/settings", settingsHandler.GetSettings)
		admin.PUT("/settings", settingsHandler.UpdateSettings)
		admin.POST("/settings/tls", settingsHandler.UploadTLS)
		admin.POST("/settings/tls/generate", settingsHandler.GenerateTLS)
		admin.GET("/users", adminHandler.ListUsers)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)
		admin.PUT("/users/:id/admin", adminHandler.ToggleAdmin)
		admin.PUT("/users/:id/password", adminHandler.ResetUserPassword)
	}

	wsGroup := v1.Group("/ws")
	wsGroup.Use(middleware.AuthMiddlewareOrQuery(jwtSecret, s))
	{
		wsGroup.GET("", wsHandler.HandleWS)
	}

	// Serve embedded Vue frontend
	distFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Printf("Warning: web/dist not embedded, frontend disabled: %v", err)
	} else {
		indexHTML, _ := fs.ReadFile(distFS, "index.html")
		fileServer := http.StripPrefix("/", http.FileServer(http.FS(distFS)))

		r.GET("/assets/*filepath", gin.WrapH(fileServer))

		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if c.Request.Method != "GET" ||
				strings.HasPrefix(path, "/v1") ||
				strings.HasPrefix(path, "/health") {
				c.JSON(404, gin.H{"error": "not found"})
				return
			}
			c.Data(200, "text/html", indexHTML)
		})
	}

	if tlsManager != nil {
		// TLS with hot-reload: GetCertificate reads cert on each connection
		fmt.Printf("cb server starting on %s (TLS)\n", addr)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		tlsConfig := &tls.Config{
			GetCertificate: tlsManager.GetCertificate,
		}
		tlsListener := tls.NewListener(listener, tlsConfig)

		if err := http.Serve(tlsListener, r); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		fmt.Printf("cb server starting on %s\n", addr)
		fmt.Println("WARNING: Running without TLS. Use CB_TLS_CERT and CB_TLS_KEY for production.")
		if err := r.Run(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

func generateRandomSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
