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
	LastFMAPIKey      string
	LastFMAPISecret   string
	SpotifyClientID   string
	SpotifyClientSecret string
	APISource         string // "lastfm" or "spotify"
	
	// Path configuration
	RockboxPath       string
	PlaylistPath      string
	
	// Playlist configuration
	Artists           []string
	MaxTracksPerArtist int
}

// New returns a new Config instance with values loaded from environment variables and command line flags
func New() (*Config, error) {
	// Load environment variables from .env file if it exists
	_ = godotenv.Load() // Ignore error if .env file doesn't exist
	
	cfg := &Config{}

	// Define command line flags
	apiSource := flag.String("api-source", getEnv("API_SOURCE", "lastfm"), "API source to use (lastfm or spotify)")
	rockboxPath := flag.String("rockbox-path", getEnv("ROCKBOX_PATH", ""), "Path to Rockbox directory")
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
	cfg.RockboxPath = *rockboxPath
	cfg.MaxTracksPerArtist = *maxTracks
	
	// Parse artist list
	if *artists != "" {
		cfg.Artists = strings.Split(*artists, ",")
		for i := range cfg.Artists {
			cfg.Artists[i] = strings.TrimSpace(cfg.Artists[i])
		}
	}
	
	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	
	// Set default playlist path if not provided
	if *playlistPath == "" {
		cfg.PlaylistPath = filepath.Join(cfg.RockboxPath, "playlists")
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
		return fmt.Errorf("Last.fm API credentials not provided")
	}
	
	if c.APISource == "spotify" && (c.SpotifyClientID == "" || c.SpotifyClientSecret == "") {
		return fmt.Errorf("Spotify API credentials not provided")
	}
	
	// Validate Rockbox path
	if c.RockboxPath == "" {
		return fmt.Errorf("Rockbox path not provided")
	}
	
	if _, err := os.Stat(c.RockboxPath); os.IsNotExist(err) {
		return fmt.Errorf("Rockbox path does not exist: %s", c.RockboxPath)
	}
	
	// Validate database files
	dbPath := filepath.Join(c.RockboxPath, "database_idx.tcd")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("Rockbox database not found at %s", dbPath)
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