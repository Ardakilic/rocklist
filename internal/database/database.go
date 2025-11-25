// Package database provides database initialization and connection management
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection
type Database struct {
	db *gorm.DB
}

// Config holds database configuration
type Config struct {
	Path      string
	InMemory  bool
	LogLevel  logger.LogLevel
}

// DefaultConfig returns the default database configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Path:     filepath.Join(homeDir, ".rocklist", "rocklist.db"),
		InMemory: false,
		LogLevel: logger.Silent,
	}
}

// New creates a new database connection
func New(cfg *Config) (*Database, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	var dsn string
	if cfg.InMemory {
		dsn = "file::memory:?cache=shared"
	} else {
		// Ensure directory exists
		dir := filepath.Dir(cfg.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		dsn = cfg.Path
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign keys
	db.Exec("PRAGMA foreign_keys = ON")

	return &Database{db: db}, nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	return d.db.AutoMigrate(
		&models.Song{},
		&models.Playlist{},
		&models.PlaylistSong{},
		&models.Config{},
	)
}

// DB returns the underlying GORM database
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WipeData removes all song and playlist data (for pre-fetched data wipe)
func (d *Database) WipeData() error {
	tx := d.db.Begin()
	
	if err := tx.Exec("DELETE FROM playlist_songs").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete playlist songs: %w", err)
	}
	
	if err := tx.Exec("DELETE FROM playlists").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete playlists: %w", err)
	}
	
	if err := tx.Exec("DELETE FROM songs").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete songs: %w", err)
	}

	// Remove last parsed timestamp
	if err := tx.Where("key = ?", "last_parsed_at").Delete(&models.Config{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete last parsed config: %w", err)
	}
	
	return tx.Commit().Error
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}
