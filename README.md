# Shelf

A single-user file hosting app. Upload files, browse them publicly, manage them from an admin dashboard.

Built with Go, Chi, Templ, Tailwind CSS, and HTMX.

## Features

- **Public file browser** — browse and download files, navigate folders
- **Admin dashboard** — upload (multi-file + drag & drop), delete, manage files
- **JSON API** — upload, list, and delete files via Bearer token auth
- **Dark/light theme** — follows OS preference, manual toggle, persisted to localStorage
- **CSRF protection** on all web forms
- **Session auth** with password login
- **API token** — generate, rotate, use for programmatic access

## Quick start

Requires [mise](https://mise.jdx.dev/) and the [Tailwind CSS standalone CLI](https://tailwindcss.com/blog/standalone-cli) (`tailwindcss` binary in project root).

```bash
# Install Go toolchain
mise install

# Initial setup
mise run setup

# Download the Tailwind standalone CLI into the project root
# macOS (arm64)
curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64 -o tailwindcss
# Linux (x64)
curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss
chmod +x tailwindcss

# Start dev server (live reload via air)
mise run dev
```

Open `http://localhost:3000`. Navigate to `/admin` and log in.

## Configuration

All via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SHELF_PASSWORD` | *(required)* | Login password |
| `PORT` | `3000` | Listen port |
| `BASE_URL` | `http://localhost:3000` | Used for API response URLs and CSRF |
| `SHELF_PAGES_DIR` | `./pages` | Directory for uploaded files |
| `SHELF_DATA_DIR` | `./data` | Directory for app data (API token) |
| `CSRF_KEY` | *(required)* | 32-byte key for CSRF protection |

Set these in `mise.toml` under `[env]` for local dev, or as environment variables in production.

## Production build

```bash
mise run build
./bin/shelf
```

## API

Authenticate with `Authorization: Bearer <token>`. Generate a token from the admin dashboard.

```bash
# List files
curl -H "Authorization: Bearer <token>" http://localhost:3000/admin/api/files

# Upload files
curl -H "Authorization: Bearer <token>" \
  -F "files=@document.pdf" \
  -F "files=@image.png" \
  -F "folder=docs" \
  http://localhost:3000/admin/api/upload

# Delete a file
curl -X DELETE -H "Authorization: Bearer <token>" \
  http://localhost:3000/admin/api/files/docs/document.pdf
```

## Project structure

```
cmd/web/main.go          Entry point, router, middleware wiring
internal/
  auth/                   Session management, password check, API tokens
  config/                 Environment config
  handlers/               HTTP handlers (public, admin, auth, API)
  middleware/              Auth, CSRF, bearer token middleware
  storage/                File operations, path sanitization
components/
  *.templ                 Templ templates (layout, login, admin, public index)
  render.go               Go wrappers for templ components
styles/input.css          Tailwind CSS config + DESIGN.md theme tokens
static/css/               Built CSS (gitignored)
```

## License

MIT
