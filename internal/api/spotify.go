package api

import (
	"context"
	"fmt"
	
	"github.com/ardakilic/rocklist/internal/api/models"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// SpotifyClient provides access to the Spotify API
type SpotifyClient struct {
	client *spotify.Client
}

// NewSpotifyClientFunc is the function type for creating a new Spotify client
type NewSpotifyClientFunc func(clientID, clientSecret string) (*SpotifyClient, error)

// NewSpotifyClient creates a new Spotify client using client credentials flow
var NewSpotifyClient NewSpotifyClientFunc = func(clientID, clientSecret string) (*SpotifyClient, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("Spotify client ID or secret not provided")
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Spotify access token: %w", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	return &SpotifyClient{client: client}, nil
}

// GetTopTracks fetches the top tracks for the given artist from Spotify
func (c *SpotifyClient) GetTopTracks(artist string, limit int) ([]models.TopTrack, error) {
	// First, search for the artist
	ctx := context.Background()
	searchResult, err := c.client.Search(ctx, artist, spotify.SearchTypeArtist)
	if err != nil {
		return nil, fmt.Errorf("failed to search for artist %s: %w", artist, err)
	}

	if searchResult.Artists == nil || len(searchResult.Artists.Artists) == 0 {
		return nil, fmt.Errorf("artist %s not found on Spotify", artist)
	}

	spotifyArtist := searchResult.Artists.Artists[0]
	
	// Get the artist's top tracks (limited to market US for consistent results)
	topTracks, err := c.client.GetArtistsTopTracks(ctx, spotifyArtist.ID, "US")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top tracks for %s: %w", artist, err)
	}

	// Limit the number of tracks if needed
	if len(topTracks) > limit {
		topTracks = topTracks[:limit]
	}

	// Convert to our TopTrack format
	tracks := make([]models.TopTrack, 0, len(topTracks))
	for i, track := range topTracks {
		albumName := ""
		if track.Album.Name != "" {
			albumName = track.Album.Name
		}

		tracks = append(tracks, models.TopTrack{
			Name:   track.Name,
			Artist: spotifyArtist.Name,
			Album:  albumName,
			Rank:   i + 1,
			Found:  false, // Will be set later when matched with local files
		})
	}

	return tracks, nil
}

// APIClient is an interface for API clients that can fetch top tracks
type APIClient interface {
	GetTopTracks(artist string, limit int) ([]models.TopTrack, error)
}

// NewAPIClient returns an appropriate API client based on the source
func NewAPIClient(source, lastFMAPIKey, lastFMAPISecret, spotifyClientID, spotifyClientSecret string) (APIClient, error) {
	switch source {
	case "lastfm":
		return NewLastFMClient(lastFMAPIKey, lastFMAPISecret)
	case "spotify":
		return NewSpotifyClient(spotifyClientID, spotifyClientSecret)
	default:
		return nil, fmt.Errorf("unsupported API source: %s", source)
	}
} 