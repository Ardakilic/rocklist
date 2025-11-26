// Package rockbox provides functionality to parse Rockbox database files
package rockbox

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Ardakilic/rocklist/internal/models"
)

const (
	// TagCacheDir is the directory containing Rockbox database files
	TagCacheDir = ".rockbox"
	// DatabaseFile is the main database index file
	DatabaseFile = "database_idx.tcd"
	// TagCacheMagic is the magic number for TagCache files
	TagCacheMagic = 0x54434801 // "TCH\x01"
)

// TagType represents the type of a tag in the Rockbox database
type TagType int

const (
	TagArtist TagType = iota
	TagAlbum
	TagGenre
	TagTitle
	TagFilename
	TagComposer
	TagComment
	TagAlbumArtist
	TagGrouping
	TagYear
	TagDiscNumber
	TagTrackNumber
	TagBitrate
	TagLength
	TagPlayCount
	TagRating
	TagPlayTime
	TagLastPlayed
	TagCommitID
	TagMTime
	TagLastElapsed
	TagLastOffset
	TagTagCount
)

// Parser handles parsing of Rockbox database files
type Parser struct {
	rockboxPath string
	logger      Logger
	mu          sync.RWMutex
	status      *models.ParseStatus
}

// Logger interface for logging parse operations
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultLogger provides a simple console logger
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

// NewParser creates a new Rockbox database parser
func NewParser(rockboxPath string, logger Logger) *Parser {
	if logger == nil {
		logger = &DefaultLogger{}
	}
	return &Parser{
		rockboxPath: rockboxPath,
		logger:      logger,
		status:      &models.ParseStatus{},
	}
}

// SetPath sets the Rockbox path
func (p *Parser) SetPath(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rockboxPath = path
}

// GetPath returns the current Rockbox path
func (p *Parser) GetPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.rockboxPath
}

// GetStatus returns the current parse status
func (p *Parser) GetStatus() *models.ParseStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	statusCopy := *p.status
	return &statusCopy
}

// ValidatePath checks if the Rockbox path is valid
func (p *Parser) ValidatePath() error {
	if p.rockboxPath == "" {
		return models.ErrRockboxPathNotSet
	}

	rockboxDir := filepath.Join(p.rockboxPath, TagCacheDir)
	if _, err := os.Stat(rockboxDir); os.IsNotExist(err) {
		return models.ErrRockboxPathInvalid
	}

	dbFile := filepath.Join(rockboxDir, DatabaseFile)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return models.ErrRockboxDatabaseNotFound
	}

	return nil
}

// Parse parses the Rockbox database and returns all songs
func (p *Parser) Parse(ctx context.Context) ([]*models.Song, error) {
	if err := p.ValidatePath(); err != nil {
		return nil, err
	}

	p.mu.Lock()
	if p.status.InProgress {
		p.mu.Unlock()
		return nil, models.ErrParseInProgress
	}
	now := time.Now()
	p.status = &models.ParseStatus{
		InProgress: true,
		StartedAt:  &now,
	}
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.status.InProgress = false
		completedAt := time.Now()
		p.status.CompletedAt = &completedAt
		p.mu.Unlock()
	}()

	p.logger.Info("Starting Rockbox database parse from: %s", p.rockboxPath)

	// Parse the database files
	songs, err := p.parseDatabase(ctx)
	if err != nil {
		p.mu.Lock()
		p.status.LastError = err.Error()
		p.status.ErrorCount++
		p.mu.Unlock()
		return nil, err
	}

	p.mu.Lock()
	p.status.TotalSongs = len(songs)
	p.status.ProcessedSongs = len(songs)
	p.mu.Unlock()

	p.logger.Info("Successfully parsed %d songs", len(songs))
	return songs, nil
}

// parseDatabase reads and parses the Rockbox TagCache database files
func (p *Parser) parseDatabase(ctx context.Context) ([]*models.Song, error) {
	rockboxDir := filepath.Join(p.rockboxPath, TagCacheDir)

	// Read all tag cache files
	entries, err := p.readTagCacheEntries(ctx, rockboxDir)
	if err != nil {
		// If we can't read the tag cache, try scanning the filesystem
		p.logger.Info("TagCache not readable, falling back to filesystem scan")
		return p.scanFilesystem(ctx)
	}

	return entries, nil
}

