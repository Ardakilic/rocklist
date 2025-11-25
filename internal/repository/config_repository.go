// Package repository provides data access layer interfaces and implementations
package repository

import (
	"context"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// ConfigKeyLastParsedAt is the key for last parsed timestamp
	ConfigKeyLastParsedAt = "last_parsed_at"
	// ConfigKeyRockboxPath is the key for Rockbox path
	ConfigKeyRockboxPath = "rockbox_path"
	// ConfigKeyLastFMAPIKey is the key for Last.fm API key
	ConfigKeyLastFMAPIKey = "lastfm_api_key"
	// ConfigKeyLastFMAPISecret is the key for Last.fm API secret
	ConfigKeyLastFMAPISecret = "lastfm_api_secret"
	// ConfigKeyLastFMEnabled is the key for Last.fm enabled status
	ConfigKeyLastFMEnabled = "lastfm_enabled"
	// ConfigKeySpotifyClientID is the key for Spotify client ID
	ConfigKeySpotifyClientID = "spotify_client_id"
	// ConfigKeySpotifyClientSecret is the key for Spotify client secret
	ConfigKeySpotifyClientSecret = "spotify_client_secret"
	// ConfigKeySpotifyEnabled is the key for Spotify enabled status
	ConfigKeySpotifyEnabled = "spotify_enabled"
	// ConfigKeyMusicBrainzUserAgent is the key for MusicBrainz user agent
	ConfigKeyMusicBrainzUserAgent = "musicbrainz_user_agent"
	// ConfigKeyMusicBrainzEnabled is the key for MusicBrainz enabled status
	ConfigKeyMusicBrainzEnabled = "musicbrainz_enabled"
)

// configRepository implements ConfigRepository
type configRepository struct {
	db *gorm.DB
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}

// Get gets a config value by key
func (r *configRepository) Get(ctx context.Context, key string) (string, error) {
	var config models.Config
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", models.ErrConfigNotFound
		}
		return "", err
	}
	return config.Value, nil
}

// Set sets a config value (upsert)
func (r *configRepository) Set(ctx context.Context, key, value string) error {
	config := models.Config{
		Key:   key,
		Value: value,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&config).Error
}

// Delete deletes a config by key
func (r *configRepository) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Config{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrConfigNotFound
	}
	return nil
}

// GetAll returns all config values as a map
func (r *configRepository) GetAll(ctx context.Context) (map[string]string, error) {
	var configs []models.Config
	err := r.db.WithContext(ctx).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}

// GetLastParsedAt returns the last parsed timestamp
func (r *configRepository) GetLastParsedAt(ctx context.Context) (*time.Time, error) {
	value, err := r.Get(ctx, ConfigKeyLastParsedAt)
	if err != nil {
		if err == models.ErrConfigNotFound {
			return nil, nil
		}
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// SetLastParsedAt sets the last parsed timestamp
func (r *configRepository) SetLastParsedAt(ctx context.Context, t time.Time) error {
	return r.Set(ctx, ConfigKeyLastParsedAt, t.Format(time.RFC3339))
}
