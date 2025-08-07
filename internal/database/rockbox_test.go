package database

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestReadIndexEntries(t *testing.T) {
	// This is a basic test structure for the readIndexEntries function
	// In a real test, you would need to create a mock database file

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "dap-root-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a rockbox directory
	rockboxDir := filepath.Join(tempDir, ".rockbox")
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rockbox directory: %v", err)
	}

	// Create a mock index file - this would be filled with actual test data
	// in a real test scenario
	mockFile := filepath.Join(rockboxDir, "database_idx.tcd")

	// Try to read the non-existent file - should return an error
	_, err = readIndexEntries(mockFile)
	if err == nil {
		t.Errorf("Expected error when reading non-existent file, got nil")
	}
}

func TestGetArtists(t *testing.T) {
	// Create a mock database
	db := &RockboxDB{
		ArtistTracks: map[string][]Track{
			"Artist1": {
				{Artist: "Artist1", Title: "Song1", Filename: "file1.mp3"},
				{Artist: "Artist1", Title: "Song2", Filename: "file2.mp3"},
			},
			"Artist2": {
				{Artist: "Artist2", Title: "Song3", Filename: "file3.mp3"},
			},
		},
	}

	// Get artists
	artists := db.GetArtists()

	// Check if we have the correct number of artists
	if len(artists) != 2 {
		t.Errorf("Expected 2 artists, got %d", len(artists))
	}

	// Check if all artists are included
	artistMap := make(map[string]bool)
	for _, artist := range artists {
		artistMap[artist] = true
	}

	expectedArtists := []string{"Artist1", "Artist2"}
	for _, artist := range expectedArtists {
		if !artistMap[artist] {
			t.Errorf("Expected artist %s not found", artist)
		}
	}
}

func TestGetTracksForArtist(t *testing.T) {
	// Create a mock database
	db := &RockboxDB{
		ArtistTracks: map[string][]Track{
			"Artist1": {
				{Artist: "Artist1", Title: "Song1", Filename: "file1.mp3"},
				{Artist: "Artist1", Title: "Song2", Filename: "file2.mp3"},
			},
			"Artist2": {
				{Artist: "Artist2", Title: "Song3", Filename: "file3.mp3"},
			},
		},
	}

	// Get tracks for Artist1
	tracks := db.GetTracksForArtist("Artist1")

	// Check if we have the correct number of tracks
	if len(tracks) != 2 {
		t.Errorf("Expected 2 tracks for Artist1, got %d", len(tracks))
	}

	// Get tracks for Artist2
	tracks = db.GetTracksForArtist("Artist2")

	// Check if we have the correct number of tracks
	if len(tracks) != 1 {
		t.Errorf("Expected 1 track for Artist2, got %d", len(tracks))
	}

	// Get tracks for non-existent artist
	tracks = db.GetTracksForArtist("NonExistentArtist")

	// Check if we have the correct number of tracks
	if len(tracks) != 0 {
		t.Errorf("Expected 0 tracks for NonExistentArtist, got %d", len(tracks))
	}
}

// Helpers for building minimal valid tagcache files
func writeSingleTagFile(t *testing.T, path string, data string, idxID int32) (int32, error) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// Tag file header
	hdr := TagcacheHeader{Magic: TagcacheMagic, DataSize: 0, EntryCount: 1}
	if err := binary.Write(f, binary.LittleEndian, hdr); err != nil {
		return 0, err
	}
	// Entry header
	tlen := int32(len(data) + 1) // include trailing NUL
	if err := binary.Write(f, binary.LittleEndian, tlen); err != nil {
		return 0, err
	}
	if err := binary.Write(f, binary.LittleEndian, idxID); err != nil {
		return 0, err
	}
	// The offset the index stores is the data start position
	// Current position after writing tlen and idxID
	pos, err := f.Seek(0, 1)
	if err != nil {
		return 0, err
	}
	offset := int32(pos)

	// Data with NUL terminator
	if _, err := f.Write([]byte(data)); err != nil {
		return 0, err
	}
	if _, err := f.Write([]byte{0}); err != nil {
		return 0, err
	}

	return offset, nil
}

