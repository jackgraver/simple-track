# e2e: Playwright suite for simple-track

Stable reference: decisions, architecture, conventions. Status and next steps live in [TASKS.md](TASKS.md).
Update both in the same PR as the code.

## Goal

Pin down what simple-track does today with end-to-end tests, so that:

1. Behaviour is pinned: a regression fails a test instead of being noticed by hand.
2. Agents (and humans) get a fast, trustworthy verification loop after each change.
3. The UI overhaul and optimistic-UI work can change internals with a safety net underneath.

Cover **golden paths and data-loss risks** (log, edit, delete, persist across reload), not every
button. 30-50 well-chosen tests beat 300 shallow ones.

## Running

```
./bin/dev e2e [playwright args]                # from the repo root: installs deps if needed, runs the suite

cd e2e
npm ci && npx playwright install chromium     # once (no --with-deps: it needs sudo)
npm test                                       # full suite, desktop + mobile
npm run smoke                                  # @smoke only: the after-every-change check
npx playwright test tests/gym                  # one module ("lane")
npx playwright test --grep @gym                # same, by tag
npm run flake -- tests/gym/logging.spec.ts     # run a spec 10x to prove it is stable
npm run report                                 # open the last HTML report
npx playwright test tests/aria --update-snapshots   # accept an intended structure change
npm run ci:history                             # CI e2e record; is it ready to block?
npm run down                                   # remove the shared e2e Postgres container
```

