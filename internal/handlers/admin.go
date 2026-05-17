package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"

	"github.com/justestif/shelf/components"
	"github.com/justestif/shelf/internal/auth"
	"github.com/justestif/shelf/internal/storage"
)

// saveResult holds the result of a successful file save.
type saveResult struct {
	Filename string
	RelPath  string
}

// saveUploadedFiles writes uploaded files to disk under the pages directory.
// Returns successfully saved files and any per-file error messages.
func (h *Handler) saveUploadedFiles(folder string, files []*multipart.FileHeader) ([]saveResult, []string) {
	var saved []saveResult
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

		if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to create directory", fh.Filename))
			continue
		}

		src, err := fh.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read", fh.Filename))
			continue
		}
		dst, err := os.Create(clean)
		if err != nil {
			src.Close()
			errors = append(errors, fmt.Sprintf("%s: failed to create", fh.Filename))
			continue
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if copyErr != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to write", fh.Filename))
			continue
		}

		rel, _ := filepath.Rel(h.cfg.PagesDir, clean)
		saved = append(saved, saveResult{
			Filename: fh.Filename,
			RelPath:  filepath.ToSlash(rel),
		})
	}

	return saved, errors
}

// Admin renders the admin dashboard.
func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	entries, err := storage.Walk(h.cfg.PagesDir)
	if err != nil {
		log.Printf("walk failed: %v", err)
	}
	token := auth.GetToken(h.cfg.DataDir)
	if err := components.AdminPage(entries, token, csrf.Token(r), "").Render(r.Context(), w); err != nil {
		log.Printf("render admin page: %v", err)
	}
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
		entries, err := storage.Walk(h.cfg.PagesDir)
		if err != nil {
			log.Printf("walk failed: %v", err)
		}
		if err := components.FileListPartial(entries, csrf.Token(r), "No files selected").Render(r.Context(), w); err != nil {
			log.Printf("render file list: %v", err)
		}
		return
	}

	_, errs := h.saveUploadedFiles(folder, files)
	errMsg := strings.Join(errs, "; ")

	entries, err := storage.Walk(h.cfg.PagesDir)
	if err != nil {
		log.Printf("walk failed: %v", err)
	}
	if err := components.FileListPartial(entries, csrf.Token(r), errMsg).Render(r.Context(), w); err != nil {
		log.Printf("render file list: %v", err)
	}
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

	entries, err := storage.Walk(h.cfg.PagesDir)
	if err != nil {
		log.Printf("walk failed: %v", err)
	}
	if err := components.FileListPartial(entries, csrf.Token(r), "").Render(r.Context(), w); err != nil {
		log.Printf("render file list: %v", err)
	}
}

// TokenGenerate creates or regenerates the API token. Returns the token display partial.
func (h *Handler) TokenGenerate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GenerateToken(h.cfg.DataDir)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	if err := components.TokenPartial(token).Render(r.Context(), w); err != nil {
		log.Printf("render token partial: %v", err)
	}
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

	saved, errs := h.saveUploadedFiles(folder, files)

	type result struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}

	var uploaded []result
	for _, s := range saved {
		uploaded = append(uploaded, result{
			Path: s.RelPath,
			URL:  h.cfg.BaseURL + "/" + s.RelPath,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"uploaded": uploaded,
		"errors":   errs,
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
