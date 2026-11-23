import path from 'node:path';
import { Client } from 'pg';
import { run, sleep } from './exec';
import { BIN_DIR, CACHE_DIR, DB_PREFIX, TEMPLATE_DB, TZ } from './paths';

/**
 * One long-lived Postgres container for every e2e run on this machine (all checkouts share it;
 * database names are prefixed per checkout). Data lives on tmpfs with durability off: fast, and
 * disposable by design. `npm run down` removes it.
 */
export const CONTAINER = 'simple-track-e2e';
const IMAGE = 'postgres:16-alpine'; // same as docker-compose.yml / prod

export const pgUrl = (port: number, db: string) =>
  `postgresql://postgres:postgres@127.0.0.1:${port}/${db}?sslmode=disable`;

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
    ]);
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
 * include the migrations and the seed) changes; the key is stored as the database's comment.
 */
export async function ensureTemplate(port: number, backendKey: string): Promise<boolean> {
  const current = await withAdmin(port, async (c) => {
    const r = await c.query(
      `SELECT shobj_description(oid, 'pg_database') AS key FROM pg_database WHERE datname = $1`,
      [TEMPLATE_DB],
    );
    return r.rows[0]?.key as string | undefined;
  });
  if (current === backendKey && process.env.E2E_REBUILD !== '1') return false;

  await withAdmin(port, async (c) => {
    await c.query(`DROP DATABASE IF EXISTS ${TEMPLATE_DB} WITH (FORCE)`);
    await c.query(`CREATE DATABASE ${TEMPLATE_DB}`);
  });
  const env = { ...process.env, DATABASE_URL: pgUrl(port, TEMPLATE_DB), TZ, APP_ENV: 'development' };
  await run(path.join(BIN_DIR, 'migrate'), ['up'], { cwd: CACHE_DIR, env });
  await run(path.join(BIN_DIR, 'e2eseed'), [], { cwd: CACHE_DIR, env });
  await withAdmin(port, (c) => c.query(`COMMENT ON DATABASE ${TEMPLATE_DB} IS '${backendKey}'`));
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
