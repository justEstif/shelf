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
func AdminPage(entries []storage.Entry, token string, csrfToken string, uploadMsg string, visibility map[string]string) templ.Component {
	return adminPage(entries, token, csrfToken, uploadMsg, visibility)
}

// FileListPartial returns the file list HTMX partial component.
func FileListPartial(entries []storage.Entry, csrfToken string, errMsg string, visibility map[string]string) templ.Component {
	return fileListPartial(entries, csrfToken, errMsg, visibility)
}

// TokenPartial returns the token display HTMX partial component.
func TokenPartial(token string) templ.Component {
	return tokenPartial(token)
}

// MarkdownPreview returns a rendered markdown preview page.
func MarkdownPreview(content string, filePath string) templ.Component {
	return markdownPreview(content, filePath)
}

// ViewerPasswordPrompt returns the protected page password prompt.
func ViewerPasswordPrompt(nextPath string, csrfToken string, errorMsg string) templ.Component {
	return viewerPasswordPrompt(nextPath, csrfToken, errorMsg)
}

// VisibilityBadge returns the visibility badge + selector for a file.
func VisibilityBadge(path string, visibility string, csrfToken string) templ.Component {
	return visibilityBadge(path, visibility, csrfToken)
}
