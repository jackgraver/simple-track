# e2e tasks

Stable rules live in [README.md](README.md). Tick boxes in the same PR as the code. Work ships as
stacked PRs; infrastructure/architecture/foundation can share one.

Order: infrastructure (A), then scenarios module by module (B), then CI and scale (C).

## A. Infrastructure

- [x] **A1. Scaffold**: `e2e/package.json` (own deps), `tsconfig`, `playwright.config.ts` (desktop +
      mobile Chromium, retries 0 local / 1 CI, trace/screenshot on failure, hermetic context options),
      `.gitignore`, npm scripts (`test`, `smoke`, `flake`, `report`, `down`).
- [x] **A2. Isolated stack**: content-hash build cache (Go `-tags e2e` binaries, vite build),
      long-lived tmpfs Postgres container, `e2e_template` (migrate + seed), front door (static dist or
      vite dev proxy, `/api` routed by cookie), per-worker DB clone + backend.
- [x] **A3. Clock**: `backend/internal/clock` used by every domain `time.Now()` and gorm's `NowFunc`;
      settable only through `/__e2e/clock`.
- [x] **A4. e2e backend package** (`-tags e2e`, no-op stub otherwise, refused in production):
      reset, clock, DB faults (gorm hooks), holds (before/after phase; release, fail, drop; safety
      timeout), arm-time validation, unused-fault check, Go tests.
- [x] **A5. Seed**: `backend/internal/e2e/seed` + `cmd/e2eseed`: dev user, exercises + active program,
      foods, diet plan; dates relative to the epoch; refuses production.
- [x] **A6. Fixtures**: worker stack, bypass auth cookie, per-test reset, clock on both sides,
      console/pageerror/5xx guard with opt-out, `db` fixture, `e2e` client (fault/hold).
- [x] **A7. Failure artifacts**: backend log slice + `pg_dump` attached on failure.
- [x] **A8. First green tests**: `@smoke` (app loads, every top-level route renders) on both projects,
      plus one example spec per failure mode (DB fault, hold before/after, offline) on grocery add,
      and `@infra` self-tests proving each guard still fails tests.
- [x] **A9. Flake guard**: `npm run flake` over smoke passes; record timings in the decisions log.
- [x] **A10. Docs pointer**: one line in the project `CLAUDE.md` pointing at `e2e/README.md`.

## B. Scenarios (golden paths + data-loss risks)

For each module: write the scenario list here first, then the specs. Every write is verified after a
reload. Add the module's API helpers to `fixtures/api.ts` as needed.

- [x] **B1. Smoke**: one write per module on top of A8; `@smoke` stays under a minute.
      `tests/smoke/writes.spec.ts`, each verified after a reload: gym (log steps), diet (custom water
      amount), finances (add account), media (add item), grocery (add item), settings (save body
      profile). Whole `@smoke` set: 40 tests in ~8s.
