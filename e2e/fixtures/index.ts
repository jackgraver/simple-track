import fs from 'node:fs';
import { test as base, expect, request as pwRequest, type Page } from '@playwright/test';
import { Client } from 'pg';
import { startBackend, type Backend } from '../infra/backend';
import { BACKEND_COOKIE } from '../infra/frontdoor';
import { workerDb } from '../infra/paths';
import { cloneTemplate, dropDatabase, dumpData, pgUrl } from '../infra/postgres';
import { Api } from './api';
import { E2E } from './e2e';

export { expect };

/**
 * The suite's default "now": a Thursday at noon in America/Toronto (EST, UTC-5 in January).
 * Mirrors seed.Epoch() in backend/internal/e2e/seed.
 */
export const E2E_EPOCH = new Date('2026-01-15T12:00:00-05:00');

/** The backend's auth cookie; the e2e backend accepts the DEV_AUTH_TOKEN bypass value in it. */
const AUTH_COOKIE = 'auth_token';

interface WorkerStack {
  db: string;
  dbUrl: string;
  backend: Backend;
}

interface Options {
  /** Console errors / page errors to tolerate in this test (README principle 10). */
  allowedErrors: RegExp[];
  /** "Now" for both the server and the browser. `null` = real time, no fake browser clock. */
  now: Date | null;
  /** Set the auth cookie (false for tests of the login UI). */
  authenticated: boolean;
  /** Data on top of the base seed: 'large' = two years of history and odd values (~0.5s per test). */
  seed: 'base' | 'large';
}

interface Fixtures {
  /** Control the backend: faults, holds, clock (backend/internal/e2e). */
  e2e: E2E;
  /** Arrange helpers over the real API, authenticated. */
  api: Api;
  /** Direct Postgres access to this worker's DB. Narrow use only: see README principle 7. */
  db: Client;
  _isolate: void;
}

const PG_PORT = () => Number(process.env.E2E_PG_PORT);
const authCookieHeader = () => `${AUTH_COOKIE}=${process.env.E2E_AUTH_TOKEN}`;

