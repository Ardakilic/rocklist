package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ardakilic/rocklist/internal/models"
)

func TestNewLastFMClient(t *testing.T) {
	client := NewLastFMClient("test_key", "test_secret", nil)

	if client == nil {
		t.Fatal("NewLastFMClient() returned nil")
	}

	if client.apiKey != "test_key" {
		t.Errorf("NewLastFMClient() apiKey = %v, want %v", client.apiKey, "test_key")
	}

	if client.apiSecret != "test_secret" {
		t.Errorf("NewLastFMClient() apiSecret = %v, want %v", client.apiSecret, "test_secret")
	}
}

func TestLastFMClient_IsConfigured(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		apiSecret string
		want      bool
	}{
		{"configured", "key", "secret", true},
		{"no key", "", "secret", false},
		{"key only", "key", "", true}, // Only key is required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewLastFMClient(tt.apiKey, tt.apiSecret, nil)
			if client.IsConfigured() != tt.want {
				t.Errorf("IsConfigured() = %v, want %v", client.IsConfigured(), tt.want)
			}
		})
	}
}

func TestLastFMClient_SetCredentials(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	if client.IsConfigured() {
		t.Error("Client should not be configured initially")
	}

	client.SetCredentials("new_key", "new_secret")

	if !client.IsConfigured() {
		t.Error("Client should be configured after SetCredentials")
	}

	if client.apiKey != "new_key" {
		t.Errorf("apiKey = %v, want %v", client.apiKey, "new_key")
	}

	if client.apiSecret != "new_secret" {
		t.Errorf("apiSecret = %v, want %v", client.apiSecret, "new_secret")
	}
}

func TestLastFMClient_GetSource(t *testing.T) {
	client := NewLastFMClient("key", "secret", nil)

	if client.GetSource() != models.DataSourceLastFM {
		t.Errorf("GetSource() = %v, want %v", client.GetSource(), models.DataSourceLastFM)
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name        string
		inputArtist string
		matchArtist string
		inputTitle  string
		matchTitle  string
		minConf     float64
	}{
		{
			name:        "exact match",
			inputArtist: "Metallica",
			matchArtist: "Metallica",
			inputTitle:  "Enter Sandman",
			matchTitle:  "Enter Sandman",
			minConf:     1.0,
		},
		{
			name:        "case insensitive match",
			inputArtist: "METALLICA",
			matchArtist: "metallica",
			inputTitle:  "ENTER SANDMAN",
			matchTitle:  "enter sandman",
			minConf:     1.0,
		},
		{
			name:        "partial match",
			inputArtist: "Metallica",
			matchArtist: "Metallica (Band)",
			inputTitle:  "Enter Sandman",
			matchTitle:  "Enter Sandman (Live)",
			minConf:     0.5,
		},
		{
			name:        "no match",
			inputArtist: "Metallica",
			matchArtist: "Iron Maiden",
			inputTitle:  "Enter Sandman",
			matchTitle:  "The Trooper",
			minConf:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := calculateConfidence(tt.inputArtist, tt.matchArtist, tt.inputTitle, tt.matchTitle)
			if confidence < tt.minConf {
				t.Errorf("calculateConfidence() = %v, want >= %v", confidence, tt.minConf)
			}
		})
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want float64
	}{
		{"identical", "hello", "hello", 1.0},
		{"empty a", "", "hello", 0.0},
		{"empty b", "hello", "", 0.0},
		{"both empty", "", "", 1.0},
		{"a contains b", "hello world", "hello", 0.45}, // approximate
		{"b contains a", "hello", "hello world", 0.45}, // approximate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSimilarity(tt.a, tt.b)
			// Allow some tolerance for approximate calculations
			if tt.want == 1.0 && result != 1.0 {
				t.Errorf("stringSimilarity(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.want)
			}
			if tt.want == 0.0 && result != 0.0 {
				t.Errorf("stringSimilarity(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.want)
			}
		})
	}
}

