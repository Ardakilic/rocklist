// Package repository provides data access layer interfaces and implementations
package repository

import (
	"context"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
)

// SongRepository defines the interface for song data access
type SongRepository interface {
	// Create creates a new song
	Create(ctx context.Context, song *models.Song) error
	// CreateBatch creates multiple songs in a batch
	CreateBatch(ctx context.Context, songs []*models.Song) error
	// Update updates an existing song
	Update(ctx context.Context, song *models.Song) error
	// Delete deletes a song by ID
	Delete(ctx context.Context, id uint) error
	// FindByID finds a song by ID
	FindByID(ctx context.Context, id uint) (*models.Song, error)
	// FindByRockboxID finds a song by Rockbox ID
	FindByRockboxID(ctx context.Context, rockboxID string) (*models.Song, error)
	// FindByPath finds a song by file path
	FindByPath(ctx context.Context, path string) (*models.Song, error)
	// FindAll returns all songs
	FindAll(ctx context.Context) ([]*models.Song, error)
	// FindByArtist returns all songs by an artist
	FindByArtist(ctx context.Context, artist string) ([]*models.Song, error)
	// FindByAlbumArtist returns all songs by an album artist
	FindByAlbumArtist(ctx context.Context, albumArtist string) ([]*models.Song, error)
	// FindByGenre returns all songs matching a genre
	FindByGenre(ctx context.Context, genre string) ([]*models.Song, error)
	// FindUnmatched returns songs without external ID matches
	FindUnmatched(ctx context.Context, source models.DataSource) ([]*models.Song, error)
	// GetUniqueArtists returns a list of unique album artists
	GetUniqueArtists(ctx context.Context) ([]string, error)
	// GetUniqueGenres returns a list of unique genres
	GetUniqueGenres(ctx context.Context) ([]string, error)
	// Count returns the total number of songs
	Count(ctx context.Context) (int64, error)
	// DeleteAll deletes all songs
	DeleteAll(ctx context.Context) error
}

// PlaylistRepository defines the interface for playlist data access
type PlaylistRepository interface {
	// Create creates a new playlist
	Create(ctx context.Context, playlist *models.Playlist) error
	// Update updates an existing playlist
	Update(ctx context.Context, playlist *models.Playlist) error
	// Delete deletes a playlist by ID
	Delete(ctx context.Context, id uint) error
	// FindByID finds a playlist by ID
	FindByID(ctx context.Context, id uint) (*models.Playlist, error)
	// FindAll returns all playlists
	FindAll(ctx context.Context) ([]*models.Playlist, error)
	// FindByType returns playlists by type
	FindByType(ctx context.Context, playlistType models.PlaylistType) ([]*models.Playlist, error)
	// FindByDataSource returns playlists by data source
	FindByDataSource(ctx context.Context, source models.DataSource) ([]*models.Playlist, error)
	// AddSongs adds songs to a playlist
	AddSongs(ctx context.Context, playlistID uint, songIDs []uint) error
	// RemoveSongs removes songs from a playlist
	RemoveSongs(ctx context.Context, playlistID uint, songIDs []uint) error
	// GetSongs returns all songs in a playlist
	GetSongs(ctx context.Context, playlistID uint) ([]*models.Song, error)
}

// ConfigRepository defines the interface for configuration data access
type ConfigRepository interface {
	// Get gets a config value by key
	Get(ctx context.Context, key string) (string, error)
	// Set sets a config value
	Set(ctx context.Context, key, value string) error
	// Delete deletes a config by key
	Delete(ctx context.Context, key string) error
	// GetAll returns all config values as a map
	GetAll(ctx context.Context) (map[string]string, error)
	// GetLastParsedAt returns the last parsed timestamp
	GetLastParsedAt(ctx context.Context) (*time.Time, error)
	// SetLastParsedAt sets the last parsed timestamp
	SetLastParsedAt(ctx context.Context, t time.Time) error
}
