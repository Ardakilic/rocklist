// Package service provides business logic services
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Ardakilic/rocklist/internal/api"
	"github.com/Ardakilic/rocklist/internal/database"
	"github.com/Ardakilic/rocklist/internal/models"
	"github.com/Ardakilic/rocklist/internal/repository"
	"github.com/Ardakilic/rocklist/internal/rockbox"
)

// AppService is the main application service that coordinates all operations
type AppService struct {
	db              *database.Database
	songRepo        repository.SongRepository
	playlistRepo    repository.PlaylistRepository
	configRepo      repository.ConfigRepository
	parser          *rockbox.Parser
	playlistService *PlaylistService
	config          *models.AppConfig
	logBuffer       *LogBuffer
	mu              sync.RWMutex

	// API clients
	lastfmClient      *api.LastFMClient
	spotifyClient     *api.SpotifyClient
	musicbrainzClient *api.MusicBrainzClient
}

// LogBuffer is a circular buffer for log messages
type LogBuffer struct {
	entries []LogEntry
	maxSize int
	mu      sync.RWMutex
}

// LogEntry represents a single log entry
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

// NewLogBuffer creates a new log buffer
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a log entry to the buffer
func (lb *LogBuffer) Add(level, message string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}

	if len(lb.entries) >= lb.maxSize {
		lb.entries = lb.entries[1:]
	}
	lb.entries = append(lb.entries, entry)
}

// GetAll returns all log entries
func (lb *LogBuffer) GetAll() []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]LogEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// Clear clears all log entries
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.entries = lb.entries[:0]
}

// AppLogger implements the Logger interface with log buffer support
type AppLogger struct {
	buffer *LogBuffer
}

// NewAppLogger creates a new app logger
func NewAppLogger(buffer *LogBuffer) *AppLogger {
	return &AppLogger{buffer: buffer}
}

// Info logs an info message
func (l *AppLogger) Info(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	l.buffer.Add("info", formatted)
	fmt.Printf("[INFO] %s\n", formatted)
}

// Error logs an error message
func (l *AppLogger) Error(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	l.buffer.Add("error", formatted)
	fmt.Printf("[ERROR] %s\n", formatted)
}

// Debug logs a debug message
func (l *AppLogger) Debug(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	l.buffer.Add("debug", formatted)
	fmt.Printf("[DEBUG] %s\n", formatted)
}

