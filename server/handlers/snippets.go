package handlers

import (
	"net/http"
	"time"

	"github.com/Morolis/cb/pkg/utils"
	"github.com/Morolis/cb/server/models"
	"github.com/Morolis/cb/server/store"
	"github.com/Morolis/cb/server/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SnippetHandler struct {
	store   *store.Store
	webhook *WebhookHandler
	hub     *ws.Hub
}

func NewSnippetHandler(s *store.Store, wh *WebhookHandler, hub *ws.Hub) *SnippetHandler {
	return &SnippetHandler{store: s, webhook: wh, hub: hub}
}

type createSnippetRequest struct {
	Content     string   `json:"content" binding:"required,max=1048576"`
	Alias       string   `json:"alias,omitempty" binding:"max=64"`
	Description string   `json:"description,omitempty" binding:"max=256"`
	TTL         string   `json:"ttl,omitempty" binding:"max=16"`
	Encrypted   bool     `json:"encrypted,omitempty"`
	Category    string   `json:"category,omitempty" binding:"max=64"`
	Language    string   `json:"language,omitempty" binding:"max=32"`
	Tags        []string `json:"tags,omitempty"`
}

type updateSnippetRequest struct {
	Content     string   `json:"content,omitempty" binding:"max=1048576"`
	Alias       string   `json:"alias,omitempty" binding:"max=64"`
	Description string   `json:"description,omitempty" binding:"max=256"`
	Category    string   `json:"category,omitempty" binding:"max=64"`
	Language    string   `json:"language,omitempty" binding:"max=32"`
	Tags        []string `json:"tags,omitempty"`
}

func (h *SnippetHandler) Create(c *gin.Context) {
	var req createSnippetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	snippet := &models.Snippet{
		ID:          uuid.New().String(),
		UserID:      userID,
		Alias:       req.Alias,
		Description: req.Description,
		Content:     req.Content,
		Encrypted:   req.Encrypted,
		Category:    req.Category,
		Language:    req.Language,
	}
	snippet.SetTags(req.Tags)

	if req.TTL != "" {
		d, err := utils.ParseDuration(req.TTL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ttl: " + err.Error()})
			return
		}
		if d > 0 {
			expires := time.Now().Add(d)
			snippet.ExpiresAt = &expires
		}
	}

	if err := h.store.CreateSnippet(snippet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create snippet"})
		return
	}

	if h.webhook != nil {
		h.webhook.DispatchEvent(userID, "created", snippet)
	}
	if h.hub != nil {
		h.hub.BroadcastToUser(userID, ws.Event{Type: "snippet.created", Payload: snippet})
	}

	c.JSON(http.StatusCreated, snippet)
}

func (h *SnippetHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	snippet, err := h.store.GetSnippet(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}
	if snippet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req updateSnippetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save version if content is changing
	if req.Content != "" && req.Content != snippet.Content {
		if err := h.store.UpdateSnippetWithVersion(snippet, req.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update snippet"})
			return
		}
	}

	if req.Alias != "" {
		snippet.Alias = req.Alias
	}
	if req.Description != "" {
		snippet.Description = req.Description
	}
	if req.Category != "" {
		snippet.Category = req.Category
	}
	if req.Language != "" {
		snippet.Language = req.Language
	}
	if req.Tags != nil {
		snippet.SetTags(req.Tags)
	}

	if err := h.store.UpdateSnippet(snippet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update snippet"})
		return
	}

	if h.webhook != nil {
		h.webhook.DispatchEvent(userID, "updated", snippet)
	}
	if h.hub != nil {
		h.hub.BroadcastToUser(userID, ws.Event{Type: "snippet.updated", Payload: snippet})
	}

	c.JSON(http.StatusOK, snippet)
}

func (h *SnippetHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := parseIntParam(l, 20); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := parseIntParam(o, 0); err == nil {
			offset = parsed
		}
	}

	category := c.Query("category")
	tag := c.Query("tag")

	snippets, total, err := h.store.ListSnippetsFiltered(userID, limit, offset, category, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list snippets"})
		return
	}

	type preview struct {
		ID          string     `json:"id"`
		Alias       string     `json:"alias,omitempty"`
		Description string     `json:"description,omitempty"`
		Preview     string     `json:"preview"`
		Encrypted   bool       `json:"encrypted"`
		Category    string     `json:"category,omitempty"`
		Language    string     `json:"language,omitempty"`
		Tags        []string   `json:"tags,omitempty"`
		ExpiresAt   *time.Time `json:"expires_at,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	items := make([]preview, len(snippets))
	for i, s := range snippets {
		p := s.Content
		if len(p) > 50 {
			p = p[:50] + "..."
		}
		items[i] = preview{
			ID:          s.ID,
			Alias:       s.Alias,
			Description: s.Description,
			Preview:     p,
			Encrypted:   s.Encrypted,
			Category:    s.Category,
			Language:    s.Language,
			Tags:        s.TagList(),
			ExpiresAt:   s.ExpiresAt,
			CreatedAt:   s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}

func (h *SnippetHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	snippet, err := h.store.GetSnippet(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}

	if snippet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, snippet)
}

func (h *SnippetHandler) GetByAlias(c *gin.Context) {
	userID := c.GetString("user_id")
	alias := c.Param("alias")

	snippet, err := h.store.GetSnippetByAlias(userID, alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}

	c.JSON(http.StatusOK, snippet)
}

func (h *SnippetHandler) GetByPrefix(c *gin.Context) {
	userID := c.GetString("user_id")
	prefix := c.Param("prefix")

	snippets, err := h.store.GetSnippetByPrefix(userID, prefix)
	if err != nil || len(snippets) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no matching snippet"})
		return
	}

	if len(snippets) > 1 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "ambiguous prefix, multiple matches",
			"matches": len(snippets),
		})
		return
	}

	c.JSON(http.StatusOK, snippets[0])
}

func (h *SnippetHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	snippet, _ := h.store.GetSnippet(id)

	if err := h.store.DeleteSnippet(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}

	if h.webhook != nil && snippet != nil {
		h.webhook.DispatchEvent(userID, "deleted", snippet)
	}
	if h.hub != nil && snippet != nil {
		h.hub.BroadcastToUser(userID, ws.Event{Type: "snippet.deleted", Payload: snippet})
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *SnippetHandler) ListVersions(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	snippet, err := h.store.GetSnippet(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}
	if snippet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	versions, err := h.store.ListVersions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list versions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": versions})
}

func (h *SnippetHandler) Rollback(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	snippet, err := h.store.GetSnippet(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snippet not found"})
		return
	}
	if snippet.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req struct {
		VersionID uint `json:"version_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version_id is required"})
		return
	}

	if err := h.store.RollbackSnippet(snippet, req.VersionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.webhook != nil {
		h.webhook.DispatchEvent(userID, "updated", snippet)
	}
	if h.hub != nil {
		h.hub.BroadcastToUser(userID, ws.Event{Type: "snippet.updated", Payload: snippet})
	}

	c.JSON(http.StatusOK, snippet)
}

func parseIntParam(s string, defaultVal int) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal, nil
		}
		n = n*10 + int(c-'0')
		if n > 1000 {
			return 1000, nil
		}
	}
	return n, nil
}
