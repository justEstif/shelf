package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/justestif/shelf/internal/auth"
	"github.com/justestif/shelf/internal/config"
	"github.com/justestif/shelf/internal/handlers"
	customMiddleware "github.com/justestif/shelf/internal/middleware"
	"github.com/justestif/shelf/internal/storage"
)

//go:embed static
var staticFS embed.FS

func main() {
	cfg := config.Load()

	if cfg.Password == "" {
		log.Fatal("SHELF_PASSWORD is required")
	}

	// Ensure storage directories exist
	if err := os.MkdirAll(cfg.PagesDir, 0o755); err != nil {
		log.Fatalf("Failed to create pages directory: %v", err)
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Session manager (in-memory, lost on restart)
	sm := auth.NewSessionManager()

	// Metadata store (visibility settings per path)
	ms := storage.NewMetadataStore(cfg.DataDir)

	// Handler
	h := handlers.New(cfg, sm, ms)

	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Session middleware (loads/saves session cookie)
	r.Use(sm.LoadAndSave)

	// CSRF setup
	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) != 32 {
		log.Fatal("CSRF_KEY must be exactly 32 bytes long")
	}
	csrfMw := customMiddleware.SetupCSRF(csrfKey, cfg.IsProduction(), cfg.BaseURL)

	// Static files (embedded CSS, favicon — no CSRF needed)
	staticSub, _ := fs.Sub(staticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFS, "static/favicon.svg")
	})

	// Login routes (CSRF, no auth required) — must be before the wildcard
	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Get("/admin/login", h.LoginForm)
		r.Post("/admin/login", h.Login)

		// Viewer password prompt (for protected pages)
		r.Get("/auth/viewer", h.ViewerPrompt)
		r.Post("/auth/viewer", h.ViewerAuth)
	})

	// Admin routes (CSRF + session auth required)
	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Use(customMiddleware.RequireAuth(sm))
		r.Get("/admin", h.Admin)
		r.Post("/admin/upload", h.Upload)
		r.Delete("/admin/delete/*", h.Delete)
		r.Post("/admin/token", h.TokenGenerate)
		r.Post("/admin/visibility", h.SetVisibility)
		r.Post("/admin/logout", h.Logout)
	})

	// API routes (bearer token auth, no CSRF)
	r.Route("/admin/api", func(r chi.Router) {
		r.Use(customMiddleware.BearerToken(cfg))
		r.Get("/files", h.APIListFiles)
		r.Post("/upload", h.APIUpload)
		r.Delete("/files/*", h.APIDelete)
		r.Post("/visibility", h.APISetVisibility)
	})

	// Public routes (GET only, catch-all — must be last)
	r.Get("/*", h.PublicServe)

	port := cfg.Port
	log.Printf("Shelf starting on http://localhost:%s", port)
	log.Printf("Pages dir: %s", cfg.PagesDir)
	log.Printf("Data dir:  %s", cfg.DataDir)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
