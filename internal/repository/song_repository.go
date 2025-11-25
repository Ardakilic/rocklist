// Package repository provides data access layer interfaces and implementations
package repository

import (
	"context"
	"fmt"

	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/gorm"
)

// songRepository implements SongRepository
type songRepository struct {
	db *gorm.DB
}

// NewSongRepository creates a new song repository
func NewSongRepository(db *gorm.DB) SongRepository {
	return &songRepository{db: db}
}

// Create creates a new song
func (r *songRepository) Create(ctx context.Context, song *models.Song) error {
	return r.db.WithContext(ctx).Create(song).Error
}

// CreateBatch creates multiple songs in a batch
func (r *songRepository) CreateBatch(ctx context.Context, songs []*models.Song) error {
	if len(songs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(songs, 100).Error
}

// Update updates an existing song
func (r *songRepository) Update(ctx context.Context, song *models.Song) error {
	return r.db.WithContext(ctx).Save(song).Error
}

// Delete deletes a song by ID
func (r *songRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Song{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrSongNotFound
	}
	return nil
}

// FindByID finds a song by ID
func (r *songRepository) FindByID(ctx context.Context, id uint) (*models.Song, error) {
	var song models.Song
	err := r.db.WithContext(ctx).First(&song, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrSongNotFound
		}
		return nil, err
	}
	return &song, nil
}

// FindByRockboxID finds a song by Rockbox ID
func (r *songRepository) FindByRockboxID(ctx context.Context, rockboxID string) (*models.Song, error) {
	var song models.Song
	err := r.db.WithContext(ctx).Where("rockbox_id = ?", rockboxID).First(&song).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrSongNotFound
		}
		return nil, err
	}
	return &song, nil
}

// FindByPath finds a song by file path
func (r *songRepository) FindByPath(ctx context.Context, path string) (*models.Song, error) {
	var song models.Song
	err := r.db.WithContext(ctx).Where("path = ?", path).First(&song).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrSongNotFound
		}
		return nil, err
	}
	return &song, nil
}

// FindAll returns all songs
func (r *songRepository) FindAll(ctx context.Context) ([]*models.Song, error) {
	var songs []*models.Song
	err := r.db.WithContext(ctx).Find(&songs).Error
	return songs, err
}

// FindByArtist returns all songs by an artist
func (r *songRepository) FindByArtist(ctx context.Context, artist string) ([]*models.Song, error) {
	var songs []*models.Song
	err := r.db.WithContext(ctx).Where("artist = ? OR album_artist = ?", artist, artist).Find(&songs).Error
	return songs, err
}

// FindByAlbumArtist returns all songs by an album artist
func (r *songRepository) FindByAlbumArtist(ctx context.Context, albumArtist string) ([]*models.Song, error) {
	var songs []*models.Song
	err := r.db.WithContext(ctx).Where("album_artist = ?", albumArtist).Find(&songs).Error
	return songs, err
}

// FindByGenre returns all songs matching a genre
func (r *songRepository) FindByGenre(ctx context.Context, genre string) ([]*models.Song, error) {
	var songs []*models.Song
	// Use LIKE for partial matching (e.g., "Death Metal" matches "Death Metal", "Melodic Death Metal")
	err := r.db.WithContext(ctx).Where("genre LIKE ?", fmt.Sprintf("%%%s%%", genre)).Find(&songs).Error
	return songs, err
}

// FindUnmatched returns songs without external ID matches
func (r *songRepository) FindUnmatched(ctx context.Context, source models.DataSource) ([]*models.Song, error) {
	var songs []*models.Song
	var query *gorm.DB

	switch source {
	case models.DataSourceLastFM:
		query = r.db.WithContext(ctx).Where("lastfm_id = '' OR lastfm_id IS NULL")
	case models.DataSourceSpotify:
		query = r.db.WithContext(ctx).Where("spotify_id = '' OR spotify_id IS NULL")
	case models.DataSourceMusicBrainz:
		query = r.db.WithContext(ctx).Where("musicbrainz_id = '' OR musicbrainz_id IS NULL")
	default:
		return nil, models.ErrInvalidDataSource
	}

	err := query.Find(&songs).Error
	return songs, err
}

// GetUniqueArtists returns a list of unique album artists
func (r *songRepository) GetUniqueArtists(ctx context.Context) ([]string, error) {
	var artists []string
	err := r.db.WithContext(ctx).
		Model(&models.Song{}).
		Distinct("album_artist").
		Where("album_artist != '' AND album_artist IS NOT NULL").
		Pluck("album_artist", &artists).Error
	return artists, err
}

// GetUniqueGenres returns a list of unique genres
func (r *songRepository) GetUniqueGenres(ctx context.Context) ([]string, error) {
	var genres []string
	err := r.db.WithContext(ctx).
		Model(&models.Song{}).
		Distinct("genre").
		Where("genre != '' AND genre IS NOT NULL").
		Pluck("genre", &genres).Error
	return genres, err
}

// Count returns the total number of songs
func (r *songRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Song{}).Count(&count).Error
	return count, err
}

// DeleteAll deletes all songs
func (r *songRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM songs").Error
}
