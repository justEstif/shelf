// worker.ts — TypeScript source for the LocalTrack DB worker.
//
// Compile with:
//   bun build worker.ts --outfile worker.js --target browser
//
// The pre-built worker.js uses the CDN import below as a fallback.
// When bundling with Bun the npm package is used instead:
//   bun add @electric-sql/pglite

import { PGlite } from '@electric-sql/pglite';

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

type Period = 'today' | 'week' | 'month' | 'all';

interface WorkerMessage {
  action: string;
  payload: Record<string, unknown>;
}

interface EntryRow {
  id: number;
  name: string;
  started_at: bigint;
  ended_at: bigint | null;
  project_id: number | null;
  project_name: string | null;
  project_color: string | null;
}

interface ProjectRow {
  id: number;
  name: string;
  color: string;
  total_ms: bigint;
}

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

function periodBound(period: Period): number | null {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (period === 'today') return today.getTime();
  if (period === 'week') {
    const d = new Date(today);
    d.setDate(d.getDate() - d.getDay());
    return d.getTime();
  }
  if (period === 'month') {
    return new Date(today.getFullYear(), today.getMonth(), 1).getTime();
  }
  return null;
}

async function upsertProject(name: string | undefined): Promise<number | null> {
  if (!name?.trim()) return null;
  const trimmed = name.trim();
  await db.query(
    `INSERT INTO projects (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`,
    [trimmed],
  );
  const r = await db.query<{ id: number }>(
    `SELECT id FROM projects WHERE name = $1`,
    [trimmed],
  );
  return r.rows[0]?.id ?? null;
}

const ENTRY_SELECT = `
  SELECT e.id, e.name, e.started_at, e.ended_at,
         p.id    AS project_id,
         p.name  AS project_name,
         p.color AS project_color
  FROM   entries e
  LEFT JOIN projects p ON p.id = e.project_id
`;

// ─────────────────────────────────────────────────────────────────────────────
// Dispatch
// ─────────────────────────────────────────────────────────────────────────────

self.onmessage = async ({ data }: MessageEvent<WorkerMessage>) => {
  const { action, payload = {} } = data;

  try {
    switch (action) {

      case 'getEntries': {
        const from = periodBound((payload.period as Period) ?? 'today');
        const { rows } = from
          ? await db.query<EntryRow>(
              ENTRY_SELECT + ` WHERE e.started_at >= $1 ORDER BY e.started_at DESC`,
              [from],
            )
          : await db.query<EntryRow>(
              ENTRY_SELECT + ` ORDER BY e.started_at DESC`,
            );
        self.postMessage({ action: 'entries', payload: rows });
        break;
      }

      case 'getProjects': {
        const { rows } = await db.query<ProjectRow>(`
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

      case 'getRunningEntry': {
        const { rows } = await db.query<EntryRow>(
          ENTRY_SELECT + ` WHERE e.ended_at IS NULL ORDER BY e.started_at DESC LIMIT 1`,
        );
        self.postMessage({ action: 'runningEntry', payload: rows[0] ?? null });
        break;
      }

      case 'startTimer': {
        await db.exec(`UPDATE entries SET ended_at = ${Date.now()} WHERE ended_at IS NULL`);
        const project_id = await upsertProject(payload.project as string);
        const started_at = Date.now();
        await db.query(
          `INSERT INTO entries (name, project_id, started_at) VALUES ($1, $2, $3)`,
          [payload.description ?? '', project_id, started_at],
        );
        self.postMessage({ action: 'timerStarted', payload: { started_at } });
        break;
      }

      case 'stopTimer': {
        const ended_at = Date.now();
        await db.exec(`UPDATE entries SET ended_at = ${ended_at} WHERE ended_at IS NULL`);
        self.postMessage({ action: 'timerStopped', payload: { ended_at } });
        break;
      }

      case 'saveEntry': {
        const project_id = await upsertProject(payload.project as string);
        await db.query(
          `INSERT INTO entries (name, project_id, started_at, ended_at) VALUES ($1, $2, $3, $4)`,
          [payload.description ?? '', project_id, payload.started_at, payload.ended_at],
        );
        self.postMessage({ action: 'entrySaved' });
        break;
      }

      case 'updateEntry': {
        const project_id = await upsertProject(payload.project as string);
        await db.query(
          `UPDATE entries
           SET    name = $1, project_id = $2, started_at = $3, ended_at = $4
           WHERE  id = $5`,
          [payload.description, project_id, payload.started_at, payload.ended_at, payload.id],
        );
        self.postMessage({ action: 'entryUpdated' });
        break;
      }

      case 'deleteEntry': {
        await db.query(`DELETE FROM entries WHERE id = $1`, [payload.id]);
        self.postMessage({ action: 'entryDeleted' });
        break;
      }

      case 'updateProject': {
        await db.query(
          `UPDATE projects SET name = $1, color = $2 WHERE id = $3`,
          [payload.name, payload.color, payload.id],
        );
        self.postMessage({ action: 'projectUpdated' });
        break;
      }

      case 'deleteProject': {
        await db.query(`UPDATE entries SET project_id = NULL WHERE project_id = $1`, [payload.id]);
        await db.query(`DELETE FROM projects WHERE id = $1`, [payload.id]);
        self.postMessage({ action: 'projectDeleted' });
        break;
      }

      case 'exportCSV': {
        const { rows } = await db.query<{
          id: number; name: string; project: string | null;
          started_at: bigint; ended_at: bigint;
        }>(`
          SELECT e.id, e.name, p.name AS project, e.started_at, e.ended_at
          FROM   entries e
          LEFT JOIN projects p ON p.id = e.project_id
          WHERE  e.ended_at IS NOT NULL
          ORDER  BY e.started_at DESC
        `);

        const esc = (s: string | null | undefined) =>
          `"${String(s ?? '').replace(/"/g, '""')}"`;

        const lines = [
          'id,description,project,started_at,ended_at,duration_minutes',
          ...rows.map((r) => {
            const dur = Math.round((Number(r.ended_at) - Number(r.started_at)) / 60_000);
            return [
              r.id, esc(r.name), esc(r.project),
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

      case 'exportJSON': {
        const { rows } = await db.query<{
          id: number; name: string; project: string | null;
          started_at: bigint; ended_at: bigint;
        }>(`
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
    self.postMessage({ action: 'error', payload: (err as Error)?.message ?? String(err) });
  }
};
