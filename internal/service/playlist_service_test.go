package service

import (
	"context"
	"testing"

	"github.com/Ardakilic/rocklist/internal/api"
	"github.com/Ardakilic/rocklist/internal/models"
)

// mockSongRepository implements repository.SongRepository for testing
type mockSongRepository struct {
	songs     []*models.Song
	findError error
}

func (m *mockSongRepository) Create(ctx context.Context, song *models.Song) error { return nil }
func (m *mockSongRepository) CreateBatch(ctx context.Context, songs []*models.Song) error {
	return nil
}
func (m *mockSongRepository) Update(ctx context.Context, song *models.Song) error { return nil }
func (m *mockSongRepository) Delete(ctx context.Context, id uint) error           { return nil }
func (m *mockSongRepository) FindByID(ctx context.Context, id uint) (*models.Song, error) {
	return nil, m.findError
}
func (m *mockSongRepository) FindByRockboxID(ctx context.Context, rockboxID string) (*models.Song, error) {
	return nil, m.findError
}
func (m *mockSongRepository) FindByPath(ctx context.Context, path string) (*models.Song, error) {
	return nil, m.findError
}
func (m *mockSongRepository) FindByArtist(ctx context.Context, artist string) ([]*models.Song, error) {
	return m.songs, m.findError
}
func (m *mockSongRepository) FindByAlbumArtist(ctx context.Context, albumArtist string) ([]*models.Song, error) {
	return m.songs, m.findError
}
func (m *mockSongRepository) FindByGenre(ctx context.Context, genre string) ([]*models.Song, error) {
	return m.songs, m.findError
}
func (m *mockSongRepository) FindUnmatched(ctx context.Context, source models.DataSource) ([]*models.Song, error) {
	return m.songs, m.findError
}
func (m *mockSongRepository) FindAll(ctx context.Context) ([]*models.Song, error) {
	return m.songs, m.findError
}
func (m *mockSongRepository) GetUniqueArtists(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
func (m *mockSongRepository) GetUniqueGenres(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
func (m *mockSongRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.songs)), nil
}
func (m *mockSongRepository) DeleteAll(ctx context.Context) error { return nil }

// mockPlaylistRepository implements repository.PlaylistRepository for testing
type mockPlaylistRepository struct {
	playlists []*models.Playlist
}

func (m *mockPlaylistRepository) Create(ctx context.Context, playlist *models.Playlist) error {
	playlist.ID = 1
	m.playlists = append(m.playlists, playlist)
	return nil
}
func (m *mockPlaylistRepository) Update(ctx context.Context, playlist *models.Playlist) error {
	return nil
}
func (m *mockPlaylistRepository) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockPlaylistRepository) FindByID(ctx context.Context, id uint) (*models.Playlist, error) {
	return nil, nil
}
func (m *mockPlaylistRepository) FindAll(ctx context.Context) ([]*models.Playlist, error) {
	return m.playlists, nil
}
func (m *mockPlaylistRepository) FindByType(ctx context.Context, playlistType models.PlaylistType) ([]*models.Playlist, error) {
	return m.playlists, nil
}
func (m *mockPlaylistRepository) FindByDataSource(ctx context.Context, source models.DataSource) ([]*models.Playlist, error) {
	return m.playlists, nil
}
func (m *mockPlaylistRepository) AddSongs(ctx context.Context, playlistID uint, songIDs []uint) error {
	return nil
}
func (m *mockPlaylistRepository) RemoveSongs(ctx context.Context, playlistID uint, songIDs []uint) error {
	return nil
}
func (m *mockPlaylistRepository) GetSongs(ctx context.Context, playlistID uint) ([]*models.Song, error) {
	return nil, nil
}

// mockLogger implements Logger for testing
type mockServiceLogger struct{}

func (l *mockServiceLogger) Info(msg string, args ...interface{})  {}
func (l *mockServiceLogger) Error(msg string, args ...interface{}) {}
func (l *mockServiceLogger) Debug(msg string, args ...interface{}) {}

// mockAPIClient implements api.Client for testing
type mockAPIClient struct {
	source     models.DataSource
	configured bool
	topTracks  []*api.TrackInfo
	trackMatch *api.TrackMatch
	artistInfo *api.ArtistInfo
}

