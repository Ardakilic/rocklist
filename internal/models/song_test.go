package models

import (
	"testing"
)

func TestSong_IsMatched(t *testing.T) {
	tests := []struct {
		name string
		song Song
		want bool
	}{
		{
			name: "not matched - empty IDs",
			song: Song{},
			want: false,
		},
		{
			name: "matched with MusicBrainz ID",
			song: Song{MusicBrainzID: "abc-123"},
			want: true,
		},
		{
			name: "matched with Spotify ID",
			song: Song{SpotifyID: "spotify:track:123"},
			want: true,
		},
		{
			name: "matched with LastFM ID",
			song: Song{LastFMID: "lastfm-123"},
			want: true,
		},
		{
			name: "matched with multiple IDs",
			song: Song{MusicBrainzID: "abc", SpotifyID: "def"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.song.IsMatched(); got != tt.want {
				t.Errorf("Song.IsMatched() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSong_GetDisplayName(t *testing.T) {
	tests := []struct {
		name string
		song Song
		want string
	}{
		{
			name: "with artist and title",
			song: Song{Artist: "Metallica", Title: "Master of Puppets"},
			want: "Metallica - Master of Puppets",
		},
		{
			name: "title only",
			song: Song{Title: "Unknown Track"},
			want: "Unknown Track",
		},
		{
			name: "path only",
			song: Song{Path: "/Music/song.mp3"},
			want: "/Music/song.mp3",
		},
		{
			name: "empty song",
			song: Song{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.song.GetDisplayName(); got != tt.want {
				t.Errorf("Song.GetDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSong_GetEffectiveArtist(t *testing.T) {
	tests := []struct {
		name string
		song Song
		want string
	}{
		{
			name: "album artist preferred",
			song: Song{Artist: "Track Artist", AlbumArtist: "Album Artist"},
			want: "Album Artist",
		},
		{
			name: "fallback to artist",
			song: Song{Artist: "Track Artist"},
			want: "Track Artist",
		},
		{
			name: "empty",
			song: Song{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.song.GetEffectiveArtist(); got != tt.want {
				t.Errorf("Song.GetEffectiveArtist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSong_TableName(t *testing.T) {
	s := Song{}
	if got := s.TableName(); got != "songs" {
		t.Errorf("Song.TableName() = %v, want songs", got)
	}
}