// readTagCacheEntries reads entries from TagCache files
func (p *Parser) readTagCacheEntries(ctx context.Context, rockboxDir string) ([]*models.Song, error) {
	dbPath := filepath.Join(rockboxDir, DatabaseFile)

	file, err := os.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Read header
	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, fmt.Errorf("failed to read database header: %w", err)
	}

	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic != TagCacheMagic {
		// Try big endian
		magic = binary.BigEndian.Uint32(header[0:4])
		if magic != TagCacheMagic {
			return nil, fmt.Errorf("invalid TagCache magic number: %x", magic)
		}
	}

	dataSize := binary.LittleEndian.Uint32(header[4:8])
	entryCount := binary.LittleEndian.Uint32(header[8:12])

	p.logger.Info("TagCache: data_size=%d, entry_count=%d", dataSize, entryCount)

	// Read tag files
	songs := make([]*models.Song, 0, entryCount)
	tagFiles := []string{
		"database_0.tcd", // Artist
		"database_1.tcd", // Album
		"database_2.tcd", // Genre
		"database_3.tcd", // Title
		"database_4.tcd", // Filename
		"database_5.tcd", // Composer
		"database_6.tcd", // Comment
		"database_7.tcd", // Album Artist
	}

	tagData := make(map[int]map[int]string)
	for i, tagFile := range tagFiles {
		data, err := p.readTagFile(filepath.Join(rockboxDir, tagFile))
		if err != nil {
			p.logger.Debug("Could not read tag file %s: %v", tagFile, err)
			continue
		}
		tagData[i] = data
	}

	// Read numeric tags from index
	numericData, err := p.readNumericTags(file, int(entryCount))
	if err != nil {
		p.logger.Debug("Could not read numeric tags: %v", err)
	}

	// Build song entries
	for i := 0; i < int(entryCount); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		song := &models.Song{}

		// Get string tags
		if artists, ok := tagData[0]; ok {
			if artist, ok := artists[i]; ok {
				song.Artist = artist
			}
		}
		if albums, ok := tagData[1]; ok {
			if album, ok := albums[i]; ok {
				song.Album = album
			}
		}
		if genres, ok := tagData[2]; ok {
			if genre, ok := genres[i]; ok {
				song.Genre = genre
			}
		}
		if titles, ok := tagData[3]; ok {
			if title, ok := titles[i]; ok {
				song.Title = title
			}
		}
		if filenames, ok := tagData[4]; ok {
			if filename, ok := filenames[i]; ok {
				song.Path = filename
			}
		}
		if albumArtists, ok := tagData[7]; ok {
			if albumArtist, ok := albumArtists[i]; ok {
				song.AlbumArtist = albumArtist
			}
		}

		// Get numeric tags
		if numericData != nil {
			if numeric, ok := numericData[i]; ok {
				song.Year = numeric.Year
				song.TrackNumber = numeric.TrackNumber
				song.DiscNumber = numeric.DiscNumber
				song.Duration = numeric.Length
				song.Bitrate = numeric.Bitrate
				song.PlayCount = numeric.PlayCount
				song.Rating = numeric.Rating
			}
		}

		// Generate Rockbox ID
		song.RockboxID = p.generateRockboxID(song)

		// Only add songs with valid paths
		if song.Path != "" {
			songs = append(songs, song)
		}

		// Update progress
		if i%100 == 0 {
			p.mu.Lock()
			p.status.ProcessedSongs = i
			p.mu.Unlock()
		}
	}

	return songs, nil
}

// NumericTagData holds numeric tag values for a song
type NumericTagData struct {
	Year        int
	DiscNumber  int
	TrackNumber int
	Bitrate     int
	Length      int
	PlayCount   int
	Rating      int
	LastPlayed  int64
}