func (m *mockAPIClient) GetSource() models.DataSource { return m.source }
func (m *mockAPIClient) IsConfigured() bool           { return m.configured }
func (m *mockAPIClient) SearchTrack(ctx context.Context, artist, title string) (*api.TrackMatch, error) {
	return m.trackMatch, nil
}
func (m *mockAPIClient) GetTopTracks(ctx context.Context, artist string, limit int) ([]*api.TrackInfo, error) {
	return m.topTracks, nil
}
func (m *mockAPIClient) GetSimilarTracks(ctx context.Context, artist, title string, limit int) ([]*api.TrackInfo, error) {
	return m.topTracks, nil
}
func (m *mockAPIClient) GetTagTracks(ctx context.Context, tag string, limit int) ([]*api.TrackInfo, error) {
	return m.topTracks, nil
}
func (m *mockAPIClient) GetArtistInfo(ctx context.Context, artist string) (*api.ArtistInfo, error) {
	return m.artistInfo, nil
}
func (m *mockAPIClient) GetSimilarArtists(ctx context.Context, artist string, limit int) ([]*api.ArtistInfo, error) {
	return []*api.ArtistInfo{}, nil
}

func TestNewPlaylistService(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	if svc == nil {
		t.Fatal("NewPlaylistService() returned nil")
	}

	if svc.playlistDir != "/playlists" {
		t.Errorf("playlistDir = %v, want /playlists", svc.playlistDir)
	}
}

func TestPlaylistService_RegisterClient(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	client := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	svc.RegisterClient(models.DataSourceLastFM, client)

	if svc.clients[models.DataSourceLastFM] != client {
		t.Error("RegisterClient() should register the client")
	}
}

func TestPlaylistService_SetPlaylistDir(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/old", logger)
	svc.SetPlaylistDir("/new")

	if svc.playlistDir != "/new" {
		t.Errorf("SetPlaylistDir() playlistDir = %v, want /new", svc.playlistDir)
	}
}

func TestPlaylistService_GeneratePlaylist_NoClient(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeTopSongs,
		Artist:     "Test Artist",
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err != models.ErrDataSourceDisabled {
		t.Errorf("GeneratePlaylist() error = %v, want ErrDataSourceDisabled", err)
	}
}

func TestPlaylistService_GeneratePlaylist_NotConfigured(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	// Register unconfigured client
	client := &mockAPIClient{source: models.DataSourceLastFM, configured: false}
	svc.RegisterClient(models.DataSourceLastFM, client)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeTopSongs,
		Artist:     "Test Artist",
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err != models.ErrDataSourceDisabled {
		t.Errorf("GeneratePlaylist() error = %v, want ErrDataSourceDisabled", err)
	}
}

func TestPlaylistService_GeneratePlaylist_InvalidRequest(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	// Invalid request (empty type)
	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err == nil {
		t.Error("GeneratePlaylist() should return error for invalid request")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Normal Name", "Normal Name"},
		{"Name/With/Slashes", "Name_With_Slashes"},
		{"Name\\With\\Backslash", "Name_With_Backslash"},
		{"Name:With:Colons", "Name_With_Colons"},
		{"Name*With*Stars", "Name_With_Stars"},
		{"Name?With?Questions", "Name_With_Questions"},
		{"Name\"With\"Quotes", "Name_With_Quotes"},
		{"Name<With>Brackets", "Name_With_Brackets"},
		{"Name|With|Pipes", "Name_With_Pipes"},
		{"Multiple///Slashes", "Multiple___Slashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		min  float64
		max  float64
	}{
		{"identical", "hello", "hello", 1.0, 1.0},
		{"empty a", "", "hello", 0.0, 0.1},
		{"empty b", "hello", "", 0.0, 0.1},
		{"both empty", "", "", 0.9, 1.0},
		{"case insensitive", "HELLO", "hello", 1.0, 1.0},
		{"a contains b", "hello world", "hello", 0.4, 0.6},
		{"remastered suffix", "song (remastered)", "song", 0.9, 1.0},
		{"completely different", "abc", "xyz", 0.0, 0.5},
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

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello world"},
		{"Song (Remastered)", "song"},
		{"Song (Remaster)", "song"},
		{"Song [Remastered]", "song"},
		{"Song - Remastered", "song"},
		{"Song (Live)", "song"},
		{"Song [Live]", "song"},
		{"  spaces  ", "spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeString(tt.input)
			if result != tt.want {
				t.Errorf("normalizeString(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a    string
		b    string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "abd", 1},
		{"abc", "xyz", 3},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.want)
			}
		})
	}
}

func TestMinOf3(t *testing.T) {
	tests := []struct {
		a, b, c, want int
	}{
		{1, 2, 3, 1},
		{2, 1, 3, 1},
		{3, 2, 1, 1},
		{1, 1, 1, 1},
		{5, 3, 4, 3},
		{5, 4, 3, 3},
	}

	for _, tt := range tests {
		result := minOf3(tt.a, tt.b, tt.c)
		if result != tt.want {
			t.Errorf("minOf3(%d, %d, %d) = %d, want %d", tt.a, tt.b, tt.c, result, tt.want)
		}
	}
}

