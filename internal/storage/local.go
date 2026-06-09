package storage

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type LocalDB struct {
	db *gorm.DB
}

type snippetCache struct {
	ID          string `gorm:"primaryKey"`
	Alias       string `gorm:"index"`
	Description string
	Content     string
	Encrypted   bool
	Category    string `gorm:"index"`
	Language    string
	Tags        string     // comma-separated
	ExpiresAt   *time.Time `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (snippetCache) TableName() string {
	return "snippet_cache"
}

func NewLocalDB(path string) (*LocalDB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&snippetCache{}); err != nil {
		return nil, err
	}

	return &LocalDB{db: db}, nil
}

func (l *LocalDB) Close() error {
	sqlDB, err := l.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func GenerateLocalID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("loc_%x", b)
}

func (l *LocalDB) SaveSnippet(alias, description, content string, encrypted bool, ttl, category, language, tags string) (*models.Snippet, error) {
	record := snippetCache{
		ID:          GenerateLocalID(),
		Alias:       alias,
		Description: description,
		Content:     content,
		Encrypted:   encrypted,
		Category:    category,
		Language:    language,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if ttl != "" {
		d, err := utils.ParseDuration(ttl)
		if err != nil {
			return nil, fmt.Errorf("invalid ttl: %w", err)
		}
		if d > 0 {
			expires := time.Now().Add(d)
			record.ExpiresAt = &expires
		}
	}

	if err := l.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return cacheToSnippet(&record), nil
}

func (l *LocalDB) CacheSnippet(s *models.Snippet) error {
	record := snippetCache{
		ID:          s.ID,
		Alias:       s.Alias,
		Description: s.Description,
		Content:     s.Content,
		Encrypted:   s.Encrypted,
		Category:    s.Category,
		Language:    s.Language,
		Tags:        strings.Join(s.Tags, ","),
		ExpiresAt:   s.ExpiresAt,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	return l.db.Save(&record).Error
}

func (l *LocalDB) UpdateSnippet(id, alias, description, content, category, language, tags string) (*models.Snippet, error) {
	var cache snippetCache
	if err := l.db.First(&cache, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if alias != "" {
		cache.Alias = alias
	}
	if description != "" {
		cache.Description = description
	}
	if content != "" {
		cache.Content = content
	}
	if category != "" {
		cache.Category = category
	}
	if language != "" {
		cache.Language = language
	}
	if tags != "" {
		cache.Tags = tags
	}
	cache.UpdatedAt = time.Now()

	if err := l.db.Save(&cache).Error; err != nil {
		return nil, err
	}
	return cacheToSnippet(&cache), nil
}

func (l *LocalDB) GetCached(id string) (*models.Snippet, error) {
	var cache snippetCache
	if err := l.db.Where("id = ?", id).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&cache).Error; err != nil {
		return nil, err
	}
	return cacheToSnippet(&cache), nil
}

func (l *LocalDB) GetByPrefix(prefix string) (*models.Snippet, error) {
	var cache snippetCache
	if err := l.db.Where("id LIKE ?", prefix+"%").
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&cache).Error; err != nil {
		return nil, err
	}
	return cacheToSnippet(&cache), nil
}

func (l *LocalDB) GetCachedByAlias(alias string) (*models.Snippet, error) {
	var cache snippetCache
	if err := l.db.Where("alias = ?", alias).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at DESC").
		First(&cache).Error; err != nil {
		return nil, err
	}
	return cacheToSnippet(&cache), nil
}

func (l *LocalDB) ListCached() ([]models.SnippetPreview, error) {
	return l.ListFiltered(100, 0, "", "")
}

func (l *LocalDB) ListFiltered(limit, offset int, category, tag string) ([]models.SnippetPreview, error) {
	query := l.db.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	var caches []snippetCache
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&caches).Error; err != nil {
		return nil, err
	}

	result := make([]models.SnippetPreview, len(caches))
	for i, c := range caches {
		result[i] = cacheToPreview(&c)
	}
	return result, nil
}

func (l *LocalDB) DeleteCached(id string) error {
	result := l.db.Delete(&snippetCache{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return fmt.Errorf("snippet not found")
	}
	return result.Error
}

func (l *LocalDB) DeleteByAlias(alias string) error {
	var cache snippetCache
	if err := l.db.Where("alias = ?", alias).Order("created_at DESC").First(&cache).Error; err != nil {
		return fmt.Errorf("snippet not found")
	}
	return l.db.Delete(&cache).Error
}

func (l *LocalDB) CleanExpired() error {
	return l.db.Delete(&snippetCache{}, "expires_at IS NOT NULL AND expires_at < ?", time.Now()).Error
}

func cacheToSnippet(c *snippetCache) *models.Snippet {
	s := &models.Snippet{
		ID:          c.ID,
		Alias:       c.Alias,
		Description: c.Description,
		Content:     c.Content,
		Encrypted:   c.Encrypted,
		Category:    c.Category,
		Language:    c.Language,
		ExpiresAt:   c.ExpiresAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Tags != "" {
		s.Tags = strings.Split(c.Tags, ",")
	}
	return s
}

func cacheToPreview(c *snippetCache) models.SnippetPreview {
	preview := c.Content
	if len(preview) > 50 {
		preview = preview[:50] + "..."
	}
	p := models.SnippetPreview{
		ID:          c.ID,
		Alias:       c.Alias,
		Description: c.Description,
		Preview:     preview,
		Encrypted:   c.Encrypted,
		Category:    c.Category,
		Language:    c.Language,
		ExpiresAt:   c.ExpiresAt,
		CreatedAt:   c.CreatedAt,
	}
	if c.Tags != "" {
		p.Tags = strings.Split(c.Tags, ",")
	}
	return p
}
