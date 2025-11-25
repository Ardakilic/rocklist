package repository

import (
	"context"
	"testing"

	"github.com/Ardakilic/rocklist/internal/database"
	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *database.Database {
	cfg := &database.Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	
	return db
}

func TestSongRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	song := &models.Song{
		RockboxID: "test-123",
		Path:      "/Music/test.mp3",
		Title:     "Test Song",
		Artist:    "Test Artist",
	}
	
	err := repo.Create(ctx, song)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	
	if song.ID == 0 {
		t.Error("Create() did not set song ID")
	}
}

func TestSongRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	// Create a song first
	song := &models.Song{
		RockboxID: "test-123",
		Path:      "/Music/test.mp3",
		Title:     "Test Song",
		Artist:    "Test Artist",
	}
	repo.Create(ctx, song)
	
	// Find it
	found, err := repo.FindByID(ctx, song.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	
	if found.RockboxID != song.RockboxID {
		t.Errorf("FindByID() RockboxID = %v, want %v", found.RockboxID, song.RockboxID)
	}
}

func TestSongRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_, err := repo.FindByID(ctx, 99999)
	if err != models.ErrSongNotFound {
		t.Errorf("FindByID() error = %v, want ErrSongNotFound", err)
	}
}

func TestSongRepository_FindByRockboxID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	song := &models.Song{
		RockboxID: "unique-rockbox-id",
		Path:      "/Music/test.mp3",
		Title:     "Test Song",
	}
	repo.Create(ctx, song)
	
	found, err := repo.FindByRockboxID(ctx, "unique-rockbox-id")
	if err != nil {
		t.Fatalf("FindByRockboxID() error = %v", err)
	}
	
	if found.ID != song.ID {
		t.Errorf("FindByRockboxID() ID = %v, want %v", found.ID, song.ID)
	}
}

func TestSongRepository_FindByArtist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	// Create songs by different artists
	repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Artist: "Metallica", Title: "Song 1"})
	repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Artist: "Metallica", Title: "Song 2"})
	repo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", Artist: "Iron Maiden", Title: "Song 3"})
	
	songs, err := repo.FindByArtist(ctx, "Metallica")
	if err != nil {
		t.Fatalf("FindByArtist() error = %v", err)
	}
	
	if len(songs) != 2 {
		t.Errorf("FindByArtist() returned %d songs, want 2", len(songs))
	}
}

func TestSongRepository_GetUniqueArtists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", AlbumArtist: "Artist A", Title: "Song 1"})
	repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", AlbumArtist: "Artist B", Title: "Song 2"})
	repo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", AlbumArtist: "Artist A", Title: "Song 3"})
	
	artists, err := repo.GetUniqueArtists(ctx)
	if err != nil {
		t.Fatalf("GetUniqueArtists() error = %v", err)
	}
	
	if len(artists) != 2 {
		t.Errorf("GetUniqueArtists() returned %d artists, want 2", len(artists))
	}
}

func TestSongRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Title: "Song 1"})
	repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Title: "Song 2"})
	
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestSongRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	song := &models.Song{RockboxID: "to-delete", Path: "/delete.mp3", Title: "Delete Me"}
	repo.Create(ctx, song)
	
	err := repo.Delete(ctx, song.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	
	_, err = repo.FindByID(ctx, song.ID)
	if err != models.ErrSongNotFound {
		t.Errorf("Song should be deleted, but FindByID returned: %v", err)
	}
}

func TestSongRepository_CreateBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	songs := []*models.Song{
		{RockboxID: "batch-1", Path: "/1.mp3", Title: "Song 1"},
		{RockboxID: "batch-2", Path: "/2.mp3", Title: "Song 2"},
		{RockboxID: "batch-3", Path: "/3.mp3", Title: "Song 3"},
	}
	
	err := repo.CreateBatch(ctx, songs)
	if err != nil {
		t.Fatalf("CreateBatch() error = %v", err)
	}
	
	count, _ := repo.Count(ctx)
	if count != 3 {
		t.Errorf("CreateBatch() created %d songs, want 3", count)
	}
}
