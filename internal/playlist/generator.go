package playlist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ardakilic/rocklist/internal/api"
	"github.com/ardakilic/rocklist/internal/api/models"
	"github.com/ardakilic/rocklist/internal/database"
)

// Generator handles playlist generation
type Generator struct {
	db          *database.RockboxDB
	apiClient   api.APIClient
	playlistDir string
	maxTracks   int
}

// NewGenerator creates a new playlist generator
func NewGenerator(db *database.RockboxDB, apiClient api.APIClient, playlistDir string, maxTracks int) *Generator {
	return &Generator{
		db:          db,
		apiClient:   apiClient,
		playlistDir: playlistDir,
		maxTracks:   maxTracks,
	}
}

// GeneratePlaylistForArtist generates a playlist for the given artist
func (g *Generator) GeneratePlaylistForArtist(artist string) error {
	// Get top tracks from API
	topTracks, err := g.apiClient.GetTopTracks(artist, g.maxTracks)
	if err != nil {
		return fmt.Errorf("failed to get top tracks for %s: %w", artist, err)
	}

	// Get artist's tracks from the database
	dbTracks := g.db.GetTracksForArtist(artist)
	if len(dbTracks) == 0 {
		return fmt.Errorf("no tracks found for artist %s in Rockbox database", artist)
	}

	// Match API tracks with database tracks
	matchedTracks := g.matchTracks(topTracks, dbTracks)
	
	// If no matches, return error
	if len(matchedTracks) == 0 {
		return fmt.Errorf("no matching tracks found for artist %s", artist)
	}
	
	// Generate playlist file
	if err := g.writePlaylistFile(artist, matchedTracks); err != nil {
		return fmt.Errorf("failed to write playlist file for %s: %w", artist, err)
	}
	
	return nil
}

// matchTracks matches API tracks with database tracks
func (g *Generator) matchTracks(topTracks []models.TopTrack, dbTracks []database.Track) []models.TopTrack {
	matched := make([]models.TopTrack, 0)
	
	// Create a map for faster lookup
	trackMap := make(map[string]database.Track)
	for _, track := range dbTracks {
		// Use lowercase title for case-insensitive matching
		key := strings.ToLower(track.Title)
		trackMap[key] = track
	}
	
	// Match tracks by title
	for _, apiTrack := range topTracks {
		key := strings.ToLower(apiTrack.Name)
		if dbTrack, ok := trackMap[key]; ok {
			apiTrack.Filename = dbTrack.Filename
			apiTrack.Found = true
			matched = append(matched, apiTrack)
		}
	}
	
	// If we didn't find matches using exact matching, try fuzzy matching
	if len(matched) == 0 {
		for _, apiTrack := range topTracks {
			normalizedAPITitle := normalizeTitle(apiTrack.Name)
			
			for _, dbTrack := range dbTracks {
				normalizedDBTitle := normalizeTitle(dbTrack.Title)
				
				// Simple fuzzy match: title starts with the same words
				if strings.HasPrefix(normalizedAPITitle, normalizedDBTitle) || 
				   strings.HasPrefix(normalizedDBTitle, normalizedAPITitle) {
					apiTrack.Filename = dbTrack.Filename
					apiTrack.Found = true
					matched = append(matched, apiTrack)
					break
				}
			}
		}
	}
	
	return matched
}

// normalizeTitle removes special characters and normalizes the title for better matching
func normalizeTitle(title string) string {
	title = strings.ToLower(title)
	
	// Remove special characters
	title = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return -1
	}, title)
	
	// Remove extra spaces
	title = strings.Join(strings.Fields(title), " ")
	
	return title
}

// writePlaylistFile writes the playlist file for the given artist
func (g *Generator) writePlaylistFile(artist string, tracks []models.TopTrack) error {
	// Create playlist directory if it doesn't exist
	if err := os.MkdirAll(g.playlistDir, 0755); err != nil {
		return fmt.Errorf("failed to create playlist directory: %w", err)
	}
	
	// Create a valid filename from the artist name
	safeArtistName := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return '_'
	}, artist)
	
	// Create playlist file
	filename := filepath.Join(g.playlistDir, fmt.Sprintf("%s-top-%d.m3u", safeArtistName, len(tracks)))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create playlist file: %w", err)
	}
	defer file.Close()
	
	// Write playlist header
	if _, err := file.WriteString("#EXTM3U\n"); err != nil {
		return fmt.Errorf("failed to write playlist header: %w", err)
	}
	
	// Write tracks
	for _, track := range tracks {
		if !track.Found {
			continue
		}
		
		// Write track info
		info := fmt.Sprintf("#EXTINF:-1,%s - %s\n", track.Artist, track.Name)
		if _, err := file.WriteString(info); err != nil {
			return fmt.Errorf("failed to write track info: %w", err)
		}
		
		// Write track path
		if _, err := file.WriteString(track.Filename + "\n"); err != nil {
			return fmt.Errorf("failed to write track path: %w", err)
		}
	}
	
	return nil
} 