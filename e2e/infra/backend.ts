import { spawn, type ChildProcess } from 'node:child_process';
import fs from 'node:fs';
import net from 'node:net';
import path from 'node:path';
import { sleep } from './exec';
import { BIN_DIR, CACHE_DIR, LOG_DIR, TZ } from './paths';

export interface Backend {
  url: string;
  port: number;
  logFile: string;
  process: ChildProcess;
  stop(): Promise<void>;
}

export function freePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = net.createServer();
    srv.unref();
    srv.on('error', reject);
    srv.listen(0, '127.0.0.1', () => {
      const { port } = srv.address() as net.AddressInfo;
      srv.close(() => resolve(port));
    });
  });
}

/** Env shared by every e2e backend process. Secrets come from global setup (one per run). */
export function backendEnv(databaseUrl: string, listen: string): NodeJS.ProcessEnv {
  return {
    PATH: process.env.PATH,
    HOME: process.env.HOME,
    TZ,
    APP_ENV: 'development',
    DATABASE_URL: databaseUrl,
    LISTEN_ADDR: listen,
    JWT_SECRET: process.env.E2E_JWT_SECRET,
    DEV_AUTH_TOKEN: process.env.E2E_AUTH_TOKEN,
    ALLOW_BYPASS: 'true',
    DEV_AUTH_USER: 'dev',
    CORS_ORIGINS: process.env.E2E_BASE_URL,
    // Login protection is per process; keep it out of the way of the auth specs.
    LOGIN_RATE_LIMIT_MAX_IP: '10000',
    LOGIN_RATE_LIMIT_MAX_USERNAME: '10000',
    LOGIN_SPRAY_MAX_USERNAMES: '10000',
  };
}

/** Start the e2e server binary against `databaseUrl` on a free port; resolves once /health is 200. */
export async function startBackend(
  name: string,
  databaseUrl: string,
  extraEnv: NodeJS.ProcessEnv = {},
): Promise<Backend> {
  fs.mkdirSync(LOG_DIR, { recursive: true });
  const logFile = path.join(LOG_DIR, `${name}.log`);
  for (let attempt = 0; ; attempt++) {
    const port = await freePort();
    const out = fs.openSync(logFile, 'w');
    const child = spawn(path.join(BIN_DIR, 'server'), [], {
      cwd: CACHE_DIR, // no .env here: everything comes from the env below
      env: { ...backendEnv(databaseUrl, `127.0.0.1:${port}`), ...extraEnv },
      stdio: ['ignore', out, out],
    });
    fs.closeSync(out);
    const url = `http://127.0.0.1:${port}`;
    let exited = false;
    child.on('exit', () => (exited = true));
    const deadline = Date.now() + 20_000;
    while (!exited && Date.now() < deadline) {
      try {
        const r = await fetch(`${url}/health`);
        if (r.ok) {
          return {
            url,
            port,
            logFile,
            process: child,
            stop: () =>
              new Promise<void>((resolve) => {
                if (child.exitCode !== null) return resolve();
                child.once('exit', () => resolve());
                child.kill('SIGTERM');
                setTimeout(() => child.kill('SIGKILL'), 3000).unref();
              }),
          };
        }
      } catch {
        // not listening yet
      }
      await sleep(25);
    }
    child.kill('SIGKILL');
    const log = fs.readFileSync(logFile, 'utf8');
    // The free port can be taken between probing and binding: retry on a fresh one.
    if (attempt < 3 && /address already in use/.test(log)) continue;
    throw new Error(`backend ${name} did not become healthy:\n${log.slice(-3000)}`);
  }
}
