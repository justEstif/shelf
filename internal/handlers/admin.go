package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"

	"github.com/justestif/shelf/internal/auth"
	"github.com/justestif/shelf/internal/storage"
)

// Admin renders the admin dashboard.
func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	entries, _ := storage.Walk(h.cfg.PagesDir)
	token := auth.GetToken(h.cfg.DataDir)
	adminPage(entries, token, csrf.Token(r), "").Render(r.Context(), w)
}

// Upload handles HTMX file uploads. Returns an HTML partial on success.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	folder := r.FormValue("folder")
	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		entries, _ := storage.Walk(h.cfg.PagesDir)
		fileListPartial(entries, csrf.Token(r), "No files selected").Render(r.Context(), w)
		return
	}

	var errors []string
	for _, fh := range files {
		destDir := h.cfg.PagesDir
		if folder != "" {
			destDir = filepath.Join(destDir, folder)
		}
		destPath := filepath.Join(destDir, fh.Filename)

		// Sanitize
		clean, err := storage.SanitizePath(h.cfg.PagesDir, filepath.Join(folder, fh.Filename))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: invalid path", fh.Filename))
			continue
		}
		destPath = clean

		// Check conflict
		if storage.Exists(destPath) {
			errors = append(errors, fmt.Sprintf("%s: already exists", fh.Filename))
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to create directory", fh.Filename))
			continue
		}

		// Save file
		src, err := fh.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read", fh.Filename))
			continue
		}
		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			errors = append(errors, fmt.Sprintf("%s: failed to create", fh.Filename))
			continue
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			errors = append(errors, fmt.Sprintf("%s: failed to write", fh.Filename))
			continue
		}
		src.Close()
		dst.Close()
	}

	errMsg := strings.Join(errors, "; ")
	entries, _ := storage.Walk(h.cfg.PagesDir)
	fileListPartial(entries, csrf.Token(r), errMsg).Render(r.Context(), w)
}

// Delete handles HTMX delete requests. Returns the updated file list partial.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// Chi wildcard: /admin/delete/*
	pathParam := strings.TrimPrefix(r.URL.Path, "/admin/delete/")
	if pathParam == "" {
		http.Error(w, "No path specified", http.StatusBadRequest)
		return
	}

	fullPath, err := storage.SanitizePath(h.cfg.PagesDir, pathParam)
	if err != nil || !storage.Exists(fullPath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if err := storage.Delete(fullPath); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	entries, _ := storage.Walk(h.cfg.PagesDir)
	fileListPartial(entries, csrf.Token(r), "").Render(r.Context(), w)
}

// TokenGenerate creates or regenerates the API token. Returns the token display partial.
func (h *Handler) TokenGenerate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GenerateToken(h.cfg.DataDir)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	tokenPartial(token).Render(r.Context(), w)
}

// --- API Handlers (JSON) ---

// APIListFiles returns all files as JSON.
func (h *Handler) APIListFiles(w http.ResponseWriter, r *http.Request) {
	entries, err := storage.Walk(h.cfg.PagesDir)
	if err != nil {
		jsonError(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"files": entries,
	})
}

// APIUpload handles file uploads via the JSON API.
func (h *Handler) APIUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		jsonError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	folder := r.FormValue("folder")
	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		jsonError(w, "No files provided", http.StatusBadRequest)
		return
	}

	type result struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}

	var uploaded []result
	var errors []string

	for _, fh := range files {
		clean, err := storage.SanitizePath(h.cfg.PagesDir, filepath.Join(folder, fh.Filename))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: invalid path", fh.Filename))
			continue
		}

		if storage.Exists(clean) {
			errors = append(errors, fmt.Sprintf("%s: already exists", fh.Filename))
			continue
		}

		if err := os.MkdirAll(filepath.Dir(clean), 0755); err != nil {
			errors = append(errors, fmt.Sprintf("%s: mkdir failed", fh.Filename))
			continue
		}

		src, err := fh.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read failed", fh.Filename))
			continue
		}
		dst, err := os.Create(clean)
		if err != nil {
			src.Close()
			errors = append(errors, fmt.Sprintf("%s: create failed", fh.Filename))
			continue
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if copyErr != nil {
			errors = append(errors, fmt.Sprintf("%s: write failed", fh.Filename))
			continue
		}

		rel, _ := filepath.Rel(h.cfg.PagesDir, clean)
		uploaded = append(uploaded, result{
			Path: filepath.ToSlash(rel),
			URL:  h.cfg.BaseURL + "/" + filepath.ToSlash(rel),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"uploaded": uploaded,
		"errors":   errors,
	})
}

// APIDelete deletes a file or folder via the JSON API.
func (h *Handler) APIDelete(w http.ResponseWriter, r *http.Request) {
	pathParam := strings.TrimPrefix(r.URL.Path, "/admin/api/files/")
	if pathParam == "" {
		jsonError(w, "No path specified", http.StatusBadRequest)
		return
	}

	fullPath, err := storage.SanitizePath(h.cfg.PagesDir, pathParam)
	if err != nil || !storage.Exists(fullPath) {
		jsonError(w, "Not found", http.StatusNotFound)
		return
	}

	if err := storage.Delete(fullPath); err != nil {
		jsonError(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"deleted": pathParam,
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
