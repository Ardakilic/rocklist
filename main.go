package main

import (
	"embed"
	"os"

	"github.com/Ardakilic/rocklist/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Pass the embedded assets to the cmd package
	cmd.SetAssets(assets)

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
