package api

import (
	"fmt"
	"github.com/ardakilic/rocklist/internal/api/models"
	"github.com/sonjek/go-lastfm/lastfm"
)

// TopTrack represents a top track from an API response
type TopTrack struct {
	Name     string
	Artist   string
	Album    string
	Rank     int
	Filename string
	Found    bool // indicates if the track was found in the Rockbox database
}

// LastFMClient provides access to the Last.fm API
type LastFMClient struct {
	api *lastfm.Api
}

// NewLastFMClientFunc is the function type for creating a new LastFM client
type NewLastFMClientFunc func(apiKey, apiSecret string) (*LastFMClient, error)

// NewLastFMClient creates a new Last.fm client
var NewLastFMClient NewLastFMClientFunc = func(apiKey, apiSecret string) (*LastFMClient, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("Last.fm API key or secret not provided")
	}

	api := lastfm.New(apiKey, apiSecret)
	return &LastFMClient{api: api}, nil
}

// GetTopTracks fetches the top tracks for the given artist from Last.fm
func (c *LastFMClient) GetTopTracks(artist string, limit int) ([]models.TopTrack, error) {
	params := lastfm.P{
		"artist": artist,
		"limit":  limit,
	}

	result, err := c.api.Artist.GetTopTracks(params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top tracks for %s: %w", artist, err)
	}

	tracks := make([]models.TopTrack, 0, len(result.Tracks))
	for i, track := range result.Tracks {
		// Use the original artist name as we don't have artist info in the track
		tracks = append(tracks, models.TopTrack{
			Name:   track.Name,
			Artist: artist, // Use the original artist name
			Rank:   i + 1,
			Found:  false, // Will be set later when matched with local files
		})
	}

	return tracks, nil
} 