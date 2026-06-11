package handlers

import (
	"net/http"
	"os"

	"github.com/Morolis/cb/server/internal/tlsutil"
	"github.com/Morolis/cb/server/store"
	"github.com/gin-gonic/gin"
)

type StartHTTPSFunc func(*tlsutil.Manager)
type StopHTTPSFunc func()

type SettingsHandler struct {
	store      *store.Store
	tlsManager *tlsutil.Manager
	certDir    string
	startHTTPS StartHTTPSFunc
	stopHTTPS  StopHTTPSFunc
}

func NewSettingsHandler(s *store.Store, tls *tlsutil.Manager, certDir string, startHTTPS StartHTTPSFunc, stopHTTPS StopHTTPSFunc) *SettingsHandler {
	return &SettingsHandler{store: s, tlsManager: tls, certDir: certDir, startHTTPS: startHTTPS, stopHTTPS: stopHTTPS}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings := map[string]interface{}{
		"cors_origin": os.Getenv("CB_CORS_ORIGIN"),
		"tls_enabled": h.tlsManager != nil,
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

type updateSettingsRequest struct {
	CORSOrigin string `json:"cors_origin,omitempty"`
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CORSOrigin != "" {
		os.Setenv("CB_CORS_ORIGIN", req.CORSOrigin)
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated."})
}

func (h *SettingsHandler) EnableTLS(c *gin.Context) {
	if h.tlsManager != nil {
		c.JSON(http.StatusOK, gin.H{"message": "TLS is already enabled."})
		return
	}

	os.MkdirAll(h.certDir, 0700)

	certPath, keyPath, err := tlsutil.GenerateSelfSignedCert(h.certDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate cert: " + err.Error()})
		return
	}

	m, err := tlsutil.NewManager(certPath, keyPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cert: " + err.Error()})
		return
	}

	h.tlsManager = m
	if h.startHTTPS != nil {
		h.startHTTPS(m)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "HTTPS enabled. HTTP requests will be redirected to HTTPS.",
		"cert_path": certPath,
		"key_path":  keyPath,
	})
}

func (h *SettingsHandler) DisableTLS(c *gin.Context) {
	if h.tlsManager == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "HTTPS is not enabled"})
		return
	}

	if h.stopHTTPS != nil {
		h.stopHTTPS()
	}

	h.tlsManager = nil
	c.JSON(http.StatusOK, gin.H{"message": "HTTPS disabled. Server is now HTTP only."})
}
