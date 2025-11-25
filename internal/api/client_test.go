package api

import (
	"testing"

	"github.com/Ardakilic/rocklist/internal/models"
)

func TestNewBaseClient(t *testing.T) {
	client := NewBaseClient(models.DataSourceLastFM, "TestAgent/1.0")

	if client == nil {
		t.Fatal("NewBaseClient() returned nil")
	}

	if client.source != models.DataSourceLastFM {
		t.Errorf("NewBaseClient() source = %v, want %v", client.source, models.DataSourceLastFM)
	}

	if client.userAgent != "TestAgent/1.0" {
		t.Errorf("NewBaseClient() userAgent = %v, want %v", client.userAgent, "TestAgent/1.0")
	}
}

func TestBaseClient_GetSource(t *testing.T) {
	tests := []struct {
		name   string
		source models.DataSource
	}{
		{"LastFM", models.DataSourceLastFM},
		{"Spotify", models.DataSourceSpotify},
		{"MusicBrainz", models.DataSourceMusicBrainz},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewBaseClient(tt.source, "Test/1.0")
			if client.GetSource() != tt.source {
				t.Errorf("GetSource() = %v, want %v", client.GetSource(), tt.source)
			}
		})
	}
}

func TestBaseClient_HTTPClient(t *testing.T) {
	client := NewBaseClient(models.DataSourceLastFM, "Test/1.0")

	httpClient := client.HTTPClient()
	if httpClient == nil {
		t.Error("HTTPClient() returned nil")
		return
	}

	if httpClient.Timeout == 0 {
		t.Error("HTTPClient() should have a timeout set")
	}
}

func TestBaseClient_UserAgent(t *testing.T) {
	userAgent := "MyApp/2.0"
	client := NewBaseClient(models.DataSourceLastFM, userAgent)

	if client.UserAgent() != userAgent {
		t.Errorf("UserAgent() = %v, want %v", client.UserAgent(), userAgent)
	}
}

func TestTrackMatch_Fields(t *testing.T) {
	match := &TrackMatch{
		ExternalID: "abc123",
		Artist:     "Test Artist",
		Title:      "Test Title",
		Album:      "Test Album",
		Confidence: 0.95,
		Source:     models.DataSourceLastFM,
		URL:        "https://example.com/track",
		Duration:   300,
		Playcount:  1000,
	}

	if match.ExternalID != "abc123" {
		t.Errorf("TrackMatch.ExternalID = %v, want %v", match.ExternalID, "abc123")
	}
	if match.Artist != "Test Artist" {
		t.Errorf("TrackMatch.Artist = %v, want %v", match.Artist, "Test Artist")
	}
	if match.Confidence != 0.95 {
		t.Errorf("TrackMatch.Confidence = %v, want %v", match.Confidence, 0.95)
	}
}

func TestTrackInfo_Fields(t *testing.T) {
	info := &TrackInfo{
		ExternalID: "xyz789",
		Artist:     "Another Artist",
		Title:      "Another Title",
		Album:      "Another Album",
		Rank:       1,
		Playcount:  5000,
		Duration:   240,
		URL:        "https://example.com/track2",
		Source:     models.DataSourceSpotify,
	}

	if info.Rank != 1 {
		t.Errorf("TrackInfo.Rank = %v, want %v", info.Rank, 1)
	}
	if info.Source != models.DataSourceSpotify {
		t.Errorf("TrackInfo.Source = %v, want %v", info.Source, models.DataSourceSpotify)
	}
}

func TestArtistInfo_Fields(t *testing.T) {
	info := &ArtistInfo{
		ExternalID: "artist123",
		Name:       "Famous Artist",
		URL:        "https://example.com/artist",
		ImageURL:   "https://example.com/image.jpg",
		Listeners:  1000000,
		Playcount:  50000000,
		Tags:       []string{"rock", "alternative"},
		Similar:    []string{"Similar Artist 1", "Similar Artist 2"},
		Source:     models.DataSourceMusicBrainz,
	}

	if info.Name != "Famous Artist" {
		t.Errorf("ArtistInfo.Name = %v, want %v", info.Name, "Famous Artist")
	}
	if len(info.Tags) != 2 {
		t.Errorf("ArtistInfo.Tags length = %v, want %v", len(info.Tags), 2)
	}
	if len(info.Similar) != 2 {
		t.Errorf("ArtistInfo.Similar length = %v, want %v", len(info.Similar), 2)
	}
}
