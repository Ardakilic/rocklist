// Package cmd provides CLI commands for Rocklist
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Ardakilic/rocklist/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse the Rockbox database",
	Long: `Parse the Rockbox database and store song information in the local database.

This command reads the TagCache files from the Rockbox device and extracts
metadata for all songs. The information is stored locally for playlist generation.

Example:
  rocklist parse --rockbox-path /Volumes/IPOD`,
	Run: func(cmd *cobra.Command, args []string) {
		usePrefetched, _ := cmd.Flags().GetBool("use-prefetched")
		runParse(usePrefetched)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.Flags().Bool("use-prefetched", false, "Use previously fetched data instead of parsing")
}

func runParse(usePrefetched bool) {
	ctx := context.Background()

	dbPath := viper.GetString("db_path")
	svc, err := service.NewAppService(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize service: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	rockboxPath := viper.GetString("rockbox_path")
	if rockboxPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --rockbox-path is required")
		os.Exit(1)
	}

	if err := svc.SetRockboxPath(rockboxPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to set Rockbox path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsing Rockbox database from: %s\n", rockboxPath)

	if err := svc.ParseRockboxDatabase(ctx, usePrefetched); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to parse database: %v\n", err)
		os.Exit(1)
	}

	count, _ := svc.GetSongCount(ctx)
	fmt.Printf("Successfully parsed %d songs\n", count)
}
