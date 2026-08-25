package web

import "embed"

// Files are embedded so the Go binary serves the same frontend in a clean directory.
//
//go:embed index.html app.js styles.css
var Files embed.FS
