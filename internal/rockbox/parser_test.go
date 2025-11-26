package rockbox

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ardakilic/rocklist/internal/models"
)

// mockLogger implements Logger for testing
type mockLogger struct {
	infos  []string
	errors []string
	debugs []string
}

func (l *mockLogger) Info(msg string, args ...interface{})  { l.infos = append(l.infos, msg) }
func (l *mockLogger) Error(msg string, args ...interface{}) { l.errors = append(l.errors, msg) }
func (l *mockLogger) Debug(msg string, args ...interface{}) { l.debugs = append(l.debugs, msg) }

func TestNewParser(t *testing.T) {
	parser := NewParser("/test/path", nil)

	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}

	if parser.rockboxPath != "/test/path" {
		t.Errorf("NewParser() rockboxPath = %v, want %v", parser.rockboxPath, "/test/path")
	}

	if parser.logger == nil {
		t.Error("NewParser() should set default logger when nil")
	}
}

func TestNewParser_WithLogger(t *testing.T) {
	logger := &mockLogger{}
	parser := NewParser("/test/path", logger)

	if parser.logger != logger {
		t.Error("NewParser() should use provided logger")
	}
}

func TestParser_SetPath(t *testing.T) {
	parser := NewParser("/initial/path", nil)
	parser.SetPath("/new/path")

	if parser.GetPath() != "/new/path" {
		t.Errorf("SetPath() path = %v, want %v", parser.GetPath(), "/new/path")
	}
}

func TestParser_GetPath(t *testing.T) {
	parser := NewParser("/test/path", nil)

	if parser.GetPath() != "/test/path" {
		t.Errorf("GetPath() = %v, want %v", parser.GetPath(), "/test/path")
	}
}

func TestParser_GetStatus(t *testing.T) {
	parser := NewParser("/test/path", nil)
	status := parser.GetStatus()

	if status == nil {
		t.Fatal("GetStatus() returned nil")
	}

	if status.InProgress {
		t.Error("Initial status should not be in progress")
	}
}

func TestParser_ValidatePath_Empty(t *testing.T) {
	parser := NewParser("", nil)
	err := parser.ValidatePath()

	if err != models.ErrRockboxPathNotSet {
		t.Errorf("ValidatePath() error = %v, want ErrRockboxPathNotSet", err)
	}
}

func TestParser_ValidatePath_Invalid(t *testing.T) {
	parser := NewParser("/nonexistent/path", nil)
	err := parser.ValidatePath()

	if err != models.ErrRockboxPathInvalid {
		t.Errorf("ValidatePath() error = %v, want ErrRockboxPathInvalid", err)
	}
}

func TestParser_ValidatePath_NoDB(t *testing.T) {
	// Create temp dir with .rockbox but no database
	tmpDir := t.TempDir()
	rockboxDir := filepath.Join(tmpDir, TagCacheDir)
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	parser := NewParser(tmpDir, nil)
	err := parser.ValidatePath()

	if err != models.ErrRockboxDatabaseNotFound {
		t.Errorf("ValidatePath() error = %v, want ErrRockboxDatabaseNotFound", err)
	}
}

