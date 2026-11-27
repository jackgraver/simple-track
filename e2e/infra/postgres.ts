import fs from 'node:fs';
import path from 'node:path';
import { Client } from 'pg';
import { run, sleep } from './exec';
import { BIN_DIR, CACHE_DIR, DB_PREFIX, ROOT, TEMPLATE_DB, TZ } from './paths';

/**
 * One long-lived Postgres container for every e2e run on this machine (all checkouts share it;
 * database names are prefixed per checkout). Data lives on tmpfs with durability off: fast, and
 * disposable by design. `npm run down` removes it.
 */
export const CONTAINER = 'simple-track-e2e';
const IMAGE = 'postgres:16-alpine'; // same as docker-compose.yml / prod

export const pgUrl = (port: number, db: string) =>
  `postgresql://postgres:postgres@127.0.0.1:${port}/${db}?sslmode=disable`;

/**
 * The server's login role (backend/internal/database AppRole), as in production: DML only, so
 * row-level security applies to it. `migrate up` (template build) gives it this password. The
 * role is cluster-wide, so every checkout sharing the container uses the same fixed password.
 * pgUrl (postgres) stays the owner for migrate, e2eseed, the /__e2e reset and the fixtures.
 */
export const APP_ROLE = 'simpletrack_app';
export const appPgUrl = (port: number, db: string) =>
  `postgresql://${APP_ROLE}:e2e_app@127.0.0.1:${port}/${db}?sslmode=disable`;

async function containerPort(): Promise<number | undefined> {
  try {
    const out = await run('docker', ['port', CONTAINER, '5432/tcp']);
    const m = out.match(/:(\d+)\s*$/m);
    return m ? Number(m[1]) : undefined;
  } catch {
    return undefined;
  }
}

async function containerState(): Promise<'running' | 'stopped' | 'missing'> {
  try {
    const out = await run('docker', ['inspect', '-f', '{{.State.Running}}', CONTAINER]);
    return out.trim() === 'true' ? 'running' : 'stopped';
  } catch {
    return 'missing';
  }
}

/** Start (or reuse) the container and wait until it accepts connections. Returns its host port. */
export async function ensurePostgres(): Promise<number> {
  const state = await containerState();
  if (state === 'stopped') {
    // tmpfs is wiped on stop, so a stopped container is as good as missing.
    await run('docker', ['rm', '-f', CONTAINER]);
  }
  if (state !== 'running') {
    // Several checkouts can get here at once; the loser of the name race just uses the winner's.
    await run('docker', [
      'run', '-d', '--name', CONTAINER,
      '-p', '127.0.0.1::5432', // random free host port; read back with `docker port`
      '--tmpfs', '/var/lib/postgresql/data:rw,size=2g',
      '--shm-size', '256m',
      '-e', 'POSTGRES_PASSWORD=postgres',
      IMAGE,
      '-c', 'fsync=off',
      '-c', 'synchronous_commit=off',
      '-c', 'full_page_writes=off',
      '-c', 'max_connections=300',
    ]).catch(async (e) => {
      if ((await containerState()) !== 'running') throw e;
    });
  }
  const port = await containerPort();
  if (!port) throw new Error(`could not read the host port of ${CONTAINER}`);
  const deadline = Date.now() + 30_000;
  for (;;) {
    const c = new Client({ connectionString: pgUrl(port, 'postgres') });
    try {
      await c.connect();
      await c.query('SELECT 1');
      await c.end();
      return port;
    } catch (e) {
      await c.end().catch(() => {});
      if (Date.now() > deadline) throw new Error(`Postgres in ${CONTAINER} not ready: ${e}`);
      await sleep(200);
    }
  }
}

export async function withAdmin<T>(port: number, fn: (c: Client) => Promise<T>): Promise<T> {
  const c = new Client({ connectionString: pgUrl(port, 'postgres') });
  await c.connect();
  try {
    return await fn(c);
  } finally {
    await c.end();
  }
}

/**
 * The template is migrations + seed. Rebuilt only when the backend key (Go + SQL sources, which
 * include the migrations and the seed) changes. The database's comment is `<key> <checkout path>`;
 * the path lets pruneOrphans() drop databases of checkouts that no longer exist.
 */
export async function ensureTemplate(port: number, backendKey: string): Promise<boolean> {
  const current = await withAdmin(port, async (c) => {
    const r = await c.query(
      `SELECT shobj_description(oid, 'pg_database') AS key FROM pg_database WHERE datname = $1`,
      [TEMPLATE_DB],
    );
    return (r.rows[0]?.key as string | undefined)?.split(' ')[0];
  });
  if (current === backendKey && process.env.E2E_REBUILD !== '1') return false;

  await withAdmin(port, async (c) => {
    await c.query(`DROP DATABASE IF EXISTS ${TEMPLATE_DB} WITH (FORCE)`);
    await c.query(`CREATE DATABASE ${TEMPLATE_DB}`);
  });
  const env = {
    ...process.env,
    DATABASE_OWNER_URL: pgUrl(port, TEMPLATE_DB),
    DATABASE_URL: appPgUrl(port, TEMPLATE_DB),
    TZ,
    APP_ENV: 'development',
  };
  await run(path.join(BIN_DIR, 'migrate'), ['up'], { cwd: CACHE_DIR, env });
  await run(path.join(BIN_DIR, 'e2eseed'), [], { cwd: CACHE_DIR, env });
  await withAdmin(port, (c) =>
    c.query(`COMMENT ON DATABASE ${TEMPLATE_DB} IS ${c.escapeLiteral(`${backendKey} ${ROOT}`)}`),
  );
  return true;
}