export const test = base.extend<Fixtures & Options, { stack: WorkerStack }>({
  allowedErrors: [[], { option: true }],
  now: [E2E_EPOCH, { option: true }],
  authenticated: [true, { option: true }],
  seed: ['base', { option: true }],

  baseURL: async ({}, use) => use(process.env.E2E_BASE_URL),

  // One database + backend per worker, cloned from the template.
  stack: [
    async ({}, use, workerInfo) => {
      const db = workerDb(workerInfo.parallelIndex);
      await cloneTemplate(PG_PORT(), db);
      const dbUrl = pgUrl(PG_PORT(), db);
      const backend = await startBackend(`w${workerInfo.parallelIndex}`, dbUrl);
      await use({ db, dbUrl, backend });
      await backend.stop();
      if (process.env.E2E_KEEP !== '1') await dropDatabase(PG_PORT(), db);
    },
    { scope: 'worker', timeout: 60_000 },
  ],

  e2e: async ({ stack }, use) => {
    const ctx = await pwRequest.newContext();
    await use(new E2E(ctx, stack.backend.url));
    await ctx.dispose();
  },

  api: async ({ stack }, use) => {
    const ctx = await pwRequest.newContext({ extraHTTPHeaders: { cookie: authCookieHeader() } });
    await use(new Api(ctx, stack.backend.url));
    await ctx.dispose();
  },

  db: async ({ stack }, use) => {
    const client = new Client({ connectionString: stack.dbUrl });
    await client.connect();
    await use(client);
    await client.end();
  },

  // Before each test: fresh seed data and a fixed clock. After: fail on leftover faults/holds,
  // and on failure attach the backend log slice and a data dump (README "Reading a failure").
  _isolate: [
    async ({ stack, e2e, now, seed }, use, testInfo) => {
      await e2e.reset();
      if (seed !== 'base') await e2e.seedProfile(seed);
      if (now) await e2e.setClock(now);
      const logStart = fs.statSync(stack.backend.logFile).size;

      await use();

      if (testInfo.status === 'passed') {
        const leftovers = await e2e.leftovers();
        if (leftovers.length) throw new Error(`e2e leftovers:\n  ${leftovers.join('\n  ')}`);
      }
      if (testInfo.status !== testInfo.expectedStatus) {
        // Written next to trace.zip in test-results/<test>/, and linked from the HTML report.
        const logPath = testInfo.outputPath('backend.log');
        fs.writeFileSync(logPath, fs.readFileSync(stack.backend.logFile).subarray(logStart));
        await testInfo.attach('backend.log', { path: logPath, contentType: 'text/plain' });
        const dumpPath = testInfo.outputPath('db.sql');
        fs.writeFileSync(dumpPath, await dumpData(stack.db).catch((e) => `pg_dump failed: ${e}`));
        await testInfo.attach('db.sql', { path: dumpPath, contentType: 'text/plain' });
      }
    },
    { auto: true },
  ],

  context: async ({ context, stack, baseURL, authenticated, _isolate }, use) => {
    void _isolate; // ordering: the DB is reset before the browser exists
    const cookies = [{ name: BACKEND_COOKIE, value: String(stack.backend.port), url: baseURL! }];
    if (authenticated) cookies.push({ name: AUTH_COOKIE, value: process.env.E2E_AUTH_TOKEN!, url: baseURL! });
    await context.addCookies(cookies);

    // Hermetic: nothing leaves the front door. External stylesheets/fonts get an empty 200 so the
    // page doesn't log load errors for them.
    await context.route(
      (url) => !url.href.startsWith(baseURL!) && url.protocol.startsWith('http'),
      (route) =>
        route.request().resourceType() === 'stylesheet'
          ? route.fulfill({ status: 200, contentType: 'text/css', body: '' })
          : route.abort('blockedbyclient'),
    );
    await use(context);
  },

  page: async ({ page, now, allowedErrors, e2e }, use) => {
    if (now) await page.clock.install({ time: now });
    const problems = watchConsole(page);
    await use(page);
    const unexpected = problems.filter(
      (p) =>
        !allowedErrors.some((re) => re.test(p)) &&
        !(e2e.expectServerErrors && /\b5\d\d\b/.test(p)) &&
        !(e2e.expectNetworkErrors && /net::ERR_|Network Error|Failed to fetch/i.test(p)),
    );
    if (unexpected.length && test.info().status === 'passed') {
      throw new Error(
        `Unexpected browser errors (allow with test.use({ allowedErrors: [/.../] })):\n  ${unexpected.join('\n  ')}`,
      );
    }
  },
});

/** Console errors, uncaught page errors and 5xx responses seen by the page. */
function watchConsole(page: Page): string[] {
  const out: string[] = [];
  page.on('console', (m) => {
    if (m.type() === 'error') out.push(`console.error: ${m.text()}`);
  });
  page.on('pageerror', (e) => out.push(`pageerror: ${e.message}`));
  page.on('response', (r) => {
    if (r.status() >= 500) out.push(`HTTP ${r.status()} ${r.request().method()} ${r.url()}`);
  });
  return out;
}

/**
 * Resolves when the page gets a successful response for `route` ("POST /tracking/grocery/items",
 * backend path, same syntax as e2e.hold; ":param" segments match anything). Start it before the
 * action, await it before a reload: an optimistic row is on screen before its request is even sent,
 * so reloading right after it appears can cancel the write.
 */
export function waitForApi(page: Page, route: string) {
  const [method, path] = route.trim().split(/\s+/);
  const re = new RegExp(`/api${path.replace(/:[^/]+/g, '[^/]+')}(\\?|$)`);
  return page.waitForResponse(
    (r) => r.request().method() === method.toUpperCase() && re.test(new URL(r.url()).pathname + '?') && r.ok(),
  );
}
