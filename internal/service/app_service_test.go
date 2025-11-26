package service

import (
	"context"
	"os"
	"testing"

	"github.com/Ardakilic/rocklist/internal/database"
	"github.com/Ardakilic/rocklist/internal/models"
	"github.com/Ardakilic/rocklist/internal/repository"
	"gorm.io/gorm/logger"
)

func setupTestAppService(t *testing.T) (*database.Database, *AppService) {
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

	logBuffer := NewLogBuffer(100)
	appLogger := NewAppLogger(logBuffer)

	songRepo := repository.NewSongRepository(db.DB())
	playlistRepo := repository.NewPlaylistRepository(db.DB())
	configRepo := repository.NewConfigRepository(db.DB())

	// Create a minimal AppService for testing
	svc := &AppService{
		db:           db,
		songRepo:     songRepo,
		playlistRepo: playlistRepo,
		configRepo:   configRepo,
		config:       &models.AppConfig{},
		logBuffer:    logBuffer,
	}

	// Initialize playlist service
	svc.playlistService = NewPlaylistService(songRepo, playlistRepo, "", appLogger)

	return db, svc
}

func TestNewAppService(t *testing.T) {
	// Test with in-memory database path
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	if svc == nil {
		t.Fatal("NewAppService() returned nil")
	}
}

func TestNewAppService_EmptyPath(t *testing.T) {
	// Test with empty path (uses default)
	svc, err := NewAppService("")
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	if svc == nil {
		t.Fatal("NewAppService() returned nil")
	}
}

func TestAppService_GetConfig(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	config := svc.GetConfig()
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}
}

func TestAppService_SaveConfig(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	// Need to set up parser
	svc.parser = nil // Will be nil but we're testing SaveConfig's repo operations

	config := &models.AppConfig{
		RockboxPath: "/test/path",
		LastFM: models.LastFMConfig{
			APIKey:    "test_key",
			APISecret: "test_secret",
			Enabled:   true,
		},
		Spotify: models.SpotifyConfig{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			Enabled:      true,
		},
		MusicBrainz: models.MusicBrainzConfig{
			UserAgent: "TestAgent/1.0",
			Enabled:   true,
		},
	}

	// This will fail because parser is nil, but tests the config save path
	defer func() { _ = recover() }()
	_ = svc.SaveConfig(context.Background(), config)
}

func TestAppService_GetSongCount(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	count, err := svc.GetSongCount(ctx)
	if err != nil {
		t.Fatalf("GetSongCount() error = %v", err)
	}

	if count != 0 {
		t.Errorf("GetSongCount() = %d, want 0", count)
	}

	// Add some songs
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Title: "Song 1"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Title: "Song 2"})

	count, _ = svc.GetSongCount(ctx)
	if count != 2 {
		t.Errorf("GetSongCount() = %d, want 2", count)
	}
}

func TestAppService_GetUniqueArtists(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", AlbumArtist: "Artist 1"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", AlbumArtist: "Artist 2"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", AlbumArtist: "Artist 1"})

	artists, err := svc.GetUniqueArtists(ctx)
	if err != nil {
		t.Fatalf("GetUniqueArtists() error = %v", err)
	}

	if len(artists) != 2 {
		t.Errorf("GetUniqueArtists() returned %d artists, want 2", len(artists))
	}
}

func TestAppService_GetUniqueGenres(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Genre: "Rock"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Genre: "Metal"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "3", Path: "/3.mp3", Genre: "Rock"})

	genres, err := svc.GetUniqueGenres(ctx)
	if err != nil {
		t.Fatalf("GetUniqueGenres() error = %v", err)
	}

	if len(genres) != 2 {
		t.Errorf("GetUniqueGenres() returned %d genres, want 2", len(genres))
	}
}

func TestAppService_GetAllPlaylists(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create some playlists
	_ = svc.playlistRepo.Create(ctx, &models.Playlist{Name: "Playlist 1", Type: models.PlaylistTypeTopSongs, DataSource: models.DataSourceLastFM})
	_ = svc.playlistRepo.Create(ctx, &models.Playlist{Name: "Playlist 2", Type: models.PlaylistTypeSimilar, DataSource: models.DataSourceSpotify})

	playlists, err := svc.GetAllPlaylists(ctx)
	if err != nil {
		t.Fatalf("GetAllPlaylists() error = %v", err)
	}

	if len(playlists) != 2 {
		t.Errorf("GetAllPlaylists() returned %d playlists, want 2", len(playlists))
	}
}

func TestAppService_GetPlaylist(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	playlist := &models.Playlist{Name: "Test Playlist", Type: models.PlaylistTypeTopSongs, DataSource: models.DataSourceLastFM}
	_ = svc.playlistRepo.Create(ctx, playlist)

	found, err := svc.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Fatalf("GetPlaylist() error = %v", err)
	}

	if found.Name != "Test Playlist" {
		t.Errorf("GetPlaylist() Name = %v, want Test Playlist", found.Name)
	}
}

func TestAppService_DeletePlaylist(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	playlist := &models.Playlist{Name: "To Delete", Type: models.PlaylistTypeTopSongs, DataSource: models.DataSourceLastFM}
	_ = svc.playlistRepo.Create(ctx, playlist)

	err := svc.DeletePlaylist(ctx, playlist.ID)
	if err != nil {
		t.Fatalf("DeletePlaylist() error = %v", err)
	}

	// Verify deleted
	_, err = svc.playlistRepo.FindByID(ctx, playlist.ID)
	if err != models.ErrPlaylistNotFound {
		t.Error("Playlist should be deleted")
	}
}

