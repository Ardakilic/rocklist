// Package api provides clients for external music APIs
package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
)

const (
	spotifyAuthURL = "https://accounts.spotify.com/api/token"
	spotifyAPIURL  = "https://api.spotify.com/v1"
)

// SpotifyClient is a client for the Spotify API
type SpotifyClient struct {
	*BaseClient
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	mu           sync.RWMutex
	logger       Logger
}

// NewSpotifyClient creates a new Spotify API client
func NewSpotifyClient(clientID, clientSecret string, logger Logger) *SpotifyClient {
	return &SpotifyClient{
		BaseClient:   NewBaseClient(models.DataSourceSpotify, "Rocklist/1.0"),
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
	}
}

// IsConfigured returns true if the client is properly configured
func (c *SpotifyClient) IsConfigured() bool {
	return c.clientID != "" && c.clientSecret != ""
}

// SetCredentials sets the API credentials
func (c *SpotifyClient) SetCredentials(clientID, clientSecret string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clientID = clientID
	c.clientSecret = clientSecret
	c.accessToken = "" // Reset token
}

// getAccessToken gets or refreshes the access token
func (c *SpotifyClient) getAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	if !c.IsConfigured() {
		return "", models.ErrAPIKeyMissing
	}

	// Request new token
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", spotifyAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent())

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return "", models.NewAPIError(models.DataSourceSpotify, 0, "auth request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", models.NewAPIError(models.DataSourceSpotify, resp.StatusCode, "auth failed", models.ErrAPIUnauthorized)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if err := json.Unmarshal(buf.Bytes(), &tokenResp); err != nil {
		return "", err
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return c.accessToken, nil
}

// makeRequest makes a request to the Spotify API
func (c *SpotifyClient) makeRequest(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(spotifyAPIURL + endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	if c.logger != nil {
		c.logger.Debug("Spotify API request: %s", u.String())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", c.UserAgent())

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, models.NewAPIError(models.DataSourceSpotify, 0, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, models.NewAPIError(models.DataSourceSpotify, 429, "rate limited", models.ErrAPIRateLimited)
	}
	if resp.StatusCode == 401 {
		// Token might be expired, reset and retry
		c.mu.Lock()
		c.accessToken = ""
		c.mu.Unlock()
		return nil, models.NewAPIError(models.DataSourceSpotify, 401, "unauthorized", models.ErrAPIUnauthorized)
	}
	if resp.StatusCode != 200 {
		return nil, models.NewAPIError(models.DataSourceSpotify, resp.StatusCode, fmt.Sprintf("unexpected status: %d", resp.StatusCode), nil)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.Bytes(), nil
}

// SearchTrack searches for a track by artist and title
func (c *SpotifyClient) SearchTrack(ctx context.Context, artist, title string) (*TrackMatch, error) {
	query := fmt.Sprintf("artist:%s track:%s", artist, title)
	
	data, err := c.makeRequest(ctx, "/search", map[string]string{
		"q":     query,
		"type":  "track",
		"limit": "1",
	})
	if err != nil {
		return nil, err
	}

	var result spotifySearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Tracks.Items) == 0 {
		return nil, models.ErrNoMatchFound
	}

	track := result.Tracks.Items[0]
	artistName := ""
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}

	confidence := calculateConfidence(artist, artistName, title, track.Name)

	return &TrackMatch{
		ExternalID: track.ID,
		Artist:     artistName,
		Title:      track.Name,
		Album:      track.Album.Name,
		Confidence: confidence,
		Source:     models.DataSourceSpotify,
		Duration:   track.DurationMS / 1000,
	}, nil
}

// GetTopTracks returns top tracks for an artist
func (c *SpotifyClient) GetTopTracks(ctx context.Context, artist string, limit int) ([]*TrackInfo, error) {
	// First, search for the artist
	artistID, err := c.searchArtistID(ctx, artist)
	if err != nil {
		return nil, err
	}

	data, err := c.makeRequest(ctx, fmt.Sprintf("/artists/%s/top-tracks", artistID), map[string]string{
		"market": "US",
	})
	if err != nil {
		return nil, err
	}

	var result spotifyTopTracksResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Tracks))
	for i, t := range result.Tracks {
		if limit > 0 && i >= limit {
			break
		}
		artistName := ""
		if len(t.Artists) > 0 {
			artistName = t.Artists[0].Name
		}
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.ID,
			Artist:     artistName,
			Title:      t.Name,
			Album:      t.Album.Name,
			Rank:       i + 1,
			Playcount:  t.Popularity,
			Duration:   t.DurationMS / 1000,
			Source:     models.DataSourceSpotify,
		})
	}

	return tracks, nil
}

