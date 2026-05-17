package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/justestif/shelf/internal/auth"
	"github.com/justestif/shelf/internal/config"
	"github.com/justestif/shelf/internal/storage"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	cfg *config.Config
	sm  *scs.SessionManager
}

// New creates a Handler with the given config and session manager.
func New(cfg *config.Config, sm *scs.SessionManager) *Handler {
	return &Handler{cfg: cfg, sm: sm}
}

// PublicServe handles all public GET requests.
//   - "/" → index page listing top-level files and folders
//   - "/<folder>/" → folder listing
//   - "/<path>" → serve the file
func (h *Handler) PublicServe(w http.ResponseWriter, r *http.Request) {
	pathParam := r.URL.Path[1:] // strip leading "/"

	// Root: show index
	if pathParam == "" || pathParam == "/" {
		entries, err := storage.List(h.cfg.PagesDir)
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}
		renderPublicIndex(w, r, entries, "")
		return
	}

	// Sanitize path
	fullPath, err := storage.SanitizePath(h.cfg.PagesDir, pathParam)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check existence
	info, err := os.Stat(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Directory: show listing
	if info.IsDir() {
		entries, err := storage.List(fullPath)
		if err != nil {
			http.Error(w, "Failed to list files", http.StatusInternalServerError)
			return
		}
		prefix := pathParam
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		renderPublicIndex(w, r, entries, prefix)
		return
	}

	// File: serve with correct content type
	w.Header().Set("Content-Type", storage.ContentType(fullPath))
	http.ServeFile(w, r, fullPath)
}

func renderPublicIndex(w http.ResponseWriter, r *http.Request, entries []storage.Entry, prefix string) {
	_ = publicIndex(entries, prefix).Render(r.Context(), w)
}
