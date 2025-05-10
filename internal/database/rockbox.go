package database

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// TagCount is the number of tags in the Rockbox database
	TagCount = 22

	// TagcacheMagic is the magic number for the Rockbox database
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
	TagSeek [TagCount]int32
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
func LoadDatabase(rockboxPath string) (*RockboxDB, error) {
	// Read index file
	idxEntries, err := readIndexEntries(filepath.Join(rockboxPath, "database_idx.tcd"))
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	// Read tag files
	tagMaps := make([]map[int32]string, TagCount)
	for i := 0; i < TagCount; i++ {
		tagFile := filepath.Join(rockboxPath, fmt.Sprintf("database_%d.tcd", i))
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

	var hdr TagcacheHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Magic != TagcacheMagic {
		return nil, fmt.Errorf("invalid magic: %x", hdr.Magic)
	}

	entries := make([]IndexEntry, hdr.EntryCount)
	if err := binary.Read(f, binary.LittleEndian, &entries); err != nil {
		return nil, err
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

	var hdr TagcacheHeader
	if err := binary.Read(f, binary.LittleEndian, &hdr); err != nil {
		return nil, err
	}
	if hdr.Magic != TagcacheMagic {
		return nil, fmt.Errorf("invalid magic: %x", hdr.Magic)
	}

	tagData := make(map[int32]string)
	for i := 0; i < int(hdr.EntryCount); i++ {
		var tlen, idxID int32
		if err := binary.Read(f, binary.LittleEndian, &tlen); err != nil {
			break
		}
		if err := binary.Read(f, binary.LittleEndian, &idxID); err != nil {
			break
		}

		buf := make([]byte, tlen)
		if _, err := f.Read(buf); err != nil {
			break
		}
		tagData[idxID] = string(bytes.Trim(buf, "\x00"))
	}
	return tagData, nil
}