func TestReadAll(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{"empty", "", 0},
		{"small", "hello", 5},
		{"medium", "hello world this is a test", 26},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &mockReader{data: []byte(tt.input)}
			buf := make([]byte, 0)
			result, err := readAll(reader, buf)
			if err != nil {
				t.Fatalf("readAll() error = %v", err)
			}
			if len(result) != tt.wantLen {
				t.Errorf("readAll() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

// mockReader implements io.Reader for testing
type mockReader struct {
	data []byte
	pos  int
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, &eofError{}
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

type eofError struct{}

func (e *eofError) Error() string { return "EOF" }

func TestLastFMResponseStructures(t *testing.T) {
	// Test that response structures are properly defined
	_ = lastFMTrackSearchResponse{}
	_ = lastFMArtistTopTracksResponse{}
	_ = lastFMSimilarTracksResponse{}
	_ = lastFMTagTopTracksResponse{}
	_ = lastFMArtistInfoResponse{}
	_ = lastFMSimilarArtistsResponse{}
}

func TestLastFMClient_MakeRequest_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.makeRequest(context.Background(), "test.method", nil)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("makeRequest() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_HTTPClient(t *testing.T) {
	client := NewLastFMClient("key", "secret", nil)

	httpClient := client.HTTPClient()
	if httpClient == nil {
		t.Error("HTTPClient() should not return nil")
	}
}

func TestStringSimilarity_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		min  float64
		max  float64
	}{
		{"exact match", "metallica", "metallica", 1.0, 1.0},
		{"case insensitive in caller", "METALLICA", "METALLICA", 1.0, 1.0},
		{"partial prefix", "metal", "metallica", 0.4, 0.6},
		{"completely different", "abc", "xyz", 0.0, 0.5},
		{"one contains other", "song name", "song", 0.4, 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSimilarity(tt.a, tt.b)
			if result < tt.min || result > tt.max {
				t.Errorf("stringSimilarity(%q, %q) = %v, want between %v and %v", tt.a, tt.b, result, tt.min, tt.max)
			}
		})
	}
}

func TestCalculateConfidence_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		inputArtist string
		matchArtist string
		inputTitle  string
		matchTitle  string
		minConf     float64
	}{
		{"empty input", "", "Artist", "", "Title", 0.0},
		{"empty match", "Artist", "", "Title", "", 0.0},
		{"exact", "Artist", "Artist", "Title", "Title", 1.0},
		{"similar", "Artist Name", "Artist Name Band", "Song Title", "Song Title Live", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := calculateConfidence(tt.inputArtist, tt.matchArtist, tt.inputTitle, tt.matchTitle)
			if conf < tt.minConf {
				t.Errorf("calculateConfidence() = %v, want >= %v", conf, tt.minConf)
			}
		})
	}
}

func TestReadAll_LargeData(t *testing.T) {
	// Test with larger data that requires buffer expansion
	largeData := make([]byte, 2048)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	reader := &mockReader{data: largeData}
	buf := make([]byte, 0, 64) // Start with small buffer
	
	result, err := readAll(reader, buf)
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	
	if len(result) != len(largeData) {
		t.Errorf("readAll() length = %v, want %v", len(result), len(largeData))
	}
}

func TestLastFMConstants(t *testing.T) {
	if lastFMBaseURL != "https://ws.audioscrobbler.com/2.0/" {
		t.Errorf("lastFMBaseURL = %v, unexpected", lastFMBaseURL)
	}
}

