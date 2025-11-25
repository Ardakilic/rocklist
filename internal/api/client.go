// Package api provides clients for external music APIs
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
)

// Client is the interface for music API clients
type Client interface {
	// GetSource returns the data source type
	GetSource() models.DataSource
	// IsConfigured returns true if the client is properly configured
	IsConfigured() bool
	// SearchTrack searches for a track by artist and title
	SearchTrack(ctx context.Context, artist, title string) (*TrackMatch, error)
	// GetTopTracks returns top tracks for an artist
	GetTopTracks(ctx context.Context, artist string, limit int) ([]*TrackInfo, error)
	// GetSimilarTracks returns similar tracks to a given track
	GetSimilarTracks(ctx context.Context, artist, title string, limit int) ([]*TrackInfo, error)
	// GetTagTracks returns top tracks for a tag/genre
	GetTagTracks(ctx context.Context, tag string, limit int) ([]*TrackInfo, error)
	// GetArtistInfo returns information about an artist
	GetArtistInfo(ctx context.Context, artist string) (*ArtistInfo, error)
	// GetSimilarArtists returns similar artists
	GetSimilarArtists(ctx context.Context, artist string, limit int) ([]*ArtistInfo, error)
}

// TrackMatch represents a matched track from an external API
type TrackMatch struct {
	ExternalID  string  `json:"external_id"`
	Artist      string  `json:"artist"`
	Title       string  `json:"title"`
	Album       string  `json:"album,omitempty"`
	Confidence  float64 `json:"confidence"`
	Source      models.DataSource `json:"source"`
	URL         string  `json:"url,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	Playcount   int     `json:"playcount,omitempty"`
}

// TrackInfo represents track information from an external API
type TrackInfo struct {
	ExternalID  string  `json:"external_id"`
	Artist      string  `json:"artist"`
	Title       string  `json:"title"`
	Album       string  `json:"album,omitempty"`
	Rank        int     `json:"rank,omitempty"`
	Playcount   int     `json:"playcount,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	URL         string  `json:"url,omitempty"`
	Source      models.DataSource `json:"source"`
}

// ArtistInfo represents artist information from an external API
type ArtistInfo struct {
	ExternalID  string   `json:"external_id"`
	Name        string   `json:"name"`
	URL         string   `json:"url,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	Listeners   int      `json:"listeners,omitempty"`
	Playcount   int      `json:"playcount,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Similar     []string `json:"similar,omitempty"`
	Source      models.DataSource `json:"source"`
}

// BaseClient provides common HTTP client functionality
type BaseClient struct {
	httpClient *http.Client
	source     models.DataSource
	userAgent  string
}

// NewBaseClient creates a new base client
func NewBaseClient(source models.DataSource, userAgent string) *BaseClient {
	return &BaseClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		source:    source,
		userAgent: userAgent,
	}
}

// GetSource returns the data source type
func (c *BaseClient) GetSource() models.DataSource {
	return c.source
}

// HTTPClient returns the underlying HTTP client
func (c *BaseClient) HTTPClient() *http.Client {
	return c.httpClient
}

// UserAgent returns the user agent string
func (c *BaseClient) UserAgent() string {
	return c.userAgent
}

// Logger interface for API clients
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}
