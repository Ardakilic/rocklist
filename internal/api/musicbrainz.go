// Package api provides clients for external music APIs
package api

import (
	"bytes"
	"context"
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
	musicBrainzAPIURL = "https://musicbrainz.org/ws/2"
	// MusicBrainz requires rate limiting: 1 request per second
	musicBrainzRateLimit = 1100 * time.Millisecond
)

// MusicBrainzClient is a client for the MusicBrainz API
type MusicBrainzClient struct {
	*BaseClient
	userAgent    string
	lastRequest  time.Time
	mu           sync.Mutex
	logger       Logger
	baseURL      string // Allow override for testing
}

// SetBaseURL sets a custom base URL (for testing)
func (c *MusicBrainzClient) SetBaseURL(url string) {
	c.baseURL = url
}

// getBaseURL returns the base URL to use
func (c *MusicBrainzClient) getBaseURL() string {
	if c.baseURL != "" {
		return c.baseURL
	}
	return musicBrainzAPIURL
}

// NewMusicBrainzClient creates a new MusicBrainz API client
func NewMusicBrainzClient(userAgent string, logger Logger) *MusicBrainzClient {
	if userAgent == "" {
		userAgent = "Rocklist/1.0.0 ( https://github.com/Ardakilic/Rocklist )"
	}
	return &MusicBrainzClient{
		BaseClient: NewBaseClient(models.DataSourceMusicBrainz, userAgent),
		userAgent:  userAgent,
		logger:     logger,
	}
}

// IsConfigured returns true if the client is properly configured
func (c *MusicBrainzClient) IsConfigured() bool {
	return c.userAgent != ""
}

// SetUserAgent sets the user agent
func (c *MusicBrainzClient) SetUserAgent(userAgent string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userAgent = userAgent
}

// rateLimitWait ensures we don't exceed the rate limit
func (c *MusicBrainzClient) rateLimitWait() {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.lastRequest)
	if elapsed < musicBrainzRateLimit {
		time.Sleep(musicBrainzRateLimit - elapsed)
	}
	c.lastRequest = time.Now()
}

