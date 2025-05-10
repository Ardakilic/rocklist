package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadIndexEntries(t *testing.T) {
	// This is a basic test structure for the readIndexEntries function
	// In a real test, you would need to create a mock database file
	
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "rockbox-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a mock index file - this would be filled with actual test data
	// in a real test scenario
	mockFile := filepath.Join(tempDir, "database_idx.tcd")
	
	// Try to read the non-existent file - should return an error
	_, err = readIndexEntries(mockFile)
	if err == nil {
		t.Errorf("Expected error when reading non-existent file, got nil")
	}
}

func TestGetArtists(t *testing.T) {
	// Create a mock database
	db := &RockboxDB{
		ArtistTracks: map[string][]Track{
			"Artist1": {
				{Artist: "Artist1", Title: "Song1", Filename: "file1.mp3"},
				{Artist: "Artist1", Title: "Song2", Filename: "file2.mp3"},
			},
			"Artist2": {
				{Artist: "Artist2", Title: "Song3", Filename: "file3.mp3"},
			},
		},
	}
	
	// Get artists
	artists := db.GetArtists()
	
	// Check if we have the correct number of artists
	if len(artists) != 2 {
		t.Errorf("Expected 2 artists, got %d", len(artists))
	}
	
	// Check if all artists are included
	artistMap := make(map[string]bool)
	for _, artist := range artists {
		artistMap[artist] = true
	}
	
	expectedArtists := []string{"Artist1", "Artist2"}
	for _, artist := range expectedArtists {
		if !artistMap[artist] {
			t.Errorf("Expected artist %s not found", artist)
		}
	}
}

func TestGetTracksForArtist(t *testing.T) {
	// Create a mock database
	db := &RockboxDB{
		ArtistTracks: map[string][]Track{
			"Artist1": {
				{Artist: "Artist1", Title: "Song1", Filename: "file1.mp3"},
				{Artist: "Artist1", Title: "Song2", Filename: "file2.mp3"},
			},
			"Artist2": {
				{Artist: "Artist2", Title: "Song3", Filename: "file3.mp3"},
			},
		},
	}
	
	// Get tracks for Artist1
	tracks := db.GetTracksForArtist("Artist1")
	
	// Check if we have the correct number of tracks
	if len(tracks) != 2 {
		t.Errorf("Expected 2 tracks for Artist1, got %d", len(tracks))
	}
	
	// Get tracks for Artist2
	tracks = db.GetTracksForArtist("Artist2")
	
	// Check if we have the correct number of tracks
	if len(tracks) != 1 {
		t.Errorf("Expected 1 track for Artist2, got %d", len(tracks))
	}
	
	// Get tracks for non-existent artist
	tracks = db.GetTracksForArtist("NonExistentArtist")
	
	// Check if we have the correct number of tracks
	if len(tracks) != 0 {
		t.Errorf("Expected 0 tracks for NonExistentArtist, got %d", len(tracks))
	}
} 