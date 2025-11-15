package babbler

import (
	"embed"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//go:embed chunks/*
var embeddedChunks embed.FS

// Babbler
type Babbler struct {
	storeService     Storer
	responseMinDelay uint // minimum delay in milliseconds
	responseMaxDelay uint // maximum delay in milliseconds
}

// Storer interface for storing and retrieving babble stats
type Storer interface {
	Increment(t string) error
	GetStats() ([]byte, error)
}

// NewBabbler creates a new Babbler instance
func NewBabbler(storeService Storer) *Babbler {
	return &Babbler{
		storeService:     storeService,
		responseMinDelay: 0,
		responseMaxDelay: 0,
	}
}

// SetResponseDelay sets the maximum response delay in milliseconds
func (b *Babbler) SetResponseDelay(minDelayMs, maxDelayMs uint) {
	b.responseMinDelay = minDelayMs
	b.responseMaxDelay = maxDelayMs
}

// Handler returns an HTTP handler that serves babble text based on type
func (b *Babbler) BabbleHandler(t string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add random delay to waste bot time
		if b.responseMaxDelay > 0 {
			delay := b.responseMinDelay + uint(rand.Intn(int(b.responseMaxDelay-b.responseMinDelay+1)))
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// Try to get specific file if provided in URL path
		tryFile := ""
		if r.URL.Path != "/" {
			// Extract only the filename from the path
			path := r.URL.Path[1:] // remove leading slash
			// Find the last slash and get everything after it
			if lastSlash := strings.LastIndex(path, "/"); lastSlash != -1 {
				tryFile = path[lastSlash+1:]
			} else {
				tryFile = path
			}
		}

		babbleBytes := b.getChunkForType(t, tryFile)
		b.storeService.Increment(t)

		// Set Header by type
		b.setHeader(w, t)
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

// getChunkForType returns the babble chunk for a given type
func (b *Babbler) getChunkForType(t string, tryfile string) []byte {
	var textChunk []byte
	switch t {
	case "php":
		textChunk, _ = b.chunkLoader("chunks/php", tryfile)
	case "env":
		textChunk, _ = b.chunkLoader("chunks/env", tryfile)
	default:
		textChunk = []byte("")
	}

	return textChunk
}

// chunkLoader loads a random chunk from the specified folder
func (b *Babbler) chunkLoader(chunkFolder string, tryfile string) ([]byte, error) {
	var chunk []byte

	// try to load the tryfile first in the chunkFolder
	if tryfile != "" {
		var err error
		chunk, err = embeddedChunks.ReadFile(chunkFolder + "/" + tryfile)
		if err == nil {
			return chunk, nil
		}
	}

	// If tryfile not found or not provided, proceed to load a random chunk
	rand.Seed(time.Now().UnixNano())

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

func (b *Babbler) setHeader(w http.ResponseWriter, t string) {
	switch t {
	case "php":
		// Set headers to mimic an old PHP server
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Server", "Apache/2.2.34 PHP/5.6.40")
		w.Header().Set("X-Powered-By", "PHP/5.6.40")
	case "env":
		w.Header().Set("Content-Type", "text/plain")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
}
