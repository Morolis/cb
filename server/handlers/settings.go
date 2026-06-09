package handlers

import (
	"net/http"
	"os"

	"github.com/Morolis/cb/server/internal/tlsutil"
	"github.com/Morolis/cb/server/store"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	store      *store.Store
	tlsManager *tlsutil.Manager
}

func NewSettingsHandler(s *store.Store, tls *tlsutil.Manager) *SettingsHandler {
	return &SettingsHandler{store: s, tlsManager: tls}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings := map[string]interface{}{
		"cors_origin": os.Getenv("CB_CORS_ORIGIN"),
	}

	if h.tlsManager != nil {
		settings["tls_enabled"] = true
		settings["tls_cert_path"] = h.tlsManager.CertFile()
		settings["tls_key_path"] = h.tlsManager.KeyFile()
	} else {
		settings["tls_enabled"] = false
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

	c.JSON(http.StatusOK, gin.H{"message": "settings updated. Some changes require restart to take effect."})
}

type uploadTLSRequest struct {
	Cert string `json:"cert" binding:"required"`
	Key  string `json:"key" binding:"required"`
}

func (h *SettingsHandler) UploadTLS(c *gin.Context) {
	var req uploadTLSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.tlsManager == nil {
		// No TLS manager exists, create one
		certDir := "certs"
		os.MkdirAll(certDir, 0700)

		certPath := certDir + "/cert.pem"
		keyPath := certDir + "/key.pem"

		if err := os.WriteFile(certPath, []byte(req.Cert), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write cert"})
			return
		}
		if err := os.WriteFile(keyPath, []byte(req.Key), 0600); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write key"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "TLS certificate saved. Restart server with CB_TLS_CERT=" + certPath + " to enable TLS.",
		})
		return
	}

	// Hot-reload: update files and reload certificate
	if err := h.tlsManager.UpdateFiles(req.Cert, req.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload certificate: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "TLS certificate updated and applied immediately. No restart needed."})
}

func (h *SettingsHandler) GenerateTLS(c *gin.Context) {
	certDir := "certs"
	os.MkdirAll(certDir, 0700)

	certPath, keyPath, err := tlsutil.GenerateSelfSignedCert(certDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate cert: " + err.Error()})
		return
	}

	if h.tlsManager != nil {
		if err := h.tlsManager.Reload(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "generated but failed to reload: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Self-signed certificate generated and applied immediately.",
			"cert_path": certPath,
			"key_path":  keyPath,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Self-signed certificate generated. Restart server with CB_TLS_CERT=" + certPath + " to enable TLS.",
		"cert_path": certPath,
		"key_path":  keyPath,
	})
}
