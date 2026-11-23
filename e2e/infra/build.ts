import fs from 'node:fs';
import path from 'node:path';
import { run } from './exec';
import { BACKEND_DIR, BIN_DIR, CACHE_DIR, DIST_DIR, FRONTEND_DIR, hashFiles, readKey } from './paths';

const force = () => process.env.E2E_REBUILD === '1';

export interface BuildResult {
  key: string;
  built: boolean;
  ms: number;
}

/**
 * Go binaries built with -tags e2e: server, migrate, e2eseed. The key covers every Go file, the
 * module files and the SQL migrations (which are embedded), so it also keys the template DB.
 */
export async function buildBackend(): Promise<BuildResult> {
  const t0 = Date.now();
  const key = hashFiles([BACKEND_DIR], (f) => /\.(go|sql)$/.test(f) || /go\.(mod|sum)$/.test(f));
  const keyFile = path.join(BIN_DIR, '.key');
  const bins = ['server', 'migrate', 'e2eseed'].map((b) => path.join(BIN_DIR, b));
  if (!force() && readKey(keyFile) === key && bins.every((b) => fs.existsSync(b))) {
    return { key, built: false, ms: Date.now() - t0 };
  }
  fs.mkdirSync(BIN_DIR, { recursive: true });
  await run('go', ['build', '-tags', 'e2e', '-o', BIN_DIR + '/', './cmd/server', './cmd/migrate', './cmd/e2eseed'], {
    cwd: BACKEND_DIR,
  });
  fs.writeFileSync(keyFile, key);
  return { key, built: true, ms: Date.now() - t0 };
}

/** Production frontend build into .cache/dist. */
export async function buildFrontend(): Promise<BuildResult> {
  const t0 = Date.now();
  const key = hashFiles(
    ['src', 'public', 'index.html', 'vite.config.ts', 'package-lock.json', 'tsconfig.json', '.env', '.env.production'].map(
      (p) => path.join(FRONTEND_DIR, p),
    ),
    () => true,
  ) + '-sourcemap';
  const keyFile = path.join(CACHE_DIR, 'dist.key');
  if (!force() && readKey(keyFile) === key && fs.existsSync(path.join(DIST_DIR, 'index.html'))) {
    return { key, built: false, ms: Date.now() - t0 };
  }
  await ensureFrontendDeps();
  // Source maps let the console guard report frontend errors as src/ file:line (infra/sourcemap.ts).
  // The bundle itself is the normal minified production build.
  await run('npx', ['vite', 'build', '--sourcemap', '--outDir', DIST_DIR, '--emptyOutDir', '--logLevel', 'warn'], {
    cwd: FRONTEND_DIR,
  });
  fs.writeFileSync(keyFile, key);
  return { key, built: true, ms: Date.now() - t0 };
}

/** A fresh worktree has no frontend/node_modules. Install it once (keyed by the lockfile). */
export async function ensureFrontendDeps(): Promise<void> {
  const lock = path.join(FRONTEND_DIR, 'package-lock.json');
  const marker = path.join(FRONTEND_DIR, 'node_modules', '.package-lock.json');
  if (fs.existsSync(marker) && fs.statSync(marker).mtimeMs >= fs.statSync(lock).mtimeMs) return;
  await run('npm', ['ci', '--no-audit', '--no-fund'], { cwd: FRONTEND_DIR });
}
