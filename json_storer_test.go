package babbler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewJSONStorer(t *testing.T) {
	storer := NewJSONStorer("./test_data")
	if storer == nil {
		t.Fatal("NewJSONStorer returned nil")
	}
	if storer.filepath != "./test_data" {
		t.Errorf("Expected filepath './test_data', got '%s'", storer.filepath)
	}
}

func TestJSONStorer_Increment(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	storer := NewJSONStorer(tempDir)

	// Test incrementing a new type
	err := storer.Increment("php")
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	// Test incrementing the same type again
	err = storer.Increment("php")
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	// Test incrementing a different type
	err = storer.Increment("env")
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	// Verify the stats file was created
	statsFile := filepath.Join(tempDir, "stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		t.Fatal("Stats file was not created")
	}

	// Read and verify the content
	data, err := os.ReadFile(statsFile)
	if err != nil {
		t.Fatalf("Failed to read stats file: %v", err)
	}

	var stats map[string]int
	err = json.Unmarshal(data, &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if stats["php"] != 2 {
		t.Errorf("Expected php count to be 2, got %d", stats["php"])
	}
	if stats["env"] != 1 {
		t.Errorf("Expected env count to be 1, got %d", stats["env"])
	}
}

func TestJSONStorer_GetStats(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	storer := NewJSONStorer(tempDir)

	// Increment some counters
	storer.Increment("php")
	storer.Increment("php")
	storer.Increment("env")

	// Get stats
	statsBytes, err := storer.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	var stats map[string]int
	err = json.Unmarshal(statsBytes, &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if stats["php"] != 2 {
		t.Errorf("Expected php count to be 2, got %d", stats["php"])
	}
	if stats["env"] != 1 {
		t.Errorf("Expected env count to be 1, got %d", stats["env"])
	}
}

func TestJSONStorer_GetStatsEmptyFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	storer := NewJSONStorer(tempDir)

	// Get stats without incrementing anything
	statsBytes, err := storer.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	var stats map[string]int
	err = json.Unmarshal(statsBytes, &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if len(stats) != 0 {
		t.Errorf("Expected empty stats, got %v", stats)
	}
}

func TestJSONStorer_ConcurrentAccess(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	storer := NewJSONStorer(tempDir)

	// Run concurrent increments
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 5; j++ {
				storer.Increment("php")
			}
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final count
	statsBytes, err := storer.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	var stats map[string]int
	err = json.Unmarshal(statsBytes, &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if stats["php"] != 50 {
		t.Errorf("Expected php count to be 50, got %d", stats["php"])
	}
}
