// Package models contains all domain models for Rocklist
package models

import (
	"time"

	"gorm.io/gorm"
)

// PlaylistType represents the type of playlist
type PlaylistType string

const (
	PlaylistTypeTopSongs    PlaylistType = "top_songs"
	PlaylistTypeMixedSongs  PlaylistType = "mixed_songs"
	PlaylistTypeSimilar     PlaylistType = "similar"
	PlaylistTypeTag         PlaylistType = "tag"
)

// String returns the string representation of PlaylistType
func (pt PlaylistType) String() string {
	return string(pt)
}

// DisplayName returns a human-readable name for the playlist type
func (pt PlaylistType) DisplayName() string {
	switch pt {
	case PlaylistTypeTopSongs:
		return "Top Songs"
	case PlaylistTypeMixedSongs:
		return "Mixed Songs"
	case PlaylistTypeSimilar:
		return "Similar Songs"
	case PlaylistTypeTag:
		return "Tag Radio"
	default:
		return string(pt)
	}
}

// DataSource represents the upstream data source
type DataSource string

const (
	DataSourceLastFM      DataSource = "lastfm"
	DataSourceSpotify     DataSource = "spotify"
	DataSourceMusicBrainz DataSource = "musicbrainz"
)

// String returns the string representation of DataSource
func (ds DataSource) String() string {
	return string(ds)
}

// DisplayName returns a human-readable name for the data source
func (ds DataSource) DisplayName() string {
	switch ds {
	case DataSourceLastFM:
		return "Last.fm"
	case DataSourceSpotify:
		return "Spotify"
	case DataSourceMusicBrainz:
		return "MusicBrainz"
	default:
		return string(ds)
	}
}

// Playlist represents a generated playlist
type Playlist struct {
	gorm.Model
	Name        string       `gorm:"not null" json:"name"`
	Description string       `json:"description,omitempty"`
	Type        PlaylistType `gorm:"not null;index" json:"type"`
	DataSource  DataSource   `gorm:"not null;index" json:"data_source"`
	Artist      string       `gorm:"index" json:"artist,omitempty"`
	Tag         string       `gorm:"index" json:"tag,omitempty"`
	FilePath    string       `json:"file_path,omitempty"` // Path to exported m3u file
	SongCount   int          `json:"song_count"`
	GeneratedAt time.Time    `json:"generated_at"`
	ExportedAt  *time.Time   `json:"exported_at,omitempty"`
	Songs       []PlaylistSong `gorm:"foreignKey:PlaylistID" json:"songs,omitempty"`
}

// TableName returns the table name for Playlist
func (Playlist) TableName() string {
	return "playlists"
}

// PlaylistSong represents a song in a playlist with ordering
type PlaylistSong struct {
	gorm.Model
	PlaylistID uint   `gorm:"not null;index" json:"playlist_id"`
	SongID     uint   `gorm:"not null;index" json:"song_id"`
	Position   int    `gorm:"not null" json:"position"`
	Song       Song   `gorm:"foreignKey:SongID" json:"song,omitempty"`
}

// TableName returns the table name for PlaylistSong
func (PlaylistSong) TableName() string {
	return "playlist_songs"
}

// PlaylistRequest represents a request to generate a playlist
type PlaylistRequest struct {
	Type       PlaylistType `json:"type"`
	DataSource DataSource   `json:"data_source"`
	Artist     string       `json:"artist,omitempty"`
	Tag        string       `json:"tag,omitempty"`
	Limit      int          `json:"limit,omitempty"` // Max songs to include
}

// Validate validates the playlist request
func (pr *PlaylistRequest) Validate() error {
	if pr.Type == "" {
		return ErrInvalidPlaylistType
	}
	if pr.DataSource == "" {
		return ErrInvalidDataSource
	}
	if pr.Type == PlaylistTypeTag && pr.Tag == "" {
		return ErrTagRequired
	}
	if pr.Limit <= 0 {
		pr.Limit = 50 // Default limit
	}
	return nil
}
