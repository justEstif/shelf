package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/shelf/components"
	"github.com/justestif/shelf/internal/auth"
)

// LoginForm renders the login page.
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	// Already logged in? Redirect to admin.
	if auth.IsAuthenticated(h.sm, r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if err := components.LoginPage(csrf.Token(r), "").Render(r.Context(), w); err != nil {
		log.Printf("render login: %v", err)
	}
}

// Login validates the password and creates a session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if !auth.CheckPassword(password, h.cfg.Password) {
		if err := components.LoginPage(csrf.Token(r), "Invalid password").Render(r.Context(), w); err != nil {
			log.Printf("render login: %v", err)
		}
		return
	}

	auth.Authenticate(h.sm, w, r)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout clears the session and redirects to login.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearAuthentication(h.sm, w, r)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// ViewerPrompt renders the password prompt for protected pages.
func (h *Handler) ViewerPrompt(w http.ResponseWriter, r *http.Request) {
	if auth.IsAuthenticated(h.sm, r) || auth.IsViewerAuthenticated(h.sm, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}
	if err := components.ViewerPasswordPrompt(next, csrf.Token(r), "").Render(r.Context(), w); err != nil {
		log.Printf("render viewer prompt: %v", err)
	}
}

// ViewerAuth validates the viewer password and redirects to the protected page.
func (h *Handler) ViewerAuth(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	next := r.FormValue("next")
	if next == "" {
		next = "/"
	}

	if h.cfg.ViewerPassword == "" || password != h.cfg.ViewerPassword {
		if err := components.ViewerPasswordPrompt(next, csrf.Token(r), "Invalid password").Render(r.Context(), w); err != nil {
			log.Printf("render viewer prompt: %v", err)
		}
		return
	}

	auth.AuthenticateViewer(h.sm, w, r)
	http.Redirect(w, r, next, http.StatusSeeOther)
}
