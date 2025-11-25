// Package api provides clients for external music APIs
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Ardakilic/rocklist/internal/models"
)

const (
	lastFMBaseURL = "https://ws.audioscrobbler.com/2.0/"
)

// LastFMClient is a client for the Last.fm API
type LastFMClient struct {
	*BaseClient
	apiKey    string
	apiSecret string
	logger    Logger
	baseURL   string // Allow override for testing
}

// SetBaseURL sets a custom base URL (for testing)
func (c *LastFMClient) SetBaseURL(url string) {
	c.baseURL = url
}

// getBaseURL returns the base URL to use
func (c *LastFMClient) getBaseURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return lastFMBaseURL
}

// NewLastFMClient creates a new Last.fm API client
func NewLastFMClient(apiKey, apiSecret string, logger Logger) *LastFMClient {
	return &LastFMClient{
		BaseClient: NewBaseClient(models.DataSourceLastFM, "Rocklist/1.0"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		logger:     logger,
	}
}

// IsConfigured returns true if the client is properly configured
func (c *LastFMClient) IsConfigured() bool {
	return c.apiKey != ""
}

// SetCredentials sets the API credentials
func (c *LastFMClient) SetCredentials(apiKey, apiSecret string) {
	c.apiKey = apiKey
	c.apiSecret = apiSecret
}

// makeRequest makes a request to the Last.fm API
func (c *LastFMClient) makeRequest(ctx context.Context, method string, params map[string]string) ([]byte, error) {
	if !c.IsConfigured() {
		return nil, models.ErrAPIKeyMissing
	}

	u, err := url.Parse(c.getBaseURL())
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("method", method)
	q.Set("api_key", c.apiKey)
	q.Set("format", "json")
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	if c.logger != nil {
		c.logger.Debug("Last.fm API request: %s %s", method, u.String())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent())

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, models.NewAPIError(models.DataSourceLastFM, 0, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, models.NewAPIError(models.DataSourceLastFM, 429, "rate limited", models.ErrAPIRateLimited)
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, models.NewAPIError(models.DataSourceLastFM, resp.StatusCode, "unauthorized", models.ErrAPIUnauthorized)
	}
	if resp.StatusCode != 200 {
		return nil, models.NewAPIError(models.DataSourceLastFM, resp.StatusCode, fmt.Sprintf("unexpected status: %d", resp.StatusCode), nil)
	}

	var buf []byte
	buf = make([]byte, 0, 1024*1024) // 1MB buffer
	buf, err = readAll(resp.Body, buf)
	if err != nil {
		return nil, err
	}

	// Check for API error
	var apiErr struct {
		Error   int    `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(buf, &apiErr) == nil && apiErr.Error != 0 {
		return nil, models.NewAPIError(models.DataSourceLastFM, apiErr.Error, apiErr.Message, nil)
	}

	return buf, nil
}

// SearchTrack searches for a track by artist and title
func (c *LastFMClient) SearchTrack(ctx context.Context, artist, title string) (*TrackMatch, error) {
	data, err := c.makeRequest(ctx, "track.search", map[string]string{
		"artist": artist,
		"track":  title,
		"limit":  "1",
	})
	if err != nil {
		return nil, err
	}

	var result lastFMTrackSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Results.TrackMatches.Track) == 0 {
		return nil, models.ErrNoMatchFound
	}

	track := result.Results.TrackMatches.Track[0]
	
	// Calculate confidence based on string similarity
	confidence := calculateConfidence(artist, track.Artist, title, track.Name)

	return &TrackMatch{
		ExternalID: track.MBID,
		Artist:     track.Artist,
		Title:      track.Name,
		Confidence: confidence,
		Source:     models.DataSourceLastFM,
		URL:        track.URL,
	}, nil
}

// GetTopTracks returns top tracks for an artist
func (c *LastFMClient) GetTopTracks(ctx context.Context, artist string, limit int) ([]*TrackInfo, error) {
	if limit <= 0 {
		limit = 50
	}

	data, err := c.makeRequest(ctx, "artist.getTopTracks", map[string]string{
		"artist": artist,
		"limit":  strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result lastFMArtistTopTracksResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.TopTracks.Track))
	for i, t := range result.TopTracks.Track {
		playcount, _ := strconv.Atoi(t.Playcount)
		duration, _ := strconv.Atoi(t.Duration)
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.MBID,
			Artist:     t.Artist.Name,
			Title:      t.Name,
			Rank:       i + 1,
			Playcount:  playcount,
			Duration:   duration,
			URL:        t.URL,
			Source:     models.DataSourceLastFM,
		})
	}

	return tracks, nil
}

// GetSimilarTracks returns similar tracks to a given track
func (c *LastFMClient) GetSimilarTracks(ctx context.Context, artist, title string, limit int) ([]*TrackInfo, error) {
	if limit <= 0 {
		limit = 50
	}

	data, err := c.makeRequest(ctx, "track.getSimilar", map[string]string{
		"artist": artist,
		"track":  title,
		"limit":  strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result lastFMSimilarTracksResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.SimilarTracks.Track))
	for i, t := range result.SimilarTracks.Track {
		playcount, _ := strconv.Atoi(t.Playcount)
		duration, _ := strconv.Atoi(t.Duration)
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.MBID,
			Artist:     t.Artist.Name,
			Title:      t.Name,
			Rank:       i + 1,
			Playcount:  playcount,
			Duration:   duration,
			URL:        t.URL,
			Source:     models.DataSourceLastFM,
		})
	}

	return tracks, nil
}

// GetTagTracks returns top tracks for a tag/genre
func (c *LastFMClient) GetTagTracks(ctx context.Context, tag string, limit int) ([]*TrackInfo, error) {
	if limit <= 0 {
		limit = 50
	}

	data, err := c.makeRequest(ctx, "tag.getTopTracks", map[string]string{
		"tag":   tag,
		"limit": strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result lastFMTagTopTracksResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Tracks.Track))
	for i, t := range result.Tracks.Track {
		duration, _ := strconv.Atoi(t.Duration)
		tracks = append(tracks, &TrackInfo{
			ExternalID: t.MBID,
			Artist:     t.Artist.Name,
			Title:      t.Name,
			Rank:       i + 1,
			Duration:   duration,
			URL:        t.URL,
			Source:     models.DataSourceLastFM,
		})
	}

	return tracks, nil
}

// GetArtistInfo returns information about an artist
func (c *LastFMClient) GetArtistInfo(ctx context.Context, artist string) (*ArtistInfo, error) {
	data, err := c.makeRequest(ctx, "artist.getInfo", map[string]string{
		"artist": artist,
	})
	if err != nil {
		return nil, err
	}

	var result lastFMArtistInfoResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	a := result.Artist
	listeners, _ := strconv.Atoi(a.Stats.Listeners)
	playcount, _ := strconv.Atoi(a.Stats.Playcount)

	tags := make([]string, 0, len(a.Tags.Tag))
	for _, t := range a.Tags.Tag {
		tags = append(tags, t.Name)
	}

	similar := make([]string, 0, len(a.Similar.Artist))
	for _, s := range a.Similar.Artist {
		similar = append(similar, s.Name)
	}

	return &ArtistInfo{
		ExternalID: a.MBID,
		Name:       a.Name,
		URL:        a.URL,
		Listeners:  listeners,
		Playcount:  playcount,
		Tags:       tags,
		Similar:    similar,
		Source:     models.DataSourceLastFM,
	}, nil
}

// GetSimilarArtists returns similar artists
func (c *LastFMClient) GetSimilarArtists(ctx context.Context, artist string, limit int) ([]*ArtistInfo, error) {
	if limit <= 0 {
		limit = 20
	}

	data, err := c.makeRequest(ctx, "artist.getSimilar", map[string]string{
		"artist": artist,
		"limit":  strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result lastFMSimilarArtistsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	artists := make([]*ArtistInfo, 0, len(result.SimilarArtists.Artist))
	for _, a := range result.SimilarArtists.Artist {
		artists = append(artists, &ArtistInfo{
			ExternalID: a.MBID,
			Name:       a.Name,
			URL:        a.URL,
			Source:     models.DataSourceLastFM,
		})
	}

	return artists, nil
}

// Last.fm API response structures
type lastFMTrackSearchResponse struct {
	Results struct {
		TrackMatches struct {
			Track []struct {
				Name   string `json:"name"`
				Artist string `json:"artist"`
				URL    string `json:"url"`
				MBID   string `json:"mbid"`
			} `json:"track"`
		} `json:"trackmatches"`
	} `json:"results"`
}

type lastFMArtistTopTracksResponse struct {
	TopTracks struct {
		Track []struct {
			Name      string `json:"name"`
			Playcount string `json:"playcount"`
			Duration  string `json:"duration"`
			URL       string `json:"url"`
			MBID      string `json:"mbid"`
			Artist    struct {
				Name string `json:"name"`
				MBID string `json:"mbid"`
			} `json:"artist"`
		} `json:"track"`
	} `json:"toptracks"`
}

type lastFMSimilarTracksResponse struct {
	SimilarTracks struct {
		Track []struct {
			Name      string `json:"name"`
			Playcount string `json:"playcount"`
			Duration  string `json:"duration"`
			URL       string `json:"url"`
			MBID      string `json:"mbid"`
			Artist    struct {
				Name string `json:"name"`
				MBID string `json:"mbid"`
			} `json:"artist"`
		} `json:"track"`
	} `json:"similartracks"`
}

type lastFMTagTopTracksResponse struct {
	Tracks struct {
		Track []struct {
			Name     string `json:"name"`
			Duration string `json:"duration"`
			URL      string `json:"url"`
			MBID     string `json:"mbid"`
			Artist   struct {
				Name string `json:"name"`
				MBID string `json:"mbid"`
			} `json:"artist"`
		} `json:"track"`
	} `json:"tracks"`
}

type lastFMArtistInfoResponse struct {
	Artist struct {
		Name  string `json:"name"`
		MBID  string `json:"mbid"`
		URL   string `json:"url"`
		Stats struct {
			Listeners string `json:"listeners"`
			Playcount string `json:"playcount"`
		} `json:"stats"`
		Tags struct {
			Tag []struct {
				Name string `json:"name"`
			} `json:"tag"`
		} `json:"tags"`
		Similar struct {
			Artist []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"artist"`
		} `json:"similar"`
	} `json:"artist"`
}

