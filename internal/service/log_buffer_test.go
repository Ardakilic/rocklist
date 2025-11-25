package service

import (
	"testing"
	"time"
)

func TestNewLogBuffer(t *testing.T) {
	buffer := NewLogBuffer(100)

	if buffer == nil {
		t.Fatal("NewLogBuffer() returned nil")
	}

	if buffer.maxSize != 100 {
		t.Errorf("NewLogBuffer() maxSize = %v, want 100", buffer.maxSize)
	}

	if len(buffer.entries) != 0 {
		t.Errorf("NewLogBuffer() should have empty entries, got %d", len(buffer.entries))
	}
}

func TestLogBuffer_Add(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add("info", "test message")

	entries := buffer.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Add() should have 1 entry, got %d", len(entries))
	}

	if entries[0].Level != "info" {
		t.Errorf("Add() Level = %v, want info", entries[0].Level)
	}

	if entries[0].Message != "test message" {
		t.Errorf("Add() Message = %v, want test message", entries[0].Message)
	}

	if entries[0].Time.IsZero() {
		t.Error("Add() Time should not be zero")
	}
}

func TestLogBuffer_Add_Overflow(t *testing.T) {
	buffer := NewLogBuffer(3)

	buffer.Add("info", "message 1")
	buffer.Add("info", "message 2")
	buffer.Add("info", "message 3")
	buffer.Add("info", "message 4") // Should remove message 1

	entries := buffer.GetAll()
	if len(entries) != 3 {
		t.Fatalf("Buffer should have max 3 entries, got %d", len(entries))
	}

	// First entry should now be "message 2"
	if entries[0].Message != "message 2" {
		t.Errorf("First entry should be 'message 2', got %v", entries[0].Message)
	}

	// Last entry should be "message 4"
	if entries[2].Message != "message 4" {
		t.Errorf("Last entry should be 'message 4', got %v", entries[2].Message)
	}
}

func TestLogBuffer_GetAll(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add("info", "test 1")
	buffer.Add("error", "test 2")

	entries := buffer.GetAll()

	if len(entries) != 2 {
		t.Fatalf("GetAll() should return 2 entries, got %d", len(entries))
	}

	// Verify it returns a copy
	entries[0].Message = "modified"
	original := buffer.GetAll()
	if original[0].Message == "modified" {
		t.Error("GetAll() should return a copy, not the original")
	}
}

func TestLogBuffer_GetAll_Empty(t *testing.T) {
	buffer := NewLogBuffer(10)

	entries := buffer.GetAll()

	if entries == nil {
		t.Error("GetAll() should not return nil")
	}

	if len(entries) != 0 {
		t.Errorf("GetAll() should return empty slice, got %d entries", len(entries))
	}
}

func TestLogBuffer_Clear(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add("info", "test 1")
	buffer.Add("info", "test 2")
	buffer.Clear()

	entries := buffer.GetAll()
	if len(entries) != 0 {
		t.Errorf("Clear() should empty buffer, got %d entries", len(entries))
	}
}

func TestLogBuffer_Concurrent(t *testing.T) {
	buffer := NewLogBuffer(100)
	done := make(chan bool, 30)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			buffer.Add("info", "message")
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = buffer.GetAll()
			done <- true
		}()
	}

	// Concurrent clears
	for i := 0; i < 10; i++ {
		go func() {
			buffer.Clear()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}

	// Should not panic
}

func TestLogEntry_Fields(t *testing.T) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   "error",
		Message: "test error",
	}

	if entry.Level != "error" {
		t.Errorf("LogEntry.Level = %v, want error", entry.Level)
	}

	if entry.Message != "test error" {
		t.Errorf("LogEntry.Message = %v, want test error", entry.Message)
	}
}

func TestNewAppLogger(t *testing.T) {
	buffer := NewLogBuffer(10)
	logger := NewAppLogger(buffer)

	if logger == nil {
		t.Fatal("NewAppLogger() returned nil")
	}

	if logger.buffer != buffer {
		t.Error("NewAppLogger() should use provided buffer")
	}
}

func TestAppLogger_Info(t *testing.T) {
	buffer := NewLogBuffer(10)
	logger := NewAppLogger(buffer)

	logger.Info("test %s %d", "message", 42)

	entries := buffer.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Info() should add entry, got %d", len(entries))
	}

	if entries[0].Level != "info" {
		t.Errorf("Info() Level = %v, want info", entries[0].Level)
	}

	if entries[0].Message != "test message 42" {
		t.Errorf("Info() Message = %v, want 'test message 42'", entries[0].Message)
	}
}

func TestAppLogger_Error(t *testing.T) {
	buffer := NewLogBuffer(10)
	logger := NewAppLogger(buffer)

	logger.Error("error %s", "occurred")

	entries := buffer.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Error() should add entry, got %d", len(entries))
	}

	if entries[0].Level != "error" {
		t.Errorf("Error() Level = %v, want error", entries[0].Level)
	}
}

func TestAppLogger_Debug(t *testing.T) {
	buffer := NewLogBuffer(10)
	logger := NewAppLogger(buffer)

	logger.Debug("debug %v", 123)

	entries := buffer.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Debug() should add entry, got %d", len(entries))
	}

	if entries[0].Level != "debug" {
		t.Errorf("Debug() Level = %v, want debug", entries[0].Level)
	}
}
