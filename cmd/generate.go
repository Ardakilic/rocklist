// Package cmd provides CLI commands for Rocklist
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Ardakilic/rocklist/internal/models"
	"github.com/Ardakilic/rocklist/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a playlist",
	Long: `Generate a playlist using external music data sources.

Available data sources:
  - lastfm: Last.fm
  - spotify: Spotify
  - musicbrainz: MusicBrainz

Available playlist types:
  - top_songs: Top songs by an artist
  - mixed_songs: Mix of top and similar songs
  - similar: Songs from similar artists
  - tag: Songs matching a genre/tag

Examples:
  rocklist generate --source lastfm --type top_songs --artist "Metallica"
  rocklist generate --source spotify --type tag --tag "death metal" --limit 100
  rocklist generate --source musicbrainz --type similar --artist "Iron Maiden"`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("source")
		playlistType, _ := cmd.Flags().GetString("type")
		artist, _ := cmd.Flags().GetString("artist")
		tag, _ := cmd.Flags().GetString("tag")
		limit, _ := cmd.Flags().GetInt("limit")

		runGenerate(source, playlistType, artist, tag, limit)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("source", "s", "lastfm", "Data source (lastfm, spotify, musicbrainz)")
	generateCmd.Flags().StringP("type", "t", "top_songs", "Playlist type (top_songs, mixed_songs, similar, tag)")
	generateCmd.Flags().StringP("artist", "a", "", "Artist name (required for artist-based playlists)")
	generateCmd.Flags().String("tag", "", "Tag/genre name (required for tag playlists)")
	generateCmd.Flags().IntP("limit", "l", 50, "Maximum number of songs to include")

	// API credentials
	generateCmd.Flags().String("lastfm-api-key", "", "Last.fm API key")
	generateCmd.Flags().String("lastfm-api-secret", "", "Last.fm API secret")
	generateCmd.Flags().String("spotify-client-id", "", "Spotify client ID")
	generateCmd.Flags().String("spotify-client-secret", "", "Spotify client secret")
	generateCmd.Flags().String("musicbrainz-user-agent", "", "MusicBrainz user agent")

	_ = viper.BindPFlag("lastfm_api_key", generateCmd.Flags().Lookup("lastfm-api-key"))
	_ = viper.BindPFlag("lastfm_api_secret", generateCmd.Flags().Lookup("lastfm-api-secret"))
	_ = viper.BindPFlag("spotify_client_id", generateCmd.Flags().Lookup("spotify-client-id"))
	_ = viper.BindPFlag("spotify_client_secret", generateCmd.Flags().Lookup("spotify-client-secret"))
	_ = viper.BindPFlag("musicbrainz_user_agent", generateCmd.Flags().Lookup("musicbrainz-user-agent"))
}

func runGenerate(source, playlistType, artist, tag string, limit int) {
	ctx := context.Background()

	// Validate inputs
	if playlistType == "tag" && tag == "" {
		fmt.Fprintln(os.Stderr, "Error: --tag is required for tag playlists")
		osExit(1)
		return
	}
	if (playlistType == "top_songs" || playlistType == "mixed_songs" || playlistType == "similar") && artist == "" {
		fmt.Fprintln(os.Stderr, "Error: --artist is required for this playlist type")
		osExit(1)
		return
	}

	dbPath := viper.GetString("db_path")
	svc, err := service.NewAppService(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize service: %v\n", err)
		osExit(1)
		return
	}
	defer func() { _ = svc.Close() }()

	// Set Rockbox path
	rockboxPath := viper.GetString("rockbox_path")
	if rockboxPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --rockbox-path is required")
		osExit(1)
		return
	}
	if err := svc.SetRockboxPath(rockboxPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to set Rockbox path: %v\n", err)
		osExit(1)
		return
	}

	// Configure API credentials
	config := svc.GetConfig()

	if apiKey := viper.GetString("lastfm_api_key"); apiKey != "" {
		config.LastFM.APIKey = apiKey
		config.LastFM.APISecret = viper.GetString("lastfm_api_secret")
		config.LastFM.Enabled = true
	}
	if clientID := viper.GetString("spotify_client_id"); clientID != "" {
		config.Spotify.ClientID = clientID
		config.Spotify.ClientSecret = viper.GetString("spotify_client_secret")
		config.Spotify.Enabled = true
	}
	if userAgent := viper.GetString("musicbrainz_user_agent"); userAgent != "" {
		config.MusicBrainz.UserAgent = userAgent
		config.MusicBrainz.Enabled = true
	}

	if err := svc.SaveConfig(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save config: %v\n", err)
	}

	// Check song count
	count, _ := svc.GetSongCount(ctx)
	if count == 0 {
		fmt.Fprintln(os.Stderr, "Error: No songs in database. Run 'rocklist parse' first.")
		osExit(1)
		return
	}

	fmt.Printf("Found %d songs in database\n", count)
	fmt.Printf("Generating %s playlist from %s...\n", playlistType, source)

	req := &models.PlaylistRequest{
		DataSource: models.DataSource(source),
		Type:       models.PlaylistType(playlistType),
		Artist:     artist,
		Tag:        tag,
		Limit:      limit,
	}

	playlist, err := svc.GeneratePlaylist(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate playlist: %v\n", err)
		osExit(1)
		return
	}

	fmt.Printf("\nPlaylist generated successfully!\n")
	fmt.Printf("  Name: %s\n", playlist.Name)
	fmt.Printf("  Songs: %d\n", playlist.SongCount)
	if playlist.FilePath != "" {
		fmt.Printf("  Exported to: %s\n", playlist.FilePath)
	}
}
