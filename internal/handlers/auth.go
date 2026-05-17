package handlers

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/shelf/internal/auth"
)

// LoginForm renders the login page.
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	// Already logged in? Redirect to admin.
	if auth.IsAuthenticated(h.sm, r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	loginPage(csrf.Token(r), "").Render(r.Context(), w)
}

// Login validates the password and creates a session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if !auth.CheckPassword(password, h.cfg.Password) {
		loginPage(csrf.Token(r), "Invalid password").Render(r.Context(), w)
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
