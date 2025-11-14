package babbler

import (
	"embed"
	"math/rand"
	"net/http"
	"time"
)

//go:embed chunks/*
var embeddedChunks embed.FS

// Babbler
type Babbler struct {
	storeService  Storer
	responseDelay uint // delay in milliseconds (0 to responseDelay)
}

// Storer interface for storing and retrieving babble stats
type Storer interface {
	Increment(t string) error
	GetStats() ([]byte, error)
}

// NewBabbler creates a new Babbler instance
func NewBabbler(storeService Storer) *Babbler {
	return &Babbler{
		storeService:  storeService,
		responseDelay: 0, // default: no delay
	}
}

// SetResponseDelay sets the maximum response delay in milliseconds
func (b *Babbler) SetResponseDelay(delayMs uint) {
	b.responseDelay = delayMs
}

// Handler returns an HTTP handler that serves babble text based on type
func (b *Babbler) BabbleHandler(t string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add random delay to waste bot time
		if b.responseDelay > 0 {
			delay := rand.Intn(int(b.responseDelay))
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		babbleBytes := b.getTextForType(t)
		b.storeService.Increment(t)
		w.Header().Set("Content-Type", "text/plain")
		w.Write(babbleBytes)
	})
}

// StatsHandler returns an HTTP handler that serves babble statistics
func (b *Babbler) StatsHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statsBytes, err := b.storeService.GetStats()
		if err != nil {
			http.Error(w, "Error retrieving stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(statsBytes)
	})
}

// getTextForType returns the babble text for a given type
func (b *Babbler) getTextForType(t string) []byte {
	var textChunk []byte
	switch t {
	case "php":
		textChunk, _ = b.chunkLoader("chunks/php")
	case "env":
		textChunk, _ = b.chunkLoader("chunks/env")
	default:
		textChunk = []byte("")
	}

	return textChunk
}

// chunkLoader loads a random chunk from the specified folder
func (b *Babbler) chunkLoader(chunkFolder string) ([]byte, error) {
	var chunk []byte

	files, err := embeddedChunks.ReadDir(chunkFolder)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []byte(""), nil
	}

	// Read a random file from the directory
	randomIndex := rand.Intn(len(files))
	randomFile := files[randomIndex]
	chunk, err = embeddedChunks.ReadFile(chunkFolder + "/" + randomFile.Name())
	if err != nil {
		return nil, err
	}

	return chunk, nil
}
