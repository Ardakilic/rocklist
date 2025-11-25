package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ardakilic/rocklist/internal/models"
)

func TestNewSpotifyClient(t *testing.T) {
	client := NewSpotifyClient("client_id", "client_secret", nil)

	if client == nil {
		t.Fatal("NewSpotifyClient() returned nil")
	}

	if client.clientID != "client_id" {
		t.Errorf("NewSpotifyClient() clientID = %v, want %v", client.clientID, "client_id")
	}

	if client.clientSecret != "client_secret" {
		t.Errorf("NewSpotifyClient() clientSecret = %v, want %v", client.clientSecret, "client_secret")
	}
}

func TestSpotifyClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		want         bool
	}{
		{"fully configured", "id", "secret", true},
		{"no client ID", "", "secret", false},
		{"no client secret", "id", "", false},
		{"nothing configured", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSpotifyClient(tt.clientID, tt.clientSecret, nil)
			if client.IsConfigured() != tt.want {
				t.Errorf("IsConfigured() = %v, want %v", client.IsConfigured(), tt.want)
			}
		})
	}
}

func TestSpotifyClient_SetCredentials(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	if client.IsConfigured() {
		t.Error("Client should not be configured initially")
	}

	// Set an initial token to verify it gets reset
	client.accessToken = "old_token"

	client.SetCredentials("new_id", "new_secret")

	if !client.IsConfigured() {
		t.Error("Client should be configured after SetCredentials")
	}

	if client.clientID != "new_id" {
		t.Errorf("clientID = %v, want %v", client.clientID, "new_id")
	}

	if client.clientSecret != "new_secret" {
		t.Errorf("clientSecret = %v, want %v", client.clientSecret, "new_secret")
	}

	if client.accessToken != "" {
		t.Error("accessToken should be reset after SetCredentials")
	}
}

func TestSpotifyClient_GetSource(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	if client.GetSource() != models.DataSourceSpotify {
		t.Errorf("GetSource() = %v, want %v", client.GetSource(), models.DataSourceSpotify)
	}
}

func TestSpotifyClient_BaseClient(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	if client.BaseClient == nil {
		t.Error("BaseClient should not be nil")
	}

	if client.HTTPClient() == nil {
		t.Error("HTTPClient() should not return nil")
	}

	if client.UserAgent() == "" {
		t.Error("UserAgent() should not be empty")
	}
}

func TestSpotifyResponseStructures(t *testing.T) {
	// Test that response structures compile and are usable
	_ = spotifySearchResponse{}
	_ = spotifyTopTracksResponse{}
	_ = spotifyRecommendationsResponse{}
	_ = spotifyArtistSearchResponse{}
	_ = spotifyRelatedArtistsResponse{}
}

