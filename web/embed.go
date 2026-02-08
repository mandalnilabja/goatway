// Package web provides the embedded web UI files.
package web

import "embed"

// FS contains the embedded web UI files (index.html, static/css, static/js).
// This is exported for use by the HTTP handler to serve the web dashboard.
//
//go:embed index.html static
var FS embed.FS
