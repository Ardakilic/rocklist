// Package models contains all domain models for Rocklist
package models

import (
	"time"

	"gorm.io/gorm"
)

// Config represents application configuration stored in the database
type Config struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null" json:"key"`
	Value string `json:"value"`
}

// TableName returns the table name for Config
func (Config) TableName() string {
	return "configs"
}

// AppConfig represents the full application configuration
type AppConfig struct {
	RockboxPath        string           `json:"rockbox_path"`
	LastParsedAt       *time.Time       `json:"last_parsed_at,omitempty"`
	LastFM             LastFMConfig     `json:"lastfm"`
	Spotify            SpotifyConfig    `json:"spotify"`
	MusicBrainz        MusicBrainzConfig `json:"musicbrainz"`
	EnabledSources     []DataSource     `json:"enabled_sources"`
	DefaultPlaylistDir string           `json:"default_playlist_dir"`
}

// LastFMConfig holds Last.fm API configuration
type LastFMConfig struct {
	Enabled   bool   `json:"enabled"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

// SpotifyConfig holds Spotify API configuration
type SpotifyConfig struct {
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// MusicBrainzConfig holds MusicBrainz API configuration
type MusicBrainzConfig struct {
	Enabled   bool   `json:"enabled"`
	UserAgent string `json:"user_agent"` // MusicBrainz requires a user agent
}

// IsSourceEnabled checks if a data source is enabled
func (ac *AppConfig) IsSourceEnabled(source DataSource) bool {
	for _, s := range ac.EnabledSources {
		if s == source {
			return true
		}
	}
	return false
}

// GetEnabledSources returns all enabled data sources
func (ac *AppConfig) GetEnabledSources() []DataSource {
	var enabled []DataSource
	if ac.LastFM.Enabled {
		enabled = append(enabled, DataSourceLastFM)
	}
	if ac.Spotify.Enabled {
		enabled = append(enabled, DataSourceSpotify)
	}
	if ac.MusicBrainz.Enabled {
		enabled = append(enabled, DataSourceMusicBrainz)
	}
	return enabled
}

// ParseStatus represents the status of a Rockbox database parse operation
type ParseStatus struct {
	InProgress    bool       `json:"in_progress"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	TotalSongs    int        `json:"total_songs"`
	ProcessedSongs int       `json:"processed_songs"`
	ErrorCount    int        `json:"error_count"`
	LastError     string     `json:"last_error,omitempty"`
}

// Progress returns the progress percentage (0-100)
func (ps *ParseStatus) Progress() float64 {
	if ps.TotalSongs == 0 {
		return 0
	}
	return float64(ps.ProcessedSongs) / float64(ps.TotalSongs) * 100
}
