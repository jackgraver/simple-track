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
- [x] **B3. Diet**: specs in `tests/diet/quick-log.spec.ts` (Quick log, the simplest logged-meal path).
      Done: log (optimistic, persists after reload, `@smoke`); edit (replaces the row, exactly one logged
      meal left); delete via confirm dialog (other meal survives reload).
      Scenario list still open: log a catalogue-food meal via `/diet/log` (multi-item, amounts);
      log a planned meal (one-tap Log, "Logged" badge, no double log); edit-and-log a planned meal;
      day navigation (`?offset=`) logs to the right day; water log; macro totals/goals update after
      log/delete; delete a still-saving row; retry after dropped response (no duplicate, `client_id`);
      DB fault on log rolls the row back with a Retry toast; targets edit persists; plan week
      (add/copy/delete planned meal); saved meals CRUD.
      Data-loss risks to cover next: editing a logged meal is replace-by-id (`replace_meal_id`), so a
      failure between create and delete could lose or duplicate it; deleting a logged meal must not
      delete shared foods or saved meals.
- [ ] **B4. Finances**
- [x] **B5. Media**: fake upstream for the lookup APIs (`infra/upstream.ts`, fixture `upstream`).
      Specs (`tests/media/lookup.spec.ts`): look up an anime, pick a match, form fills, save, persisted
      with `source`/`external_id`; upstream 500 shows "Lookup failed" and keeps the form.
      Not yet: TMDB / Open Library / iTunes fakes (add handlers to `Upstream`), manual add/edit/delete,
      list filters and status tabs, rating, progress, cover image (needs a fake image route).
- [x] **B4. Finances** (`tests/finances/transactions.spec.ts`, helpers `createAccount`/`addTransaction`).
      The product has no "edit transaction" (only add and delete); edit-shaped paths are account edit
      and re-recording a balance. Scenarios:
      - [x] add account (opening balance) -> app opens it -> persists across reload and in the list
      - [x] deposit with "also update balance": row, balance and exactly one DB row survive reload
      - [x] delete transaction: row gone after reload, balance unchanged (by design)
      - [x] edit account name via dialog; persists, old name gone
      - [x] data-loss: deposit whose balance update fails (`test.fail`, see Findings)
      - [ ] later: withdrawal, record balance over an existing day (upsert), balance delete, archive/
            restore (net worth), delete account (cascade), contribution-room rules, update-balances
            dialog, liability accounts (no transactions section), validation errors (400)
