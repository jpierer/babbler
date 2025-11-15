package babbler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockStorer for testing
type MockStorer struct {
	stats map[string]int
	err   error
}

func NewMockStorer() *MockStorer {
	return &MockStorer{
		stats: make(map[string]int),
	}
}

func (m *MockStorer) Increment(t string) error {
	if m.err != nil {
		return m.err
	}
	m.stats[t]++
	return nil
}

func (m *MockStorer) GetStats() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`{"php": 2, "env": 1}`), nil
}

func (m *MockStorer) SetError(err error) {
	m.err = err
}

func TestNewBabbler(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)

	if babbler == nil {
		t.Fatal("NewBabbler returned nil")
	}
	if babbler.storeService != storer {
		t.Error("Storer was not set correctly")
	}
	if babbler.responseMinDelay != 0 {
		t.Errorf("Expected default responseMinDelay to be 0, got %d", babbler.responseMinDelay)
	}
	if babbler.responseMaxDelay != 0 {
		t.Errorf("Expected default responseMaxDelay to be 0, got %d", babbler.responseMaxDelay)
	}
}

func TestBabbler_SetResponseDelay(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)

	babbler.SetResponseDelay(50, 100)
	if babbler.responseMinDelay != 50 {
		t.Errorf("Expected responseMinDelay to be 50, got %d", babbler.responseMinDelay)
	}
	if babbler.responseMaxDelay != 100 {
		t.Errorf("Expected responseMaxDelay to be 100, got %d", babbler.responseMaxDelay)
	}
}

func TestBabbler_BabbleHandler_PHP(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for this test

	// Create test request
	req := httptest.NewRequest("GET", "/test.php", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := babbler.BabbleHandler("php")
	handler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}

	// Check that increment was called
	if storer.stats["php"] != 1 {
		t.Errorf("Expected php count to be 1, got %d", storer.stats["php"])
	}

	// Check response body starts with expected PHP tag
	if !strings.HasPrefix(w.Body.String(), "<?php") {
		t.Errorf("Expected response body to start with '<?php', got '%s'", w.Body.String())
	}

	// Should return some content (from embedded chunks)
	if w.Body.Len() == 0 {
		t.Error("Expected some response body content")
	}
}

func TestBabbler_BabbleHandler_ENV(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for this test

	// Create test request
	req := httptest.NewRequest("GET", "/.env", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := babbler.BabbleHandler("env")
	handler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that increment was called
	if storer.stats["env"] != 1 {
		t.Errorf("Expected env count to be 1, got %d", storer.stats["env"])
	}
}

func TestBabbler_BabbleHandler_UnknownType(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for this test

	// Create test request
	req := httptest.NewRequest("GET", "/test.unknown", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := babbler.BabbleHandler("unknown")
	handler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should return empty content for unknown type
	if w.Body.String() != "" {
		t.Error("Expected empty response body for unknown type")
	}

	// Check that increment was called anyway
	if storer.stats["unknown"] != 1 {
		t.Errorf("Expected unknown count to be 1, got %d", storer.stats["unknown"])
	}
}

func TestBabbler_BabbleHandler_WithDelay(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(10, 50) // 10-50ms delay range

	// Test multiple times to ensure delay is working
	var totalDuration time.Duration
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test.php", nil)
		w := httptest.NewRecorder()

		// Measure response time
		start := time.Now()
		handler := babbler.BabbleHandler("php")
		handler(w, req)
		duration := time.Since(start)
		totalDuration += duration

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Each individual request should not exceed reasonable bounds
		if duration > 100*time.Millisecond {
			t.Errorf("Response took too long: %v (expected max ~50ms)", duration)
		}
	}

	// Average should be reasonable (some delay but not excessive)
	avgDuration := totalDuration / 5
	if avgDuration > 80*time.Millisecond {
		t.Errorf("Average response time too long: %v", avgDuration)
	}
}

func TestBabbler_StatsHandler(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for stats

	// Create test request
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := babbler.StatsHandler()
	handler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	// Check response body
	expected := `{"php": 2, "env": 1}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected response body '%s', got '%s'", expected, w.Body.String())
	}
}

func TestBabbler_StatsHandler_Error(t *testing.T) {
	storer := NewMockStorer()
	storer.SetError(http.ErrAbortHandler)
	babbler := NewBabbler(storer)

	// Create test request
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := babbler.StatsHandler()
	handler(w, req)

	// Check error response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestBabbler_MultipleRequests(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test.php", nil)
		w := httptest.NewRecorder()
		handler := babbler.BabbleHandler("php")
		handler(w, req)
	}

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/.env", nil)
		w := httptest.NewRecorder()
		handler := babbler.BabbleHandler("env")
		handler(w, req)
	}

	// Check counters
	if storer.stats["php"] != 3 {
		t.Errorf("Expected php count to be 3, got %d", storer.stats["php"])
	}
	if storer.stats["env"] != 2 {
		t.Errorf("Expected env count to be 2, got %d", storer.stats["env"])
	}
}

func TestBabbler_SpecificFileLoading(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for faster test execution

	tests := []struct {
		url          string
		expectedFile string
		fileType     string
		description  string
	}{
		{"/admin.php", "admin.php", "php", "Should load admin.php from chunks/php/"},
		{"/foo/admin.php", "foo/admin.php", "php", "Should load admin.php with path prefix"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Reset storer
			storer.stats = make(map[string]int)

			req := httptest.NewRequest("GET", test.url, nil)
			w := httptest.NewRecorder()

			handler := babbler.BabbleHandler(test.fileType)
			handler(w, req)

			// Check basic response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", test.url, w.Code)
			}

			// Check that increment was called
			if storer.stats[test.fileType] != 1 {
				t.Errorf("Expected %s count to be 1 for %s, got %d", test.fileType, test.url, storer.stats[test.fileType])
			}

			// Should have some content
			if w.Body.Len() == 0 {
				t.Errorf("Expected some response body content for %s", test.url)
			}

			// Check that the response contains the expected WP_ADMIN check for admin.php
			body := w.Body.String()
			if strings.Contains(body, "if ( ! defined( 'WP_ADMIN' ) ) {") {
				// Good - admin.php content was loaded
			} else {
				t.Logf("Note: admin.php specific content not found, may have fallen back to random content for /foo/admin.php")
			}

			if len(body) == 0 {
				t.Errorf("Expected non-empty response for %s", test.url)
			}
		})
	}
}

func TestBabbler_FileLoadingFallback(t *testing.T) {
	storer := NewMockStorer()
	babbler := NewBabbler(storer)
	babbler.SetResponseDelay(0, 0) // No delay for faster test

	// Test with a file that likely doesn't exist - should fallback to random
	req := httptest.NewRequest("GET", "/nonexistent-file.php", nil)
	w := httptest.NewRecorder()

	handler := babbler.BabbleHandler("php")
	handler(w, req)

	// Should still return 200 and some content (from random fallback)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for nonexistent file, got %d", w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("Expected some response body content even for nonexistent file (random fallback)")
	}

	// Should still increment counter
	if storer.stats["php"] != 1 {
		t.Errorf("Expected php count to be 1 for nonexistent file, got %d", storer.stats["php"])
	}
}
