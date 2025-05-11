package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/ardakilic/rocklist/internal/api"
	"github.com/ardakilic/rocklist/internal/config"
	"github.com/ardakilic/rocklist/internal/database"
	"github.com/ardakilic/rocklist/internal/playlist"
	"github.com/ardakilic/rocklist/pkg/util"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Create API client
	apiClient, err := api.NewAPIClient(
		cfg.APISource,
		cfg.LastFMAPIKey,
		cfg.LastFMAPISecret,
		cfg.SpotifyClientID,
		cfg.SpotifyClientSecret,
	)
	if err != nil {
		log.Fatalf("Error creating API client: %v", err)
	}

	// Load Rockbox database from .rockbox directory inside DAP root
	rockboxDir := filepath.Join(cfg.DapRootPath, ".rockbox")
	db, err := database.LoadDatabase(rockboxDir)
	if err != nil {
		log.Fatalf("Error loading Rockbox database: %v", err)
	}

	// Ensure playlist directory exists
	if err := util.EnsureDirectoryExists(cfg.PlaylistPath); err != nil {
		log.Fatalf("Error creating playlist directory: %v", err)
	}

	// Determine which artists to process
	var artists []string
	if len(cfg.Artists) > 0 {
		artists = cfg.Artists
	} else {
		artists = db.GetArtists()
	}

	if len(cfg.Artists) > 0 {
		fmt.Printf("Processing %d specified artists\n", len(artists))
	} else {
		fmt.Printf("Found %d artists in the Rockbox database\n", len(artists))
	}
	fmt.Printf("Using %s API for top tracks\n", cfg.APISource)
	fmt.Printf("Generating playlists with up to %d tracks per artist\n", cfg.MaxTracksPerArtist)

	// Create playlist generator
	generator := playlist.NewGenerator(db, apiClient, cfg.PlaylistPath, cfg.MaxTracksPerArtist)

	// Process artists in parallel with a limit of 5 concurrent operations
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	// Keep track of successful and failed artists
	var (
		successMutex sync.Mutex
		failMutex    sync.Mutex
		successful   []string
		failed       map[string]string = make(map[string]string)
	)

	for _, artist := range artists {
		wg.Add(1)
		go func(artist string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			fmt.Printf("Processing artist: %s\n", artist)
			if err := generator.GeneratePlaylistForArtist(artist); err != nil {
				fmt.Printf("Error generating playlist for %s: %v\n", artist, err)
				failMutex.Lock()
				failed[artist] = err.Error()
				failMutex.Unlock()
				return
			}

			successMutex.Lock()
			successful = append(successful, artist)
			successMutex.Unlock()
		}(artist)
	}

	wg.Wait()

	// Print summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("Generated playlists for %d artists\n", len(successful))
	fmt.Printf("Failed to generate playlists for %d artists\n", len(failed))

	if len(failed) > 0 {
		fmt.Printf("\nFailed artists:\n")
		for artist, reason := range failed {
			fmt.Printf("- %s: %s\n", artist, reason)
		}
	}

	// Copy playlists to the Rockbox playlists directory if needed
	playlistDir := filepath.Join(cfg.DapRootPath, "Playlists")
	if cfg.PlaylistPath != playlistDir {
		fmt.Printf("\nCopying playlists to Rockbox playlists directory: %s\n", playlistDir)
		if err := copyPlaylists(cfg.PlaylistPath, playlistDir); err != nil {
			fmt.Printf("Error copying playlists: %v\n", err)
		}
	}
}

// copyPlaylists copies playlist files from src to dst
func copyPlaylists(src, dst string) error {
	// Ensure the destination directory exists
	if err := util.EnsureDirectoryExists(dst); err != nil {
		return err
	}

	// Read playlist files
	files, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read playlist directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".m3u" {
			srcFile := filepath.Join(src, file.Name())
			dstFile := filepath.Join(dst, file.Name())
			if err := util.CopyFile(srcFile, dstFile); err != nil {
				return fmt.Errorf("failed to copy playlist %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}
