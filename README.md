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
    "github.com/jpierer/babbler"
    "net/http"

    "github.com/go-chi/chi/middleware"
    "github.com/go-chi/chi/v5"
)

func main() {
    // Initialize Chi router
    r := chi.NewRouter()

    // Add logging middleware to see all requests
    r.Use(middleware.Logger)

    // Your normal application route
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("normal app"))
    })

    // Initialize Babbler with JSON storage backend
    // This stores simple counters for how often each file type is requested
    babblerStorer := babbler.NewJSONStorer("./data")
    babbler := babbler.NewBabbler(babblerStorer)

    // Configure response delay to waste more bot time (0-200ms random delay)
    babbler.SetResponseDelay(200)

    // Stats endpoint to view request counters by file type
    r.Get("/babbler/stats", babbler.StatsHandler())

    // Honeypot routes that catch common attack patterns and waste bot time
    r.Get("/*.php", babbler.BabbleHandler("php"))     // Serves fake PHP files to bots
    r.Get("/*.env", babbler.BabbleHandler("env"))     // Serves fake .env files to bots

    // Start the server
    http.ListenAndServe(":3000", r)
}
```

### Support Me

Give a star if this project was helpful in any way!

### License

The code is released under the [MIT LICENSE](/LICENSE).
