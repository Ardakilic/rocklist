package models

import (
	"testing"
)

func TestPlaylistType_String(t *testing.T) {
	tests := []struct {
		pt   PlaylistType
		want string
	}{
		{PlaylistTypeTopSongs, "top_songs"},
		{PlaylistTypeMixedSongs, "mixed_songs"},
		{PlaylistTypeSimilar, "similar"},
		{PlaylistTypeTag, "tag"},
	}

	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := tt.pt.String(); got != tt.want {
				t.Errorf("PlaylistType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlaylistType_DisplayName(t *testing.T) {
	tests := []struct {
		pt   PlaylistType
		want string
	}{
		{PlaylistTypeTopSongs, "Top Songs"},
		{PlaylistTypeMixedSongs, "Mixed Songs"},
		{PlaylistTypeSimilar, "Similar Songs"},
		{PlaylistTypeTag, "Tag Radio"},
		{PlaylistType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := tt.pt.DisplayName(); got != tt.want {
				t.Errorf("PlaylistType.DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataSource_String(t *testing.T) {
	tests := []struct {
		ds   DataSource
		want string
	}{
		{DataSourceLastFM, "lastfm"},
		{DataSourceSpotify, "spotify"},
		{DataSourceMusicBrainz, "musicbrainz"},
	}

	for _, tt := range tests {
		t.Run(string(tt.ds), func(t *testing.T) {
			if got := tt.ds.String(); got != tt.want {
				t.Errorf("DataSource.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataSource_DisplayName(t *testing.T) {
	tests := []struct {
		ds   DataSource
		want string
	}{
		{DataSourceLastFM, "Last.fm"},
		{DataSourceSpotify, "Spotify"},
		{DataSourceMusicBrainz, "MusicBrainz"},
		{DataSource("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.ds), func(t *testing.T) {
			if got := tt.ds.DisplayName(); got != tt.want {
				t.Errorf("DataSource.DisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlaylistRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaylistRequest
		wantErr error
	}{
		{
			name:    "valid top songs request",
			req:     PlaylistRequest{Type: PlaylistTypeTopSongs, DataSource: DataSourceLastFM},
			wantErr: nil,
		},
		{
			name:    "valid tag request",
			req:     PlaylistRequest{Type: PlaylistTypeTag, DataSource: DataSourceSpotify, Tag: "metal"},
			wantErr: nil,
		},
		{
			name:    "missing type",
			req:     PlaylistRequest{DataSource: DataSourceLastFM},
			wantErr: ErrInvalidPlaylistType,
		},
		{
			name:    "missing data source",
			req:     PlaylistRequest{Type: PlaylistTypeTopSongs},
			wantErr: ErrInvalidDataSource,
		},
		{
			name:    "tag request without tag",
			req:     PlaylistRequest{Type: PlaylistTypeTag, DataSource: DataSourceLastFM},
			wantErr: ErrTagRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if err != tt.wantErr {
				t.Errorf("PlaylistRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaylistRequest_Validate_DefaultLimit(t *testing.T) {
	req := PlaylistRequest{Type: PlaylistTypeTopSongs, DataSource: DataSourceLastFM}
	err := req.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Limit != 50 {
		t.Errorf("expected default limit 50, got %d", req.Limit)
	}
}

func TestPlaylist_TableName(t *testing.T) {
	p := Playlist{}
	if got := p.TableName(); got != "playlists" {
		t.Errorf("Playlist.TableName() = %v, want playlists", got)
	}
}

func TestPlaylistSong_TableName(t *testing.T) {
	ps := PlaylistSong{}
	if got := ps.TableName(); got != "playlist_songs" {
		t.Errorf("PlaylistSong.TableName() = %v, want playlist_songs", got)
	}
}

func TestPlaylistRequest_UseAlbumArtist(t *testing.T) {
	tests := []struct {
		name           string
		req            PlaylistRequest
		wantErr        error
		useAlbumArtist bool
	}{
		{
			name: "request with UseAlbumArtist true",
			req: PlaylistRequest{
				Type:           PlaylistTypeTopSongs,
				DataSource:     DataSourceLastFM,
				Artist:         "Metallica",
				UseAlbumArtist: true,
			},
			wantErr:        nil,
			useAlbumArtist: true,
		},
		{
			name: "request with UseAlbumArtist false (default)",
			req: PlaylistRequest{
				Type:       PlaylistTypeTopSongs,
				DataSource: DataSourceLastFM,
				Artist:     "Metallica",
			},
			wantErr:        nil,
			useAlbumArtist: false,
		},
		{
			name: "tag request with UseAlbumArtist",
			req: PlaylistRequest{
				Type:           PlaylistTypeTag,
				DataSource:     DataSourceLastFM,
				Tag:            "metal",
				UseAlbumArtist: true,
			},
			wantErr:        nil,
			useAlbumArtist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if err != tt.wantErr {
				t.Errorf("PlaylistRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.req.UseAlbumArtist != tt.useAlbumArtist {
				t.Errorf("PlaylistRequest.UseAlbumArtist = %v, want %v", tt.req.UseAlbumArtist, tt.useAlbumArtist)
			}
		})
	}
}
