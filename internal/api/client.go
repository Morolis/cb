package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
)

type Client struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg:        cfg,
	}
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s%s", c.cfg.APIURL(), path)
}

func (c *Client) token() string {
	return c.cfg.Token()
}

func (c *Client) do(req *http.Request, out interface{}) error {
	if c.token() != "" {
		req.Header.Set("Authorization", "Bearer "+c.token())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr models.APIError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
			return fmt.Errorf("api error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("api error (%d): %s", resp.StatusCode, string(body))
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// Auth

func (c *Client) Register(username, password string) (*models.AuthResponse, error) {
	hashed := crypto.PreHashPassword(username, password)
	payload, _ := json.Marshal(models.AuthRequest{Username: username, Password: hashed})
	req, _ := http.NewRequest("POST", c.url("/auth/register"), bytes.NewReader(payload))

	var resp models.AuthResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Login(username, password string) (*models.AuthResponse, error) {
	hashed := crypto.PreHashPassword(username, password)
	payload, _ := json.Marshal(models.AuthRequest{Username: username, Password: hashed})
	req, _ := http.NewRequest("POST", c.url("/auth/login"), bytes.NewReader(payload))

	var resp models.AuthResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Snippets

func (c *Client) CreateSnippet(content, alias, ttl string, encrypted bool) (*models.Snippet, error) {
	return c.CreateSnippetFull(content, alias, "", ttl, encrypted)
}

func (c *Client) CreateSnippetFull(content, alias, description, ttl string, encrypted bool) (*models.Snippet, error) {
	payload, _ := json.Marshal(models.CreateSnippetRequest{
		Content:     content,
		Alias:       alias,
		Description: description,
		TTL:         ttl,
		Encrypted:   encrypted,
	})
	req, _ := http.NewRequest("POST", c.url("/snippets"), bytes.NewReader(payload))

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (c *Client) ListSnippets(limit, offset int) ([]models.SnippetPreview, error) {
	url := fmt.Sprintf("/snippets?limit=%d&offset=%d", limit, offset)
	req, _ := http.NewRequest("GET", c.url(url), nil)

	var resp struct {
		Items []models.SnippetPreview `json:"items"`
	}
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetSnippet(id string) (*models.Snippet, error) {
	req, _ := http.NewRequest("GET", c.url("/snippets/"+id), nil)

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (c *Client) GetSnippetByPrefix(prefix string) (*models.Snippet, error) {
	req, _ := http.NewRequest("GET", c.url("/snippets/prefix/"+prefix), nil)

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (c *Client) GetSnippetByAlias(alias string) (*models.Snippet, error) {
	req, _ := http.NewRequest("GET", c.url("/snippets/alias/"+alias), nil)

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (c *Client) DeleteSnippet(id string) error {
	req, _ := http.NewRequest("DELETE", c.url("/snippets/"+id), nil)
	return c.do(req, nil)
}

func (c *Client) UpdateSnippet(id, content string) (*models.Snippet, error) {
	payload, _ := json.Marshal(map[string]string{"content": content})
	req, _ := http.NewRequest("PUT", c.url("/snippets/"+id), bytes.NewReader(payload))

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (c *Client) ListVersions(snippetID string) ([]models.SnippetVersion, error) {
	req, _ := http.NewRequest("GET", c.url("/snippets/"+snippetID+"/versions"), nil)

	var resp struct {
		Items []models.SnippetVersion `json:"items"`
	}
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) Rollback(snippetID string, versionID uint) (*models.Snippet, error) {
	payload, _ := json.Marshal(map[string]uint{"version_id": versionID})
	req, _ := http.NewRequest("POST", c.url("/snippets/"+snippetID+"/rollback"), bytes.NewReader(payload))

	var snippet models.Snippet
	if err := c.do(req, &snippet); err != nil {
		return nil, err
	}
	return &snippet, nil
}

// Webhooks

type WebhookResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	Events       []string `json:"events"`
	BodyTemplate string   `json:"body_template,omitempty"`
	Active       bool     `json:"active"`
	CreatedAt    string   `json:"created_at"`
}

type WebhookLogResponse struct {
	ID         uint   `json:"id"`
	WebhookID  string `json:"webhook_id"`
	EventType  string `json:"event_type"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"created_at"`
}

func (c *Client) CreateWebhook(name, url string, events []string, bodyTemplate string) (*WebhookResponse, error) {
	payload, _ := json.Marshal(map[string]any{
		"name":          name,
		"url":           url,
		"events":        events,
		"body_template": bodyTemplate,
	})
	req, _ := http.NewRequest("POST", c.url("/webhooks"), bytes.NewReader(payload))

	var resp WebhookResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListWebhooks() ([]WebhookResponse, error) {
	req, _ := http.NewRequest("GET", c.url("/webhooks"), nil)

	var resp struct {
		Items []WebhookResponse `json:"items"`
	}
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) DeleteWebhook(id string) error {
	req, _ := http.NewRequest("DELETE", c.url("/webhooks/"+id), nil)
	return c.do(req, nil)
}

func (c *Client) ToggleWebhook(id string) (bool, error) {
	req, _ := http.NewRequest("PUT", c.url("/webhooks/"+id+"/toggle"), nil)

	var resp struct {
		Active bool `json:"active"`
	}
	if err := c.do(req, &resp); err != nil {
		return false, err
	}
	return resp.Active, nil
}

func (c *Client) ListWebhookLogs(webhookID string) ([]WebhookLogResponse, error) {
	req, _ := http.NewRequest("GET", c.url("/webhooks/"+webhookID+"/logs"), nil)

	var resp struct {
		Items []WebhookLogResponse `json:"items"`
	}
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