// NewAppService creates a new application service
func NewAppService(dbPath string) (*AppService, error) {
	// Initialize database
	dbConfig := database.DefaultConfig()
	if dbPath != "" {
		dbConfig.Path = dbPath
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logBuffer := NewLogBuffer(1000)
	logger := NewAppLogger(logBuffer)

	// Create repositories
	songRepo := repository.NewSongRepository(db.DB())
	playlistRepo := repository.NewPlaylistRepository(db.DB())
	configRepo := repository.NewConfigRepository(db.DB())

	// Create services
	parser := rockbox.NewParser("", logger)
	playlistService := NewPlaylistService(songRepo, playlistRepo, "", logger)

	app := &AppService{
		db:              db,
		songRepo:        songRepo,
		playlistRepo:    playlistRepo,
		configRepo:      configRepo,
		parser:          parser,
		playlistService: playlistService,
		config:          &models.AppConfig{},
		logBuffer:       logBuffer,
	}

	// Initialize API clients
	app.lastfmClient = api.NewLastFMClient("", "", logger)
	app.spotifyClient = api.NewSpotifyClient("", "", logger)
	app.musicbrainzClient = api.NewMusicBrainzClient("", logger)

	// Load configuration
	if err := app.loadConfig(context.Background()); err != nil {
		logger.Debug("No existing config found, using defaults")
	}

	return app, nil
}

// Close closes all resources
func (s *AppService) Close() error {
	return s.db.Close()
}

// loadConfig loads configuration from database
func (s *AppService) loadConfig(ctx context.Context) error {
	configs, err := s.configRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	if path, ok := configs[repository.ConfigKeyRockboxPath]; ok {
		s.config.RockboxPath = path
		s.parser.SetPath(path)
	}

	// Last.fm config
	if apiKey, ok := configs[repository.ConfigKeyLastFMAPIKey]; ok {
		s.config.LastFM.APIKey = apiKey
	}
	if apiSecret, ok := configs[repository.ConfigKeyLastFMAPISecret]; ok {
		s.config.LastFM.APISecret = apiSecret
	}
	if enabled, ok := configs[repository.ConfigKeyLastFMEnabled]; ok {
		s.config.LastFM.Enabled = enabled == "true"
	}

	// Spotify config
	if clientID, ok := configs[repository.ConfigKeySpotifyClientID]; ok {
		s.config.Spotify.ClientID = clientID
	}
	if clientSecret, ok := configs[repository.ConfigKeySpotifyClientSecret]; ok {
		s.config.Spotify.ClientSecret = clientSecret
	}
	if enabled, ok := configs[repository.ConfigKeySpotifyEnabled]; ok {
		s.config.Spotify.Enabled = enabled == "true"
	}

	// MusicBrainz config
	if userAgent, ok := configs[repository.ConfigKeyMusicBrainzUserAgent]; ok {
		s.config.MusicBrainz.UserAgent = userAgent
	}
	if enabled, ok := configs[repository.ConfigKeyMusicBrainzEnabled]; ok {
		s.config.MusicBrainz.Enabled = enabled == "true"
	}

	// Update clients
	s.updateClients()

	// Load last parsed timestamp
	lastParsed, err := s.configRepo.GetLastParsedAt(ctx)
	if err == nil && lastParsed != nil {
		s.config.LastParsedAt = lastParsed
	}

	return nil
}

// updateClients updates API clients with current config
func (s *AppService) updateClients() {
	s.lastfmClient.SetCredentials(s.config.LastFM.APIKey, s.config.LastFM.APISecret)
	s.spotifyClient.SetCredentials(s.config.Spotify.ClientID, s.config.Spotify.ClientSecret)

	userAgent := s.config.MusicBrainz.UserAgent
	if userAgent == "" {
		userAgent = "Rocklist/1.0.0 ( https://github.com/Ardakilic/rocklist )"
	}
	s.musicbrainzClient.SetUserAgent(userAgent)

	// Register clients with playlist service
	if s.config.LastFM.Enabled {
		s.playlistService.RegisterClient(models.DataSourceLastFM, s.lastfmClient)
	}
	if s.config.Spotify.Enabled {
		s.playlistService.RegisterClient(models.DataSourceSpotify, s.spotifyClient)
	}
	if s.config.MusicBrainz.Enabled {
		s.playlistService.RegisterClient(models.DataSourceMusicBrainz, s.musicbrainzClient)
	}
}

// GetConfig returns the current configuration
func (s *AppService) GetConfig() *models.AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SaveConfig saves the configuration
func (s *AppService) SaveConfig(ctx context.Context, config *models.AppConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save to database
	if err := s.configRepo.Set(ctx, repository.ConfigKeyRockboxPath, config.RockboxPath); err != nil {
		return err
	}

	// Last.fm
	if err := s.configRepo.Set(ctx, repository.ConfigKeyLastFMAPIKey, config.LastFM.APIKey); err != nil {
		return err
	}
	if err := s.configRepo.Set(ctx, repository.ConfigKeyLastFMAPISecret, config.LastFM.APISecret); err != nil {
		return err
	}
	enabledStr := "false"
	if config.LastFM.Enabled {
		enabledStr = "true"
	}
	if err := s.configRepo.Set(ctx, repository.ConfigKeyLastFMEnabled, enabledStr); err != nil {
		return err
	}

	// Spotify
	if err := s.configRepo.Set(ctx, repository.ConfigKeySpotifyClientID, config.Spotify.ClientID); err != nil {
		return err
	}
	if err := s.configRepo.Set(ctx, repository.ConfigKeySpotifyClientSecret, config.Spotify.ClientSecret); err != nil {
		return err
	}
	enabledStr = "false"
	if config.Spotify.Enabled {
		enabledStr = "true"
	}
	if err := s.configRepo.Set(ctx, repository.ConfigKeySpotifyEnabled, enabledStr); err != nil {
		return err
	}

	// MusicBrainz
	if err := s.configRepo.Set(ctx, repository.ConfigKeyMusicBrainzUserAgent, config.MusicBrainz.UserAgent); err != nil {
		return err
	}
	enabledStr = "false"
	if config.MusicBrainz.Enabled {
		enabledStr = "true"
	}
	if err := s.configRepo.Set(ctx, repository.ConfigKeyMusicBrainzEnabled, enabledStr); err != nil {
		return err
	}

	s.config = config
	s.parser.SetPath(config.RockboxPath)
	s.updateClients()

	return nil
}

// SetRockboxPath sets the Rockbox device path
func (s *AppService) SetRockboxPath(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.RockboxPath = path
	s.parser.SetPath(path)

	// Update playlist directory
	playlistDir := filepath.Join(path, "Playlists")
	s.playlistService.SetPlaylistDir(playlistDir)

	return s.configRepo.Set(context.Background(), repository.ConfigKeyRockboxPath, path)
}

// ParseRockboxDatabase parses the Rockbox database
func (s *AppService) ParseRockboxDatabase(ctx context.Context, usePrefetched bool) error {
	logger := NewAppLogger(s.logBuffer)

	if usePrefetched {
		// Check if we have pre-fetched data
		count, err := s.songRepo.Count(ctx)
		if err != nil {
			return err
		}
		if count > 0 {
			logger.Info("Using pre-fetched data: %d songs", count)
			return nil
		}
		return models.ErrNoPreFetchedData
	}

	// Parse database
	songs, err := s.parser.Parse(ctx)
	if err != nil {
		return err
	}

	logger.Info("Parsed %d songs, saving to database...", len(songs))

	// Clear existing songs and save new ones
	if err := s.songRepo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("failed to clear existing songs: %w", err)
	}

	if err := s.songRepo.CreateBatch(ctx, songs); err != nil {
		return fmt.Errorf("failed to save songs: %w", err)
	}

	// Update last parsed timestamp
	now := time.Now()
	s.config.LastParsedAt = &now
	if err := s.configRepo.SetLastParsedAt(ctx, now); err != nil {
		logger.Error("Failed to save last parsed timestamp: %v", err)
	}

	logger.Info("Successfully saved %d songs to database", len(songs))
	return nil
}

