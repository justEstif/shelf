# LocalTrack

A local-first time tracker. All data lives in your browser's **Origin Private File System** via Postgres-in-WASM. No server. No auth. No cloud.

Inspired by [localtrack](https://github.com/arthurcornil/localtrack) (Go+WASM), rebuilt with Bun + TypeScript.

## Architecture

```
index.html  (Alpine.js UI)
    │
    │  postMessage({ action, payload })
    ▼
worker.js   (PGlite DB thread)
    │
    │  OPFS — opfs://localtrack
    ▼
browser filesystem  (persistent, private to origin)
```

### Why a Worker?

PGlite initializes a WASM Postgres instance. Both WASM compilation and SQL execution can briefly block the JS thread. Running it in a **dedicated Worker** keeps every keypress and animation frame on the main thread unaffected.

### Message protocol

Every message in both directions uses:

```js
{ action: string, payload: object }
```

The worker signals readiness with `{ action: 'ready' }`. Errors use `{ action: 'error', payload: string }`.

### DB schema

```sql
CREATE TABLE projects (
  id    SERIAL PRIMARY KEY,
  name  TEXT   NOT NULL UNIQUE,
  color TEXT   NOT NULL DEFAULT '#788C5D'
);

CREATE TABLE entries (
  id          SERIAL  PRIMARY KEY,
  name        TEXT    NOT NULL DEFAULT '',
  project_id  INTEGER REFERENCES projects(id) ON DELETE SET NULL,
  started_at  BIGINT  NOT NULL,   -- epoch ms
  ended_at    BIGINT              -- NULL while timer is running
);
```

Projects are created on-demand via upsert whenever a project name is typed. Deleting a project sets all its entries' `project_id` to NULL (entries are preserved).

## Running

### Option A — pre-built worker (CDN, no build step)

`worker.js` is already checked-in and imports PGlite from jsDelivr. Serve with any static HTTP server (browsers block `file://` for OPFS):

```bash
bunx serve . --cors
# or:  npx serve . --cors
# or:  python3 -m http.server 3000
```

Open `http://localhost:3000`.

### Option B — build from TypeScript source

```bash
bun install          # installs @electric-sql/pglite + types
bun run build        # compiles worker.ts → worker.js
bunx serve . --cors
```

During active development, keep the TypeScript compiler watching:

```bash
bun run watch &
bunx serve . --cors
```

## OPFS persistence

Data is stored in the browser's **Origin Private File System** — a sandboxed, origin-scoped filesystem that persists across page reloads and browser restarts. It is completely invisible to other origins and to the user's regular filesystem.

**Reset all data:** `chrome://settings/content/all` → find your origin → Clear storage. Or in DevTools: Application → Storage → Clear site data.

OPFS does **not** require `Cross-Origin-Embedder-Policy`/`Cross-Origin-Opener-Policy` headers, so any plain static server works.

## Features

| Feature | Notes |
|---------|-------|
| Timer | Start/stop with description + project |
| Manual entry | Description, project, start/end datetime pickers |
| Period filter | Today / This week / This month / All |
| Project summary | Time bars, color picker, all-time totals |
| Edit entries | Click any row to edit description, project, times |
| Delete entries | From the edit dialog |
| Project management | Rename, recolor, delete (via Projects dialog) |
| **Export CSV** | `id,description,project,started_at,ended_at,duration_minutes` |
| **Export JSON** | Array of entry objects with ISO timestamps |

Export was missing from the original localtrack — it's added here as a first-class feature accessible from both the header and the Projects dialog.

## Export format

**CSV:**
```
id,description,project,started_at,ended_at,duration_minutes
1,"Fix login bug","Backend","2025-06-01T09:00:00.000Z","2025-06-01T10:30:00.000Z",90
```

**JSON:**
```json
[
  {
    "id": 1,
    "description": "Fix login bug",
    "project": "Backend",
    "started_at": "2025-06-01T09:00:00.000Z",
    "ended_at": "2025-06-01T10:30:00.000Z",
    "duration_minutes": 90
  }
]
```

Running (stopped) timers are excluded from exports.

## Extending

**Add tags** — extend the schema with a `tags` table and a junction `entry_tags`. Add `getTags` / `updateEntryTags` worker actions.

**Sync on reconnect** — listen for `navigator.onLine` events; on reconnect, read all rows from OPFS and push to your API. The OPFS database is the source of truth.

**Charts** — the `projects` payload already includes `total_ms`. Feed it to a canvas library or build a simple SVG chart directly.

**Multi-device** — replace `opfs://localtrack` with a CRDT sync adapter like [electric-sql](https://electric-sql.com) or [powersync](https://www.powersync.com) and your data becomes collaborative without changing the schema.
