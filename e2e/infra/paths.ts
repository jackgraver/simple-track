import { createHash } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

export const E2E_DIR = path.resolve(__dirname, '..');
export const ROOT = path.resolve(E2E_DIR, '..');
export const BACKEND_DIR = path.join(ROOT, 'backend');
export const FRONTEND_DIR = path.join(ROOT, 'frontend');
export const CACHE_DIR = path.join(E2E_DIR, '.cache');
export const BIN_DIR = path.join(CACHE_DIR, 'bin');
export const DIST_DIR = path.join(CACHE_DIR, 'dist');
export const LOG_DIR = path.join(CACHE_DIR, 'logs');

/** Server-local time zone for the backend and the browser. Deliberately not UTC. */
export const TZ = 'America/Toronto';

/**
 * Database names are prefixed per checkout so several worktrees can share the one e2e Postgres
 * container without touching each other's databases.
 */
export const DB_PREFIX = `e2e_${createHash('sha1').update(ROOT).digest('hex').slice(0, 6)}`;
export const TEMPLATE_DB = `${DB_PREFIX}_template`;
export const workerDb = (i: number) => `${DB_PREFIX}_w${i}`;

const SKIP_DIRS = new Set(['node_modules', 'dist', 'tmp', '.cache', 'storybook-static', 'test-results']);

/** sha256 over the relative paths and contents of every matching file under `roots` (files or dirs). */
export function hashFiles(roots: string[], include: (file: string) => boolean): string {
  const files: string[] = [];
  const walk = (p: string) => {
    if (!fs.existsSync(p)) return;
    const st = fs.statSync(p);
    if (st.isFile()) {
      if (include(p)) files.push(p);
      return;
    }
    for (const e of fs.readdirSync(p, { withFileTypes: true })) {
      if (e.isDirectory() && SKIP_DIRS.has(e.name)) continue;
      walk(path.join(p, e.name));
    }
  };
  roots.forEach(walk);
  files.sort();
  const h = createHash('sha256');
  for (const f of files) {
    h.update(path.relative(ROOT, f));
    h.update('\0');
    h.update(fs.readFileSync(f));
    h.update('\0');
  }
  return h.digest('hex').slice(0, 16);
}

export function readKey(file: string): string | undefined {
  try {
    return fs.readFileSync(file, 'utf8').trim();
  } catch {
    return undefined;
  }
}