type lastFMSimilarArtistsResponse struct {
	SimilarArtists struct {
		Artist []struct {
			Name string `json:"name"`
			MBID string `json:"mbid"`
			URL  string `json:"url"`
		} `json:"artist"`
	} `json:"similarartists"`
}

// calculateConfidence calculates a confidence score based on string similarity
func calculateConfidence(inputArtist, matchArtist, inputTitle, matchTitle string) float64 {
	artistSimilarity := stringSimilarity(strings.ToLower(inputArtist), strings.ToLower(matchArtist))
	titleSimilarity := stringSimilarity(strings.ToLower(inputTitle), strings.ToLower(matchTitle))
	return (artistSimilarity + titleSimilarity) / 2
}

// stringSimilarity calculates a simple similarity score between two strings
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple containment check
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

	// Levenshtein-like ratio
	matches := 0
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] == b[i] {
			matches++
		}
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	return float64(matches) / float64(maxLen)
}

// readAll reads all data from reader into buffer
func readAll(r interface{ Read([]byte) (int, error) }, buf []byte) ([]byte, error) {
	for {
		if len(buf) == cap(buf) {
			newBuf := make([]byte, len(buf), 2*cap(buf)+512)
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := r.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err != nil {
			if err.Error() == "EOF" {
				return buf, nil
			}
			return buf, err
		}
	}
}