`npm run down` removes the container for **every** checkout (other worktrees' runs included); a run
recreates it. Overriding the reporter (`--reporter=dot`) also drops the failure summary below.

```
```

Node comes from mise: prefix with `mise exec --` if `node` is not on PATH. One run per checkout at a
time (a lock in `.cache/`); different worktrees can run at once. The suite never touches the dev
stack, so it runs whether or not `dev start` is up.

Timings on the dev machine (24 cores, 6 workers): setup is ~60ms when everything is cached, ~3s for a
frontend rebuild, ~1.3s for a cold Postgres container + template. Each test costs ~0.5s (reset
~60ms included). The smoke set takes ~2.5s, and 280 tests take ~28s (~21s with 12 workers).

Environment knobs:

| Var | Effect |
|---|---|
| `E2E_WORKERS` | Parallel workers (default 6 locally, 2 in CI). |
| `E2E_FRONTEND=dev` | Serve the frontend from a vite dev server instead of the cached production build. Skips the ~3s rebuild after frontend edits, but every page load is slower (unbundled modules): measured ~2.5x slower per test. It rarely pays off; the production build is the default. |
| `E2E_KEEP=1` | Leave the per-worker databases in place after the run, for poking with psql. |
| `E2E_REBUILD=1` | Ignore build caches (frontend, Go binaries, template DB). |
| `E2E_BLOCKING=1` | `ci.sh` only: fail the pipeline step when the suite fails. |

## Architecture

```
globalSetup (once per run; each step cached by a content hash)
  ├─ go build -tags e2e  → .cache/bin/{server,migrate,e2eseed}
  ├─ vite build          → .cache/dist
  ├─ Postgres: long-lived container `simple-track-e2e` (tmpfs, fsync off)
  ├─ template DB `e2e_template` = migrations + e2eseed
  └─ front door: one small Node server, serves dist/ (or proxies to vite dev) and
     forwards /api to the backend named by the `e2e_backend` cookie

per worker (once per Playwright worker, ~1s)
  ├─ CREATE DATABASE e2e_w<N> TEMPLATE e2e_template
  └─ start .cache/bin/server on a free port against it (logs to a file)

per test
  ├─ POST /__e2e/reset: truncate + reseed (~tens of ms), clears clock/faults/holds
  ├─ browser context gets the auth cookie + the `e2e_backend` cookie
  └─ after: unused faults fail the test; on failure attach backend log slice + pg_dump
```

### Why a database per worker (not a user per worker)

The app is single-tenant: no domain table has a `user_id`, and several tables are unique per date
(`days`, `body_weight_logs`, `step_logs`, `account_balances`). Two tests touching "today" would collide
in a shared database. A database per worker gives full parallelism with no shared state. If
per-user data lands later, fixtures gain a `user` option; the database model stays.

### The long-lived Postgres

One container (`simple-track-e2e`, its own port and volume, never the dev DB) stays up between runs
so a run doesn't pay container boot. Its data lives on tmpfs (RAM) with `fsync`, `synchronous_commit`
and `full_page_writes` off: Postgres stops forcing writes to disk, which makes cloning, resets and
every write fast. Safe only because the data is disposable. Never use these settings elsewhere.

### Test-only backend surface (`-tags e2e`)

`backend/internal/e2e` is compiled only with the `e2e` build tag; the prod binary contains a no-op
stub. On top of that the server refuses to start with `E2E` set in production. It provides:

| Endpoint | Purpose |
|---|---|
| `POST /__e2e/reset` | Truncate every table, reseed, clear clock/faults/holds. |
| `POST /__e2e/clock` | Set the server's "now" (`backend/internal/clock`). |
| `POST /__e2e/faults` | Make DB statements fail (table + op), optionally `times: n`. |
| `POST /__e2e/holds` | Hold matching requests until the test releases, fails or drops them. |

## Principles

1. **Real stack, no mocks.** Vue, then API, then Postgres. No `route.fulfill` with made-up bodies.
2. **Trigger the real cause, never the response.** Every non-2xx has a cause; the test produces the
   cause and the real handler produces the response:

   | Cause | Example | How |
   |---|---|---|
   | Input | malformed body → 400 | send it |
   | State | duplicate name, missing id | arrange it (API, or `db` when the API can't) |
   | Dependency | DB error, slow or dropped response, AniList down | `e2e.fault`, `e2e.hold`, fake upstream |
   | Environment | expired session, date rules | fixed clock, real expired cookie |
   | Network | offline, connection refused | `context.setOffline()`, `route.abort()` |
   | Unreachable otherwise | failure between two writes | named failpoint (add only when needed) |

   Faults and holds are validated when armed (unknown route/table fails immediately) and an armed
   fault that never fired fails the test, so a typo can't silently pass.
3. **Isolated, deterministic data.** Each test starts from the seed. Never run against
   `simpletracker_dev`, never sync prod data (personal, non-deterministic, opaque).
4. **No logins.** The fixture sets the `DEV_AUTH_TOKEN` bypass cookie. Only `tests/auth/` drives the
   real login form.
5. **Control time on both sides.** Default "now" is the seed epoch (`E2E_EPOCH` in `fixtures/`), set
   on the server via `/__e2e/clock` and in the browser via `page.clock`. Backend `TZ` and Playwright
   `timezoneId` match (`America/Toronto`), deliberately not UTC so day-boundary bugs show.
6. **Arrange via API, act and assert via UI.** API helpers set up state. The action under test goes
   through the UI. Always verify persistence with a **reload** after a write.
7. **`db` fixture: narrow use only.** Read-only invariant checks the UI/API can't show ("double
   submit created exactly one row", "delete set `deleted_at`"), or writes for states the API can't
   produce. Never for ordinary arrangement.
8. **Role-first selectors.** `getByRole` / `getByLabel` first, `data-testid` as an escape hatch,
   never CSS or class selectors.
9. **Web-first assertions only.** `await expect(locator)...` auto-retries. No `waitForTimeout`. For
   slow/optimistic paths, use a hold instead of a delay: assert the in-flight state, then release.
   Optimistic rows appear before their request is even sent, so before a persistence reload, await
   `waitForApi(page, 'POST /...')` (started before the action). Skip retry backoff with
   `page.clock.runFor()` after the failed response lands, not before.
10. **Clean console.** Console errors, page errors and unexpected 5xx fail the test. Opt out per test
    with `test.use({ allowedErrors: [/pattern/] })`.
11. **Hermetic.** Requests to anything but the front door are blocked (Google Fonts included).
    Service workers blocked, `reducedMotion: 'reduce'`.
12. **Flakiness is a bug.** `retries: 0` locally, `1` in CI with `--fail-on-flaky-tests`. Run every new
    spec through `npm run flake` (10 repeats) before merging.
13. **Snapshots: structure, not pixels.** A few `toMatchAriaSnapshot` checks in `tests/aria/` (main
    nav, gym week strip), per project. Keep the set small while the UI overhaul churns markup.

## CI

`.cicd.yaml` runs `e2e/ci.sh` after the unit tests: the whole suite with `CI=1` (retries 1,
`--fail-on-flaky-tests`, dot + JSON reporters). Each run appends a line to
`~/.local/state/simple-track-e2e/ci-history.jsonl`, and a failed run's `test-results/` and
`playwright-report/` are kept next to it in `runs/` (the pipeline's checkout is thrown away).

The step is **non-blocking** (exits 0 on a red suite) until `E2E_BLOCKING=1` is set on it.
`npm run ci:history` prints the record and says when the bar is met: the last 20 runs all green
with no flaky tests.

The container is shared by every checkout, CI's snapshot checkouts included. Database names carry
a hash of the checkout path and each template's comment records that path, so setup drops the
databases of checkouts that no longer exist, never anyone else's.

## Data

| Layer | What | Why |
|---|---|---|
| Seed (`backend/internal/e2e/seed`) | Dev user, exercise library + active program, food catalogue, diet plan. Fixed dates relative to the epoch, fixed IDs (sequences restart). | Stable reference context the app needs to function. |
| Per test (API helpers) | Anything the test mutates or asserts on. | Isolation and readable failures. |

If a test writes it or asserts on its exact value, the test creates it. Never assert on seed contents
beyond "it exists".

**Seed profiles** add data on top of the seed for scale and edge cases: `test.use({ seed: 'large' })`
loads two years of history in every module plus awkward values (unicode and very long names, zero
and huge numbers, an overdrawn account), in ~300ms per test. `tests/scale/` runs every page against
it. The same data can go into any non-production database: `go run ./cmd/e2eseed -profile large`
(it empties the database first).

## The fixtures in one screen

```ts
import { expect, test, waitForApi } from '../../fixtures';

test('…', { tag: ['@gym'] }, async ({ page, api, e2e, db }) => {
  // page: authenticated, clock at E2E_EPOCH, hermetic, console-guarded
  // api:  arrange through the real API (fixtures/api.ts)
  // e2e:  e2e.fault({ table, op, times? , code? }), e2e.hold('POST /path', { phase }), e2e.setClock(date)
  // db:   pg Client on this worker's DB (narrow use)
});
test.use({ now: new Date('…') });         // a different "now" (null = real time)
test.use({ allowedErrors: [/…/] });       // tolerate specific browser errors
test.use({ authenticated: false });       // no auth cookie (login UI tests)
```

Worked examples of every failure mode: `tests/examples/failure-modes.spec.ts`. Self-tests that prove
the guards still fail tests: `tests/examples/guards.spec.ts` (`@infra`).

## Reading a failure

Start with **`test-results/failures.md`**: one block per failed or flaky test with the error, the
failing line, a rerun command, the last API calls, backend errors and the artifact paths. Each failed
test also gets, in `test-results/<test>/` and the HTML report:

- `trace.zip`: DOM snapshots, network (with response bodies), console, one step per action.
  `npx playwright show-trace <file>`.
- `backend.log`: the worker backend's log lines while that test ran.
- `db.sql`: `pg_dump --data-only` of the worker DB at the moment of failure.

Every test starts from the seed with a fixed clock, so re-running the single test reproduces it.

## Layout

```
e2e/
  playwright.config.ts
  global-setup.ts             # builds, Postgres, template DB, front door
  ci.sh                       # the j-cicd step (history, kept reports, E2E_BLOCKING)
  infra/                      # build cache, postgres, backend process, front door, summary reporter
  .cache/                     # (gitignored) binaries, dist, backend logs, run lock
  fixtures/
    index.ts                  # extended `test`: worker stack, auth, reset, clock, guards, artifacts
    api.ts                    # typed arrange helpers, grown per module
    e2e.ts                    # /__e2e client: reset, clock, fault, hold
  pages/                      # page objects ONLY for complex screens (gym logging)
  tests/
    smoke/  examples/  scale/  aria/  auth/  gym/  diet/  finances/  media/  grocery/  settings/
```

Mirrors `frontend/src/pages`, so an agent knows which lane to run for the module it touched. Prefer
fixtures over page objects; don't wrap every locator.

## Adding a test

1. Put it under `tests/<module>/`, tag it (`{ tag: ['@gym'] }`, plus `@smoke` if it belongs there).
2. Arrange with `api` helpers (add one to `fixtures/api.ts` if missing), act through the UI, assert,
   await the write (`waitForApi`), then `page.reload()` and assert again.
3. `npm run flake -- <spec>` before committing.