// makeRequest makes a request to the MusicBrainz API
func (c *MusicBrainzClient) makeRequest(ctx context.Context, endpoint string, params map[string]string) ([]byte, error) {
	c.rateLimitWait()

	u, err := url.Parse(c.getBaseURL() + endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("fmt", "json")
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	if c.logger != nil {
		c.logger.Debug("MusicBrainz API request: %s", u.String())
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return nil, models.NewAPIError(models.DataSourceMusicBrainz, 0, "request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 || resp.StatusCode == 503 {
		return nil, models.NewAPIError(models.DataSourceMusicBrainz, resp.StatusCode, "rate limited", models.ErrAPIRateLimited)
	}
	if resp.StatusCode != 200 {
		return nil, models.NewAPIError(models.DataSourceMusicBrainz, resp.StatusCode, fmt.Sprintf("unexpected status: %d", resp.StatusCode), nil)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SearchTrack searches for a track by artist and title
func (c *MusicBrainzClient) SearchTrack(ctx context.Context, artist, title string) (*TrackMatch, error) {
	query := fmt.Sprintf("recording:\"%s\" AND artist:\"%s\"", escapeQuery(title), escapeQuery(artist))
	
	data, err := c.makeRequest(ctx, "/recording", map[string]string{
		"query": query,
		"limit": "1",
	})
	if err != nil {
		return nil, err
	}

	var result mbRecordingSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if len(result.Recordings) == 0 {
		return nil, models.ErrNoMatchFound
	}

	rec := result.Recordings[0]
	artistName := ""
	if len(rec.ArtistCredit) > 0 {
		artistName = rec.ArtistCredit[0].Name
	}

	confidence := calculateConfidence(artist, artistName, title, rec.Title)

	return &TrackMatch{
		ExternalID: rec.ID,
		Artist:     artistName,
		Title:      rec.Title,
		Confidence: confidence,
		Source:     models.DataSourceMusicBrainz,
		Duration:   rec.Length / 1000,
	}, nil
}

// GetTopTracks returns top tracks for an artist (based on recording count)
func (c *MusicBrainzClient) GetTopTracks(ctx context.Context, artist string, limit int) ([]*TrackInfo, error) {
	// First, find the artist
	artistID, err := c.searchArtistID(ctx, artist)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	// Get recordings by this artist, sorted by popularity (using release count as proxy)
	data, err := c.makeRequest(ctx, "/recording", map[string]string{
		"query": fmt.Sprintf("arid:%s", artistID),
		"limit": strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result mbRecordingSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Recordings))
	for i, rec := range result.Recordings {
		artistName := artist
		if len(rec.ArtistCredit) > 0 {
			artistName = rec.ArtistCredit[0].Name
		}
		tracks = append(tracks, &TrackInfo{
			ExternalID: rec.ID,
			Artist:     artistName,
			Title:      rec.Title,
			Rank:       i + 1,
			Duration:   rec.Length / 1000,
			Source:     models.DataSourceMusicBrainz,
		})
	}

	return tracks, nil
}

// GetSimilarTracks returns similar tracks (MusicBrainz doesn't directly support this)
func (c *MusicBrainzClient) GetSimilarTracks(ctx context.Context, artist, title string, limit int) ([]*TrackInfo, error) {
	// MusicBrainz doesn't have a direct similar tracks feature
	// We can try to find tracks from similar artists or same album
	
	// First, find the recording
	match, err := c.SearchTrack(ctx, artist, title)
	if err != nil {
		return nil, err
	}

	// Get the recording details to find the release
	data, err := c.makeRequest(ctx, fmt.Sprintf("/recording/%s", match.ExternalID), map[string]string{
		"inc": "releases+artists",
	})
	if err != nil {
		return nil, err
	}

	var rec mbRecordingResponse
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}

	// Get other tracks from the same artist
	return c.GetTopTracks(ctx, artist, limit)
}

// GetTagTracks returns tracks for a tag/genre
func (c *MusicBrainzClient) GetTagTracks(ctx context.Context, tag string, limit int) ([]*TrackInfo, error) {
	if limit <= 0 {
		limit = 50
	}

	// Search for recordings with this tag
	data, err := c.makeRequest(ctx, "/recording", map[string]string{
		"query": fmt.Sprintf("tag:\"%s\"", escapeQuery(tag)),
		"limit": strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result mbRecordingSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tracks := make([]*TrackInfo, 0, len(result.Recordings))
	for i, rec := range result.Recordings {
		artistName := ""
		if len(rec.ArtistCredit) > 0 {
			artistName = rec.ArtistCredit[0].Name
		}
		tracks = append(tracks, &TrackInfo{
			ExternalID: rec.ID,
			Artist:     artistName,
			Title:      rec.Title,
			Rank:       i + 1,
			Duration:   rec.Length / 1000,
			Source:     models.DataSourceMusicBrainz,
		})
	}

	return tracks, nil
}

// GetArtistInfo returns information about an artist
func (c *MusicBrainzClient) GetArtistInfo(ctx context.Context, artist string) (*ArtistInfo, error) {
	artistID, err := c.searchArtistID(ctx, artist)
	if err != nil {
		return nil, err
	}

	data, err := c.makeRequest(ctx, fmt.Sprintf("/artist/%s", artistID), map[string]string{
		"inc": "tags+url-rels",
	})
	if err != nil {
		return nil, err
	}

	var result mbArtistResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	tags := make([]string, 0, len(result.Tags))
	for _, t := range result.Tags {
		tags = append(tags, t.Name)
	}

	return &ArtistInfo{
		ExternalID: result.ID,
		Name:       result.Name,
		Tags:       tags,
		Source:     models.DataSourceMusicBrainz,
	}, nil
}

// GetSimilarArtists returns similar artists (MusicBrainz doesn't directly support this)
func (c *MusicBrainzClient) GetSimilarArtists(ctx context.Context, artist string, limit int) ([]*ArtistInfo, error) {
	// Get artist info to find tags
	info, err := c.GetArtistInfo(ctx, artist)
	if err != nil {
		return nil, err
	}

	if len(info.Tags) == 0 {
		return []*ArtistInfo{}, nil
	}

	// Search for artists with similar tags
	tagQuery := strings.Join(info.Tags[:min(3, len(info.Tags))], " OR ")
	data, err := c.makeRequest(ctx, "/artist", map[string]string{
		"query": fmt.Sprintf("tag:(%s)", tagQuery),
		"limit": strconv.Itoa(limit),
	})
	if err != nil {
		return nil, err
	}

	var result mbArtistSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	artists := make([]*ArtistInfo, 0, len(result.Artists))
	for _, a := range result.Artists {
		if a.Name == artist {
			continue // Skip the original artist
		}
		tags := make([]string, 0, len(a.Tags))
		for _, t := range a.Tags {
			tags = append(tags, t.Name)
		}
		artists = append(artists, &ArtistInfo{
			ExternalID: a.ID,
			Name:       a.Name,
			Tags:       tags,
			Source:     models.DataSourceMusicBrainz,
		})
	}

	return artists, nil
}

// searchArtistID searches for an artist and returns their ID
func (c *MusicBrainzClient) searchArtistID(ctx context.Context, artist string) (string, error) {
	data, err := c.makeRequest(ctx, "/artist", map[string]string{
		"query": fmt.Sprintf("artist:\"%s\"", escapeQuery(artist)),
		"limit": "1",
	})
	if err != nil {
		return "", err
	}

	var result mbArtistSearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if len(result.Artists) == 0 {
		return "", models.ErrNoMatchFound
	}

	return result.Artists[0].ID, nil
}

// escapeQuery escapes special characters in MusicBrainz queries
func escapeQuery(s string) string {
	// Escape special Lucene characters
	special := []string{"+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "\\", "/"}
	result := s
	for _, c := range special {
		result = strings.ReplaceAll(result, c, "\\"+c)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MusicBrainz API response structures
type mbRecordingSearchResponse struct {
	Recordings []mbRecording `json:"recordings"`
}

type mbRecording struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Length       int    `json:"length"`
	ArtistCredit []struct {
		Name   string `json:"name"`
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"artist-credit"`
	Releases []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"releases"`
}

type mbRecordingResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Length       int    `json:"length"`
	ArtistCredit []struct {
		Name   string `json:"name"`
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"artist-credit"`
	Releases []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"releases"`
}

type mbArtistSearchResponse struct {
	Artists []mbArtist `json:"artists"`
}

type mbArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tags []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
}

type mbArtistResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tags []struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tags"`
	Relations []struct {
		Type string `json:"type"`
		URL  struct {
			Resource string `json:"resource"`
		} `json:"url"`
	} `json:"relations"`
}
