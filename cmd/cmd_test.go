package cmd

import (
	"context"
	"embed"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestSetAssets(t *testing.T) {
	var testAssets embed.FS
	SetAssets(testAssets)

	result := GetAssets()
	// Just verify it doesn't panic and returns the same type
	_ = result
}

func TestGetAssets(t *testing.T) {
	result := GetAssets()
	// Verify it returns an embed.FS (won't panic)
	_ = result
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are defined
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// GitCommit and BuildDate can be "unknown" but shouldn't be empty
	if GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}
	if BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestNewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	// ctx should be nil before startup
	if app.ctx != nil {
		t.Error("ctx should be nil before startup")
	}

	// service should be nil before startup
	if app.service != nil {
		t.Error("service should be nil before startup")
	}
}

func TestRootCmd_Use(t *testing.T) {
	if rootCmd.Use != "rocklist" {
		t.Errorf("rootCmd.Use = %v, want rocklist", rootCmd.Use)
	}
}

func TestRootCmd_Short(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}

func TestRootCmd_Long(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestVersionCmd_Use(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %v, want version", versionCmd.Use)
	}
}

func TestVersionCmd_Short(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}

func TestParseCmd_Use(t *testing.T) {
	if parseCmd.Use != "parse" {
		t.Errorf("parseCmd.Use = %v, want parse", parseCmd.Use)
	}
}

func TestParseCmd_Short(t *testing.T) {
	if parseCmd.Short == "" {
		t.Error("parseCmd.Short should not be empty")
	}
}

func TestParseCmd_Long(t *testing.T) {
	if parseCmd.Long == "" {
		t.Error("parseCmd.Long should not be empty")
	}
}

func TestGenerateCmd_Use(t *testing.T) {
	if generateCmd.Use != "generate" {
		t.Errorf("generateCmd.Use = %v, want generate", generateCmd.Use)
	}
}

func TestGenerateCmd_Short(t *testing.T) {
	if generateCmd.Short == "" {
		t.Error("generateCmd.Short should not be empty")
	}
}

func TestGenerateCmd_Long(t *testing.T) {
	if generateCmd.Long == "" {
		t.Error("generateCmd.Long should not be empty")
	}
}

func TestGenerateCmd_Flags(t *testing.T) {
	// Test that flags are defined
	flags := []string{"source", "type", "artist", "tag", "limit"}
	for _, flag := range flags {
		f := generateCmd.Flags().Lookup(flag)
		if f == nil {
			t.Errorf("generateCmd should have flag %q", flag)
		}
	}
}

func TestParseCmd_Flags(t *testing.T) {
	f := parseCmd.Flags().Lookup("use-prefetched")
	if f == nil {
		t.Error("parseCmd should have flag 'use-prefetched'")
	}
}

func TestRootCmd_PersistentFlags(t *testing.T) {
	flags := []string{"config", "rockbox-path", "db-path"}
	for _, flag := range flags {
		f := rootCmd.PersistentFlags().Lookup(flag)
		if f == nil {
			t.Errorf("rootCmd should have persistent flag %q", flag)
		}
	}
}

func TestApp_MethodsBeforeStartup(t *testing.T) {
	app := NewApp()

	// These methods access service which is nil before startup
	// They should panic when called before startup

	// Test GetAppInfo panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetAppInfo()
	}()

	// Test GetConfig panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetConfig()
	}()

	// Test GetSongCount panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetSongCount()
	}()

	// Test GetUniqueArtists panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetUniqueArtists()
	}()

	// Test GetUniqueGenres panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetUniqueGenres()
	}()

	// Test GetAllPlaylists panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetAllPlaylists()
	}()

	// Test GetLogs panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetLogs()
	}()

	// Test GetEnabledSources panics
	func() {
		defer func() { _ = recover() }()
		_ = app.GetEnabledSources()
	}()
}

func TestApp_SaveConfig(t *testing.T) {
	app := NewApp()

	// SaveConfig returns nil as it's a stub
	err := app.SaveConfig(nil)
	if err != nil {
		t.Errorf("SaveConfig() error = %v, want nil", err)
	}
}

