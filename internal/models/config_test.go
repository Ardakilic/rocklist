package models

import (
	"testing"
)

func TestConfig_TableName(t *testing.T) {
	c := Config{}
	if got := c.TableName(); got != "configs" {
		t.Errorf("Config.TableName() = %v, want configs", got)
	}
}

func TestAppConfig_IsSourceEnabled(t *testing.T) {
	config := &AppConfig{
		EnabledSources: []DataSource{DataSourceLastFM, DataSourceSpotify},
	}

	tests := []struct {
		source DataSource
		want   bool
	}{
		{DataSourceLastFM, true},
		{DataSourceSpotify, true},
		{DataSourceMusicBrainz, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if got := config.IsSourceEnabled(tt.source); got != tt.want {
				t.Errorf("AppConfig.IsSourceEnabled(%v) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestAppConfig_GetEnabledSources(t *testing.T) {
	tests := []struct {
		name   string
		config AppConfig
		want   int
	}{
		{
			name:   "no sources enabled",
			config: AppConfig{},
			want:   0,
		},
		{
			name: "lastfm enabled",
			config: AppConfig{
				LastFM: LastFMConfig{Enabled: true},
			},
			want: 1,
		},
		{
			name: "all enabled",
			config: AppConfig{
				LastFM:      LastFMConfig{Enabled: true},
				Spotify:     SpotifyConfig{Enabled: true},
				MusicBrainz: MusicBrainzConfig{Enabled: true},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetEnabledSources()
			if len(got) != tt.want {
				t.Errorf("AppConfig.GetEnabledSources() returned %d sources, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseStatus_Progress(t *testing.T) {
	tests := []struct {
		name   string
		status ParseStatus
		want   float64
	}{
		{
			name:   "zero total",
			status: ParseStatus{TotalSongs: 0, ProcessedSongs: 0},
			want:   0,
		},
		{
			name:   "half processed",
			status: ParseStatus{TotalSongs: 100, ProcessedSongs: 50},
			want:   50,
		},
		{
			name:   "fully processed",
			status: ParseStatus{TotalSongs: 100, ProcessedSongs: 100},
			want:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Progress(); got != tt.want {
				t.Errorf("ParseStatus.Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}
