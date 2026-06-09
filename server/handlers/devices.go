package handlers

import (
	"net/http"
	"time"

	"github.com/Morolis/cb/server/models"
	"github.com/Morolis/cb/server/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeviceHandler struct {
	store *store.Store
}

func NewDeviceHandler(s *store.Store) *DeviceHandler {
	return &DeviceHandler{store: s}
}

type heartbeatRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type"` // "cli" or "web"
}

func (h *DeviceHandler) Heartbeat(c *gin.Context) {
	userID := c.GetString("user_id")

	var req heartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "cli"
	}

	device := &models.Device{
		ID:       uuid.New().String(),
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		LastSeen: time.Now(),
	}

	if err := h.store.UpsertDevice(device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *DeviceHandler) ListOnline(c *gin.Context) {
	userID := c.GetString("user_id")

	devices, err := h.store.ListOnlineDevices(userID, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": devices})
}
