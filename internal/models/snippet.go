package models

import (
	"strings"
	"time"
)

type Snippet struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id,omitempty"`
	Alias       string     `json:"alias,omitempty"`
	Description string     `json:"description,omitempty"`
	Content     string     `json:"content,omitempty"`
	Encrypted   bool       `json:"encrypted"`
	Category    string     `json:"category,omitempty"`
	Language    string     `json:"language,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type SnippetPreview struct {
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

type AuthRequest struct {
	Username string `json:"username" binding:"required,min=2,max=32"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type CreateSnippetRequest struct {
	Content     string   `json:"content" binding:"required"`
	Alias       string   `json:"alias,omitempty"`
	Description string   `json:"description,omitempty"`
	TTL         string   `json:"ttl,omitempty"`
	Encrypted   bool     `json:"encrypted,omitempty"`
	Category    string   `json:"category,omitempty"`
	Language    string   `json:"language,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type SnippetVersion struct {
	ID        uint      `json:"id"`
	SnippetID string    `json:"snippet_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func IsLocalID(id string) bool {
	return strings.HasPrefix(id, "loc_")
}

func ShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func SanitizePreview(text string) string {
	replacer := strings.NewReplacer(
		"\r\n", " ",
		"\n", " ",
		"\r", " ",
		"\t", " ",
	)
	text = replacer.Replace(text)
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}
