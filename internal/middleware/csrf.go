package middleware

import (
	"log"
	"net/http"
	"net/url"

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
func SetupCSRF(key []byte, secure bool, baseURL string) func(http.Handler) http.Handler {
	opts := []csrf.Option{
		csrf.Secure(secure),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.FieldName("gorilla.csrf.Token"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reason := csrf.FailureReason(r)
			log.Printf("CSRF rejected: %v | Origin=%s Referer=%s Host=%s", reason, r.Header.Get("Origin"), r.Header.Get("Referer"), r.Host)
			http.Error(w, reason.Error(), http.StatusForbidden)
		})),
	}

	// Trust the configured host (needed when TLS is terminated upstream, e.g. Cloudflare)
	u, err := url.Parse(baseURL)
	if err == nil && u.Host != "" {
		opts = append(opts, csrf.TrustedOrigins([]string{u.Host}))
	}
	return csrf.Protect(key, opts...)
}
