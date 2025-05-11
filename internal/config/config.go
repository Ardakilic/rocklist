package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// API configuration
	LastFMAPIKey        string
	LastFMAPISecret     string
	SpotifyClientID     string
	SpotifyClientSecret string
	APISource           string // "lastfm" or "spotify"

	// Path configuration
	DapRootPath  string // Root path for the digital audio player
	PlaylistPath string // Path to playlists (inside DapRootPath)

	// Playlist configuration
	Artists            []string
	MaxTracksPerArtist int
}

// New returns a new Config instance with values loaded from environment variables and command line flags
func New() (*Config, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load() // Ignore error if .env file doesn't exist

	cfg := &Config{}

	// Define command line flags
	apiSource := flag.String("api-source", getEnv("API_SOURCE", "lastfm"), "API source to use (lastfm or spotify)")
	dapRootPath := flag.String("dap-root", getEnv("DAP_ROOT", ""), "Path to digital audio player's root directory")
	playlistPath := flag.String("playlist-path", getEnv("PLAYLIST_PATH", ""), "Path to output playlist directory")
	artists := flag.String("artist", getEnv("ARTISTS", ""), "Comma-separated list of artists to generate playlists for")
	maxTracks := flag.Int("tracks", 10, "Maximum number of tracks per artist")

	// Parse command line flags
	flag.Parse()

	// Set configuration values
	cfg.LastFMAPIKey = getEnv("LASTFM_API_KEY", "")
	cfg.LastFMAPISecret = getEnv("LASTFM_API_SECRET", "")
	cfg.SpotifyClientID = getEnv("SPOTIFY_CLIENT_ID", "")
	cfg.SpotifyClientSecret = getEnv("SPOTIFY_CLIENT_SECRET", "")
	cfg.APISource = *apiSource
	cfg.MaxTracksPerArtist = *maxTracks

	// Parse artist list
	if *artists != "" {
		cfg.Artists = strings.Split(*artists, ",")
		for i := range cfg.Artists {
			cfg.Artists[i] = strings.TrimSpace(cfg.Artists[i])
		}
	}

	// Set DAP root path
	cfg.DapRootPath = *dapRootPath

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Set default playlist path if not provided
	if *playlistPath == "" {
		cfg.PlaylistPath = filepath.Join(cfg.DapRootPath, "Playlists")
	} else {
		cfg.PlaylistPath = *playlistPath
	}

	return cfg, nil
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Validate API source
	if c.APISource != "lastfm" && c.APISource != "spotify" {
		return fmt.Errorf("invalid API source: %s. Must be 'lastfm' or 'spotify'", c.APISource)
	}

	// Validate API credentials
	if c.APISource == "lastfm" && (c.LastFMAPIKey == "" || c.LastFMAPISecret == "") {
		return fmt.Errorf("last.fm API credentials not provided")
	}

	if c.APISource == "spotify" && (c.SpotifyClientID == "" || c.SpotifyClientSecret == "") {
		return fmt.Errorf("spotify API credentials not provided")
	}

	// Validate DAP root path
	if c.DapRootPath == "" {
		return fmt.Errorf("DAP root path not provided")
	}

	if _, err := os.Stat(c.DapRootPath); os.IsNotExist(err) {
		return fmt.Errorf("DAP root path does not exist: %s", c.DapRootPath)
	}

	// Ensure .rockbox directory exists within DAP root
	rockboxDir := filepath.Join(c.DapRootPath, ".rockbox")
	if _, err := os.Stat(rockboxDir); os.IsNotExist(err) {
		return fmt.Errorf(".rockbox directory does not exist: %s", rockboxDir)
	}

	// Validate database files
	dbPath := filepath.Join(rockboxDir, "database_idx.tcd")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("rockbox database not found at %s", dbPath)
	}

	return nil
}

// getEnv returns the value of the environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