// GetParseStatus returns the current parse status
func (s *AppService) GetParseStatus() *models.ParseStatus {
	return s.parser.GetStatus()
}

// GetLastParsedAt returns the last parsed timestamp
func (s *AppService) GetLastParsedAt(ctx context.Context) (*time.Time, error) {
	return s.configRepo.GetLastParsedAt(ctx)
}

// GeneratePlaylist generates a playlist
func (s *AppService) GeneratePlaylist(ctx context.Context, req *models.PlaylistRequest) (*models.Playlist, error) {
	playlist, err := s.playlistService.GeneratePlaylist(ctx, req)
	if err != nil {
		return nil, err
	}

	// Export to Rockbox
	_, err = s.playlistService.ExportPlaylist(ctx, playlist.ID)
	if err != nil {
		NewAppLogger(s.logBuffer).Error("Failed to export playlist: %v", err)
	}

	return playlist, nil
}

// GetAllSongs returns all songs
func (s *AppService) GetAllSongs(ctx context.Context) ([]*models.Song, error) {
	return s.songRepo.FindAll(ctx)
}

// GetSongCount returns the total number of songs
func (s *AppService) GetSongCount(ctx context.Context) (int64, error) {
	return s.songRepo.Count(ctx)
}

// GetUniqueArtists returns all unique album artists
func (s *AppService) GetUniqueArtists(ctx context.Context) ([]string, error) {
	return s.songRepo.GetUniqueArtists(ctx)
}

// GetUniqueGenres returns all unique genres
func (s *AppService) GetUniqueGenres(ctx context.Context) ([]string, error) {
	return s.songRepo.GetUniqueGenres(ctx)
}

// GetAllPlaylists returns all playlists
func (s *AppService) GetAllPlaylists(ctx context.Context) ([]*models.Playlist, error) {
	return s.playlistRepo.FindAll(ctx)
}

// GetPlaylist returns a playlist by ID
func (s *AppService) GetPlaylist(ctx context.Context, id uint) (*models.Playlist, error) {
	return s.playlistRepo.FindByID(ctx, id)
}

// DeletePlaylist deletes a playlist
func (s *AppService) DeletePlaylist(ctx context.Context, id uint) error {
	// Get playlist to find exported file
	playlist, err := s.playlistRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete exported file if exists
	if playlist.FilePath != "" {
		_ = os.Remove(playlist.FilePath)
	}

	return s.playlistRepo.Delete(ctx, id)
}

// WipeData wipes all pre-fetched data
func (s *AppService) WipeData(ctx context.Context) error {
	NewAppLogger(s.logBuffer).Info("Wiping all pre-fetched data...")
	return s.db.WipeData()
}

// GetLogs returns the current logs
func (s *AppService) GetLogs() []LogEntry {
	return s.logBuffer.GetAll()
}

// ClearLogs clears all logs
func (s *AppService) ClearLogs() {
	s.logBuffer.Clear()
}

// GetEnabledSources returns the list of enabled data sources
func (s *AppService) GetEnabledSources() []models.DataSource {
	return s.config.GetEnabledSources()
}

// ExportLogsToFile exports logs to a file
func (s *AppService) ExportLogsToFile(path string) error {
	logs := s.GetLogs()
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetAppInfo returns application information
func (s *AppService) GetAppInfo() map[string]string {
	return map[string]string{
		"name":        "Rocklist",
		"version":     "1.0.0",
		"author":      "Arda Kılıçdağı",
		"email":       "arda@kilicdagi.com",
		"repository":  "https://github.com/Ardakilic/rocklist",
		"license":     "MIT",
		"description": "A tool for creating playlists for Rockbox firmware devices using external music data sources",
	}
}
