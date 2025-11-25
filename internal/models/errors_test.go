package models

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name: "with underlying error",
			apiError: &APIError{
				Source:     DataSourceLastFM,
				StatusCode: 401,
				Message:    "unauthorized",
				Err:        ErrAPIUnauthorized,
			},
			want: "Last.fm: API unauthorized",
		},
		{
			name: "without underlying error",
			apiError: &APIError{
				Source:     DataSourceSpotify,
				StatusCode: 500,
				Message:    "internal server error",
				Err:        nil,
			},
			want: "Spotify: internal server error",
		},
		{
			name: "MusicBrainz error",
			apiError: &APIError{
				Source:     DataSourceMusicBrainz,
				StatusCode: 429,
				Message:    "rate limited",
				Err:        ErrAPIRateLimited,
			},
			want: "MusicBrainz: API rate limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			if got != tt.want {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	underlyingErr := ErrAPIUnauthorized
	apiErr := &APIError{
		Source:     DataSourceLastFM,
		StatusCode: 401,
		Message:    "unauthorized",
		Err:        underlyingErr,
	}

	if apiErr.Unwrap() != underlyingErr {
		t.Errorf("APIError.Unwrap() = %v, want %v", apiErr.Unwrap(), underlyingErr)
	}
}

func TestAPIError_Unwrap_Nil(t *testing.T) {
	apiErr := &APIError{
		Source:     DataSourceLastFM,
		StatusCode: 500,
		Message:    "error",
		Err:        nil,
	}

	if apiErr.Unwrap() != nil {
		t.Errorf("APIError.Unwrap() = %v, want nil", apiErr.Unwrap())
	}
}

func TestNewAPIError(t *testing.T) {
	apiErr := NewAPIError(DataSourceSpotify, 403, "forbidden", ErrAPIUnauthorized)

	if apiErr.Source != DataSourceSpotify {
		t.Errorf("NewAPIError() Source = %v, want %v", apiErr.Source, DataSourceSpotify)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("NewAPIError() StatusCode = %v, want %v", apiErr.StatusCode, 403)
	}
	if apiErr.Message != "forbidden" {
		t.Errorf("NewAPIError() Message = %v, want %v", apiErr.Message, "forbidden")
	}
	if apiErr.Err != ErrAPIUnauthorized {
		t.Errorf("NewAPIError() Err = %v, want %v", apiErr.Err, ErrAPIUnauthorized)
	}
}

func TestAPIError_ErrorsIs(t *testing.T) {
	apiErr := NewAPIError(DataSourceLastFM, 401, "unauthorized", ErrAPIUnauthorized)

	if !errors.Is(apiErr, ErrAPIUnauthorized) {
		t.Error("errors.Is() should return true for underlying error")
	}
}

func TestDomainErrors_Defined(t *testing.T) {
	// Test that all domain errors are defined and not nil
	domainErrors := []error{
		ErrDatabaseNotInitialized,
		ErrSongNotFound,
		ErrPlaylistNotFound,
		ErrConfigNotFound,
		ErrRockboxPathNotSet,
		ErrRockboxPathInvalid,
		ErrRockboxDatabaseNotFound,
		ErrParseInProgress,
		ErrNoPreFetchedData,
		ErrAPINotConfigured,
		ErrAPIKeyMissing,
		ErrAPIRequestFailed,
		ErrAPIRateLimited,
		ErrAPIUnauthorized,
		ErrDataSourceDisabled,
		ErrInvalidPlaylistType,
		ErrInvalidDataSource,
		ErrTagRequired,
		ErrNoMatchingSongs,
		ErrPlaylistExportFailed,
		ErrNoMatchFound,
		ErrMultipleMatchesFound,
		ErrInvalidInput,
		ErrOperationCancelled,
	}

	for i, err := range domainErrors {
		if err == nil {
			t.Errorf("Domain error at index %d is nil", i)
		}
	}
}

func TestDomainErrors_Messages(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{ErrSongNotFound, "song not found"},
		{ErrPlaylistNotFound, "playlist not found"},
		{ErrConfigNotFound, "config not found"},
		{ErrRockboxPathNotSet, "rockbox path not set"},
		{ErrAPIKeyMissing, "API key is missing"},
		{ErrAPIRateLimited, "API rate limited"},
		{ErrNoMatchFound, "no match found"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error message = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestAPIError_ImplementsError(t *testing.T) {
	var _ error = &APIError{}
}
