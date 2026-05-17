package components

import (
	"github.com/a-h/templ"
	"github.com/justestif/shelf/internal/storage"
)

// PublicIndex returns the public file listing page component.
func PublicIndex(entries []storage.Entry, prefix string) templ.Component {
	return publicIndex(entries, prefix)
}

// LoginPage returns the login page component.
func LoginPage(csrfToken string, errorMsg string) templ.Component {
	return loginPage(csrfToken, errorMsg)
}

// AdminPage returns the admin dashboard component.
func AdminPage(entries []storage.Entry, token string, csrfToken string, uploadMsg string) templ.Component {
	return adminPage(entries, token, csrfToken, uploadMsg)
}

// FileListPartial returns the file list HTMX partial component.
func FileListPartial(entries []storage.Entry, csrfToken string, errMsg string) templ.Component {
	return fileListPartial(entries, csrfToken, errMsg)
}

// TokenPartial returns the token display HTMX partial component.
func TokenPartial(token string) templ.Component {
	return tokenPartial(token)
}