// GetSimilarTracks returns similar tracks based on track features
func (c *SpotifyClient) GetSimilarTracks(ctx context.Context, artist, title string, limit int) ([]*TrackInfo, error) {
	// Search for the seed track
	match, err := c.SearchTrack(ctx, artist, title)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	// Get recommendations based on the track
	data, err := c.makeRequest(ctx, "/recommendations", map[string]string{
		"seed_tracks": match.ExternalID,
		"limit":       strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result spotifyRecommendationsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Tracks))
	for i, t := range result.Tracks {
		artistName := ""
		if len(t.Artists) > 0 {
			artistName = t.Artists[0].Name
		}
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.ID,
			Artist:     artistName,
			Title:      t.Name,
			Album:      t.Album.Name,
			Rank:       i + 1,
			Duration:   t.DurationMS / 1000,
			Source:     models.DataSourceSpotify,
		})
	}

	return tracks, nil
}

// GetTagTracks returns tracks for a genre
func (c *SpotifyClient) GetTagTracks(ctx context.Context, tag string, limit int) ([]*TrackInfo, error) {
	if limit <= 0 {
		limit = 50
	}

	// Spotify uses seed_genres for recommendations
	genre := strings.ToLower(strings.ReplaceAll(tag, " ", "-"))

	data, err := c.makeRequest(ctx, "/recommendations", map[string]string{
		"seed_genres": genre,
		"limit":       strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result spotifyRecommendationsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Tracks))
	for i, t := range result.Tracks {
		artistName := ""
		if len(t.Artists) > 0 {
			artistName = t.Artists[0].Name
		}
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.ID,
			Artist:     artistName,
			Title:      t.Name,
			Album:      t.Album.Name,
			Rank:       i + 1,
			Duration:   t.DurationMS / 1000,
			Source:     models.DataSourceSpotify,
		})
	}

	return tracks, nil
}

// GetArtistInfo returns information about an artist
func (c *SpotifyClient) GetArtistInfo(ctx context.Context, artist string) (*ArtistInfo, error) {
	artistID, err := c.searchArtistID(ctx, artist)
	if err != nil {
		return nil, err
	}

	data, err := c.makeRequest(ctx, fmt.Sprintf("/artists/%s", artistID), nil)
	if err != nil {
		return nil, err
	}

	var result spotifyArtistResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	imageURL := ""
	if len(result.Images) > 0 {
		imageURL = result.Images[0].URL
	}

	return &ArtistInfo{
		ExternalID: result.ID,
		Name:       result.Name,
		URL:        result.ExternalURLs.Spotify,
		ImageURL:   imageURL,
		Listeners:  result.Followers.Total,
		Tags:       result.Genres,
		Source:     models.DataSourceSpotify,
	}, nil
}

// GetSimilarArtists returns similar artists
func (c *SpotifyClient) GetSimilarArtists(ctx context.Context, artist string, limit int) ([]*ArtistInfo, error) {
	artistID, err := c.searchArtistID(ctx, artist)
	if err != nil {
		return nil, err
	}

	data, err := c.makeRequest(ctx, fmt.Sprintf("/artists/%s/related-artists", artistID), nil)
	if err != nil {
		return nil, err
	}

	var result spotifyRelatedArtistsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	artists := make([]*ArtistInfo, 0, len(result.Artists))
	for i, a := range result.Artists {
		if limit > 0 && i >= limit {
			break
		}
		imageURL := ""
		if len(a.Images) > 0 {
			imageURL = a.Images[0].URL
		}
		artists = append(artists, &ArtistInfo{
			ExternalID: a.ID,
			Name:       a.Name,
			URL:        a.ExternalURLs.Spotify,
			ImageURL:   imageURL,
			Listeners:  a.Followers.Total,
			Tags:       a.Genres,
			Source:     models.DataSourceSpotify,
		})
	}

	return artists, nil
}

// searchArtistID searches for an artist and returns their ID
func (c *SpotifyClient) searchArtistID(ctx context.Context, artist string) (string, error) {
	data, err := c.makeRequest(ctx, "/search", map[string]string{
		"q":     artist,
		"type":  "artist",
		"limit": "1",
	})
	if err != nil {
		return "", err
	}

	var result spotifyArtistSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if len(result.Artists.Items) == 0 {
		return "", models.ErrNoMatchFound
	}

	return result.Artists.Items[0].ID, nil
}

// Spotify API response structures
type spotifySearchResponse struct {
	Tracks struct {
		Items []spotifyTrack `json:"items"`
	} `json:"tracks"`
}

type spotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	DurationMS int    `json:"duration_ms"`
	Popularity int    `json:"popularity"`
	Artists    []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"album"`
}

type spotifyTopTracksResponse struct {
	Tracks []spotifyTrack `json:"tracks"`
}

type spotifyRecommendationsResponse struct {
	Tracks []spotifyTrack `json:"tracks"`
}

type spotifyArtistSearchResponse struct {
	Artists struct {
		Items []spotifyArtistResponse `json:"items"`
	} `json:"artists"`
}

type spotifyArtistResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Genres       []string `json:"genres"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

type spotifyRelatedArtistsResponse struct {
	Artists []spotifyArtistResponse `json:"artists"`
}
