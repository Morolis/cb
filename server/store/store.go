package store

import (
	"fmt"
	"os"
	"time"

	"github.com/Morolis/cb/server/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db     *gorm.DB
	dbPath string
}

func New(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Snippet{},
		&models.SnippetVersion{},
		&models.Device{},
		&models.SystemSetting{},
		&models.Webhook{},
		&models.WebhookLog{},
	); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	s := &Store{db: db, dbPath: dbPath}
	go s.cleanupExpiredLoop()

	return s, nil
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// User operations

func (s *Store) CreateUser(user *models.User) error {
	return s.db.Create(user).Error
}

func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) UserCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (s *Store) ListUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) DeleteUser(id string) error {
	result := s.db.Delete(&models.User{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return result.Error
}

func (s *Store) UpdateUserPassword(id, newHash string) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("password_hash", newHash).Error
}

func (s *Store) SetUserAdmin(id string, isAdmin bool) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("is_admin", isAdmin).Error
}

func (s *Store) IsFirstUser() (bool, error) {
	count, err := s.UserCount()
	return count == 0, err
}

// SystemSetting operations

func (s *Store) GetSetting(key string) (string, error) {
	var setting models.SystemSetting
	if err := s.db.First(&setting, "key = ?", key).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *Store) SetSetting(key, value string) error {
	return s.db.Save(&models.SystemSetting{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
	}).Error
}

func (s *Store) ListSettings() ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Store) DeleteSetting(key string) error {
	return s.db.Delete(&models.SystemSetting{}, "key = ?", key).Error
}

// Snippet operations

func (s *Store) CreateSnippet(snippet *models.Snippet) error {
	return s.db.Create(snippet).Error
}

func (s *Store) GetSnippet(id string) (*models.Snippet, error) {
	var snippet models.Snippet
	if err := s.db.First(&snippet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (s *Store) GetSnippetByPrefix(userID, prefix string) ([]models.Snippet, error) {
	var snippets []models.Snippet
	if err := s.db.Where("user_id = ? AND id LIKE ?", userID, prefix+"%").
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").Limit(5).Find(&snippets).Error; err != nil {
		return nil, err
	}
	return snippets, nil
}

func (s *Store) GetSnippetByAlias(userID, alias string) (*models.Snippet, error) {
	var snippet models.Snippet
	if err := s.db.Where("user_id = ? AND alias = ?", userID, alias).
		Order("created_at DESC").First(&snippet).Error; err != nil {
		return nil, err
	}
	return &snippet, nil
}

func (s *Store) ListSnippets(userID string, limit, offset int) ([]models.Snippet, error) {
	var snippets []models.Snippet
	query := s.db.Where("user_id = ?", userID)
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&snippets).Error; err != nil {
		return nil, err
	}
	return snippets, nil
}

func (s *Store) ListSnippetsFiltered(userID string, limit, offset int, category, tag string) ([]models.Snippet, int64, error) {
	query := s.db.Where("user_id = ?", userID)
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	var total int64
	query.Model(&models.Snippet{}).Count(&total)

	var snippets []models.Snippet
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&snippets).Error; err != nil {
		return nil, 0, err
	}
	return snippets, total, nil
}

func (s *Store) SnippetCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.Snippet{}).Count(&count).Error
	return count, err
}

func (s *Store) UpdateSnippet(snippet *models.Snippet) error {
	return s.db.Save(snippet).Error
}

func (s *Store) UpdateSnippetWithVersion(snippet *models.Snippet, newContent string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		version := models.SnippetVersion{
			SnippetID: snippet.ID,
			Content:   snippet.Content,
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		snippet.Content = newContent
		snippet.UpdatedAt = time.Now()
		return tx.Save(snippet).Error
	})
}

func (s *Store) DeleteSnippet(id, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Snippet{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("snippet not found")
	}
	return result.Error
}

// Version operations

func (s *Store) CreateVersion(version *models.SnippetVersion) error {
	return s.db.Create(version).Error
}

func (s *Store) ListVersions(snippetID string) ([]models.SnippetVersion, error) {
	var versions []models.SnippetVersion
	if err := s.db.Where("snippet_id = ?", snippetID).
		Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func (s *Store) RollbackSnippet(snippet *models.Snippet, versionID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var version models.SnippetVersion
		if err := tx.Where("id = ? AND snippet_id = ?", versionID, snippet.ID).First(&version).Error; err != nil {
			return fmt.Errorf("version not found")
		}
		// Archive current content
		archive := models.SnippetVersion{
			SnippetID: snippet.ID,
			Content:   snippet.Content,
		}
		if err := tx.Create(&archive).Error; err != nil {
			return err
		}
		snippet.Content = version.Content
		snippet.UpdatedAt = time.Now()
		return tx.Save(snippet).Error
	})
}

// Device operations

func (s *Store) UpsertDevice(device *models.Device) error {
	var existing models.Device
	err := s.db.Where("user_id = ? AND name = ?", device.UserID, device.Name).First(&existing).Error
	if err == nil {
		existing.LastSeen = device.LastSeen
		existing.Type = device.Type
		return s.db.Save(&existing).Error
	}
	return s.db.Create(device).Error
}

func (s *Store) ListOnlineDevices(userID string, since time.Duration) ([]models.Device, error) {
	var devices []models.Device
	cutoff := time.Now().Add(-since)
	if err := s.db.Where("user_id = ? AND last_seen > ?", userID, cutoff).
		Order("last_seen DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *Store) ListAllDevices(userID string) ([]models.Device, error) {
	var devices []models.Device
	if err := s.db.Where("user_id = ?", userID).
		Order("last_seen DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *Store) DeviceCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.Device{}).Count(&count).Error
	return count, err
}

func (s *Store) CleanupExpired() (int64, error) {
	result := s.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&models.Snippet{})
	return result.RowsAffected, result.Error
}

func (s *Store) DBSize() int64 {
	if s.dbPath == "" {
		return 0
	}
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (s *Store) cleanupExpiredLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		count, err := s.CleanupExpired()
		if err != nil {
			fmt.Printf("cleanup error: %v\n", err)
		}
		if count > 0 {
			fmt.Printf("cleaned up %d expired snippets\n", count)
		}
	}
}

// Webhook operations

func (s *Store) CreateWebhook(webhook *models.Webhook) error {
	return s.db.Create(webhook).Error
}

func (s *Store) GetWebhook(id string) (*models.Webhook, error) {
	var w models.Webhook
	if err := s.db.First(&w, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListWebhooks(userID string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (s *Store) UpdateWebhook(webhook *models.Webhook) error {
	return s.db.Save(webhook).Error
}

func (s *Store) DeleteWebhook(id, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Webhook{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}
	return result.Error
}

func (s *Store) GetActiveWebhooksForUser(userID string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	if err := s.db.Where("user_id = ? AND active = ?", userID, true).Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (s *Store) CreateWebhookLog(log *models.WebhookLog) error {
	return s.db.Create(log).Error
}

func (s *Store) ListWebhookLogs(webhookID string, limit int) ([]models.WebhookLog, error) {
	var logs []models.WebhookLog
	if err := s.db.Where("webhook_id = ?", webhookID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