// readTagFile reads a single tag file and returns a map of index -> value
func (p *Parser) readTagFile(path string) (map[int]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// Read header
	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return nil, err
	}

	magic := binary.LittleEndian.Uint32(header[0:4])
	if magic != TagCacheMagic {
		return nil, fmt.Errorf("invalid magic number")
	}

	dataSize := binary.LittleEndian.Uint32(header[4:8])
	entryCount := binary.LittleEndian.Uint32(header[8:12])

	result := make(map[int]string, entryCount)

	// Read string data
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, err
	}

	// Parse null-terminated strings
	idx := 0
	offset := 0
	for offset < len(data) && idx < int(entryCount) {
		end := offset
		for end < len(data) && data[end] != 0 {
			end++
		}
		if end > offset {
			result[idx] = string(data[offset:end])
		}
		idx++
		offset = end + 1
	}

	return result, nil
}

// readNumericTags reads numeric tags from the index file
func (p *Parser) readNumericTags(file *os.File, entryCount int) (map[int]*NumericTagData, error) {
	result := make(map[int]*NumericTagData, entryCount)

	// Skip to numeric data section (after header)
	// Each entry has fixed-size numeric fields
	entrySize := 32 // Approximate size per entry

	for i := 0; i < entryCount; i++ {
		data := make([]byte, entrySize)
		n, err := file.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if n < entrySize {
			break
		}

		result[i] = &NumericTagData{
			Year:        int(binary.LittleEndian.Uint16(data[0:2])),
			DiscNumber:  int(data[2]),
			TrackNumber: int(data[3]),
			Bitrate:     int(binary.LittleEndian.Uint16(data[4:6])),
			Length:      int(binary.LittleEndian.Uint32(data[6:10])),
			PlayCount:   int(binary.LittleEndian.Uint32(data[10:14])),
			Rating:      int(data[14]),
		}
	}

	return result, nil
}

// scanFilesystem scans the filesystem for audio files
func (p *Parser) scanFilesystem(ctx context.Context) ([]*models.Song, error) {
	p.logger.Info("Scanning filesystem for audio files...")

	var songs []*models.Song
	audioExts := map[string]bool{
		".mp3":  true,
		".flac": true,
		".ogg":  true,
		".m4a":  true,
		".aac":  true,
		".wav":  true,
		".wma":  true,
		".ape":  true,
		".mpc":  true,
		".opus": true,
	}

	err := filepath.Walk(p.rockboxPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && path != p.rockboxPath {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}

		// Create relative path for Rockbox
		relPath, err := filepath.Rel(p.rockboxPath, path)
		if err != nil {
			relPath = path
		}
		// Convert to Rockbox path format (forward slashes)
		rockboxPath := "/" + strings.ReplaceAll(relPath, "\\", "/")

		song := &models.Song{
			Path:     rockboxPath,
			FileSize: info.Size(),
		}

		// Extract metadata from filename if no other info
		baseName := strings.TrimSuffix(info.Name(), ext)
		parts := strings.SplitN(baseName, " - ", 2)
		if len(parts) == 2 {
			song.Artist = strings.TrimSpace(parts[0])
			song.Title = strings.TrimSpace(parts[1])
		} else {
			song.Title = baseName
		}

		song.RockboxID = p.generateRockboxID(song)
		songs = append(songs, song)

		if len(songs)%100 == 0 {
			p.mu.Lock()
			p.status.ProcessedSongs = len(songs)
			p.mu.Unlock()
			p.logger.Debug("Scanned %d files...", len(songs))
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("filesystem scan failed: %w", err)
	}

	p.logger.Info("Found %d audio files", len(songs))
	return songs, nil
}

// generateRockboxID generates a unique ID for a song based on its path
func (p *Parser) generateRockboxID(song *models.Song) string {
	hash := md5.New()
	hash.Write([]byte(song.Path))
	return hex.EncodeToString(hash.Sum(nil))
}

// GetPlaylistPath returns the path for playlists in the Rockbox device
func (p *Parser) GetPlaylistPath() string {
	return filepath.Join(p.rockboxPath, "Playlists")
}

// EnsurePlaylistDir ensures the playlist directory exists
func (p *Parser) EnsurePlaylistDir() error {
	playlistDir := p.GetPlaylistPath()
	return os.MkdirAll(playlistDir, 0755)
}
