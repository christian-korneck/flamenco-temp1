package web

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed static
var webStaticFS embed.FS

// WebAppHandler returns a HTTP handler to serve the static files of the Flamenco Manager web app.
func WebAppHandler() (http.Handler, error) {
	// Strip the 'static/' directory off of the embedded filesystem.
	fs, err := fs.Sub(webStaticFS, "static")
	if err != nil {
		return nil, fmt.Errorf("unable to wrap embedded filesystem: %w", err)
	}

	// Serve `index.html` from the root directory if the requested file cannot be
	// found.
	wrappedFS := WrapFS(fs, "index.html")

	// Windows doesn't know this mime type. Web browsers won't load the webapp JS
	// file when it's served as text/plain.
	if err := mime.AddExtensionType(".js", "application/javascript"); err != nil {
		return nil, fmt.Errorf("registering mime type for JavaScript files: %w", err)
	}

	return http.FileServer(http.FS(wrappedFS)), nil
}

// FSWrapper wraps a filesystem and falls back to serving a specific file when
// the requested file cannot be found.
//
// This is necesasry for compatibility with Vue Router, as that generates URL
// paths to files that don't exist on the filesystem, like
// `/workers/c441766a-5d28-47cb-9589-b0caa4269065`. Serving `/index.html` in
// such cases makes Vue Router understand what's going on again.
type FSWrapper struct {
	fs       fs.FS
	fallback string
}

func (w *FSWrapper) Open(name string) (fs.File, error) {
	file, err := w.fs.Open(name)

	switch {
	case err == nil:
		return file, nil
	case errors.Is(err, fs.ErrNotExist):
		fallbackFile, fallbackErr := w.fs.Open(w.fallback)
		if fallbackErr != nil {
			log.Error().
				Str("name", name).
				Str("fallback", w.fallback).
				Err(err).
				Str("fallbackErr", fallbackErr.Error()).
				Msg("static web server: error opening fallback file")
			return file, err
		}
		return fallbackFile, nil
	}

	return file, err
}

func WrapFS(fs fs.FS, fallback string) *FSWrapper {
	return &FSWrapper{fs: fs, fallback: fallback}
}
