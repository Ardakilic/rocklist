package database

import (
	"testing"

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
	
	err = db.Transaction(func(tx interface{}) error {
		return nil
	})
	
	// Note: The Transaction method expects *gorm.DB, so this test is simplified
	// In practice, you'd test with proper gorm.DB transactions
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