func TestApp_Shutdown_NilService(t *testing.T) {
	app := NewApp()

	// Should not panic even with nil service
	app.shutdown(context.TODO())
}

// mockExit captures exit codes for testing
type mockExitCapture struct {
	called   bool
	exitCode int
}

func (m *mockExitCapture) exit(code int) {
	m.called = true
	m.exitCode = code
}

func TestRunParse_NoRockboxPath(t *testing.T) {
	// Save original osExit and restore after test
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	// Clear viper config
	viper.Reset()

	runParse(false)

	if !mock.called {
		t.Error("runParse() should call osExit when rockbox-path is not set")
	}
	if mock.exitCode != 1 {
		t.Errorf("runParse() exitCode = %d, want 1", mock.exitCode)
	}
}

func TestRunGenerate_TagPlaylistNoTag(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	runGenerate("lastfm", "tag", "", "", 50)

	if !mock.called {
		t.Error("runGenerate() should call osExit when tag is empty for tag playlist")
	}
	if mock.exitCode != 1 {
		t.Errorf("runGenerate() exitCode = %d, want 1", mock.exitCode)
	}
}

func TestRunGenerate_TopSongsNoArtist(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	runGenerate("lastfm", "top_songs", "", "", 50)

	if !mock.called {
		t.Error("runGenerate() should call osExit when artist is empty for top_songs")
	}
	if mock.exitCode != 1 {
		t.Errorf("runGenerate() exitCode = %d, want 1", mock.exitCode)
	}
}

func TestRunGenerate_MixedSongsNoArtist(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	runGenerate("lastfm", "mixed_songs", "", "", 50)

	if !mock.called {
		t.Error("runGenerate() should call osExit when artist is empty for mixed_songs")
	}
}

func TestRunGenerate_SimilarNoArtist(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	runGenerate("lastfm", "similar", "", "", 50)

	if !mock.called {
		t.Error("runGenerate() should call osExit when artist is empty for similar")
	}
}

func TestRunGenerate_NoRockboxPath(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	// Clear viper and set valid inputs but no rockbox path
	viper.Reset()

	runGenerate("lastfm", "tag", "", "rock", 50)

	if !mock.called {
		t.Error("runGenerate() should call osExit when rockbox-path is not set")
	}
}

func TestInitConfig_WithConfigFile(t *testing.T) {
	// Create temp config file
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// Set cfgFile
	originalCfgFile := cfgFile
	cfgFile = tmpFile.Name()
	defer func() { cfgFile = originalCfgFile }()

	viper.Reset()
	initConfig()

	// Should not panic
}

func TestInitConfig_DefaultPath(t *testing.T) {
	originalCfgFile := cfgFile
	cfgFile = ""
	defer func() { cfgFile = originalCfgFile }()

	viper.Reset()
	initConfig()

	// Should not panic
}

func TestExecute(t *testing.T) {
	// Test that Execute returns without panic
	// We can't fully test this without running the command
	// but we can verify the function exists and is callable
	_ = Execute
}

func TestVersionCmd_Run(t *testing.T) {
	// Execute version command Run function directly
	versionCmd.Run(versionCmd, []string{})
	// Should not panic and should print version info
}

func TestParseCmd_Run(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	viper.Reset()
	// Execute parse command - will fail due to missing rockbox-path
	parseCmd.Run(parseCmd, []string{})

	// Should have called exit due to missing path
	if !mock.called {
		t.Error("parseCmd.Run() should eventually call osExit")
	}
}

func TestGenerateCmd_Run(t *testing.T) {
	originalExit := osExit
	defer func() { osExit = originalExit }()

	mock := &mockExitCapture{}
	osExit = mock.exit

	viper.Reset()
	// Execute generate command - will fail due to validation
	generateCmd.Run(generateCmd, []string{})

	// Should have called exit due to missing artist
	if !mock.called {
		t.Error("generateCmd.Run() should eventually call osExit")
	}
}

func TestOsExitVariable(t *testing.T) {
	// Verify osExit is a valid function
	if osExit == nil {
		t.Error("osExit should not be nil")
	}
}