func TestSanitizeFilename_Long(t *testing.T) {
	// Test filename truncation (max 200 chars)
	longName := ""
	for i := 0; i < 250; i++ {
		longName += "a"
	}

	result := sanitizeFilename(longName)
	if len(result) > 200 {
		t.Errorf("sanitizeFilename() should truncate to 200 chars, got %d", len(result))
	}
}

func TestPlaylistService_MultipleClients(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	// Register multiple clients
	client1 := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	client2 := &mockAPIClient{source: models.DataSourceSpotify, configured: true}
	client3 := &mockAPIClient{source: models.DataSourceMusicBrainz, configured: true}

	svc.RegisterClient(models.DataSourceLastFM, client1)
	svc.RegisterClient(models.DataSourceSpotify, client2)
	svc.RegisterClient(models.DataSourceMusicBrainz, client3)

	if len(svc.clients) != 3 {
		t.Errorf("Should have 3 clients, got %d", len(svc.clients))
	}
}

func TestCalculateMatchScore(t *testing.T) {
	tests := []struct {
		name     string
		track    *api.TrackInfo
		song     *models.Song
		minScore float64
	}{
		{
			"exact match",
			&api.TrackInfo{Artist: "Metallica", Title: "Enter Sandman"},
			&models.Song{Artist: "Metallica", Title: "Enter Sandman"},
			0.9,
		},
		{
			"case insensitive",
			&api.TrackInfo{Artist: "METALLICA", Title: "ENTER SANDMAN"},
			&models.Song{Artist: "metallica", Title: "enter sandman"},
			0.9,
		},
		{
			"different",
			&api.TrackInfo{Artist: "Artist A", Title: "Song X"},
			&models.Song{Artist: "Artist B", Title: "Song Y"},
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMatchScore(tt.track, tt.song)
			if score < tt.minScore {
				t.Errorf("calculateMatchScore() = %v, want >= %v", score, tt.minScore)
			}
		})
	}
}

func TestPlaylistService_GeneratePlaylist_TagRequired(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	client := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	svc.RegisterClient(models.DataSourceLastFM, client)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeTag,
		Tag:        "", // Empty tag
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err != models.ErrTagRequired {
		t.Errorf("GeneratePlaylist() error = %v, want ErrTagRequired", err)
	}
}

func TestPlaylistService_GeneratePlaylist_TopSongs_NoArtist(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	client := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	svc.RegisterClient(models.DataSourceLastFM, client)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeTopSongs,
		Artist:     "", // Empty artist
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err == nil {
		t.Error("GeneratePlaylist() should return error when artist is empty for top_songs")
	}
}

func TestPlaylistService_GeneratePlaylist_MixedSongs_NoArtist(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	client := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	svc.RegisterClient(models.DataSourceLastFM, client)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeMixedSongs,
		Artist:     "", // Empty artist
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err == nil {
		t.Error("GeneratePlaylist() should return error when artist is empty for mixed_songs")
	}
}

func TestPlaylistService_GeneratePlaylist_Similar_NoArtist(t *testing.T) {
	songRepo := &mockSongRepository{}
	playlistRepo := &mockPlaylistRepository{}
	logger := &mockServiceLogger{}

	svc := NewPlaylistService(songRepo, playlistRepo, "/playlists", logger)

	client := &mockAPIClient{source: models.DataSourceLastFM, configured: true}
	svc.RegisterClient(models.DataSourceLastFM, client)

	req := &models.PlaylistRequest{
		DataSource: models.DataSourceLastFM,
		Type:       models.PlaylistTypeSimilar,
		Artist:     "", // Empty artist
		Limit:      10,
	}

	_, err := svc.GeneratePlaylist(context.Background(), req)
	if err == nil {
		t.Error("GeneratePlaylist() should return error when artist is empty for similar")
	}
}

func TestMatchStats_MatchRate(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		matched  int
		expected float64
	}{
		{"all matched", 10, 10, 1.0},
		{"half matched", 10, 5, 0.5},
		{"none matched", 10, 0, 0.0},
		{"zero total", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &MatchStats{
				Total:     tt.total,
				Matched:   tt.matched,
				Unmatched: tt.total - tt.matched,
			}
			rate := stats.MatchRate()
			if rate != tt.expected {
				t.Errorf("MatchStats.MatchRate() = %v, want %v", rate, tt.expected)
			}
		})
	}
}