- [ ] **B5. Media**: needs a fake upstream for the lookup APIs (AniList etc., URLs are configurable).
- [x] **B6. Grocery** (`tests/grocery/list.spec.ts`; add-failure modes are A8's, not repeated).
      Scenarios: empty state; add several items, persisted after reload; check off (leaves the list,
      stays gone after reload, `completed_at` set); delete via API is gone from the UI (soft delete).
      Not yet: Paste list dialog (`/items/parse`, `/items/bulk`, de-dupe), suggestions dropdown,
      offline complete, blank-name input.
- [ ] **B7. Settings and auth**: login UI (no bypass cookie), logout, session expiry (via clock).
- [ ] **B6. Grocery**
- [x] **B7. Settings and auth**: login UI (no bypass cookie), logout, session expiry (via clock).
      Specs in `tests/auth/login.spec.ts` (`@auth`, `authenticated: false`): unauthenticated visit
      redirects to sign-in with `redirect=` and returns there after login (+ reload keeps session,
      `@smoke`); wrong password shows an alert; logout (API call, no UI control) ends the session;
      expired session (server clock +366d) redirects to sign-in on direct load and on next full load.
      Not yet: settings profile (body profile form) save + reload, register flow, rate-limit lockout.
- [ ] **B8. Optimistic UI**: hold-based assertions of instant state and post-reload state.
- [x] **B8. Optimistic UI**: hold-based assertions of instant state and post-reload state.
      Pattern (`tests/optimistic/`): hold the request, assert the instant state, release, wait for the
      UI to settle (`aria-busy` gone, no network wait), reload, assert again. Scenarios:
      grocery add (row visible + `aria-busy` while held, persists), grocery complete (checked at once,
      gone after reload; hold by template because the client may use `client_id` as the ref), weight save
      (overwrite prompt quotes the optimistic value while held and the server value after reload; the
      chart is a canvas, so the cache is probed through that prompt). Not yet: rollback with Retry
      (covered for grocery add in examples), gym set logging and diet/finances (their lanes B2-B4 own
      them; reuse this pattern).

## C. CI and scale

- [ ] **C1. CI step**: non-blocking `e2e` step in `.cicd.yaml` (production build, `--fail-on-flaky-tests`).
- [ ] **C2. Promote to blocking** after a stretch of stable runs.
- [ ] **C3. `large` seed profile** (or bulk `/__e2e/seed/*` endpoints) for long histories and scale.
- [ ] **C4. Aria snapshots** for key screens, if they earn their keep.
- [ ] **C5. Failure summary reporter** (`results/failures.md`), after reviewing how agents read failures.

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
- Adding a transaction with "also update the balance" is two client requests (`POST .../transactions`
  then `PUT .../balances`), not one atomic call. If the second fails, the transaction is saved, the
  balance is stale and the user sees an error, so retrying double-counts the deposit. Pinned by a
  `test.fail` spec in `tests/finances`; fix = one server endpoint doing both in a DB transaction, then
  drop the `test.fail`.
- Grocery has no delete control in the UI (only `DELETE /tracking/grocery/items/:id`), so the delete
  scenario arranges via API and asserts the UI. Decide whether a UI delete is intended.
- Grocery checkbox accessible name is the item name ("Milk"), not "Complete Milk" (the `label` prop is
  not exposed by `Checkbox` when the slot is used).
- The UI has no logout control: `useAuth().logout` is never called from a component. The logout spec
  calls `POST /api/auth/logout` instead. Add the button, then drive it through the UI.
- `ListRow` already sets `aria-busy="true"` on pending rows (grocery uses it), so the "mark pending
  rows" idea only needs extending to the other optimistic lists (gym sets, diet, finances) before
  specs there can drop `waitForApi`. Specs for grocery use it now.
- `Checkbox.vue` takes `label` but the slot content replaces it, so grocery checkboxes are named by the
  item name (`Eggs`), not `Complete Eggs` as the `label` prop suggests.

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
- 2026-11-22: Example specs use grocery add: it is optimistic, retries twice (TanStack `retry: 2`) and
  is idempotent by `client_id`, so it exercises every failure mode.
- 2026-11-22: Third-party lookups (media) are replaced by a per-worker fake upstream HTTP server in the
  Playwright worker (`infra/upstream.ts`), not by `route.fulfill`: the backend's own HTTP client,
  parsing and error mapping run. The backend takes `LOOKUP_*_URL` env overrides (ignored in
  production); `startBackend` accepts extra env. Add a source by adding a handler plus its env var.
- 2026-11-22: B1 smoke writes wait for their own response (URL regex) before reloading. Where the
  page doesn't prefill after reload (steps form) or navigates after create (finances account page,
  which hides its title on mobile), persistence is checked via `api.get` / a fresh `goto` of the list.
- 2026-11-22: Gym specs drive the per-exercise logger directly (`/gym/logging/<id>`); sets are arranged
  with `api.logSet`. The reps slider is scroll-driven (a tick click smooth-scrolls and scroll events set
  the value), so `setReps` waits for the scroll to settle before trusting the value.
- 2026-11-22: Diet specs start with Quick log: no catalogue dependency, one row, and it covers the
  optimistic log/edit/delete flow. Delete asks for confirmation ("Remove quick log" dialog).
- 2026-11-22: Finances specs create accounts through the API with a balance date before `E2E_EPOCH`:
  the form only offers "also update the balance" when the transaction date is on/after the latest
  balance. After creating an account the app navigates to it, so specs wait for the URL before reloading.
- 2026-11-22: Grocery specs use `waitForApi` with `:id` placeholders (not `*`) for path params.
- 2026-11-22: Real logins stay few: one per test, all as the seeded `dev`, one wrong password per run (the
  limiter is per backend process; only failures should count toward 5/user). Session expiry moves the
  server clock past the 365d token lifetime; the browser cookie's own expiry uses real time.
- 2026-11-22: Optimistic specs wait for the UI to settle (`aria-busy` gone) before reloading, instead of
  `waitForApi`, where the app exposes it; holds on completes use a route template (ref may be `client_id`).
