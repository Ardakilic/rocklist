// Package cmd provides CLI commands for Rocklist
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Ardakilic/rocklist/internal/models"
	"github.com/Ardakilic/rocklist/internal/service"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main Wails application
type App struct {
	ctx     context.Context
	service *service.AppService
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dbPath := viper.GetString("db_path")
	svc, err := service.NewAppService(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app service: %v\n", err)
		return
	}
	a.service = svc

	// Set Rockbox path if provided
	if path := viper.GetString("rockbox_path"); path != "" {
		_ = a.service.SetRockboxPath(path)
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.service != nil {
		_ = a.service.Close()
	}
}

// GetAppInfo returns application information
func (a *App) GetAppInfo() map[string]string {
	return a.service.GetAppInfo()
}

// GetConfig returns the current configuration
func (a *App) GetConfig() interface{} {
	return a.service.GetConfig()
}

// SaveConfig saves the configuration
func (a *App) SaveConfig(config map[string]interface{}) error {
	// Convert map to AppConfig
	// This is a simplified version - in production you'd want proper marshaling
	return nil // Handled by individual setters
}

// SetRockboxPath sets the Rockbox device path
func (a *App) SetRockboxPath(path string) error {
	return a.service.SetRockboxPath(path)
}

// SelectDirectory opens a directory picker dialog and returns the selected path
func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Rockbox Device",
	})
}

// SetLastFMCredentials sets Last.fm API credentials
func (a *App) SetLastFMCredentials(apiKey, apiSecret string, enabled bool) error {
	config := a.service.GetConfig()
	config.LastFM.APIKey = apiKey
	config.LastFM.APISecret = apiSecret
	config.LastFM.Enabled = enabled
	return a.service.SaveConfig(a.ctx, config)
}

// SetSpotifyCredentials sets Spotify API credentials
func (a *App) SetSpotifyCredentials(clientID, clientSecret string, enabled bool) error {
	config := a.service.GetConfig()
	config.Spotify.ClientID = clientID
	config.Spotify.ClientSecret = clientSecret
	config.Spotify.Enabled = enabled
	return a.service.SaveConfig(a.ctx, config)
}

// SetMusicBrainzCredentials sets MusicBrainz settings
func (a *App) SetMusicBrainzCredentials(userAgent string, enabled bool) error {
	config := a.service.GetConfig()
	config.MusicBrainz.UserAgent = userAgent
	config.MusicBrainz.Enabled = enabled
	return a.service.SaveConfig(a.ctx, config)
}

// ParseDatabase parses the Rockbox database
func (a *App) ParseDatabase(usePrefetched bool) error {
	return a.service.ParseRockboxDatabase(a.ctx, usePrefetched)
}

// GetParseStatus returns the current parse status
func (a *App) GetParseStatus() interface{} {
	return a.service.GetParseStatus()
}

// GetLastParsedAt returns the last parsed timestamp
func (a *App) GetLastParsedAt() interface{} {
	t, _ := a.service.GetLastParsedAt(a.ctx)
	return t
}

// GeneratePlaylist generates a playlist
// useAlbumArtist: when true, prioritizes album artist field for matching (with fallback to artist if empty)
func (a *App) GeneratePlaylist(dataSource, playlistType, artist, tag string, limit int, useAlbumArtist bool) (interface{}, error) {
	req := &models.PlaylistRequest{
		DataSource:     models.DataSource(dataSource),
		Type:           models.PlaylistType(playlistType),
		Artist:         artist,
		Tag:            tag,
		Limit:          limit,
		UseAlbumArtist: useAlbumArtist,
	}
	return a.service.GeneratePlaylist(a.ctx, req)
}

// GetSongCount returns the number of songs in the database
func (a *App) GetSongCount() int64 {
	count, _ := a.service.GetSongCount(a.ctx)
	return count
}

// GetUniqueArtists returns all unique artists
func (a *App) GetUniqueArtists() []string {
	artists, _ := a.service.GetUniqueArtists(a.ctx)
	return artists
}

// GetUniqueGenres returns all unique genres
func (a *App) GetUniqueGenres() []string {
	genres, _ := a.service.GetUniqueGenres(a.ctx)
	return genres
}

// GetAllPlaylists returns all playlists
func (a *App) GetAllPlaylists() interface{} {
	playlists, _ := a.service.GetAllPlaylists(a.ctx)
	return playlists
}

// DeletePlaylist deletes a playlist
func (a *App) DeletePlaylist(id uint) error {
	return a.service.DeletePlaylist(a.ctx, id)
}

// WipeData wipes all pre-fetched data
func (a *App) WipeData() error {
	return a.service.WipeData(a.ctx)
}

// GetLogs returns the current logs
func (a *App) GetLogs() interface{} {
	return a.service.GetLogs()
}

// ClearLogs clears all logs
func (a *App) ClearLogs() {
	a.service.ClearLogs()
}

// GetEnabledSources returns enabled data sources
func (a *App) GetEnabledSources() []string {
	sources := a.service.GetEnabledSources()
	result := make([]string, len(sources))
	for i, s := range sources {
		result[i] = string(s)
	}
	return result
}

// runGUI starts the Wails GUI application
func runGUI() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Rocklist",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			About: &mac.AboutInfo{
				Title:   "Rocklist",
				Message: "Playlist generator for Rockbox devices\n\nAuthor: Arda Kılıçdağı\nEmail: arda@kilicdagi.com\n\nhttps://github.com/Ardakilic/Rocklist",
			},
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
		return
	}
}
