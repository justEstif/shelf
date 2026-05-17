# Local-First Browser Pattern

Core idea from “ship the backend to the user”: if the work is single-user, personal/team-local, or workshop-only, the browser can be the runtime. Ship static files; let the user’s machine hold state; avoid server complexity until collaboration, shared secrets, or heavy compute require it.

## JS-first stance

Prefer JavaScript-only local-first apps:

- UI on main thread.
- State in plain JS objects.
- Persistence through URL params, `localStorage`, IndexedDB, or OPFS.
- Optional Web Worker for expensive parsing/export/scoring so UI stays responsive.
- Optional Service Worker only when offline caching is explicitly useful.

Do not use Go/Rust/C/WASM just because the article demonstrates it. WASM is an escape hatch for existing non-JS logic, CPU-heavy work, or real SQLite/database semantics. For most playgrounds, JS is enough.

## Persistence ladder

Choose lowest sufficient persistence:

1. **No persistence** — demo or manager doc.
2. **URL params** — tiny shareable configuration.
3. **localStorage** — workshop drafts, checklists, text fields, small JSON state.
4. **IndexedDB** — larger structured data, multiple saved sessions, pasted outputs, files/blobs.
5. **OPFS with JS file APIs** — local files, generated exports, larger private artifacts that should persist in-browser.
6. **SQLite WASM / OPFS VFS** — only when the browser needs real DB semantics, large offline datasets, relational queries, migrations, or import/export of `.db` files.

Do not jump to WASM/SQLite because it is interesting. Use it only when localStorage/IndexedDB are insufficient.

## Browser-as-backend architecture

For richer playgrounds, split responsibilities without creating a server:

```text
Main thread
  -> renders UI
  -> owns Alpine/component state
  -> sends heavy work to Worker

Web Worker
  -> parses files
  -> transforms data
  -> generates exports
  -> runs scoring/checks
  -> returns structured result

Browser storage
  -> localStorage / IndexedDB / OPFS
  -> never leaves device unless user exports or copies
```

Use `postMessage()` for Worker communication. Keep messages structured and small:

```js
worker.postMessage({ type: 'generateExport', payload: state });
worker.onmessage = (event) => {
  if (event.data.type === 'exportReady') download(event.data.markdown);
};
```

## AI integration ladder

1. **Prebaked outputs** — most reliable for facilitation.
2. **Copy/paste prompts** — safest for org-approved AI tools.
3. **Bring-your-own-key** — acceptable only for personal demos, never default for company workshops.
4. **Backend proxy** — required for shared API keys, auth, logging, or policy controls.

Default for company workshops: copy/paste plus prebaked fallback.

## Export/import pattern

Every lab with user input should export one of:
- markdown summary
- JSON state
- clipboard text
- downloadable `.md` / `.json`

Export should include enough context to make sense outside the app: scenario, choices, outputs, checklist results, and open questions.

For longer-lived local-first apps, add import too. Browser storage is origin/browser-bound; switching browsers or clearing site data loses state unless users export.

## Decision rules

- If data must sync across users in real time, use a backend.
- If the app needs shared secrets/API keys, use a backend.
- If data is local to a workshop participant/team, keep it in browser.
- If state is small JSON, use `localStorage`.
- If state is many records/files, use IndexedDB.
- If app needs relational queries, then consider SQLite WASM.
- If UI stutters from heavy work, move that work to a Web Worker before adding backend complexity.