- [ ] **B2. Gym** (first: most active and riskiest): log a set, edit, delete, history, persistence
- [ ] **B1. Smoke**: one write per module on top of A8; `@smoke` stays under a minute.
- [x] **B2. Gym** (first: most active and riskiest): log a set, edit, delete, history, persistence
      across reload, mobile sticky log bar, retry after dropped response (no duplicate set).
      Specs: `tests/gym/logging.spec.ts` (per-exercise logger, Back Squat on today's "Legs" plan).
      Scenarios (done unless marked):
      - log a set: shows at once, one row in DB, survives reload
      - edit: tap a saved set, change reps, save; persists
      - delete: removes only that set; persists
      - history: last session's set shows as "Last" next to today's
      - mobile: on a short phone the Log button is pinned above the tab bar (mobile project only)
      - dropped response (hold after + drop): retry with the same client_id, exactly one set
      - todo: weight/plates edit, notes, finish + "Up next", cardio/mobility, add/remove exercise,
        offline log, persistent save failure ("Didn't save" + Retry), plans pages, progression, weight, steps
- [ ] **B3. Diet**
- [ ] **B4. Finances**
- [ ] **B5. Media**: needs a fake upstream for the lookup APIs (AniList etc., URLs are configurable).
- [ ] **B6. Grocery**
- [ ] **B7. Settings and auth**: login UI (no bypass cookie), logout, session expiry (via clock).
- [ ] **B8. Optimistic UI**: hold-based assertions of instant state and post-reload state.

## C. CI and scale

- [x] **C1. CI step**: `e2e` step in `.cicd.yaml` runs `e2e/ci.sh`: CI settings, run history in
      `~/.local/state/simple-track-e2e/ci-history.jsonl`, failed runs' reports kept. Non-blocking.
- [ ] **C2. Promote to blocking**: set `E2E_BLOCKING=1` on the step once `npm run ci:history` reports
      the last 20 runs clean (green, no flaky). Waiting on run history.
- [x] **C3. `large` seed profile**: `seed.Large` (two years of history in every module, odd values),
      `test.use({ seed: 'large' })`, `POST /__e2e/seed/large`, `e2eseed -profile large`; `tests/scale/`
      renders every page against it. Timing budgets wait for the performance project.
- [x] **C4. Aria snapshots**: `tests/aria/` covers the main nav and the gym week strip, per project.
      Grow only for regions where losing an item would go unnoticed.
- [x] **C5. Failure summary reporter**: `test-results/failures.md`. Revisit its contents after agents
      have used it on real failures.

## Findings (bugs the suite surfaced; not fixed here)

- Optimistic rows appear before their request is sent, so a reload right after the row shows can
  cancel the write (found as a mobile-only flake in the first golden-path spec). Not a realistic user
  path, but specs must `waitForApi` before reloading. Idea: mark pending rows with `aria-busy="true"`;
  it is accessible and gives tests a UI signal to wait on instead of a network one.

- `POST /diet/foods` with an existing name returns **500**: the unique index `idx_foods_name` rejects it
  and `PostFood` maps every error to `apierr.Internal`. Decide the intended status (409?), then pin it.

- Gym reps slider (`NumberSlider`): `syncingScroll` is cleared after one animation frame, but a
  smooth scroll outlasts it, so mid-scroll events can overwrite the tapped value (a tap on 9 saved 11 in
  ~1 of 3 runs before the spec waited for scrolling to stop). Worth making the tap deterministic.

## Decisions log

- 2026-11-22: Playwright Test, Chromium only (desktop + mobile projects). Own `e2e/package.json`.
- 2026-11-22: Real stack, no mocks. Do not sync prod data. Seed + per-test data.
- 2026-11-22: Isolation is a **database per worker** cloned from a cached template, reset before every
  test. Not per user: the app is single-tenant and several tables are unique per date.
- 2026-11-22: Long-lived tmpfs Postgres container with durability off; cached builds keyed by content hash.
- 2026-11-22: Auth via the `DEV_AUTH_TOKEN` bypass cookie (login is rate-limited: 5/user/15min, and
  >5 usernames/IP/10min blocks the IP for 24h). Only `tests/auth/` logs in for real.
- 2026-11-22: Server clock package + test-only `/__e2e/*` endpoints behind the `e2e` build tag, refused
  in production.
- 2026-11-22: Failures are produced by triggering their real cause (input, state, DB fault, hold,
  network), never by faking responses. Holds instead of delays for slow paths.
- 2026-11-22: Production build by default; `E2E_FRONTEND=dev` for fast iteration.
- 2026-11-22: No attach mode against the dev stack (shared dev DB breaks isolation).
- 2026-11-22: Docs stay in `e2e/` (`docs/projects/` is going away).
- 2026-11-22: Measured (24 cores, 6 workers): cached setup ~60ms, cold container + template ~1.3s,
  frontend rebuild ~3s, ~0.5s per test, 280 tests in ~28s with 0 flaky over 5 repeats. Vite dev mode
  ~2.5x slower per test, so the production build stays the default.
- 2026-11-22: Worker DBs are named per checkout (`e2e_<hash of path>_w<N>`) so worktrees share the one
  container; one run per checkout at a time (lock file).
- 2026-11-22: CI is non-blocking through `ci.sh` (j-cicd steps have no allow-failure option); the
  switch to blocking is a data-backed decision (`npm run ci:history`, 20 clean runs).
- 2026-11-22: Orphan pruning drops only databases whose recorded checkout path no longer exists. An
  earlier version also dropped databases without a recorded path and broke concurrent runs in other
  worktrees: never prune what can't be proven orphaned.
- 2026-11-22: Example specs use grocery add: it is optimistic, retries twice (TanStack `retry: 2`) and
  is idempotent by `client_id`, so it exercises every failure mode.
- 2026-11-22: B1 smoke writes wait for their own response (URL regex) before reloading. Where the
  page doesn't prefill after reload (steps form) or navigates after create (finances account page,
  which hides its title on mobile), persistence is checked via `api.get` / a fresh `goto` of the list.
- 2026-11-22: Gym specs drive the per-exercise logger directly (`/gym/logging/<id>`); sets are arranged
  with `api.logSet`. The reps slider is scroll-driven (a tick click smooth-scrolls and scroll events set
  the value), so `setReps` waits for the scroll to settle before trusting the value.
