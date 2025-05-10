package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDirectoryExists ensures that the given directory exists
func EnsureDirectoryExists(path string) error {
	if path == "" {
		return fmt.Errorf("empty directory path")
	}
	
	// Check if directory exists
	info, err := os.Stat(path)
	if err == nil {
		// Path exists, check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		return nil
	}
	
	// Create directory if it doesn't exist
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	
	return err
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", src, err)
	}
	
	// Ensure the destination directory exists
	dstDir := filepath.Dir(dst)@
	if err := EnsureDirectoryExists(dstDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
	}
	
	// Write the destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}
	
	return nil
} 