/** Drop and re-create a worker's database from the template. Retries: concurrent clones of one
 * template can briefly collide ("source database is being accessed by other users"). */
export async function cloneTemplate(port: number, db: string): Promise<void> {
  for (let attempt = 0; ; attempt++) {
    try {
      await withAdmin(port, async (c) => {
        await c.query(`DROP DATABASE IF EXISTS ${db} WITH (FORCE)`);
        await c.query(`CREATE DATABASE ${db} TEMPLATE ${TEMPLATE_DB}`);
      });
      return;
    } catch (e) {
      if (attempt >= 20) throw e;
      await sleep(50 + Math.random() * 150);
    }
  }
}

export async function dropDatabase(port: number, db: string): Promise<void> {
  await withAdmin(port, (c) => c.query(`DROP DATABASE IF EXISTS ${db} WITH (FORCE)`));
}

/** Plain-SQL data dump of one database (for failure artifacts). */
export function dumpData(db: string): Promise<string> {
  return run('docker', ['exec', CONTAINER, 'pg_dump', '-U', 'postgres', '--data-only', '--inserts', '--no-owner', db]);
}

/** Every database this checkout owns (template + workers). */
export async function ownDatabases(port: number): Promise<string[]> {
  return withAdmin(port, async (c) => {
    const r = await c.query(`SELECT datname FROM pg_database WHERE datname LIKE $1`, [`${DB_PREFIX}_%`]);
    return r.rows.map((x) => x.datname as string);
  });
}

/**
 * Drop every e2e database (template + workers) whose checkout path is gone: deleted worktrees and
 * CI snapshot checkouts would otherwise pile up in the container's tmpfs.
 */
export async function pruneOrphans(port: number): Promise<string[]> {
  return withAdmin(port, async (c) => {
    const r = await c.query(
      `SELECT datname, shobj_description(oid, 'pg_database') AS note FROM pg_database WHERE datname ~ '^e2e_[0-9a-f]{6}_template$'`,
    );
    const dropped: string[] = [];
    for (const { datname, note } of r.rows as { datname: string; note: string | null }[]) {
      const prefix = datname.replace(/_template$/, '');
      if (prefix === DB_PREFIX) continue;
      // Only drop what is provably orphaned: the comment names a checkout and that path is gone.
      // No path (older suite versions) means unknown owner: keep it.
      const root = note?.split(' ').slice(1).join(' ');
      if (!root || fs.existsSync(root)) continue;
      const all = await c.query(`SELECT datname FROM pg_database WHERE datname LIKE $1`, [`${prefix}_%`]);
      for (const { datname: db } of all.rows) {
        await c.query(`DROP DATABASE IF EXISTS ${db} WITH (FORCE)`);
        dropped.push(db);
      }
    }
    return dropped;
  });
}

/**
 * Per-test "what did this test change" support. At test start the fixtures record each table's
 * max(id); on failure, changedRows() returns rows the test created (id above the start max),
 * updated (updated_at <> created_at: seeded rows are written with both equal) or soft-deleted.
 * Small by construction, so it goes straight into failures.md.
 */
export interface TableInfo {
  name: string;
  hasTimestamps: boolean;
  hasDeletedAt: boolean;
}

export async function tablesWithId(c: Client): Promise<TableInfo[]> {
  const r = await c.query(`
    SELECT table_name AS name,
           bool_or(column_name = 'updated_at') AND bool_or(column_name = 'created_at') AS "hasTimestamps",
           bool_or(column_name = 'deleted_at') AS "hasDeletedAt"
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name <> 'schema_migrations'
    GROUP BY table_name HAVING bool_or(column_name = 'id') ORDER BY table_name`);
  return r.rows;
}

export async function maxIds(c: Client, tables: TableInfo[]): Promise<Map<string, number>> {
  if (!tables.length) return new Map();
  const sql = tables.map((t) => `SELECT '${t.name}' AS t, coalesce(max(id), 0)::bigint AS m FROM "${t.name}"`).join(' UNION ALL ');
  const r = await c.query(sql);
  return new Map(r.rows.map((x) => [x.t as string, Number(x.m)]));
}

export async function changedRows(c: Client, tables: TableInfo[], start: Map<string, number>, limit = 10): Promise<string> {
  const out: string[] = [];
  for (const t of tables) {
    const conds = [`id > ${start.get(t.name) ?? 0}`];
    if (t.hasTimestamps) conds.push('updated_at <> created_at');
    if (t.hasDeletedAt) conds.push('deleted_at IS NOT NULL');
    const r = await c.query(
      `SELECT to_jsonb(x) - 'created_at' - 'updated_at' AS row, count(*) OVER () AS n
       FROM "${t.name}" x WHERE ${conds.join(' OR ')} ORDER BY id LIMIT ${limit}`,
    );
    if (!r.rows.length) continue;
    const n = Number(r.rows[0].n);
    out.push(`${t.name} (${n} row${n === 1 ? '' : 's'}${n > limit ? `, first ${limit}` : ''}):`);
    for (const { row } of r.rows) out.push('  ' + JSON.stringify(row));
  }
  return out.length ? out.join('\n') : '(no rows created, updated or deleted)';
}
