package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/Morolis/cb/server/models"
	"github.com/Morolis/cb/server/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	store *store.Store
}

func NewWebhookHandler(s *store.Store) *WebhookHandler {
	return &WebhookHandler{store: s}
}

type createWebhookRequest struct {
	Name         string   `json:"name" binding:"required,max=64"`
	URL          string   `json:"url" binding:"required,url,max=2048"`
	Events       []string `json:"events" binding:"required"`
	Secret       string   `json:"secret,omitempty" binding:"max=128"`
	BodyTemplate string   `json:"body_template,omitempty" binding:"max=8192"`
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req createWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isPrivateURL(req.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook URL cannot point to a private/internal address"})
		return
	}

	// Validate events
	validEvents := map[string]bool{"created": true, "updated": true, "deleted": true}
	for _, e := range req.Events {
		if !validEvents[e] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event: " + e + " (valid: created, updated, deleted)"})
			return
		}
	}

	webhook := &models.Webhook{
		ID:           uuid.New().String(),
		UserID:       c.GetString("user_id"),
		Name:         req.Name,
		URL:          req.URL,
		Events:       strings.Join(req.Events, ","),
		Secret:       req.Secret,
		BodyTemplate: req.BodyTemplate,
		Active:       true,
	}

	if err := h.store.CreateWebhook(webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            webhook.ID,
		"name":          webhook.Name,
		"url":           webhook.URL,
		"events":        req.Events,
		"body_template": webhook.BodyTemplate,
		"active":        true,
	})
}

func (h *WebhookHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	webhooks, err := h.store.ListWebhooks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list webhooks"})
		return
	}

	type webhookView struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		URL          string    `json:"url"`
		Events       []string  `json:"events"`
		BodyTemplate string    `json:"body_template,omitempty"`
		Active       bool      `json:"active"`
		CreatedAt    time.Time `json:"created_at"`
	}

	items := make([]webhookView, len(webhooks))
	for i, w := range webhooks {
		items[i] = webhookView{
			ID:           w.ID,
			Name:         w.Name,
			URL:          w.URL,
			Events:       w.EventList(),
			BodyTemplate: w.BodyTemplate,
			Active:       w.Active,
			CreatedAt:    w.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.store.DeleteWebhook(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook deleted"})
}

func (h *WebhookHandler) Toggle(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	webhook, err := h.store.GetWebhook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	if webhook.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	webhook.Active = !webhook.Active
	if err := h.store.UpdateWebhook(webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"active": webhook.Active})
}

func (h *WebhookHandler) ListLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	webhook, err := h.store.GetWebhook(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}
	if webhook.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	logs, err := h.store.ListWebhookLogs(id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": logs})
}

func (h *WebhookHandler) DispatchEvent(userID, eventType string, snippet *models.Snippet) {
	go func() {
		webhooks, err := h.store.GetActiveWebhooksForUser(userID)
		if err != nil {
			log.Printf("webhook dispatch error: %v", err)
			return
		}

		data := webhookTemplateData{
			Event:    eventType,
			DateTime: time.Now().UTC().Format(time.RFC3339),
			Snippet:  snippet,
		}

		for _, wh := range webhooks {
			if !wh.HasEvent(eventType) {
				continue
			}
			go h.sendWebhook(&wh, eventType, data)
		}
	}()
}

type webhookTemplateData struct {
	Event    string
	DateTime string
	Snippet  *models.Snippet
}

func templateJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (h *WebhookHandler) sendWebhook(wh *models.Webhook, eventType string, data webhookTemplateData) {
	var body []byte

	if wh.BodyTemplate != "" {
		tmpl, err := template.New("webhook").Funcs(template.FuncMap{
			"json": templateJSON,
		}).Parse(wh.BodyTemplate)
		if err != nil {
			h.saveLog(wh.ID, eventType, 0, "", "template parse error: "+err.Error())
			return
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			h.saveLog(wh.ID, eventType, 0, "", "template execute error: "+err.Error())
			return
		}
		body = buf.Bytes()
	} else {
		payload := map[string]interface{}{
			"event":    data.Event,
			"snippet":  data.Snippet,
			"datetime": data.DateTime,
		}
		body, _ = json.Marshal(payload)
	}

	req, err := http.NewRequest("POST", wh.URL, bytes.NewReader(body))
	if err != nil {
		h.saveLog(wh.ID, eventType, 0, "", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Cb-Event", eventType)
	req.Header.Set("X-Cb-Delivery", uuid.New().String())

	if wh.Secret != "" {
		mac := hmac.New(sha256.New, []byte(wh.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Cb-Signature", "sha256="+sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.saveLog(wh.ID, eventType, 0, "", err.Error())
		return
	}
	defer resp.Body.Close()

	h.saveLog(wh.ID, eventType, resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode), "")
}

func isPrivateURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	host := u.Hostname()
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified()
	}
	// Block common internal hostnames
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".internal") || strings.HasSuffix(lower, ".local") {
		return true
	}
	return false
}

func (h *WebhookHandler) saveLog(webhookID, eventType string, statusCode int, response, errMsg string) {
	h.store.CreateWebhookLog(&models.WebhookLog{
		WebhookID:  webhookID,
		EventType:  eventType,
		StatusCode: statusCode,
		Response:   response,
		Error:      errMsg,
	})
}
