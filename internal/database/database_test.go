package database

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	
	if cfg.Path == "" {
		t.Error("DefaultConfig().Path is empty")
	}
	
	if cfg.InMemory {
		t.Error("DefaultConfig().InMemory should be false by default")
	}
	
	if cfg.LogLevel != logger.Silent {
		t.Errorf("DefaultConfig().LogLevel = %v, want Silent", cfg.LogLevel)
	}
}

func TestNew_InMemory(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	if db.DB() == nil {
		t.Error("New() returned nil DB")
	}
}

func TestDatabase_Migrate(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	if err := db.Migrate(); err != nil {
		t.Errorf("Migrate() error = %v", err)
	}
}

func TestDatabase_WipeData(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	
	if err := db.WipeData(); err != nil {
		t.Errorf("WipeData() error = %v", err)
	}
}

func TestDatabase_Transaction(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	
	err = db.Transaction(func(tx *gorm.DB) error {
		return nil
	})
	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}
}

func TestDatabase_Close(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNew_WithNilConfig(t *testing.T) {
	db, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}
	defer db.Close()
	
	if db.DB() == nil {
		t.Error("New(nil) returned nil DB")
	}
}

func TestNew_WithFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		Path:     tmpDir + "/test.db",
		InMemory: false,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	if db.DB() == nil {
		t.Error("New() returned nil DB")
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Path:     "/test/path.db",
		InMemory: true,
		LogLevel: logger.Info,
	}
	
	if cfg.Path != "/test/path.db" {
		t.Errorf("Config.Path = %v, want /test/path.db", cfg.Path)
	}
	if !cfg.InMemory {
		t.Error("Config.InMemory should be true")
	}
	if cfg.LogLevel != logger.Info {
		t.Errorf("Config.LogLevel = %v, want Info", cfg.LogLevel)
	}
}

func TestDatabase_DB(t *testing.T) {
	cfg := &Config{
		InMemory: true,
		LogLevel: logger.Silent,
	}
	
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	gormDB := db.DB()
	if gormDB == nil {
		t.Error("DB() returned nil")
	}
	
	// Verify it's a valid gorm.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("gorm.DB.DB() error = %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("DB ping failed: %v", err)
	}
}
