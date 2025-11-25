package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Ardakilic/rocklist/internal/database"
	"github.com/Ardakilic/rocklist/internal/models"
	"gorm.io/gorm/logger"
)

func setupConfigTestDB(t *testing.T) *database.Database {
	cfg := &database.Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestConfigRepository_Set(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	err := repo.Set(ctx, "test_key", "test_value")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
}

func TestConfigRepository_Get(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	_ = repo.Set(ctx, "test_key", "test_value")

	value, err := repo.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if value != "test_value" {
		t.Errorf("Get() = %v, want %v", value, "test_value")
	}
}

func TestConfigRepository_Get_NotFound(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent_key")
	if err != models.ErrConfigNotFound {
		t.Errorf("Get() error = %v, want ErrConfigNotFound", err)
	}
}

func TestConfigRepository_Set_Upsert(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	_ = repo.Set(ctx, "test_key", "original_value")
	_ = repo.Set(ctx, "test_key", "updated_value")

	value, _ := repo.Get(ctx, "test_key")
	if value != "updated_value" {
		t.Errorf("Set() upsert failed, got %v, want %v", value, "updated_value")
	}
}

func TestConfigRepository_Delete(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	_ = repo.Set(ctx, "to_delete", "value")

	err := repo.Delete(ctx, "to_delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.Get(ctx, "to_delete")
	if err != models.ErrConfigNotFound {
		t.Errorf("Key should be deleted, but Get returned: %v", err)
	}
}

func TestConfigRepository_Delete_NotFound(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent_key")
	if err != models.ErrConfigNotFound {
		t.Errorf("Delete() error = %v, want ErrConfigNotFound", err)
	}
}

func TestConfigRepository_GetAll(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	_ = repo.Set(ctx, "key1", "value1")
	_ = repo.Set(ctx, "key2", "value2")
	_ = repo.Set(ctx, "key3", "value3")

	configs, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("GetAll() returned %d configs, want 3", len(configs))
	}

	if configs["key1"] != "value1" {
		t.Errorf("GetAll() key1 = %v, want %v", configs["key1"], "value1")
	}
}

func TestConfigRepository_GetAll_Empty(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	configs, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("GetAll() returned %d configs, want 0", len(configs))
	}
}

func TestConfigRepository_GetLastParsedAt(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	_ = repo.SetLastParsedAt(ctx, now)

	parsed, err := repo.GetLastParsedAt(ctx)
	if err != nil {
		t.Fatalf("GetLastParsedAt() error = %v", err)
	}

	if parsed == nil {
		t.Fatal("GetLastParsedAt() returned nil")
	}

	if !parsed.Equal(now) {
		t.Errorf("GetLastParsedAt() = %v, want %v", parsed, now)
	}
}

func TestConfigRepository_GetLastParsedAt_NotSet(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	parsed, err := repo.GetLastParsedAt(ctx)
	if err != nil {
		t.Fatalf("GetLastParsedAt() error = %v", err)
	}

	if parsed != nil {
		t.Errorf("GetLastParsedAt() should return nil when not set, got %v", parsed)
	}
}

func TestConfigRepository_SetLastParsedAt(t *testing.T) {
	db := setupConfigTestDB(t)
	defer db.Close()

	repo := NewConfigRepository(db.DB())
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	err := repo.SetLastParsedAt(ctx, now)
	if err != nil {
		t.Fatalf("SetLastParsedAt() error = %v", err)
	}

	// Verify it was stored correctly
	value, _ := repo.Get(ctx, ConfigKeyLastParsedAt)
	if value == "" {
		t.Error("SetLastParsedAt() did not store value")
	}
}

func TestConfigKeyConstants(t *testing.T) {
	// Test that all config key constants are defined
	keys := []string{
		ConfigKeyLastParsedAt,
		ConfigKeyRockboxPath,
		ConfigKeyLastFMAPIKey,
		ConfigKeyLastFMAPISecret,
		ConfigKeyLastFMEnabled,
		ConfigKeySpotifyClientID,
		ConfigKeySpotifyClientSecret,
		ConfigKeySpotifyEnabled,
		ConfigKeyMusicBrainzUserAgent,
		ConfigKeyMusicBrainzEnabled,
	}

	for _, key := range keys {
		if key == "" {
			t.Errorf("Config key constant is empty")
		}
	}
}