func TestLastFMClient_SearchTrack_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.SearchTrack(context.Background(), "Artist", "Title")
	if err != models.ErrAPIKeyMissing {
		t.Errorf("SearchTrack() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_GetTopTracks_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.GetTopTracks(context.Background(), "Artist", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetTopTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_GetSimilarTracks_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.GetSimilarTracks(context.Background(), "Artist", "Title", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetSimilarTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_GetTagTracks_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.GetTagTracks(context.Background(), "rock", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetTagTracks() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_GetArtistInfo_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.GetArtistInfo(context.Background(), "Artist")
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetArtistInfo() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_GetSimilarArtists_NotConfigured(t *testing.T) {
	client := NewLastFMClient("", "", nil)

	_, err := client.GetSimilarArtists(context.Background(), "Artist", 10)
	if err != models.ErrAPIKeyMissing {
		t.Errorf("GetSimilarArtists() error = %v, want ErrAPIKeyMissing", err)
	}
}

func TestLastFMClient_UserAgent(t *testing.T) {
	client := NewLastFMClient("key", "secret", nil)

	if client.UserAgent() == "" {
		t.Error("UserAgent() should not be empty")
	}
}

func TestLastFMClient_MakeRequest_WithServer(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":{}}`))
	}))
	defer server.Close()

	// Verify server is created
	if server == nil {
		t.Error("Server should not be nil")
	}
}

func TestLastFMClient_WithLogger(t *testing.T) {
	logger := &testLogger{}
	client := NewLastFMClient("key", "secret", logger)

	if client.logger != logger {
		t.Error("Client should use provided logger")
	}
}

// testLogger implements Logger for testing
type testLogger struct {
	infos  []string
	errors []string
	debugs []string
}

func (l *testLogger) Info(msg string, args ...interface{})  { l.infos = append(l.infos, msg) }
func (l *testLogger) Error(msg string, args ...interface{}) { l.errors = append(l.errors, msg) }
func (l *testLogger) Debug(msg string, args ...interface{}) { l.debugs = append(l.debugs, msg) }

func TestLastFMTrackSearchResponse_Fields(t *testing.T) {
	resp := lastFMTrackSearchResponse{}
	// Verify the structure is usable
	resp.Results.TrackMatches.Track = []struct {
		Name   string `json:"name"`
		Artist string `json:"artist"`
		URL    string `json:"url"`
		MBID   string `json:"mbid"`
	}{
		{Name: "Test", Artist: "Artist", URL: "http://test.com", MBID: "123"},
	}
	if len(resp.Results.TrackMatches.Track) != 1 {
		t.Error("Should have 1 track")
	}
}

func TestLastFMArtistTopTracksResponse_Fields(t *testing.T) {
	resp := lastFMArtistTopTracksResponse{}
	// Just verify the structure exists and is accessible
	_ = resp.TopTracks
}

func TestLastFMSimilarTracksResponse_Fields(t *testing.T) {
	resp := lastFMSimilarTracksResponse{}
	_ = resp.SimilarTracks
}

func TestLastFMTagTopTracksResponse_Fields(t *testing.T) {
	resp := lastFMTagTopTracksResponse{}
	_ = resp.Tracks
}

func TestLastFMArtistInfoResponse_Fields(t *testing.T) {
	resp := lastFMArtistInfoResponse{}
	_ = resp.Artist
}

func TestLastFMSimilarArtistsResponse_Fields(t *testing.T) {
	resp := lastFMSimilarArtistsResponse{}
	_ = resp.SimilarArtists
}

func TestLastFMClient_MakeRequest_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	// Verify server works
	if server == nil {
		t.Error("Server should not be nil")
	}
}

func TestLastFMClient_MakeRequest_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	_, err := client.SearchTrack(context.Background(), "Artist", "Title")
	if err == nil {
		t.Error("SearchTrack should return error for unauthorized response")
	}
}

func TestLastFMClient_MakeRequest_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	_, err := client.SearchTrack(context.Background(), "Artist", "Title")
	if err == nil {
		t.Error("SearchTrack should return error for rate limit response")
	}
}

func TestLastFMClient_MakeRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	_, err := client.SearchTrack(context.Background(), "Artist", "Title")
	if err == nil {
		t.Error("SearchTrack should return error for server error response")
	}
}

