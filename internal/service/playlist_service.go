// Package service provides business logic services
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ardakilic/rocklist/internal/api"
	"github.com/Ardakilic/rocklist/internal/models"
	"github.com/Ardakilic/rocklist/internal/repository"
)

// PlaylistService handles playlist generation and management
type PlaylistService struct {
	songRepo     repository.SongRepository
	playlistRepo repository.PlaylistRepository
	clients      map[models.DataSource]api.Client
	playlistDir  string
	logger       Logger
}

// Logger interface for services
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// NewPlaylistService creates a new playlist service
func NewPlaylistService(
	songRepo repository.SongRepository,
	playlistRepo repository.PlaylistRepository,
	playlistDir string,
	logger Logger,
) *PlaylistService {
	return &PlaylistService{
		songRepo:     songRepo,
		playlistRepo: playlistRepo,
		clients:      make(map[models.DataSource]api.Client),
		playlistDir:  playlistDir,
		logger:       logger,
	}
}

// RegisterClient registers an API client for a data source
func (s *PlaylistService) RegisterClient(source models.DataSource, client api.Client) {
	s.clients[source] = client
}

// SetPlaylistDir sets the playlist directory
func (s *PlaylistService) SetPlaylistDir(dir string) {
	s.playlistDir = dir
}

// GeneratePlaylist generates a playlist based on the request
func (s *PlaylistService) GeneratePlaylist(ctx context.Context, req *models.PlaylistRequest) (*models.Playlist, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	client, ok := s.clients[req.DataSource]
	if !ok || !client.IsConfigured() {
		return nil, models.ErrDataSourceDisabled
	}

	s.logger.Info("Generating %s playlist from %s", req.Type.DisplayName(), req.DataSource.DisplayName())

	var externalTracks []*api.TrackInfo
	var err error
	var playlistName string

	switch req.Type {
	case models.PlaylistTypeTopSongs:
		if req.Artist == "" {
			return nil, fmt.Errorf("artist is required for top songs playlist")
		}
		externalTracks, err = client.GetTopTracks(ctx, req.Artist, req.Limit)
		playlistName = fmt.Sprintf("Top Songs - %s (%s)", req.Artist, req.DataSource.DisplayName())

	case models.PlaylistTypeMixedSongs:
		if req.Artist == "" {
			return nil, fmt.Errorf("artist is required for mixed songs playlist")
		}
		externalTracks, err = s.getMixedSongs(ctx, client, req.Artist, req.Limit)
		playlistName = fmt.Sprintf("Mixed Songs - %s (%s)", req.Artist, req.DataSource.DisplayName())

	case models.PlaylistTypeSimilar:
		if req.Artist == "" {
			return nil, fmt.Errorf("artist is required for similar songs playlist")
		}
		externalTracks, err = s.getSimilarArtistTracks(ctx, client, req.Artist, req.Limit)
		playlistName = fmt.Sprintf("Similar to %s (%s)", req.Artist, req.DataSource.DisplayName())

	case models.PlaylistTypeTag:
		if req.Tag == "" {
			return nil, models.ErrTagRequired
		}
		externalTracks, err = client.GetTagTracks(ctx, req.Tag, req.Limit)
		playlistName = fmt.Sprintf("%s Radio (%s)", req.Tag, req.DataSource.DisplayName())

	default:
		return nil, models.ErrInvalidPlaylistType
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get tracks from %s: %w", req.DataSource.DisplayName(), err)
	}

	if len(externalTracks) == 0 {
		return nil, models.ErrNoMatchingSongs
	}

	s.logger.Info("Found %d tracks from %s, matching with local library...", len(externalTracks), req.DataSource.DisplayName())

	// Match external tracks to local songs
	matchedSongs, matchStats := s.matchTracks(ctx, externalTracks)

	s.logger.Info("Matched %d/%d tracks (%.1f%% match rate)",
		matchStats.Matched, matchStats.Total, matchStats.MatchRate()*100)

	if len(matchedSongs) == 0 {
		return nil, models.ErrNoMatchingSongs
	}

	// Create playlist
	now := time.Now()
	playlist := &models.Playlist{
		Name:        playlistName,
		Description: fmt.Sprintf("Generated from %s on %s", req.DataSource.DisplayName(), now.Format("2006-01-02 15:04")),
		Type:        req.Type,
		DataSource:  req.DataSource,
		Artist:      req.Artist,
		Tag:         req.Tag,
		SongCount:   len(matchedSongs),
		GeneratedAt: now,
	}

	if err := s.playlistRepo.Create(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	// Add songs to playlist
	songIDs := make([]uint, len(matchedSongs))
	for i, song := range matchedSongs {
		songIDs[i] = song.ID
	}
	if err := s.playlistRepo.AddSongs(ctx, playlist.ID, songIDs); err != nil {
		return nil, fmt.Errorf("failed to add songs to playlist: %w", err)
	}

	return playlist, nil
}

// ExportPlaylist exports a playlist to an M3U file
func (s *PlaylistService) ExportPlaylist(ctx context.Context, playlistID uint) (string, error) {
	playlist, err := s.playlistRepo.FindByID(ctx, playlistID)
	if err != nil {
		return "", err
	}

	songs, err := s.playlistRepo.GetSongs(ctx, playlistID)
	if err != nil {
		return "", err
	}

	if len(songs) == 0 {
		return "", models.ErrNoMatchingSongs
	}

	// Create playlist directory if needed
	if err := os.MkdirAll(s.playlistDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create playlist directory: %w", err)
	}

	// Generate filename
	safeName := sanitizeFilename(playlist.Name)
	filename := filepath.Join(s.playlistDir, safeName+".m3u8")

	// Create M3U file
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create playlist file: %w", err)
	}
	defer file.Close()

	// Write M3U header
	file.WriteString("#EXTM3U\n")
	file.WriteString(fmt.Sprintf("#PLAYLIST:%s\n", playlist.Name))
	file.WriteString("\n")

	// Write tracks
	for _, song := range songs {
		// Write extended info
		duration := song.Duration
		if duration == 0 {
			duration = -1
		}
		displayName := song.GetDisplayName()
		file.WriteString(fmt.Sprintf("#EXTINF:%d,%s\n", duration, displayName))
		file.WriteString(song.Path + "\n")
	}

	// Update playlist with export info
	now := time.Now()
	playlist.FilePath = filename
	playlist.ExportedAt = &now
	if err := s.playlistRepo.Update(ctx, playlist); err != nil {
		s.logger.Error("Failed to update playlist export info: %v", err)
	}

	s.logger.Info("Exported playlist to: %s", filename)
	return filename, nil
}

