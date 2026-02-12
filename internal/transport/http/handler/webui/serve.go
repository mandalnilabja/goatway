package webui

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/mandalnilabja/goatway/web"
)

// WebUIHandler creates an HTTP handler for serving the embedded web UI.
// It serves static files from the embedded filesystem and falls back to
// index.html for SPA routing (History API).
// The handler expects to be mounted at /web/ prefix.
func (h *Handlers) WebUIHandler() http.Handler {
	// Get the embedded filesystem
	staticFS, err := fs.Sub(web.FS, ".")
	if err != nil {
		// This should never happen with a valid embed
		panic("failed to create sub filesystem: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Strip /web prefix to get the actual file path
		filePath := strings.TrimPrefix(path, "/web")
		if filePath == "" {
			filePath = "/"
		}

		// Serve static files directly
		if strings.HasPrefix(filePath, "/static/") {
			r.URL.Path = filePath
			fileServer.ServeHTTP(w, r)
			return
		}

		// For SPA routes, check if file exists; if not, serve index.html
		if filePath != "/" {
			// Try to open the file to see if it exists
			_, err := fs.Stat(staticFS, strings.TrimPrefix(filePath, "/"))
			if err != nil {
				// File doesn't exist - serve index.html for SPA routing
				filePath = "/"
			}
		}

		r.URL.Path = filePath
		fileServer.ServeHTTP(w, r)
	})
}

// ServeWebUI is a convenience method that returns the WebUI handler.
// Use this in route registration.
func (h *Handlers) ServeWebUI() http.Handler {
	return h.WebUIHandler()
}