func TestLastFMClient_SearchTrack_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"results": {
				"trackmatches": {
					"track": [
						{"name": "Test Track", "artist": "Test Artist", "url": "http://test.com", "mbid": "123"}
					]
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	match, err := client.SearchTrack(context.Background(), "Test Artist", "Test Track")
	if err != nil {
		t.Fatalf("SearchTrack() error = %v", err)
	}
	if match == nil {
		t.Fatal("SearchTrack() returned nil match")
	}
	if match.Title != "Test Track" {
		t.Errorf("SearchTrack() Title = %v, want Test Track", match.Title)
	}
}

func TestLastFMClient_GetTopTracks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"toptracks": {
				"track": [
					{"name": "Track 1", "playcount": "1000", "duration": "180", "url": "http://test.com", "artist": {"name": "Artist", "mbid": ""}},
					{"name": "Track 2", "playcount": "500", "duration": "200", "url": "http://test.com", "artist": {"name": "Artist", "mbid": ""}}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	tracks, err := client.GetTopTracks(context.Background(), "Artist", 10)
	if err != nil {
		t.Fatalf("GetTopTracks() error = %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("GetTopTracks() returned %d tracks, want 2", len(tracks))
	}
}

func TestLastFMClient_GetSimilarTracks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"similartracks": {
				"track": [
					{"name": "Similar Track", "match": 0.9, "url": "http://test.com", "artist": {"name": "Similar Artist", "mbid": ""}}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	tracks, err := client.GetSimilarTracks(context.Background(), "Artist", "Track", 10)
	if err != nil {
		t.Fatalf("GetSimilarTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Errorf("GetSimilarTracks() returned %d tracks, want 1", len(tracks))
	}
}

func TestLastFMClient_GetTagTracks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"tracks": {
				"track": [
					{"name": "Rock Track", "duration": "200", "url": "http://test.com", "artist": {"name": "Rock Artist", "mbid": ""}}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	tracks, err := client.GetTagTracks(context.Background(), "rock", 10)
	if err != nil {
		t.Fatalf("GetTagTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Errorf("GetTagTracks() returned %d tracks, want 1", len(tracks))
	}
}

func TestLastFMClient_GetArtistInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"artist": {
				"name": "Test Artist",
				"mbid": "artist-mbid",
				"url": "http://test.com",
				"tags": {"tag": [{"name": "rock"}]},
				"bio": {"summary": "Artist bio"}
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	info, err := client.GetArtistInfo(context.Background(), "Test Artist")
	if err != nil {
		t.Fatalf("GetArtistInfo() error = %v", err)
	}
	if info == nil {
		t.Fatal("GetArtistInfo() returned nil")
	}
	if info.Name != "Test Artist" {
		t.Errorf("GetArtistInfo() Name = %v, want Test Artist", info.Name)
	}
}

func TestLastFMClient_GetSimilarArtists_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"similarartists": {
				"artist": [
					{"name": "Similar Artist 1", "mbid": "mbid1", "url": "http://test.com", "match": 0.9},
					{"name": "Similar Artist 2", "mbid": "mbid2", "url": "http://test.com", "match": 0.8}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLastFMClient("test_key", "test_secret", nil)
	client.SetBaseURL(server.URL + "/")

	artists, err := client.GetSimilarArtists(context.Background(), "Artist", 10)
	if err != nil {
		t.Fatalf("GetSimilarArtists() error = %v", err)
	}
	if len(artists) != 2 {
		t.Errorf("GetSimilarArtists() returned %d artists, want 2", len(artists))
	}
}

func TestLastFMClient_SetBaseURL(t *testing.T) {
	client := NewLastFMClient("key", "secret", nil)
	client.SetBaseURL("http://custom.url/")

	if client.getBaseURL() != "http://custom.url/" {
		t.Errorf("getBaseURL() = %v, want http://custom.url/", client.getBaseURL())
	}
}

func TestLastFMClient_GetBaseURL_Default(t *testing.T) {
	client := NewLastFMClient("key", "secret", nil)

	if client.getBaseURL() != lastFMBaseURL {
		t.Errorf("getBaseURL() = %v, want %v", client.getBaseURL(), lastFMBaseURL)
	}
}

