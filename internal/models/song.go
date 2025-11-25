// Package models contains all domain models for Rocklist
package models

import (
	"time"

	"gorm.io/gorm"
)

// Song represents a track in the Rockbox database
type Song struct {
	gorm.Model
	RockboxID       string  `gorm:"uniqueIndex;not null" json:"rockbox_id"`
	Path            string  `gorm:"not null" json:"path"`
	Title           string  `gorm:"index" json:"title"`
	Artist          string  `gorm:"index" json:"artist"`
	AlbumArtist     string  `gorm:"index" json:"album_artist"`
	Album           string  `gorm:"index" json:"album"`
	Genre           string  `gorm:"index" json:"genre"`
	Year            int     `json:"year"`
	TrackNumber     int     `json:"track_number"`
	DiscNumber      int     `json:"disc_number"`
	Duration        int     `json:"duration"` // in seconds
	Bitrate         int     `json:"bitrate"`
	Frequency       int     `json:"frequency"`
	FileSize        int64   `json:"file_size"`
	Rating          int     `json:"rating"`
	PlayCount       int     `json:"play_count"`
	LastPlayed      *time.Time `json:"last_played,omitempty"`
	// External IDs for matching
	MusicBrainzID   string  `gorm:"index" json:"musicbrainz_id,omitempty"`
	SpotifyID       string  `gorm:"index" json:"spotify_id,omitempty"`
	LastFMID        string  `gorm:"index" json:"lastfm_id,omitempty"`
	// Matching status
	MatchedSource   string  `json:"matched_source,omitempty"` // lastfm, spotify, musicbrainz
	MatchedAt       *time.Time `json:"matched_at,omitempty"`
	MatchConfidence float64 `json:"match_confidence,omitempty"` // 0.0 to 1.0
}

// TableName returns the table name for Song
func (Song) TableName() string {
	return "songs"
}

// IsMatched returns true if the song has been matched to an external source
func (s *Song) IsMatched() bool {
	return s.MusicBrainzID != "" || s.SpotifyID != "" || s.LastFMID != ""
}

// GetDisplayName returns a formatted display name for the song
func (s *Song) GetDisplayName() string {
	if s.Artist != "" && s.Title != "" {
		return s.Artist + " - " + s.Title
	}
	if s.Title != "" {
		return s.Title
	}
	return s.Path
}

// GetEffectiveArtist returns album artist if available, otherwise artist
func (s *Song) GetEffectiveArtist() string {
	if s.AlbumArtist != "" {
		return s.AlbumArtist
	}
	return s.Artist
}
