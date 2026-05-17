package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// SetupCSRF creates CSRF protection middleware.
//
// In templates, access the token with:
//
//	csrf.Token(r)
//
// In forms, include:
//
//	<input type="hidden" name="gorilla.csrf.Token" value={ csrfToken }/>
//
// IMPORTANT:
// - CSRF_KEY must be 32 bytes
// - Set secure=true in production (HTTPS only)
// - Token automatically validated on POST/PUT/DELETE
func SetupCSRF(key []byte, secure bool) func(http.Handler) http.Handler {
	return csrf.Protect(
		key,
		csrf.Secure(secure),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
}
