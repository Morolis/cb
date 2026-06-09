package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	IsAdmin      bool           `gorm:"default:false" json:"is_admin"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Snippet struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string         `gorm:"index;not null" json:"user_id"`
	Alias       string         `gorm:"index" json:"alias,omitempty"`
	Description string         `json:"description,omitempty"`
	Content     string         `gorm:"not null" json:"content"`
	Encrypted   bool           `gorm:"default:false" json:"encrypted"`
	Category    string         `gorm:"index" json:"category,omitempty"`
	Language    string         `json:"language,omitempty"`
	Tags        string         `json:"-"`
	ExpiresAt   *time.Time     `gorm:"index" json:"expires_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (s *Snippet) TagList() []string {
	if s.Tags == "" {
		return nil
	}
	return strings.Split(s.Tags, ",")
}

func (s *Snippet) SetTags(tags []string) {
	s.Tags = strings.Join(tags, ",")
}

type SnippetVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SnippetID string    `gorm:"index;not null" json:"snippet_id"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Device struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

type SystemSetting struct {
	Key       string    `gorm:"primaryKey;type:varchar(64)" json:"key"`
	Value     string    `gorm:"not null" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Webhook struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID       string    `gorm:"index;not null" json:"user_id"`
	Name         string    `gorm:"not null" json:"name"`
	URL          string    `gorm:"not null" json:"url"`
	Events       string    `gorm:"not null" json:"-"` // comma-separated: created,updated,deleted
	Secret       string    `json:"-"`                 // HMAC signing secret
	BodyTemplate string    `json:"body_template,omitempty"`
	Active       bool      `gorm:"default:true" json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (w *Webhook) EventList() []string {
	if w.Events == "" {
		return nil
	}
	return strings.Split(w.Events, ",")
}

func (w *Webhook) HasEvent(event string) bool {
	for _, e := range w.EventList() {
		if e == event {
			return true
		}
	}
	return false
}

type WebhookLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	WebhookID  string    `gorm:"index;not null" json:"webhook_id"`
	EventType  string    `gorm:"not null" json:"event_type"`
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response,omitempty"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
