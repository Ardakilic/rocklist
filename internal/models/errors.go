// Package models contains all domain models for Rocklist
package models

import "errors"

// Domain errors
var (
	// Database errors
	ErrDatabaseNotInitialized = errors.New("database not initialized")
	ErrSongNotFound           = errors.New("song not found")
	ErrPlaylistNotFound       = errors.New("playlist not found")
	ErrConfigNotFound         = errors.New("config not found")

	// Rockbox errors
	ErrRockboxPathNotSet      = errors.New("rockbox path not set")
	ErrRockboxPathInvalid     = errors.New("rockbox path is invalid")
	ErrRockboxDatabaseNotFound = errors.New("rockbox database not found")
	ErrParseInProgress        = errors.New("parse operation already in progress")
	ErrNoPreFetchedData       = errors.New("no pre-fetched data available")

	// API errors
	ErrAPINotConfigured       = errors.New("API not configured")
	ErrAPIKeyMissing          = errors.New("API key is missing")
	ErrAPIRequestFailed       = errors.New("API request failed")
	ErrAPIRateLimited         = errors.New("API rate limited")
	ErrAPIUnauthorized        = errors.New("API unauthorized")
	ErrDataSourceDisabled     = errors.New("data source is disabled")

	// Playlist errors
	ErrInvalidPlaylistType    = errors.New("invalid playlist type")
	ErrInvalidDataSource      = errors.New("invalid data source")
	ErrTagRequired            = errors.New("tag is required for tag playlist")
	ErrNoMatchingSongs        = errors.New("no matching songs found")
	ErrPlaylistExportFailed   = errors.New("playlist export failed")

	// Matching errors
	ErrNoMatchFound           = errors.New("no match found")
	ErrMultipleMatchesFound   = errors.New("multiple matches found, confidence too low")

	// General errors
	ErrInvalidInput           = errors.New("invalid input")
	ErrOperationCancelled     = errors.New("operation cancelled")
)

// APIError represents an error from an external API
type APIError struct {
	Source     DataSource
	StatusCode int
	Message    string
	Err        error
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Source.DisplayName() + ": " + e.Err.Error()
	}
	return e.Source.DisplayName() + ": " + e.Message
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new API error
func NewAPIError(source DataSource, statusCode int, message string, err error) *APIError {
	return &APIError{
		Source:     source,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}
