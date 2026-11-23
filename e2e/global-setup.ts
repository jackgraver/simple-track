import { spawn } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import type { FullConfig } from '@playwright/test';
import { buildBackend, buildFrontend, ensureFrontendDeps } from './infra/build';
import { sleep } from './infra/exec';
import { startFrontDoor } from './infra/frontdoor';
import { CACHE_DIR, DIST_DIR, FRONTEND_DIR } from './infra/paths';
import { ensurePostgres, ensureTemplate } from './infra/postgres';
import { freePort } from './infra/backend';

const LOCK = path.join(CACHE_DIR, 'run.lock');

/** One run per checkout at a time: worker databases are named by checkout + worker index. */
function takeLock(): void {
  fs.mkdirSync(CACHE_DIR, { recursive: true });
  try {
    const pid = Number(fs.readFileSync(LOCK, 'utf8'));
    process.kill(pid, 0);
    throw new Error(`another e2e run (pid ${pid}) is active in this checkout`);
  } catch (e) {
    if (e instanceof Error && e.message.startsWith('another e2e run')) throw e;
  }
  fs.writeFileSync(LOCK, String(process.pid));
}

async function startViteDev(): Promise<{ url: string; stop: () => void }> {
  await ensureFrontendDeps();
  const port = await freePort();
  const child = spawn('npx', ['vite', '--host', '127.0.0.1', '--port', String(port), '--strictPort'], {
    cwd: FRONTEND_DIR,
    stdio: 'ignore',
  });
  const url = `http://127.0.0.1:${port}`;
  for (let i = 0; i < 300; i++) {
    try {
      if ((await fetch(url)).ok) return { url, stop: () => child.kill() };
    } catch {
      // starting
    }
    await sleep(100);
  }
  child.kill();
  throw new Error('vite dev server did not start');
}

export default async function globalSetup(_config: FullConfig) {
  const t0 = Date.now();
  takeLock();
  process.env.E2E_AUTH_TOKEN ??= randomBytes(24).toString('hex');
  process.env.E2E_JWT_SECRET ??= randomBytes(32).toString('hex');

  const devMode = process.env.E2E_FRONTEND === 'dev';
  const [backend, frontend, pgPort] = await Promise.all([
    buildBackend(),
    devMode ? startViteDev() : buildFrontend(),
    ensurePostgres(),
  ]);
  const templateRebuilt = await ensureTemplateTimed(pgPort, backend.key);

  const door = await startFrontDoor('url' in frontend ? { devServer: frontend.url } : { dist: DIST_DIR });
  process.env.E2E_BASE_URL = door.url;
  process.env.E2E_PG_PORT = String(pgPort);

  const fe = 'url' in frontend ? 'vite dev' : frontend.built ? `built ${frontend.ms}ms` : 'cached';
  console.log(
    `e2e: setup ${Date.now() - t0}ms (backend ${backend.built ? `built ${backend.ms}ms` : 'cached'}, ` +
      `frontend ${fe}, template ${templateRebuilt}) → ${door.url}`,
  );

  return async () => {
    await door.close();
    if ('stop' in frontend) frontend.stop();
    fs.rmSync(LOCK, { force: true });
  };
}

async function ensureTemplateTimed(port: number, key: string): Promise<string> {
  const t = Date.now();
  return (await ensureTemplate(port, key)) ? `rebuilt ${Date.now() - t}ms` : 'cached';
}
