# babbler

Babbler is a Go library that serves fake sensitive files (like `.env`, `config.php`, etc.) to potential attackers, wasting their time and slowing down automated scanners.

## Why?

**Annoy script kiddies**: Let bots waste CPU cycles downloading your fake files instead of getting quick 404s.

**Mess with scanners**: Serve them realistic fake .env files with bogus database credentials and API keys.

**Waste their time**: Add delays so bots sit there waiting while your server laughs at them.

**Super easy**: Just drop in a few routes and watch the chaos unfold.

**See who's knocking**: Count how many times bots fall for your traps.

## Install

    go get github.com/jpierer/babbler@main

## Example Usage

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jpierer/babbler"
)

func main() {
	// Initialize Chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize Babbler
	storagePath := "./data"
	jsonStorer := babbler.NewJSONStorer(storagePath)
	babblerInstance := babbler.NewBabbler(jsonStorer)
	babblerInstance.SetResponseDelay(500, 2000) // 500ms to 2000ms delay

	// Normal application routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Babbler server is running", "status": "online"}`))
	})

	// Stats endpoint
	r.Get("/babbler/stats", babblerInstance.StatsHandler())

	// === Babble routes ===
	// Honeypot routes that catch common attack patterns and waste bot time
	// These need to be BEFORE any general slug routes to catch .php and .env files

	// First: catch files in single subdirectory level (e.g., /wp-content/file.php)
	r.Route("/{segment}", func(sr chi.Router) {
		sr.Get("/{file}.php", babblerInstance.BabbleHandler("php"))
		sr.Get("/{file}.env", babblerInstance.BabbleHandler("env"))
	})

	// Second: catch files in root directory
	r.Get("/{file}.php", babblerInstance.BabbleHandler("php"))
	r.Get("/{file}.env", babblerInstance.BabbleHandler("env"))

	// Or you can use a wildcard route to catch files at any depth
	// r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
	// 	path := r.URL.Path
	// 	if strings.HasSuffix(path, ".php") {
	// 		babblerInstance.BabbleHandler("php")(w, r)
	// 		return
	// 	}
	// 	if strings.HasSuffix(path, ".env") {
	// 		babblerInstance.BabbleHandler("env")(w, r)
	// 		return
	// 	}
	// 	// If not a honeypot file, return 404
	// 	http.NotFound(w, r)
	// })

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Printf("Babbler Stats available at: http://localhost%s/babbler/stats", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

### Support Me

Give a :star: if this project was helpful in any way!

### License

The code is released under the [MIT LICENSE](/LICENSE).
