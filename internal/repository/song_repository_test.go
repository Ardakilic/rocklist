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
	_ = repo.Create(ctx, song)
	
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
	_ = repo.Create(ctx, song)
	
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
	_ = repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Artist: "Metallica", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Artist: "Metallica", Title: "Song 2"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", Artist: "Iron Maiden", Title: "Song 3"})
	
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
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", AlbumArtist: "Artist A", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", AlbumArtist: "Artist B", Title: "Song 2"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", AlbumArtist: "Artist A", Title: "Song 3"})
	
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
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Title: "Song 2"})
	
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
	_ = repo.Create(ctx, song)
	
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

func TestSongRepository_CreateBatch_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	err := repo.CreateBatch(ctx, []*models.Song{})
	if err != nil {
		t.Errorf("CreateBatch() with empty slice should not error, got: %v", err)
	}
}

func TestSongRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	song := &models.Song{RockboxID: "update-1", Path: "/update.mp3", Title: "Original"}
	_ = repo.Create(ctx, song)
	
	song.Title = "Updated"
	err := repo.Update(ctx, song)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	
	found, _ := repo.FindByID(ctx, song.ID)
	if found.Title != "Updated" {
		t.Errorf("Update() Title = %v, want Updated", found.Title)
	}
}

func TestSongRepository_FindByPath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	song := &models.Song{RockboxID: "path-1", Path: "/unique/path.mp3", Title: "Path Test"}
	_ = repo.Create(ctx, song)
	
	found, err := repo.FindByPath(ctx, "/unique/path.mp3")
	if err != nil {
		t.Fatalf("FindByPath() error = %v", err)
	}
	if found.Title != "Path Test" {
		t.Errorf("FindByPath() Title = %v, want Path Test", found.Title)
	}
}

func TestSongRepository_FindByPath_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_, err := repo.FindByPath(ctx, "/nonexistent.mp3")
	if err != models.ErrSongNotFound {
		t.Errorf("FindByPath() error = %v, want ErrSongNotFound", err)
	}
}

func TestSongRepository_FindByAlbumArtist(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "aa1", Path: "/1.mp3", AlbumArtist: "Album Artist 1", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "aa2", Path: "/2.mp3", AlbumArtist: "Album Artist 1", Title: "Song 2"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "aa3", Path: "/3.mp3", AlbumArtist: "Album Artist 2", Title: "Song 3"})
	
	songs, err := repo.FindByAlbumArtist(ctx, "Album Artist 1")
	if err != nil {
		t.Fatalf("FindByAlbumArtist() error = %v", err)
	}
	if len(songs) != 2 {
		t.Errorf("FindByAlbumArtist() returned %d songs, want 2", len(songs))
	}
}

func TestSongRepository_FindByGenre(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "g1", Path: "/1.mp3", Genre: "Rock", Title: "Rock Song"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "g2", Path: "/2.mp3", Genre: "Hard Rock", Title: "Hard Rock Song"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "g3", Path: "/3.mp3", Genre: "Metal", Title: "Metal Song"})
	
	songs, err := repo.FindByGenre(ctx, "Rock")
	if err != nil {
		t.Fatalf("FindByGenre() error = %v", err)
	}
	// Should match "Rock" and "Hard Rock"
	if len(songs) != 2 {
		t.Errorf("FindByGenre() returned %d songs, want 2", len(songs))
	}
}

func TestSongRepository_FindUnmatched_Spotify(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "u1", Path: "/1.mp3", Title: "Unmatched"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "u2", Path: "/2.mp3", Title: "Matched", SpotifyID: "sp123"})
	
	songs, err := repo.FindUnmatched(ctx, models.DataSourceSpotify)
	if err != nil {
		t.Fatalf("FindUnmatched() error = %v", err)
	}
	if len(songs) != 1 {
		t.Errorf("FindUnmatched() returned %d songs, want 1", len(songs))
	}
}

func TestSongRepository_FindUnmatched_InvalidSource(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_, err := repo.FindUnmatched(ctx, "invalid")
	if err != models.ErrInvalidDataSource {
		t.Errorf("FindUnmatched() error = %v, want ErrInvalidDataSource", err)
	}
}

func TestSongRepository_GetUniqueGenres(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "ug1", Path: "/1.mp3", Genre: "Rock", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "ug2", Path: "/2.mp3", Genre: "Metal", Title: "Song 2"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "ug3", Path: "/3.mp3", Genre: "Rock", Title: "Song 3"})
	
	genres, err := repo.GetUniqueGenres(ctx)
	if err != nil {
		t.Fatalf("GetUniqueGenres() error = %v", err)
	}
	if len(genres) != 2 {
		t.Errorf("GetUniqueGenres() returned %d genres, want 2", len(genres))
	}
}

func TestSongRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "fa1", Path: "/1.mp3", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "fa2", Path: "/2.mp3", Title: "Song 2"})
	
	songs, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(songs) != 2 {
		t.Errorf("FindAll() returned %d songs, want 2", len(songs))
	}
}

func TestSongRepository_DeleteAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	_ = repo.Create(ctx, &models.Song{RockboxID: "da1", Path: "/1.mp3", Title: "Song 1"})
	_ = repo.Create(ctx, &models.Song{RockboxID: "da2", Path: "/2.mp3", Title: "Song 2"})
	
	err := repo.DeleteAll(ctx)
	if err != nil {
		t.Fatalf("DeleteAll() error = %v", err)
	}
	
	count, _ := repo.Count(ctx)
	if count != 0 {
		t.Errorf("DeleteAll() should delete all songs, got count %d", count)
	}
}

func TestSongRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewSongRepository(db.DB())
	ctx := context.Background()
	
	err := repo.Delete(ctx, 99999)
	if err != models.ErrSongNotFound {
		t.Errorf("Delete() error = %v, want ErrSongNotFound", err)
	}
}
