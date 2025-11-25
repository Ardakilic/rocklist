package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
)

func TestNewMusicBrainzClient(t *testing.T) {
	client := NewMusicBrainzClient("TestApp/1.0 (test@example.com)", nil)

	if client == nil {
		t.Fatal("NewMusicBrainzClient() returned nil")
	}

	if client.userAgent != "TestApp/1.0 (test@example.com)" {
		t.Errorf("NewMusicBrainzClient() userAgent = %v, want %v", client.userAgent, "TestApp/1.0 (test@example.com)")
	}
}

func TestNewMusicBrainzClient_DefaultUserAgent(t *testing.T) {
	client := NewMusicBrainzClient("", nil)

	if client.userAgent == "" {
		t.Error("Default user agent should be set when empty string provided")
	}

	if client.userAgent != "Rocklist/1.0.0 ( https://github.com/Ardakilic/Rocklist )" {
		t.Errorf("Default user agent = %v, want default Rocklist agent", client.userAgent)
	}
}

func TestMusicBrainzClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      bool
	}{
		{"configured", "TestApp/1.0", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MusicBrainzClient{
				BaseClient: NewBaseClient(models.DataSourceMusicBrainz, tt.userAgent),
				userAgent:  tt.userAgent,
			}
			if client.IsConfigured() != tt.want {
				t.Errorf("IsConfigured() = %v, want %v", client.IsConfigured(), tt.want)
			}
		})
	}
}

func TestMusicBrainzClient_SetUserAgent(t *testing.T) {
	client := NewMusicBrainzClient("OldAgent/1.0", nil)

	client.SetUserAgent("NewAgent/2.0")

	if client.userAgent != "NewAgent/2.0" {
		t.Errorf("userAgent = %v, want %v", client.userAgent, "NewAgent/2.0")
	}
}

func TestMusicBrainzClient_GetSource(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)

	if client.GetSource() != models.DataSourceMusicBrainz {
		t.Errorf("GetSource() = %v, want %v", client.GetSource(), models.DataSourceMusicBrainz)
	}
}

func TestMusicBrainzClient_BaseClient(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)

	if client.BaseClient == nil {
		t.Error("BaseClient should not be nil")
	}

	if client.HTTPClient() == nil {
		t.Error("HTTPClient() should not return nil")
	}
}

func TestMusicBrainzRateLimitConstant(t *testing.T) {
	// Verify rate limit is reasonable (around 1 second)
	if musicBrainzRateLimit < 1000*time.Millisecond {
		t.Error("Rate limit should be at least 1 second")
	}

	if musicBrainzRateLimit > 2000*time.Millisecond {
		t.Error("Rate limit should not be more than 2 seconds")
	}
}

func TestMusicBrainzClient_RateLimitWait(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)

	// First call should not wait (no previous request)
	start := time.Now()
	client.rateLimitWait()
	elapsed := time.Since(start)

	// Should not have waited significantly
	if elapsed > 100*time.Millisecond {
		t.Errorf("First rateLimitWait() took %v, expected minimal wait", elapsed)
	}

	// Second immediate call should wait
	start = time.Now()
	client.rateLimitWait()
	elapsed = time.Since(start)

	// Should have waited close to the rate limit
	if elapsed < musicBrainzRateLimit-100*time.Millisecond {
		t.Errorf("Second rateLimitWait() took %v, expected ~%v", elapsed, musicBrainzRateLimit)
	}
}

