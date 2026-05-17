// worker.js — DB worker for LocalTrack
//
// Runs in a dedicated Worker (type: 'module') so ES module imports and
// top-level await are available. PGlite provides Postgres-in-WASM; the
// 'opfs://localtrack' URI persists the database in the browser's Origin
// Private File System (survives page reloads, isolated to this origin).
//
// TypeScript source lives in worker.ts — build with:
//   bun build worker.ts --outfile worker.js --target browser
// This .js file is the pre-built, CDN-ready version.

import { PGlite } from 'https://cdn.jsdelivr.net/npm/@electric-sql/pglite/dist/index.js';

// ─────────────────────────────────────────────────────────────────────────────
// Init
// ─────────────────────────────────────────────────────────────────────────────

const db = new PGlite('opfs://localtrack');

await db.exec(`
  CREATE TABLE IF NOT EXISTS projects (
    id    SERIAL  PRIMARY KEY,
    name  TEXT    NOT NULL UNIQUE,
    color TEXT    NOT NULL DEFAULT '#788C5D'
  );

  CREATE TABLE IF NOT EXISTS entries (
    id          SERIAL   PRIMARY KEY,
    name        TEXT     NOT NULL DEFAULT '',
    project_id  INTEGER  REFERENCES projects(id) ON DELETE SET NULL,
    started_at  BIGINT   NOT NULL,
    ended_at    BIGINT
  );
`);

self.postMessage({ action: 'ready' });

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

/**
 * Returns the epoch-ms floor for the given period, or null for 'all'.
 * @param {'today'|'week'|'month'|'all'} period
 * @returns {number|null}
 */
function periodBound(period) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (period === 'today') return today.getTime();
  if (period === 'week') {
    const d = new Date(today);
    d.setDate(d.getDate() - d.getDay()); // Sunday-start week
    return d.getTime();
  }
  if (period === 'month') {
    return new Date(today.getFullYear(), today.getMonth(), 1).getTime();
  }
  return null; // 'all'
}

/**
 * Upsert a project by name and return its id, or null if name is empty.
 * @param {string|undefined} name
 * @returns {Promise<number|null>}
 */
async function upsertProject(name) {
  if (!name || !name.trim()) return null;
  await db.query(
    `INSERT INTO projects (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`,
    [name.trim()],
  );
  const r = await db.query(`SELECT id FROM projects WHERE name = $1`, [name.trim()]);
  return r.rows[0]?.id ?? null;
}

// ─────────────────────────────────────────────────────────────────────────────
// Entry query fragment (avoids repeating the JOIN)
// ─────────────────────────────────────────────────────────────────────────────

const ENTRY_SELECT = `
  SELECT e.id, e.name, e.started_at, e.ended_at,
         p.id    AS project_id,
         p.name  AS project_name,
         p.color AS project_color
  FROM   entries e
  LEFT JOIN projects p ON p.id = e.project_id
`;

// ─────────────────────────────────────────────────────────────────────────────
// Message dispatch
// ─────────────────────────────────────────────────────────────────────────────

