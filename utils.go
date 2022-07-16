package main

import (
	"net/http"
	"os"
	"path/filepath"
)

// Handler for serving static assets with modified time to help
// caching
func serveResources(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(".", r.URL.Path))
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	modTime := fi.ModTime()

	http.ServeContent(w, r, r.URL.Path, modTime, f)
}

func handleUnexpectedError(w http.ResponseWriter, err error) {
	http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
}
