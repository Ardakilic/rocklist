package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test with non-existent environment variable
	if val := getEnv("NON_EXISTENT_ENV_VAR", "default"); val != "default" {
		t.Errorf("Expected default value for non-existent env var, got %s", val)
	}

	// Test with existing environment variable
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	if val := getEnv("TEST_ENV_VAR", "default"); val != "test_value" {
		t.Errorf("Expected test_value for existing env var, got %s", val)
	}
}

func TestConfig_Validate(t *testing.T) {
	// Create a temporary directory to simulate a DAP root directory
	tempDir, err := os.MkdirTemp("", "dap-root-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock .rockbox directory
	rockboxDir := filepath.Join(tempDir, ".rockbox")
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("Failed to create mock .rockbox directory: %v", err)
	}

	// Create a mock database file
	dbFile := filepath.Join(rockboxDir, "database_idx.tcd")
	if err := os.WriteFile(dbFile, []byte("mock data"), 0644); err != nil {
		t.Fatalf("Failed to create mock database file: %v", err)
	}

	// Test cases
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "Valid LastFM Config with DAP Root",
			config: Config{
				APISource:       "lastfm",
				LastFMAPIKey:    "key",
				LastFMAPISecret: "secret",
				DapRootPath:     tempDir,
			},
			expectError: false,
		},
		{
			name: "Valid Spotify Config with DAP Root",
			config: Config{
				APISource:           "spotify",
				SpotifyClientID:     "id",
				SpotifyClientSecret: "secret",
				DapRootPath:         tempDir,
			},
			expectError: false,
		},
		{
			name: "Invalid API Source",
			config: Config{
				APISource:   "invalid",
				DapRootPath: tempDir,
			},
			expectError: true,
		},
		{
			name: "Missing LastFM Credentials",
			config: Config{
				APISource:   "lastfm",
				DapRootPath: tempDir,
			},
			expectError: true,
		},
		{
			name: "Missing Spotify Credentials",
			config: Config{
				APISource:   "spotify",
				DapRootPath: tempDir,
			},
			expectError: true,
		},
		{
			name: "Empty DAP Root Path",
			config: Config{
				APISource:       "lastfm",
				LastFMAPIKey:    "key",
				LastFMAPISecret: "secret",
				DapRootPath:     "",
			},
			expectError: true,
		},
		{
			name: "Non-existent DAP Root Path",
			config: Config{
				APISource:       "lastfm",
				LastFMAPIKey:    "key",
				LastFMAPISecret: "secret",
				DapRootPath:     "/non/existent/path",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.expectError {
				t.Errorf("validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