self.onmessage = async ({ data }) => {
  const { action, payload = {} } = data;

  try {
    switch (action) {

      // ── List entries (filtered by period) ──────────────────────────────────
      case 'getEntries': {
        const from = periodBound(payload.period ?? 'today');
        const { rows } = from
          ? await db.query(
              ENTRY_SELECT + ` WHERE e.started_at >= $1 ORDER BY e.started_at DESC`,
              [from],
            )
          : await db.query(
              ENTRY_SELECT + ` ORDER BY e.started_at DESC`,
            );

        self.postMessage({ action: 'entries', payload: rows });
        break;
      }

      // ── Projects with all-time totals ──────────────────────────────────────
      case 'getProjects': {
        const { rows } = await db.query(`
          SELECT p.id, p.name, p.color,
                 COALESCE(SUM(
                   CASE WHEN e.ended_at IS NOT NULL
                        THEN e.ended_at - e.started_at
                        ELSE 0
                   END
                 ), 0) AS total_ms
          FROM   projects p
          LEFT JOIN entries e ON e.project_id = p.id
          GROUP  BY p.id, p.name, p.color
          ORDER  BY p.name
        `);
        self.postMessage({ action: 'projects', payload: rows });
        break;
      }

      // ── Currently running entry ────────────────────────────────────────────
      case 'getRunningEntry': {
        const { rows } = await db.query(
          ENTRY_SELECT + ` WHERE e.ended_at IS NULL ORDER BY e.started_at DESC LIMIT 1`,
        );
        self.postMessage({ action: 'runningEntry', payload: rows[0] ?? null });
        break;
      }

      // ── Start timer ────────────────────────────────────────────────────────
      case 'startTimer': {
        // Close any orphaned open entries
        await db.exec(`UPDATE entries SET ended_at = ${Date.now()} WHERE ended_at IS NULL`);

        const project_id = await upsertProject(payload.project);
        const started_at = Date.now();

        await db.query(
          `INSERT INTO entries (name, project_id, started_at) VALUES ($1, $2, $3)`,
          [payload.description ?? '', project_id, started_at],
        );

        self.postMessage({ action: 'timerStarted', payload: { started_at } });
        break;
      }

      // ── Stop timer ─────────────────────────────────────────────────────────
      case 'stopTimer': {
        const ended_at = Date.now();
        await db.exec(`UPDATE entries SET ended_at = ${ended_at} WHERE ended_at IS NULL`);
        self.postMessage({ action: 'timerStopped', payload: { ended_at } });
        break;
      }

      // ── Save manual entry ──────────────────────────────────────────────────
      case 'saveEntry': {
        const project_id = await upsertProject(payload.project);
        await db.query(
          `INSERT INTO entries (name, project_id, started_at, ended_at) VALUES ($1, $2, $3, $4)`,
          [payload.description ?? '', project_id, payload.started_at, payload.ended_at],
        );
        self.postMessage({ action: 'entrySaved' });
        break;
      }

      // ── Update entry ───────────────────────────────────────────────────────
      case 'updateEntry': {
        const project_id = await upsertProject(payload.project);
        await db.query(
          `UPDATE entries
           SET    name = $1, project_id = $2, started_at = $3, ended_at = $4
           WHERE  id = $5`,
          [payload.description, project_id, payload.started_at, payload.ended_at, payload.id],
        );
        self.postMessage({ action: 'entryUpdated' });
        break;
      }

      // ── Delete entry ───────────────────────────────────────────────────────
      case 'deleteEntry': {
        await db.query(`DELETE FROM entries WHERE id = $1`, [payload.id]);
        self.postMessage({ action: 'entryDeleted' });
        break;
      }

      // ── Update project (name + color) ──────────────────────────────────────
      case 'updateProject': {
        await db.query(
          `UPDATE projects SET name = $1, color = $2 WHERE id = $3`,
          [payload.name, payload.color, payload.id],
        );
        self.postMessage({ action: 'projectUpdated' });
        break;
      }

      // ── Delete project (un-assigns all entries first) ──────────────────────
      case 'deleteProject': {
        await db.query(`UPDATE entries SET project_id = NULL WHERE project_id = $1`, [payload.id]);
        await db.query(`DELETE FROM projects WHERE id = $1`, [payload.id]);
        self.postMessage({ action: 'projectDeleted' });
        break;
      }

      // ── Export CSV ─────────────────────────────────────────────────────────
      case 'exportCSV': {
        const { rows } = await db.query(`
          SELECT e.id, e.name, p.name AS project, e.started_at, e.ended_at
          FROM   entries e
          LEFT JOIN projects p ON p.id = e.project_id
          WHERE  e.ended_at IS NOT NULL
          ORDER  BY e.started_at DESC
        `);

        const esc = (s) => `"${String(s ?? '').replace(/"/g, '""')}"`;
        const lines = [
          'id,description,project,started_at,ended_at,duration_minutes',
          ...rows.map((r) => {
            const dur = Math.round((Number(r.ended_at) - Number(r.started_at)) / 60_000);
            return [
              r.id,
              esc(r.name),
              esc(r.project),
              new Date(Number(r.started_at)).toISOString(),
              new Date(Number(r.ended_at)).toISOString(),
              dur,
            ].join(',');
          }),
        ];

        self.postMessage({
          action: 'exportReady',
          payload: { format: 'csv', data: lines.join('\n') },
        });
        break;
      }

      // ── Export JSON ────────────────────────────────────────────────────────
      case 'exportJSON': {
        const { rows } = await db.query(`
          SELECT e.id, e.name, p.name AS project, e.started_at, e.ended_at
          FROM   entries e
          LEFT JOIN projects p ON p.id = e.project_id
          WHERE  e.ended_at IS NOT NULL
          ORDER  BY e.started_at DESC
        `);

        const entries = rows.map((r) => ({
          id:               r.id,
          description:      r.name,
          project:          r.project ?? null,
          started_at:       new Date(Number(r.started_at)).toISOString(),
          ended_at:         new Date(Number(r.ended_at)).toISOString(),
          duration_minutes: Math.round((Number(r.ended_at) - Number(r.started_at)) / 60_000),
        }));

        self.postMessage({
          action: 'exportReady',
          payload: { format: 'json', data: JSON.stringify(entries, null, 2) },
        });
        break;
      }

      default:
        console.warn('[worker] unknown action:', action);
    }
  } catch (err) {
    self.postMessage({ action: 'error', payload: err?.message ?? String(err) });
  }
};
