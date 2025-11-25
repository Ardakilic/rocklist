// Package repository provides data access layer interfaces and implementations
package repository

import (
	"context"

	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/gorm"
)

// playlistRepository implements PlaylistRepository
type playlistRepository struct {
	db *gorm.DB
}

// NewPlaylistRepository creates a new playlist repository
func NewPlaylistRepository(db *gorm.DB) PlaylistRepository {
	return &playlistRepository{db: db}
}

// Create creates a new playlist
func (r *playlistRepository) Create(ctx context.Context, playlist *models.Playlist) error {
	return r.db.WithContext(ctx).Create(playlist).Error
}

// Update updates an existing playlist
func (r *playlistRepository) Update(ctx context.Context, playlist *models.Playlist) error {
	return r.db.WithContext(ctx).Save(playlist).Error
}

// Delete deletes a playlist by ID
func (r *playlistRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete playlist songs first
		if err := tx.Where("playlist_id = ?", id).Delete(&models.PlaylistSong{}).Error; err != nil {
			return err
		}
		// Delete playlist
		result := tx.Delete(&models.Playlist{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return models.ErrPlaylistNotFound
		}
		return nil
	})
}

// FindByID finds a playlist by ID
func (r *playlistRepository) FindByID(ctx context.Context, id uint) (*models.Playlist, error) {
	var playlist models.Playlist
	err := r.db.WithContext(ctx).First(&playlist, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrPlaylistNotFound
		}
		return nil, err
	}
	return &playlist, nil
}

// FindAll returns all playlists
func (r *playlistRepository) FindAll(ctx context.Context) ([]*models.Playlist, error) {
	var playlists []*models.Playlist
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&playlists).Error
	return playlists, err
}

// FindByType returns playlists by type
func (r *playlistRepository) FindByType(ctx context.Context, playlistType models.PlaylistType) ([]*models.Playlist, error) {
	var playlists []*models.Playlist
	err := r.db.WithContext(ctx).Where("type = ?", playlistType).Order("created_at DESC").Find(&playlists).Error
	return playlists, err
}

// FindByDataSource returns playlists by data source
func (r *playlistRepository) FindByDataSource(ctx context.Context, source models.DataSource) ([]*models.Playlist, error) {
	var playlists []*models.Playlist
	err := r.db.WithContext(ctx).Where("data_source = ?", source).Order("created_at DESC").Find(&playlists).Error
	return playlists, err
}

// AddSongs adds songs to a playlist
func (r *playlistRepository) AddSongs(ctx context.Context, playlistID uint, songIDs []uint) error {
	if len(songIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current max position
		var maxPosition int
		tx.Model(&models.PlaylistSong{}).
			Where("playlist_id = ?", playlistID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxPosition)

		// Create playlist songs
		playlistSongs := make([]models.PlaylistSong, len(songIDs))
		for i, songID := range songIDs {
			playlistSongs[i] = models.PlaylistSong{
				PlaylistID: playlistID,
				SongID:     songID,
				Position:   maxPosition + i + 1,
			}
		}

		if err := tx.CreateInBatches(playlistSongs, 100).Error; err != nil {
			return err
		}

		// Update song count
		return tx.Model(&models.Playlist{}).Where("id = ?", playlistID).
			Update("song_count", gorm.Expr("song_count + ?", len(songIDs))).Error
	})
}

// RemoveSongs removes songs from a playlist
func (r *playlistRepository) RemoveSongs(ctx context.Context, playlistID uint, songIDs []uint) error {
	if len(songIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("playlist_id = ? AND song_id IN ?", playlistID, songIDs).
			Delete(&models.PlaylistSong{})
		if result.Error != nil {
			return result.Error
		}

		// Update song count
		return tx.Model(&models.Playlist{}).Where("id = ?", playlistID).
			Update("song_count", gorm.Expr("song_count - ?", result.RowsAffected)).Error
	})
}

// GetSongs returns all songs in a playlist
func (r *playlistRepository) GetSongs(ctx context.Context, playlistID uint) ([]*models.Song, error) {
	var songs []*models.Song
	err := r.db.WithContext(ctx).
		Joins("JOIN playlist_songs ON playlist_songs.song_id = songs.id").
		Where("playlist_songs.playlist_id = ?", playlistID).
		Order("playlist_songs.position ASC").
		Find(&songs).Error
	return songs, err
}