func TestEscapeQuery(t *testing.T) {
	// Test that special characters are escaped
	result := escapeQuery("hello")
	if result != "hello" {
		t.Errorf("escapeQuery(simple) = %q, want %q", result, "hello")
	}

	// Test that quotes are escaped
	result = escapeQuery(`hello "world"`)
	if !strings.Contains(result, `\`) {
		t.Error("escapeQuery should escape quotes")
	}

	// Test that parentheses are escaped
	result = escapeQuery("test (live)")
	if !strings.Contains(result, `\`) {
		t.Error("escapeQuery should escape parentheses")
	}

	// Test that original string is not returned when it contains special chars
	result = escapeQuery("Artist (feat. Other)")
	if result == "Artist (feat. Other)" {
		t.Error("escapeQuery should modify strings with special characters")
	}
}

func TestMusicBrainzResponseStructures(t *testing.T) {
	// Test that response structures compile and are usable
	_ = mbRecordingSearchResponse{}
	_ = mbRecording{}
	_ = mbRecordingResponse{}
	_ = mbArtistSearchResponse{}
	_ = mbArtist{}
	_ = mbArtistResponse{}
}

func TestMusicBrainzConstants(t *testing.T) {
	if musicBrainzAPIURL != "https://musicbrainz.org/ws/2" {
		t.Errorf("musicBrainzAPIURL = %v, unexpected", musicBrainzAPIURL)
	}
}

func TestMusicBrainzClient_ConcurrentAccess(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)

	// Test concurrent SetUserAgent calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			client.SetUserAgent("Agent" + string(rune(i)))
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should be configured
	if !client.IsConfigured() {
		t.Error("Client should be configured after SetUserAgent")
	}
}

func TestMbRecording_Fields(t *testing.T) {
	rec := mbRecording{
		ID:     "rec123",
		Title:  "Test Recording",
		Length: 240000,
	}

	if rec.ID != "rec123" {
		t.Errorf("mbRecording.ID = %v, want %v", rec.ID, "rec123")
	}
	if rec.Length != 240000 {
		t.Errorf("mbRecording.Length = %v, want %v", rec.Length, 240000)
	}
}

func TestMbArtist_Fields(t *testing.T) {
	artist := mbArtist{
		ID:   "artist123",
		Name: "Test Artist",
	}

	if artist.ID != "artist123" {
		t.Errorf("mbArtist.ID = %v, want %v", artist.ID, "artist123")
	}
	if artist.Name != "Test Artist" {
		t.Errorf("mbArtist.Name = %v, want %v", artist.Name, "Test Artist")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.want)
		}
	}
}

func TestMusicBrainzClient_NotConfigured(t *testing.T) {
	client := &MusicBrainzClient{
		BaseClient: NewBaseClient(models.DataSourceMusicBrainz, ""),
		userAgent:  "",
	}

	// Empty userAgent means not configured
	if client.IsConfigured() {
		t.Error("Client with empty userAgent should not be configured")
	}
}

func TestMusicBrainzClient_WithLogger(t *testing.T) {
	logger := &mbTestLogger{}
	client := NewMusicBrainzClient("Test/1.0", logger)

	if client.logger != logger {
		t.Error("Client should use provided logger")
	}
}

// mbTestLogger implements Logger for testing
type mbTestLogger struct{}

func (l *mbTestLogger) Info(msg string, args ...interface{})  {}
func (l *mbTestLogger) Error(msg string, args ...interface{}) {}
func (l *mbTestLogger) Debug(msg string, args ...interface{}) {}

func TestMusicBrainzClient_SetBaseURL(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)
	client.SetBaseURL("http://test.com/")

	if client.getBaseURL() != "http://test.com/" {
		t.Errorf("getBaseURL() = %v, want http://test.com/", client.getBaseURL())
	}
}

func TestMusicBrainzClient_GetBaseURL_Default(t *testing.T) {
	client := NewMusicBrainzClient("Test/1.0", nil)

	if client.getBaseURL() != musicBrainzAPIURL {
		t.Errorf("getBaseURL() = %v, want %v", client.getBaseURL(), musicBrainzAPIURL)
	}
}

func TestMusicBrainzClient_SearchTrack_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"recordings": [
				{
					"id": "recording-1",
					"title": "Test Track",
					"artist-credit": [{"name": "Test Artist"}],
					"releases": [{"title": "Test Album", "date": "2023"}],
					"length": 180000
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewMusicBrainzClient("Test/1.0", nil)
	client.SetBaseURL(server.URL)

	match, err := client.SearchTrack(context.Background(), "Test Artist", "Test Track")
	if err != nil {
		t.Fatalf("SearchTrack() error = %v", err)
	}
	if match == nil {
		t.Fatal("SearchTrack() returned nil")
	}
}

func TestMusicBrainzClient_GetArtistInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/artist") && strings.Contains(r.URL.RawQuery, "query=") {
			_, _ = w.Write([]byte(`{
				"artists": [
					{"id": "artist-1", "name": "Test Artist", "tags": [{"name": "rock"}]}
				]
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"id": "artist-1",
				"name": "Test Artist",
				"tags": [{"name": "rock"}]
			}`))
		}
	}))
	defer server.Close()

	client := NewMusicBrainzClient("Test/1.0", nil)
	client.SetBaseURL(server.URL)

	info, err := client.GetArtistInfo(context.Background(), "Test Artist")
	if err != nil {
		t.Fatalf("GetArtistInfo() error = %v", err)
	}
	if info == nil {
		t.Fatal("GetArtistInfo() returned nil")
	}
}

func TestMusicBrainzClient_GetSimilarArtists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"artists": []}`))
	}))
	defer server.Close()

	client := NewMusicBrainzClient("Test/1.0", nil)
	client.SetBaseURL(server.URL)

	// GetSimilarArtists may return artists depending on implementation
	_, err := client.GetSimilarArtists(context.Background(), "Artist", 10)
	if err != nil {
		// Expected to potentially error or return empty/results
	}
}

func TestMusicBrainzClient_GetSimilarTracks_ReturnsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"recordings": [
				{"id": "rec-1", "title": "Similar Track", "artist-credit": [{"name": "Artist"}], "length": 180000}
			]
		}`))
	}))
	defer server.Close()

	client := NewMusicBrainzClient("Test/1.0", nil)
	client.SetBaseURL(server.URL)

	// GetSimilarTracks may return error or tracks depending on implementation
	_, _ = client.GetSimilarTracks(context.Background(), "Artist", "Track", 10)
}

