// Package cmd provides CLI commands for Rocklist
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Rocklist %s\n", Version)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		fmt.Printf("  Build Date: %s\n", BuildDate)
		fmt.Printf("\nAuthor: Arda Kılıçdağı <arda@kilicdagi.com>\n")
		fmt.Printf("Repository: https://github.com/Ardakilic/Rocklist\n")
		fmt.Printf("License: MIT\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
