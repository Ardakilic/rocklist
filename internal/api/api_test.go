package api

import (
	"fmt"
	"testing"

	"github.com/ardakilic/rocklist/internal/api/models"
)

// MockAPIClient is a mock implementation of the APIClient interface
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

// Mock the API client creation functions for testing
func init() {
	// Override the real implementations with test stubs
	originalNewLastFMClient := NewLastFMClient
	originalNewSpotifyClient := NewSpotifyClient
	
	// Replace with test versions
	NewLastFMClient = func(apiKey, apiSecret string) (*LastFMClient, error) {
		if apiKey == "" || apiSecret == "" {
			return nil, fmt.Errorf("Last.fm API key or secret not provided")
		}
		return &LastFMClient{}, nil
	}
	
	NewSpotifyClient = func(clientID, clientSecret string) (*SpotifyClient, error) {
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("Spotify client ID or secret not provided")
		}
		return &SpotifyClient{}, nil
	}
	
	// Restore after tests
	// Note: This won't actually run in the test, but shows the intention
	// In a real project, we would use a testing cleanup function
	_ = originalNewLastFMClient
	_ = originalNewSpotifyClient
}

func TestNewAPIClient(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		lastFMKey   string
		lastFMSecret string
		spotifyID   string
		spotifySecret string
		expectError bool
	}{
		{
			name:        "LastFM Valid",
			source:      "lastfm",
			lastFMKey:   "key",
			lastFMSecret: "secret",
			expectError: false,
		},
		{
			name:         "LastFM Missing Credentials",
			source:       "lastfm",
			expectError:  true,
		},
		{
			name:          "Spotify Valid",
			source:        "spotify",
			spotifyID:     "id",
			spotifySecret: "secret",
			expectError:   false,
		},
		{
			name:         "Spotify Missing Credentials",
			source:       "spotify",
			expectError:  true,
		},
		{
			name:         "Invalid Source",
			source:       "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAPIClient(tt.source, tt.lastFMKey, tt.lastFMSecret, tt.spotifyID, tt.spotifySecret)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if client == nil {
					t.Errorf("Expected client, got nil")
				}
			}
		})
	}
} 