# Shelf

A single-user file hosting app. Upload files via web UI or CLI, browse publicly, manage from an admin dashboard. Markdown preview with Mermaid diagrams. Per-file access control.

## Deploy

The fastest way to run Shelf:

```bash
docker run -d \
  -p 3000:3000 \
  -e SHELF_PASSWORD=your-password \
  -e CSRF_KEY=$(openssl rand -hex 16) \
  -v shelf-pages:/shelf/pages \
  -v shelf-data:/shelf/data \
  --name shelf \
  ghcr.io/justestif/shelf:latest
```

Open `http://localhost:3000`. Go to `/admin` and log in.

### docker compose

A [`docker-compose.yml`](docker-compose.yml) is included:

```bash
# Copy and edit passwords
cp docker-compose.yml docker-compose.override.yml

# Start
docker compose up -d
```

### Dokku / Heroku

Shelf reads `PORT` as a fallback for `SHELF_PORT`, so it works with platform-assigned ports:

```bash
dokku apps:create shelf
dokku config:set shelf SHELF_PASSWORD=your-password CSRF_KEY=$(openssl rand -hex 16)
dokku storage:mount shelf /var/lib/shelf/pages:/shelf/pages
dokku storage:mount shelf /var/lib/shelf/data:/shelf/data
git push dokku main
```

### From source

Requires [Go](https://go.dev/dl/) and the [Tailwind CSS standalone CLI](https://tailwindcss.com/blog/standalone-cli):

```bash
go install github.com/a-h/templ/cmd/templ@latest

# Download tailwindcss binary for your platform into project root
curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss
chmod +x tailwindcss

# Build
./tailwindcss -i styles/input.css -o cmd/web/static/css/tailwind.css --minify
templ generate
go build -o bin/shelf ./cmd/web

# Run
SHELF_PASSWORD=your-password CSRF_KEY=$(openssl rand -hex 16) ./bin/shelf
```

## Features

- **Markdown preview** — `.md` files rendered with [marked.js](https://marked.js.org), [Mermaid](https://mermaid.js.org) diagrams, styled with the Shelf theme. YAML front matter collapsed in a `<details>` element.
- **Access control** — per-file visibility: **public** (default), **private** (admin login required), **protected** (viewer password). Inherited from parent directories.
- **Public file browser** — browse and download files, navigate folders
- **Admin dashboard** — upload (multi-file + drag & drop), delete, manage visibility
- **JSON API** — upload, list, delete, set visibility via Bearer token
- **Dark/light theme** — follows OS preference, manual toggle, persisted to localStorage
- **Single binary** — static files embedded, zero runtime file dependencies
- **Companion CLI** — [shelf-cli](https://github.com/justEstif/shelf-cli) for `shelf upload`, `shelf ls`, `shelf rm`

## CLI

```
go install github.com/justestif/shelf-cli@latest
```

```bash
export SHELF_API_TOKEN=your-token
export SHELF_URL=http://localhost:3000

# Upload and set visibility
shelf upload report.html -v protected

# Upload multiple files into a folder
shelf upload -f my-project src/index.html src/app.js

# List files
shelf ls

# Delete
shelf rm old-file.html

# Open in browser
shelf open report.html
```

See [justEstif/shelf-cli](https://github.com/justEstif/shelf-cli) for full docs.

## API

Authenticate with `Authorization: Bearer <token>`. Generate a token from the admin dashboard.

```bash
TOKEN=your-token
URL=http://localhost:3000

# List files
curl -H "Authorization: Bearer $TOKEN" $URL/admin/api/files

# Upload files
curl -H "Authorization: Bearer $TOKEN" \
  -F "files=@report.html" \
  -F "folder=docs" \
  $URL/admin/api/upload

# Delete a file
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  $URL/admin/api/files/docs/report.html

# Set visibility
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -F "path=docs/report.html" \
  -F "visibility=protected" \
  $URL/admin/api/visibility
```

## Configuration

All via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SHELF_PASSWORD` | *(required)* | Admin login password |
| `CSRF_KEY` | *(required)* | 32-byte key for CSRF protection |
| `SHELF_PORT` | `3000` | Listen port (falls back to `PORT` for Dokku/Heroku) |
| `SHELF_BASE_URL` | `http://localhost:3000` | Used for API response URLs and CSRF origin |
| `SHELF_PAGES_DIR` | `./pages` | Directory for uploaded files |
| `SHELF_DATA_DIR` | `./data` | Directory for app data (API token, metadata) |
| `SHELF_VIEWER_PASSWORD` | *(none)* | Password for "protected" pages. If unset, protected pages require admin login instead. |
| `SHELF_MAX_FILE_SIZE` | `20` | Max upload size per file (MB) |
| `SHELF_MAX_VOLUME_SIZE` | `1024` | Max total upload volume (MB) |

## Playground skill

The [playground](./skills/playground/) agent skill teaches your coding agent to build polished, self-contained HTML artifacts (specs, diagrams, prototypes, slide decks, reports) that you can drop straight into Shelf.

```bash
npx skills add justEstif/shelf
```

## Project structure

```
cmd/web/
  main.go              Entry point, router, middleware wiring
  static/              Embedded static files (favicon, CSS)
internal/
  auth/                Session management, password check, API tokens
  config/              Environment config
  handlers/            HTTP handlers (public, admin, auth, API)
  middleware/           Auth, CSRF, bearer token middleware
  storage/             File operations, metadata store, path sanitization
components/
  *.templ              Templ templates (layout, login, admin, markdown preview, viewer prompt)
  render.go            Go wrappers for templ components
styles/input.css       Tailwind CSS config + DESIGN.md theme tokens
```

## License

MIT