func TestSpotifyClient_GetAccessToken_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.getAccessToken(context.Background())
	if err != models.ErrAPIKeyMissing {
		t.Errorf("getAccessToken() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_TokenCaching(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	// Token should be empty initially
	if client.accessToken != "" {
		t.Error("Initial accessToken should be empty")
	}

	// After setting credentials, token should be reset
	client.accessToken = "old_token"
	client.SetCredentials("new_id", "new_secret")

	if client.accessToken != "" {
		t.Error("accessToken should be empty after SetCredentials")
	}
}

func TestSpotifyClient_ConcurrentAccess(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	// Test concurrent SetCredentials calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			client.SetCredentials("id"+string(rune(i)), "secret"+string(rune(i)))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should be configured
	if !client.IsConfigured() {
		t.Error("Client should be configured after SetCredentials")
	}
}

func TestSpotifyTrack_Fields(t *testing.T) {
	track := spotifyTrack{
		ID:         "track123",
		Name:       "Test Track",
		Popularity: 80,
		DurationMS: 240000,
	}

	if track.ID != "track123" {
		t.Errorf("spotifyTrack.ID = %v, want %v", track.ID, "track123")
	}
	if track.DurationMS != 240000 {
		t.Errorf("spotifyTrack.DurationMS = %v, want %v", track.DurationMS, 240000)
	}
}

func TestSpotifyConstants(t *testing.T) {
	if spotifyAuthURL != "https://accounts.spotify.com/api/token" {
		t.Errorf("spotifyAuthURL = %v, unexpected", spotifyAuthURL)
	}
	if spotifyAPIURL != "https://api.spotify.com/v1" {
		t.Errorf("spotifyAPIURL = %v, unexpected", spotifyAPIURL)
	}
}

func TestSpotifyClient_SearchTrack_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.SearchTrack(context.Background(), "Artist", "Title")
	if err != models.ErrAPIKeyMissing {
		t.Errorf("SearchTrack() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_GetTopTracks_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.GetTopTracks(context.Background(), "Artist", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetTopTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_GetSimilarTracks_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.GetSimilarTracks(context.Background(), "Artist", "Title", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetSimilarTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_GetTagTracks_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.GetTagTracks(context.Background(), "rock", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetTagTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_GetArtistInfo_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.GetArtistInfo(context.Background(), "Artist")
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetArtistInfo() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_GetSimilarArtists_NotConfigured(t *testing.T) {
	client := NewSpotifyClient("", "", nil)

	_, err := client.GetSimilarArtists(context.Background(), "Artist", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetSimilarArtists() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestSpotifyClient_UserAgent(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	if client.UserAgent() == "" {
		t.Error("UserAgent() should not be empty")
	}
}

func TestSpotifyArtistResponse_Fields(t *testing.T) {
	resp := spotifyArtistResponse{
		ID:     "artist123",
		Name:   "Test Artist",
		Genres: []string{"rock", "metal"},
	}

	if resp.ID != "artist123" {
		t.Errorf("spotifyArtistResponse.ID = %v, want artist123", resp.ID)
	}
	if len(resp.Genres) != 2 {
		t.Errorf("spotifyArtistResponse.Genres length = %v, want 2", len(resp.Genres))
	}
}

func TestSpotifyClient_SetBaseURLs(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	client.SetAuthURL("http://auth.test/")
	client.SetAPIURL("http://api.test/")

	if client.getAuthURL() != "http://auth.test/" {
		t.Errorf("getAuthURL() = %v, want http://auth.test/", client.getAuthURL())
	}
	if client.getAPIURL() != "http://api.test/" {
		t.Errorf("getAPIURL() = %v, want http://api.test/", client.getAPIURL())
	}
}

func TestSpotifyClient_GetBaseURLs_Default(t *testing.T) {
	client := NewSpotifyClient("id", "secret", nil)

	if client.getAuthURL() != spotifyAuthURL {
		t.Errorf("getAuthURL() = %v, want %v", client.getAuthURL(), spotifyAuthURL)
	}
	if client.getAPIURL() != spotifyAPIURL {
		t.Errorf("getAPIURL() = %v, want %v", client.getAPIURL(), spotifyAPIURL)
	}
}

func TestSpotifyClient_GetAccessToken_Success(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "test_token", "token_type": "Bearer", "expires_in": 3600}`))
	}))
	defer authServer.Close()

	client := NewSpotifyClient("client_id", "client_secret", nil)
	client.SetAuthURL(authServer.URL)

	token, err := client.getAccessToken(context.Background())
	if err != nil {
		t.Fatalf("getAccessToken() error = %v", err)
	}
	if token != "test_token" {
		t.Errorf("getAccessToken() = %v, want test_token", token)
	}
}

func TestSpotifyClient_GetAccessToken_Unauthorized(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer authServer.Close()

	client := NewSpotifyClient("client_id", "client_secret", nil)
	client.SetAuthURL(authServer.URL)

	_, err := client.getAccessToken(context.Background())
	if err == nil {
		t.Error("getAccessToken() should return error for unauthorized")
	}
}

func TestSpotifyClient_SearchTrack_Success(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "test_token", "token_type": "Bearer", "expires_in": 3600}`))
	}))
	defer authServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tracks": {
				"items": [
					{
						"id": "track1",
						"name": "Test Track",
						"artists": [{"id": "artist1", "name": "Test Artist"}],
						"album": {"name": "Test Album", "release_date": "2023-01-01"},
						"duration_ms": 180000,
						"popularity": 80
					}
				]
			}
		}`))
	}))
	defer apiServer.Close()

	client := NewSpotifyClient("client_id", "client_secret", nil)
	client.SetAuthURL(authServer.URL)
	client.SetAPIURL(apiServer.URL)

	match, err := client.SearchTrack(context.Background(), "Test Artist", "Test Track")
	if err != nil {
		t.Fatalf("SearchTrack() error = %v", err)
	}
	if match == nil {
		t.Fatal("SearchTrack() returned nil")
	}
	if match.Title != "Test Track" {
		t.Errorf("SearchTrack() Title = %v, want Test Track", match.Title)
	}
}

func TestSpotifyClient_GetTopTracks_Success(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "test_token", "token_type": "Bearer", "expires_in": 3600}`))
	}))
	defer authServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/search") && strings.Contains(r.URL.RawQuery, "type=artist") {
			_, _ = w.Write([]byte(`{"artists": {"items": [{"id": "artist1", "name": "Test Artist"}]}}`))
		} else if strings.Contains(r.URL.Path, "/artists/") && strings.Contains(r.URL.Path, "/top-tracks") {
			_, _ = w.Write([]byte(`{
				"tracks": [
					{"id": "track1", "name": "Track 1", "artists": [{"name": "Artist"}], "album": {"name": "Album"}, "duration_ms": 180000, "popularity": 90},
					{"id": "track2", "name": "Track 2", "artists": [{"name": "Artist"}], "album": {"name": "Album"}, "duration_ms": 200000, "popularity": 80}
				]
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	client := NewSpotifyClient("client_id", "client_secret", nil)
	client.SetAuthURL(authServer.URL)
	client.SetAPIURL(apiServer.URL)

	tracks, err := client.GetTopTracks(context.Background(), "Test Artist", 10)
	if err != nil {
		t.Fatalf("GetTopTracks() error = %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("GetTopTracks() returned %d tracks, want 2", len(tracks))
	}
}