// getMixedSongs gets a mix of top tracks and similar tracks for an artist
func (s *PlaylistService) getMixedSongs(ctx context.Context, client api.Client, artist string, limit int) ([]*api.TrackInfo, error) {
	topLimit := limit / 2
	similarLimit := limit - topLimit

	// Get top tracks
	topTracks, err := client.GetTopTracks(ctx, artist, topLimit)
	if err != nil {
		return nil, err
	}

	// Try to get similar tracks from the artist's top songs
	var similarTracks []*api.TrackInfo
	if len(topTracks) > 0 {
		similar, err := client.GetSimilarTracks(ctx, artist, topTracks[0].Title, similarLimit)
		if err == nil {
			similarTracks = similar
		}
	}

	// Combine tracks
	result := make([]*api.TrackInfo, 0, len(topTracks)+len(similarTracks))
	result = append(result, topTracks...)
	result = append(result, similarTracks...)

	return result, nil
}

// getSimilarArtistTracks gets tracks from similar artists
func (s *PlaylistService) getSimilarArtistTracks(ctx context.Context, client api.Client, artist string, limit int) ([]*api.TrackInfo, error) {
	similarArtists, err := client.GetSimilarArtists(ctx, artist, 5)
	if err != nil {
		return nil, err
	}

	tracksPerArtist := limit / len(similarArtists)
	if tracksPerArtist < 1 {
		tracksPerArtist = 1
	}

	var result []*api.TrackInfo
	for _, similarArtist := range similarArtists {
		tracks, err := client.GetTopTracks(ctx, similarArtist.Name, tracksPerArtist)
		if err != nil {
			s.logger.Debug("Failed to get tracks for similar artist %s: %v", similarArtist.Name, err)
			continue
		}
		result = append(result, tracks...)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// MatchStats holds matching statistics
type MatchStats struct {
	Total     int
	Matched   int
	Unmatched int
}

// MatchRate returns the match rate as a percentage (0-1)
func (ms *MatchStats) MatchRate() float64 {
	if ms.Total == 0 {
		return 0
	}
	return float64(ms.Matched) / float64(ms.Total)
}

// matchTracks matches external tracks to local songs
func (s *PlaylistService) matchTracks(ctx context.Context, tracks []*api.TrackInfo) ([]*models.Song, *MatchStats) {
	stats := &MatchStats{Total: len(tracks)}
	matched := make([]*models.Song, 0, len(tracks))
	seen := make(map[uint]bool) // Avoid duplicates

	for _, track := range tracks {
		// Try to find matching song in local library
		songs, err := s.songRepo.FindByArtist(ctx, track.Artist)
		if err != nil || len(songs) == 0 {
			s.logger.Debug("No songs found for artist: %s", track.Artist)
			stats.Unmatched++
			continue
		}

		// Find best match
		var bestMatch *models.Song
		bestScore := 0.0
		for _, song := range songs {
			score := calculateMatchScore(track, song)
			if score > bestScore && score >= 0.5 {
				bestScore = score
				bestMatch = song
			}
		}

		if bestMatch != nil && !seen[bestMatch.ID] {
			seen[bestMatch.ID] = true
			matched = append(matched, bestMatch)
			stats.Matched++
			s.logger.Debug("Matched: %s - %s (score: %.2f)", track.Artist, track.Title, bestScore)
		} else {
			stats.Unmatched++
			s.logger.Debug("No match for: %s - %s", track.Artist, track.Title)
		}
	}

	return matched, stats
}

// calculateMatchScore calculates a match score between an external track and a local song
func calculateMatchScore(track *api.TrackInfo, song *models.Song) float64 {
	titleScore := stringSimilarity(
		strings.ToLower(track.Title),
		strings.ToLower(song.Title),
	)
	artistScore := stringSimilarity(
		strings.ToLower(track.Artist),
		strings.ToLower(song.GetEffectiveArtist()),
	)

	// Title is more important than artist
	return titleScore*0.6 + artistScore*0.4
}

// stringSimilarity calculates string similarity (0-1)
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Normalize strings
	a = normalizeString(a)
	b = normalizeString(b)

	if a == b {
		return 1.0
	}

	// Check containment
	if strings.Contains(a, b) || strings.Contains(b, a) {
		shorter := len(a)
		if len(b) < shorter {
			shorter = len(b)
		}
		longer := len(a)
		if len(b) > longer {
			longer = len(b)
		}
		return float64(shorter) / float64(longer)
	}

	// Calculate Levenshtein distance ratio
	distance := levenshteinDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	return 1.0 - float64(distance)/float64(maxLen)
}

// normalizeString normalizes a string for comparison
func normalizeString(s string) string {
	s = strings.ToLower(s)
	// Remove common suffixes/prefixes
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "(remastered)")
	s = strings.TrimSuffix(s, "(remaster)")
	s = strings.TrimSuffix(s, "[remastered]")
	s = strings.TrimSuffix(s, " - remastered")
	s = strings.TrimSuffix(s, " (live)")
	s = strings.TrimSuffix(s, " [live]")
	return strings.TrimSpace(s)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = minOf3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func minOf3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// sanitizeFilename removes invalid characters from a filename
func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, c := range invalid {
		result = strings.ReplaceAll(result, c, "_")
	}
	// Limit length
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}