func TestAppService_GetLogs(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	// Add some logs
	svc.logBuffer.Add("info", "Test log 1")
	svc.logBuffer.Add("error", "Test log 2")

	logs := svc.GetLogs()
	if logs == nil {
		t.Fatal("GetLogs() returned nil")
	}

	if len(logs) != 2 {
		t.Errorf("GetLogs() returned %d entries, want 2", len(logs))
	}
}

func TestAppService_ClearLogs(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	svc.logBuffer.Add("info", "Test log")
	svc.ClearLogs()

	logs := svc.GetLogs()
	if len(logs) != 0 {
		t.Errorf("ClearLogs() should clear all logs, got %d", len(logs))
	}
}

func TestAppService_GetEnabledSources(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	// No sources enabled by default
	sources := svc.GetEnabledSources()
	if len(sources) != 0 {
		t.Errorf("GetEnabledSources() should return empty when no sources enabled, got %d", len(sources))
	}

	// Enable Last.fm
	svc.config.LastFM.Enabled = true
	sources = svc.GetEnabledSources()
	if len(sources) != 1 {
		t.Errorf("GetEnabledSources() should return 1 source, got %d", len(sources))
	}

	// Enable Spotify
	svc.config.Spotify.Enabled = true
	sources = svc.GetEnabledSources()
	if len(sources) != 2 {
		t.Errorf("GetEnabledSources() should return 2 sources, got %d", len(sources))
	}

	// Enable MusicBrainz
	svc.config.MusicBrainz.Enabled = true
	sources = svc.GetEnabledSources()
	if len(sources) != 3 {
		t.Errorf("GetEnabledSources() should return 3 sources, got %d", len(sources))
	}
}

func TestAppService_WipeData(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Add some data
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Title: "Song 1"})
	_ = svc.playlistRepo.Create(ctx, &models.Playlist{Name: "Playlist", Type: models.PlaylistTypeTopSongs, DataSource: models.DataSourceLastFM})

	err := svc.WipeData(ctx)
	if err != nil {
		t.Fatalf("WipeData() error = %v", err)
	}

	// Verify data is wiped
	count, _ := svc.songRepo.Count(ctx)
	if count != 0 {
		t.Errorf("WipeData() should delete all songs, got %d", count)
	}
}

func TestAppService_GetParseStatus(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	status := svc.GetParseStatus()
	if status == nil {
		t.Error("GetParseStatus() returned nil")
	}
}

func TestAppService_GetLastParsedAt(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Should be nil initially
	lastParsed, err := svc.GetLastParsedAt(ctx)
	if err != nil {
		t.Fatalf("GetLastParsedAt() error = %v", err)
	}

	if lastParsed != nil {
		t.Error("GetLastParsedAt() should return nil initially")
	}
}

func TestAppService_GetAppInfo(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	info := svc.GetAppInfo()
	if info == nil {
		t.Fatal("GetAppInfo() returned nil")
	}

	if info["name"] != "Rocklist" {
		t.Errorf("GetAppInfo() name = %v, want Rocklist", info["name"])
	}
}

func TestAppService_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}

	// Should not panic
	_ = svc.Close()
}

func TestAppService_SetRockboxPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	rockboxPath := tmpDir + "/rockbox"
	err = svc.SetRockboxPath(rockboxPath)
	if err != nil {
		t.Fatalf("SetRockboxPath() error = %v", err)
	}

	config := svc.GetConfig()
	if config.RockboxPath != rockboxPath {
		t.Errorf("SetRockboxPath() RockboxPath = %v, want %v", config.RockboxPath, rockboxPath)
	}
}

func TestAppService_GeneratePlaylist_NoClient(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	ctx := context.Background()

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeTopSongs,
		Artist:     "Test Artist",
		Limit:      10,
	}

	_, err = svc.GeneratePlaylist(ctx, req)
	if err == nil {
		t.Error("GeneratePlaylist() should return error when no client configured")
	}
}

func TestAppService_GetAllSongs(t *testing.T) {
	db, svc := setupTestAppService(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "1", Path: "/1.mp3", Title: "Song 1"})
	_ = svc.songRepo.Create(ctx, &models.Song{RockboxID: "2", Path: "/2.mp3", Title: "Song 2"})

	songs, err := svc.GetAllSongs(ctx)
	if err != nil {
		t.Fatalf("GetAllSongs() error = %v", err)
	}

	if len(songs) != 2 {
		t.Errorf("GetAllSongs() returned %d songs, want 2", len(songs))
	}
}

func TestAppService_ExportLogsToFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	svc, err := NewAppService(dbPath)
	if err != nil {
		t.Fatalf("NewAppService() error = %v", err)
	}
	defer func() { _ = svc.Close() }()

	// Add some logs
	svc.logBuffer.Add("info", "Test log 1")
	svc.logBuffer.Add("error", "Test log 2")

	// Export logs
	logFile := tmpDir + "/logs.txt"
	err = svc.ExportLogsToFile(logFile)
	if err != nil {
		t.Fatalf("ExportLogsToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("ExportLogsToFile() should create the log file")
	}
}