func TestParser_ValidatePath_Valid(t *testing.T) {
	// Create temp dir with .rockbox and database file
	tmpDir := t.TempDir()
	rockboxDir := filepath.Join(tmpDir, TagCacheDir)
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	dbFile := filepath.Join(rockboxDir, DatabaseFile)
	if err := os.WriteFile(dbFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser(tmpDir, nil)
	err := parser.ValidatePath()

	if err != nil {
		t.Errorf("ValidatePath() error = %v, want nil", err)
	}
}

func TestParser_Parse_InvalidPath(t *testing.T) {
	parser := NewParser("", nil)
	_, err := parser.Parse(context.Background())

	if err != models.ErrRockboxPathNotSet {
		t.Errorf("Parse() error = %v, want ErrRockboxPathNotSet", err)
	}
}

func TestParser_Parse_InProgress(t *testing.T) {
	tmpDir := t.TempDir()
	rockboxDir := filepath.Join(tmpDir, TagCacheDir)
	_ = os.MkdirAll(rockboxDir, 0755)
	_ = os.WriteFile(filepath.Join(rockboxDir, DatabaseFile), []byte("test"), 0644)

	parser := NewParser(tmpDir, nil)

	// Simulate in-progress state
	parser.status.InProgress = true

	_, err := parser.Parse(context.Background())

	if err != models.ErrParseInProgress {
		t.Errorf("Parse() error = %v, want ErrParseInProgress", err)
	}
}

func TestParser_generateRockboxID(t *testing.T) {
	parser := NewParser("/test", nil)

	song1 := &models.Song{Path: "/Music/song1.mp3"}
	song2 := &models.Song{Path: "/Music/song2.mp3"}
	song3 := &models.Song{Path: "/Music/song1.mp3"}

	id1 := parser.generateRockboxID(song1)
	id2 := parser.generateRockboxID(song2)
	id3 := parser.generateRockboxID(song3)

	if id1 == "" {
		t.Error("generateRockboxID() should not return empty string")
	}

	if id1 == id2 {
		t.Error("Different paths should generate different IDs")
	}

	if id1 != id3 {
		t.Error("Same paths should generate same IDs")
	}
}

func TestParser_GetPlaylistPath(t *testing.T) {
	parser := NewParser("/test/device", nil)
	path := parser.GetPlaylistPath()

	expected := filepath.Join("/test/device", "Playlists")
	if path != expected {
		t.Errorf("GetPlaylistPath() = %v, want %v", path, expected)
	}
}

func TestParser_EnsurePlaylistDir(t *testing.T) {
	tmpDir := t.TempDir()
	parser := NewParser(tmpDir, nil)

	err := parser.EnsurePlaylistDir()
	if err != nil {
		t.Fatalf("EnsurePlaylistDir() error = %v", err)
	}

	playlistDir := parser.GetPlaylistPath()
	if _, err := os.Stat(playlistDir); os.IsNotExist(err) {
		t.Error("EnsurePlaylistDir() did not create playlist directory")
	}
}

func TestDefaultLogger(t *testing.T) {
	logger := &DefaultLogger{}

	// These should not panic
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")
}

func TestTagType_Constants(t *testing.T) {
	// Verify tag type constants are defined
	tags := []TagType{
		TagArtist, TagAlbum, TagGenre, TagTitle,
		TagFilename, TagComposer, TagComment, TagAlbumArtist,
		TagGrouping, TagYear, TagDiscNumber, TagTrackNumber,
		TagBitrate, TagLength, TagPlayCount, TagRating,
		TagPlayTime, TagLastPlayed, TagCommitID, TagMTime,
		TagLastElapsed, TagLastOffset, TagTagCount,
	}

	for i, tag := range tags {
		if int(tag) != i {
			t.Errorf("Tag constant at index %d has unexpected value %d", i, tag)
		}
	}
}

func TestConstants(t *testing.T) {
	if TagCacheDir != ".rockbox" {
		t.Errorf("TagCacheDir = %v, want .rockbox", TagCacheDir)
	}

	if DatabaseFile != "database_idx.tcd" {
		t.Errorf("DatabaseFile = %v, want database_idx.tcd", DatabaseFile)
	}

	if TagCacheMagic != 0x54434801 {
		t.Errorf("TagCacheMagic = %x, want 0x54434801", TagCacheMagic)
	}
}

func TestNumericTagData_Fields(t *testing.T) {
	data := &NumericTagData{
		Year:        2024,
		DiscNumber:  1,
		TrackNumber: 5,
		Bitrate:     320,
		Length:      240,
		PlayCount:   10,
		Rating:      5,
		LastPlayed:  1700000000,
	}

	if data.Year != 2024 {
		t.Errorf("Year = %v, want 2024", data.Year)
	}
	if data.TrackNumber != 5 {
		t.Errorf("TrackNumber = %v, want 5", data.TrackNumber)
	}
}

func TestParser_ScanFilesystem(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some audio files
	musicDir := filepath.Join(tmpDir, "Music")
	_ = os.MkdirAll(musicDir, 0755)

	testFiles := []string{
		"Artist - Song.mp3",
		"Another Artist - Another Song.flac",
		"Simple Title.ogg",
		"not_audio.txt",
	}

	for _, f := range testFiles {
		_ = os.WriteFile(filepath.Join(musicDir, f), []byte("test"), 0644)
	}

	// Create hidden dir that should be skipped
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	_ = os.MkdirAll(hiddenDir, 0755)
	_ = os.WriteFile(filepath.Join(hiddenDir, "hidden.mp3"), []byte("test"), 0644)

	logger := &mockLogger{}
	parser := NewParser(tmpDir, logger)

	songs, err := parser.scanFilesystem(context.Background())
	if err != nil {
		t.Fatalf("scanFilesystem() error = %v", err)
	}

	// Should find 3 audio files (not the txt or hidden)
	if len(songs) != 3 {
		t.Errorf("scanFilesystem() found %d songs, want 3", len(songs))
	}

	// Check metadata extraction from filename
	for _, song := range songs {
		if song.Path == "" {
			t.Error("Song path should not be empty")
		}
		if song.RockboxID == "" {
			t.Error("Song RockboxID should not be empty")
		}
	}
}

func TestParser_ScanFilesystem_Cancelled(t *testing.T) {
	tmpDir := t.TempDir()
	musicDir := filepath.Join(tmpDir, "Music")
	_ = os.MkdirAll(musicDir, 0755)
	_ = os.WriteFile(filepath.Join(musicDir, "test.mp3"), []byte("test"), 0644)

	parser := NewParser(tmpDir, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := parser.scanFilesystem(ctx)
	// Error is wrapped, so check if it's non-nil
	if err == nil {
		t.Error("scanFilesystem() with cancelled context should return error")
	}
}

func TestParser_ReadTagFile_NotFound(t *testing.T) {
	parser := NewParser("/test", nil)

	_, err := parser.readTagFile("/nonexistent/file.tcd")
	if err == nil {
		t.Error("readTagFile() should return error for nonexistent file")
	}
}

func TestParser_ReadTagFile_InvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	tagFile := filepath.Join(tmpDir, "test.tcd")

	// Write file with invalid magic
	data := make([]byte, 12)
	binary.LittleEndian.PutUint32(data[0:4], 0x12345678) // Invalid magic
	_ = os.WriteFile(tagFile, data, 0644)

	parser := NewParser(tmpDir, nil)
	_, err := parser.readTagFile(tagFile)

	if err == nil {
		t.Error("readTagFile() should return error for invalid magic")
	}
}

func TestParser_ReadTagFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	tagFile := filepath.Join(tmpDir, "test.tcd")

	// Create valid tag file
	header := make([]byte, 12)
	binary.LittleEndian.PutUint32(header[0:4], TagCacheMagic)
	binary.LittleEndian.PutUint32(header[4:8], 12) // data size
	binary.LittleEndian.PutUint32(header[8:12], 2) // entry count

	data := []byte("test1\x00test2\x00")

	fullData := append(header, data...)
	_ = os.WriteFile(tagFile, fullData, 0644)

	parser := NewParser(tmpDir, nil)
	result, err := parser.readTagFile(tagFile)

	if err != nil {
		t.Fatalf("readTagFile() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("readTagFile() returned %d entries, want 2", len(result))
	}

	if result[0] != "test1" {
		t.Errorf("readTagFile() entry 0 = %v, want test1", result[0])
	}
}

func TestParser_ReadNumericTags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "numeric.tcd")

	// Create file with numeric data
	data := make([]byte, 64) // 2 entries

	// Entry 0
	binary.LittleEndian.PutUint16(data[0:2], 2024)    // Year
	data[2] = 1                                       // DiscNumber
	data[3] = 5                                       // TrackNumber
	binary.LittleEndian.PutUint16(data[4:6], 320)     // Bitrate
	binary.LittleEndian.PutUint32(data[6:10], 240000) // Length
	binary.LittleEndian.PutUint32(data[10:14], 10)    // PlayCount
	data[14] = 5                                      // Rating

	_ = os.WriteFile(testFile, data, 0644)

	file, _ := os.Open(testFile)
	defer func() { _ = file.Close() }()

	parser := NewParser(tmpDir, nil)
	result, err := parser.readNumericTags(file, 2)

	if err != nil {
		t.Fatalf("readNumericTags() error = %v", err)
	}

	if result[0].Year != 2024 {
		t.Errorf("Year = %v, want 2024", result[0].Year)
	}
	if result[0].TrackNumber != 5 {
		t.Errorf("TrackNumber = %v, want 5", result[0].TrackNumber)
	}
}

func TestParser_ParseWithFilesystemFallback(t *testing.T) {
	tmpDir := t.TempDir()
	rockboxDir := filepath.Join(tmpDir, TagCacheDir)
	_ = os.MkdirAll(rockboxDir, 0755)

	// Create invalid database file to trigger fallback
	dbFile := filepath.Join(rockboxDir, DatabaseFile)
	_ = os.WriteFile(dbFile, []byte("invalid"), 0644)

	// Create some audio files
	musicDir := filepath.Join(tmpDir, "Music")
	_ = os.MkdirAll(musicDir, 0755)
	_ = os.WriteFile(filepath.Join(musicDir, "test.mp3"), []byte("test"), 0644)

	logger := &mockLogger{}
	parser := NewParser(tmpDir, logger)

	songs, err := parser.Parse(context.Background())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have found audio files via filesystem scan
	if len(songs) != 1 {
		t.Errorf("Parse() found %d songs, want 1", len(songs))
	}

	// Verify status was updated
	status := parser.GetStatus()
	if status.InProgress {
		t.Error("Status should not be in progress after parse")
	}
}
