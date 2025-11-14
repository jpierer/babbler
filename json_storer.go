package babbler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// JSONStorer implements the Storer interface using JSON
type JSONStorer struct {
	m        sync.Mutex
	filepath string
}

// NewJSONStorer creates a new JSONStorer instance
func NewJSONStorer(filepath string) *JSONStorer {
	return &JSONStorer{filepath: filepath}
}

// loadStats loads the stats from the JSON file
func (s *JSONStorer) loadStats() (map[string]int, error) {
	statsFile := filepath.Join(s.filepath, "stats.json")

	data, err := os.ReadFile(statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]int), nil
		}
		return nil, err
	}

	var stats map[string]int
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// saveStats saves the stats to the JSON file
func (s *JSONStorer) saveStats(stats map[string]int) error {
	// Ensure directory exists
	if err := os.MkdirAll(s.filepath, os.ModePerm); err != nil {
		return err
	}

	statsFile := filepath.Join(s.filepath, "stats.json")

	data, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(statsFile, data, 0644)
}

// Increment increments the count for a given babble type
func (s *JSONStorer) Increment(t string) error {
	s.m.Lock()
	defer s.m.Unlock()

	stats, err := s.loadStats()
	if err != nil {
		return err
	}

	stats[t]++

	return s.saveStats(stats)
}

// GetStats retrieves the count for a given babble type
func (s *JSONStorer) GetStats() ([]byte, error) {
	s.m.Lock()
	defer s.m.Unlock()

	stats, err := s.loadStats()
	if err != nil {
		return nil, err
	}

	return json.Marshal(stats)
}
