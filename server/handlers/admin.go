package handlers

import (
	"net/http"
	"time"

	"github.com/Morolis/cb/server/store"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	store     *store.Store
	startedAt time.Time
}

func NewAdminHandler(s *store.Store) *AdminHandler {
	return &AdminHandler{store: s, startedAt: time.Now()}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.store.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	type userView struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		IsAdmin   bool      `json:"is_admin"`
		CreatedAt time.Time `json:"created_at"`
	}

	items := make([]userView, len(users))
	for i, u := range users {
		items[i] = userView{
			ID:        u.ID,
			Username:  u.Username,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	targetID := c.Param("id")
	currentUserID := c.GetString("user_id")

	if targetID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	if err := h.store.DeleteUser(targetID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.store.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong current password"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.store.UpdateUserPassword(userID, string(newHash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed. Please login again."})
}

func (h *AdminHandler) ToggleAdmin(c *gin.Context) {
	targetID := c.Param("id")
	currentUserID := c.GetString("user_id")

	if targetID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own admin status"})
		return
	}

	user, err := h.store.GetUserByID(targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	newAdmin := !user.IsAdmin
	if err := h.store.SetUserAdmin(targetID, newAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	msg := "user promoted to admin"
	if !newAdmin {
		msg = "admin privileges revoked"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg, "is_admin": newAdmin})
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	targetID := c.Param("id")

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.store.UpdateUserPassword(targetID, string(newHash)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

func (h *AdminHandler) SystemInfo(c *gin.Context) {
	userCount, _ := h.store.UserCount()
	snippetCount, _ := h.store.SnippetCount()
	deviceCount, _ := h.store.DeviceCount()
	dbSize := h.store.DBSize()
	uptime := time.Since(h.startedAt)

	c.JSON(http.StatusOK, gin.H{
		"user_count":     userCount,
		"snippet_count":  snippetCount,
		"device_count":   deviceCount,
		"db_size_bytes":  dbSize,
		"uptime_seconds": int(uptime.Seconds()),
		"started_at":     h.startedAt.Format(time.RFC3339),
	})
}
