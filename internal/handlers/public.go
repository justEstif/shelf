package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/justestif/shelf/components"
	"github.com/justestif/shelf/internal/auth"
	"github.com/justestif/shelf/internal/config"
	"github.com/justestif/shelf/internal/storage"
)

// isMarkdown returns true if the path has a .md extension.
func isMarkdown(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	cfg *config.Config
	sm  *scs.SessionManager
	ms  *storage.MetadataStore
}

// New creates a Handler with the given config, session manager, and metadata store.
func New(cfg *config.Config, sm *scs.SessionManager, ms *storage.MetadataStore) *Handler {
	return &Handler{cfg: cfg, sm: sm, ms: ms}
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
		if err := components.PublicIndex(entries, "").Render(r.Context(), w); err != nil {
			log.Printf("render index: %v", err)
		}
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

	// Enforce visibility
	visibility := h.ms.GetVisibility(pathParam)
	switch visibility {
	case storage.VisibilityPrivate:
		if !auth.IsAuthenticated(h.sm, r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
	case storage.VisibilityProtected:
		if !auth.IsAuthenticated(h.sm, r) && !auth.IsViewerAuthenticated(h.sm, r) {
			http.Redirect(w, r, "/auth/viewer?next=/"+pathParam, http.StatusSeeOther)
			return
		}
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
		if err := components.PublicIndex(entries, prefix).Render(r.Context(), w); err != nil {
			log.Printf("render index: %v", err)
		}
		return
	}

	// Markdown file: render preview (unless ?raw requested)
	if isMarkdown(pathParam) && !r.URL.Query().Has("raw") {
		data, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		if err := components.MarkdownPreview(string(data), pathParam).Render(r.Context(), w); err != nil {
			log.Printf("render markdown: %v", err)
		}
		return
	}

	// File: serve with correct content type
	w.Header().Set("Content-Type", storage.ContentType(fullPath))
	http.ServeFile(w, r, fullPath)
}
