package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirectoryExists(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "fileutil-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test for empty directory path
	if err := EnsureDirectoryExists(""); err == nil {
		t.Error("Expected error for empty directory path, got nil")
	}

	// Test for existing directory
	if err := EnsureDirectoryExists(tempDir); err != nil {
		t.Errorf("Expected no error for existing directory, got: %v", err)
	}

	// Test for non-existent directory
	newDir := filepath.Join(tempDir, "new_directory")
	if err := EnsureDirectoryExists(newDir); err != nil {
		t.Errorf("Expected no error for creating new directory, got: %v", err)
	}

	// Test if the directory was actually created
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", newDir)
	}

	// Test for a path that exists but is not a directory
	filePath := filepath.Join(tempDir, "test_file")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := EnsureDirectoryExists(filePath); err == nil {
		t.Error("Expected error for path that exists but is not a directory, got nil")
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "fileutil-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a source file
	srcFile := filepath.Join(tempDir, "src_file")
	content := []byte("test content")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copying to a destination file
	dstFile := filepath.Join(tempDir, "dst_file")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Errorf("Failed to copy file: %v", err)
	}

	// Verify the content of the destination file
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Expected content %q, got %q", content, dstContent)
	}

	// Test copying to a destination in a non-existent directory
	dstDir := filepath.Join(tempDir, "new_dir")
	dstFile = filepath.Join(dstDir, "dst_file")
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Errorf("Failed to copy file to new directory: %v", err)
	}

	// Verify the content of the destination file
	dstContent, err = os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Expected content %q, got %q", content, dstContent)
	}

	// Test with a non-existent source file
	nonExistentSrc := filepath.Join(tempDir, "non_existent")
	if err := CopyFile(nonExistentSrc, dstFile); err == nil {
		t.Error("Expected error for non-existent source file, got nil")
	}
} 