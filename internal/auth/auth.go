package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
)

const SessionAuthKey = "authenticated"

// NewSessionManager creates an in-memory session manager.
// Sessions are lost on server restart — re-login required.
// Acceptable for a single-user app.
func NewSessionManager() *scs.SessionManager {
	sm := scs.New()
	sm.Lifetime = 7 * 24 * time.Hour
	sm.Cookie.Name = "shelf_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = false // set true in production
	return sm
}

// IsAuthenticated checks if the current session is authenticated.
func IsAuthenticated(sm *scs.SessionManager, r *http.Request) bool {
	return sm.GetBool(r.Context(), SessionAuthKey)
}

// Authenticate marks the session as authenticated.
func Authenticate(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) {
	sm.Put(r.Context(), SessionAuthKey, true)
}

// ClearAuthentication removes the auth flag from the session.
func ClearAuthentication(sm *scs.SessionManager, w http.ResponseWriter, r *http.Request) {
	sm.Remove(r.Context(), SessionAuthKey)
}

// CheckPassword compares a password against the configured value.
func CheckPassword(password, configured string) bool {
	return password == configured
}

// --- API Token ---

// GenerateToken creates a cryptographically random token and saves it to disk.
// Returns the plaintext token (shown once to the user).
func GenerateToken(dataDir string) (string, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)
	path := filepath.Join(dataDir, "api_token")

	return token, os.WriteFile(path, []byte(token), 0600)
}

// GetToken reads the stored API token from disk.
// Returns empty string if no token exists.
func GetToken(dataDir string) string {
	data, err := os.ReadFile(filepath.Join(dataDir, "api_token"))
	if err != nil {
		return ""
	}
	return string(data)
}

// ValidateToken checks if a provided token matches the stored one.
func ValidateToken(dataDir, provided string) bool {
	stored := GetToken(dataDir)
	if stored == "" || provided == "" {
		return false
	}
	return stored == provided
}