func writeIndexFile(t *testing.T, path string, entries []IndexEntry) error {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	type indexHeader struct {
		Magic      int32
		DataSize   int32
		EntryCount int32
		Serial     int32
		CommitID   int32
		Dirty      int32
	}

	// Each entry is 23*4 + 4 bytes
	dataSize := int32(len(entries)) * (IndexFieldsBeforeFlags*4 + 4)
	hdr := indexHeader{Magic: TagcacheMagic, DataSize: dataSize, EntryCount: int32(len(entries))}
	if err := binary.Write(f, binary.LittleEndian, hdr); err != nil {
		return err
	}

	for _, e := range entries {
		for i := 0; i < IndexFieldsBeforeFlags; i++ {
			if err := binary.Write(f, binary.LittleEndian, e.TagSeek[i]); err != nil {
				return err
			}
		}
		if err := binary.Write(f, binary.LittleEndian, e.Flag); err != nil {
			return err
		}
	}
	return nil
}

func TestLoadDatabase_WithValidMockData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rbdb-test")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rockboxDir := filepath.Join(tempDir, ".rockbox")
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create tag files with a single entry each so data offset is predictable
	// artist (0), album (1), title (3), filename (4)
	artistOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_0.tcd"), "Artist X", 0)
	if err != nil {
		t.Fatalf("artist tag: %v", err)
	}
	albumOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_1.tcd"), "Album Y", 0)
	if err != nil {
		t.Fatalf("album tag: %v", err)
	}
	titleOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_3.tcd"), "Title Z", 0)
	if err != nil {
		t.Fatalf("title tag: %v", err)
	}
	fileOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_4.tcd"), "/music/a.mp3", 0)
	if err != nil {
		t.Fatalf("filename tag: %v", err)
	}

	// Build one valid entry and one DELETED entry
	var valid IndexEntry
	// zero-initialized TagSeek
	valid.TagSeek[0] = artistOffset
	valid.TagSeek[1] = albumOffset
	valid.TagSeek[3] = titleOffset
	valid.TagSeek[4] = fileOffset
	valid.Flag = 0 // not deleted

	var deleted IndexEntry
	deleted.TagSeek[0] = artistOffset
	deleted.TagSeek[1] = albumOffset
	deleted.TagSeek[3] = titleOffset
	deleted.TagSeek[4] = fileOffset
	deleted.Flag = 1 // FLAG_DELETED

	if err := writeIndexFile(t, filepath.Join(rockboxDir, "database_idx.tcd"), []IndexEntry{valid, deleted}); err != nil {
		t.Fatalf("index file: %v", err)
	}

	db, err := LoadDatabase(rockboxDir)
	if err != nil {
		t.Fatalf("LoadDatabase error: %v", err)
	}

	if len(db.Tracks) != 1 {
		t.Fatalf("expected 1 track (skip deleted), got %d", len(db.Tracks))
	}
	got := db.Tracks[0]
	if got.Artist != "Artist X" || got.Album != "Album Y" || got.Title != "Title Z" || got.Filename != "/music/a.mp3" {
		t.Fatalf("unexpected track: %#v", got)
	}
}

func TestLoadDatabase_TrackWithoutFilenameExcluded(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rbdb-test")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rockboxDir := filepath.Join(tempDir, ".rockbox")
	if err := os.MkdirAll(rockboxDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// artist and title only; no filename tag file present
	artistOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_0.tcd"), "Artist Only", 0)
	if err != nil {
		t.Fatalf("artist tag: %v", err)
	}
	titleOffset, err := writeSingleTagFile(t, filepath.Join(rockboxDir, "database_3.tcd"), "Title Only", 0)
	if err != nil {
		t.Fatalf("title tag: %v", err)
	}

	var entry IndexEntry
	entry.TagSeek[0] = artistOffset
	entry.TagSeek[3] = titleOffset
	entry.Flag = 0

	if err := writeIndexFile(t, filepath.Join(rockboxDir, "database_idx.tcd"), []IndexEntry{entry}); err != nil {
		t.Fatalf("index file: %v", err)
	}

	db, err := LoadDatabase(rockboxDir)
	if err != nil {
		t.Fatalf("LoadDatabase error: %v", err)
	}

	if len(db.Tracks) != 0 {
		t.Fatalf("expected 0 tracks when filename missing, got %d", len(db.Tracks))
	}
}
