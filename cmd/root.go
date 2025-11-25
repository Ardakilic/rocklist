// Package cmd provides CLI commands for Rocklist
package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	assets  embed.FS
)

// SetAssets sets the embedded frontend assets
func SetAssets(a embed.FS) {
	assets = a
}

// GetAssets returns the embedded frontend assets
func GetAssets() embed.FS {
	return assets
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rocklist",
	Short: "Rocklist - Playlist generator for Rockbox devices",
	Long: `Rocklist is a tool for creating playlists for Rockbox firmware devices.

It parses the Rockbox database and creates playlists using external music
data sources like Last.fm, Spotify, and MusicBrainz.

Playlist types:
  - Top songs: Most popular songs by an artist
  - Mixed songs: A mix of top and similar songs
  - Similar songs: Songs from similar artists
  - Tag/Genre radio: Songs matching a specific genre/tag

Run without arguments to start the GUI, or use subcommands for CLI mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is given, run the GUI
		runGUI()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rocklist/config.yaml)")
	rootCmd.PersistentFlags().String("rockbox-path", "", "Path to Rockbox device root")
	rootCmd.PersistentFlags().String("db-path", "", "Path to database file")

	viper.BindPFlag("rockbox_path", rootCmd.PersistentFlags().Lookup("rockbox-path"))
	viper.BindPFlag("db_path", rootCmd.PersistentFlags().Lookup("db-path"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		configDir := home + "/.rocklist"
		os.MkdirAll(configDir, 0755)

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("ROCKLIST")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
