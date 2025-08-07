package database

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// IndexFieldsBeforeFlags is the number of 32-bit fields in an index entry that precede
	// the flags field (indices 0..22). The flags field is stored separately after these.
	IndexFieldsBeforeFlags = 23

	// TagCount is the total number of tag indices we may attempt to read tag files for.
	// Rockbox defines tag files for 0..8 and 12 specifically. Others are numeric-only.
	TagCount = 24

	// TagcacheMagic is the magic/version for the Rockbox database (TCH\x10)
	TagcacheMagic = 0x54434810

	// ArtistTagIndex is the index of the artist tag in the Rockbox database
	ArtistTagIndex = 0

	// AlbumTagIndex is the index of the album tag in the Rockbox database
	AlbumTagIndex = 1

	// TitleTagIndex is the index of the title tag in the Rockbox database
	TitleTagIndex = 3

	// FilenameTagIndex is the index of the filename tag in the Rockbox database
	FilenameTagIndex = 4
)

// TagNames maps index positions to tag names for reference (not used in code but kept for documentation)
// Commenting out to avoid linting errors
/*
var tagNames = []string{
	"artist", "album", "genre", "title", "filename", "composer", "comment",
	"albumartist", "grouping", "year", "discnumber", "tracknumber",
	"canonicalartist", "bitrate", "length", "playcount", "rating", "playtime",
	"lastplayed", "commitid", "mtime", "lastelapsed", "lastoffset",
}
*/

// IndexEntry represents an entry in the Rockbox database index
type IndexEntry struct {
	TagSeek [IndexFieldsBeforeFlags]int32
	Flag    int32
}

// TagcacheHeader represents the header of a Rockbox database file
type TagcacheHeader struct {
	Magic      int32
	DataSize   int32
	EntryCount int32
}

// Track represents a track in the Rockbox database
type Track struct {
	Artist   string
	Album    string
	Title    string
	Filename string
}

// RockboxDB represents the Rockbox database
type RockboxDB struct {
	Tracks       []Track
	ArtistTracks map[string][]Track // Tracks grouped by artist
}

// LoadDatabase loads the Rockbox database from the given path
// rockboxDir should point to the .rockbox directory where database files are located
func LoadDatabase(rockboxDir string) (*RockboxDB, error) {
	// Read index file
	idxEntries, err := readIndexEntries(filepath.Join(rockboxDir, "database_idx.tcd"))
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	// Read tag files into maps keyed by byte offset of the start of the data portion in each tag file
	tagMaps := make([]map[int32]string, TagCount)
	for i := 0; i < TagCount; i++ {
		tagFile := filepath.Join(rockboxDir, fmt.Sprintf("database_%d.tcd", i))
		tagMap, err := readTagFile(tagFile)
		if err != nil {
			// Skip missing tag files
			continue
		}
		tagMaps[i] = tagMap
	}

	// Process entries
	db := &RockboxDB{
		Tracks:       make([]Track, 0, len(idxEntries)),
		ArtistTracks: make(map[string][]Track),
	}

	for _, entry := range idxEntries {
		// Skip entries flagged as deleted
		if entry.Flag&1 == 1 {
			continue
		}
		var track Track

		// Get artist
		if artistSeek := entry.TagSeek[ArtistTagIndex]; artistSeek != 0 {
			if val, ok := tagMaps[ArtistTagIndex][artistSeek]; ok {
				track.Artist = val
			}
		}

		// Get album
		if albumSeek := entry.TagSeek[AlbumTagIndex]; albumSeek != 0 {
			if val, ok := tagMaps[AlbumTagIndex][albumSeek]; ok {
				track.Album = val
			}
		}

		// Get title
		if titleSeek := entry.TagSeek[TitleTagIndex]; titleSeek != 0 {
			if val, ok := tagMaps[TitleTagIndex][titleSeek]; ok {
				track.Title = val
			}
		}

		// Get filename
		if filenameSeek := entry.TagSeek[FilenameTagIndex]; filenameSeek != 0 {
			if val, ok := tagMaps[FilenameTagIndex][filenameSeek]; ok {
				track.Filename = val
			}
		}

		// Add track to the database if it has an artist and filename
		if track.Artist != "" && track.Filename != "" {
			db.Tracks = append(db.Tracks, track)
			db.ArtistTracks[track.Artist] = append(db.ArtistTracks[track.Artist], track)
		}
	}

	return db, nil
}

// GetArtists returns all artists in the Rockbox database
func (db *RockboxDB) GetArtists() []string {
	artists := make([]string, 0, len(db.ArtistTracks))
	for artist := range db.ArtistTracks {
		artists = append(artists, artist)
	}
	return artists
}

// GetTracksForArtist returns all tracks for the given artist
func (db *RockboxDB) GetTracksForArtist(artist string) []Track {
	return db.ArtistTracks[artist]
}

// readIndexEntries reads index entries from the given path
func readIndexEntries(path string) ([]IndexEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read the full index header as defined by tagcache: Magic, DataSize, EntryCount,
	// Serial, CommitID, Dirty. We only use the first three but need to consume all
	// fields to position the file cursor correctly.
	type indexHeader struct {
		Magic      int32
		DataSize   int32
		EntryCount int32
		Serial     int32
		CommitID   int32
		Dirty      int32
	}

	var hdr indexHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Magic != TagcacheMagic {
		return nil, fmt.Errorf("invalid magic: %x", hdr.Magic)
	}
	if hdr.EntryCount < 0 {
		return nil, fmt.Errorf("invalid entry count: %d", hdr.EntryCount)
	}

	// Each entry consists of IndexFieldsBeforeFlags int32 values in TagSeek and a Flag int32 at the end.
	// We read them one by one to avoid struct padding issues.
	entries := make([]IndexEntry, 0, hdr.EntryCount)
	for i := 0; i < int(hdr.EntryCount); i++ {
		var entry IndexEntry
		// Read TagSeek
		for j := 0; j < IndexFieldsBeforeFlags; j++ {
			if err := binary.Read(f, binary.LittleEndian, &entry.TagSeek[j]); err != nil {
				return nil, err
			}
		}
		// Read Flag
		if err := binary.Read(f, binary.LittleEndian, &entry.Flag); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// readTagFile reads tag data from the given path
func readTagFile(path string) (map[int32]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Tag file header is the same first three fields
	var hdr TagcacheHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Magic != TagcacheMagic {
		return nil, fmt.Errorf("invalid magic: %x", hdr.Magic)
	}

	tagData := make(map[int32]string)
	// After reading tlen and index id, record the current data start offset
	// which is used in the index entries as the seek value for text tags.
	// Except filenames (tag 4) are not padded; other tags are padded to 4 + 8*n.
	for i := 0; i < int(hdr.EntryCount); i++ {
		var tlen, idxID int32
		if err := binary.Read(f, binary.LittleEndian, &tlen); err != nil {
			break
		}
		if err := binary.Read(f, binary.LittleEndian, &idxID); err != nil {
			break
		}

		// Record the absolute offset of the data portion
		dataStart, _ := f.Seek(0, 1)

		buf := make([]byte, tlen)
		if _, err := f.Read(buf); err != nil {
			break
		}

		// Key by the byte offset within this tag file
		tagData[int32(dataStart)] = string(bytes.Trim(buf, "\x00"))

		// For padded tags, the file may have extra padding beyond tlen that we already consumed
		// but Rockbox builds padding into tlen, so simply advancing by tlen bytes is enough here.
	}
	return tagData, nil
}
