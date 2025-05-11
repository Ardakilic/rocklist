package playlist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ardakilic/rocklist/internal/api/models"
	"github.com/ardakilic/rocklist/internal/database"
)

// MockAPIClient is a mock implementation of the APIClient interface for testing
type MockAPIClient struct {
	tracks []models.TopTrack
	err    error
}

func (m *MockAPIClient) GetTopTracks(artist string, limit int) ([]models.TopTrack, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tracks, nil
}

func TestMatchTracks(t *testing.T) {
	// Create a mock database
	db := &database.RockboxDB{
		ArtistTracks: map[string][]database.Track{
			"Artist1": {
				{Artist: "Artist1", Title: "Song1", Filename: "file1.mp3"},
				{Artist: "Artist1", Title: "Song2", Filename: "file2.mp3"},
				{Artist: "Artist1", Title: "Different Song", Filename: "file3.mp3"},
			},
		},
	}

	// Create a mock API client
	apiClient := &MockAPIClient{
		tracks: []models.TopTrack{
			{Name: "Song1", Artist: "Artist1", Rank: 1},
			{Name: "Song2", Artist: "Artist1", Rank: 2},
			{Name: "Song3", Artist: "Artist1", Rank: 3}, // This one won't match
		},
	}

	// Create a temporary directory for playlists
	tempDir, err := os.MkdirTemp("", "dap-root-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a playlist directory
	playlistDir := filepath.Join(tempDir, "Playlists")
	if err := os.MkdirAll(playlistDir, 0755); err != nil {
		t.Fatalf("Failed to create Playlists directory: %v", err)
	}

	// Create a playlist generator
	generator := NewGenerator(db, apiClient, playlistDir, 10)

	// Test matchTracks
	matched := generator.matchTracks(apiClient.tracks, db.GetTracksForArtist("Artist1"))

	// Check if we matched the correct number of tracks
	if len(matched) != 2 {
		t.Errorf("Expected 2 matched tracks, got %d", len(matched))
	}

	// Check if the matched tracks have the correct filenames
	for _, track := range matched {
		if !track.Found {
			t.Errorf("Track %s should be marked as found", track.Name)
		}

		switch track.Name {
		case "Song1":
			if track.Filename != "file1.mp3" {
				t.Errorf("Expected filename file1.mp3 for Song1, got %s", track.Filename)
			}
		case "Song2":
			if track.Filename != "file2.mp3" {
				t.Errorf("Expected filename file2.mp3 for Song2, got %s", track.Filename)
			}
		}
	}
}

func TestWritePlaylistFile(t *testing.T) {
	// Create a temporary directory for playlists
	tempDir, err := os.MkdirTemp("", "dap-root-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a playlist directory
	playlistDir := filepath.Join(tempDir, "Playlists")
	if err := os.MkdirAll(playlistDir, 0755); err != nil {
		t.Fatalf("Failed to create Playlists directory: %v", err)
	}

	// Create a mock database
	db := &database.RockboxDB{}

	// Create a mock API client
	apiClient := &MockAPIClient{}

	// Create a playlist generator
	generator := NewGenerator(db, apiClient, playlistDir, 10)

	// Create mock tracks - only include found tracks
	tracks := []models.TopTrack{
		{Name: "Song1", Artist: "Artist1", Rank: 1, Filename: "file1.mp3", Found: true},
		{Name: "Song2", Artist: "Artist1", Rank: 2, Filename: "file2.mp3", Found: true},
	}

	// Test writePlaylistFile
	if err := generator.writePlaylistFile("Artist1", tracks); err != nil {
		t.Fatalf("Failed to write playlist file: %v", err)
	}

	// Check if the playlist file was created
	playlistFile := filepath.Join(playlistDir, "Artist1-top-2.m3u")
	if _, err := os.Stat(playlistFile); os.IsNotExist(err) {
		t.Errorf("Playlist file %s not created", playlistFile)
	}

	// Read the playlist file
	content, err := os.ReadFile(playlistFile)
	if err != nil {
		t.Fatalf("Failed to read playlist file: %v", err)
	}

	// Check the content of the playlist file
	expectedContent := "#EXTM3U\n" +
		"#EXTINF:-1,Artist1 - Song1\n" +
		"file1.mp3\n" +
		"#EXTINF:-1,Artist1 - Song2\n" +
		"file2.mp3\n"

	if string(content) != expectedContent {
		t.Errorf("Expected playlist content:\n%s\n\nGot:\n%s", expectedContent, string(content))
	}
}
