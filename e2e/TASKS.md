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
- [x] **B2. Gym** (first: most active and riskiest): log a set, edit, delete, history, persistence
      across reload, mobile sticky log bar, retry after dropped response (no duplicate set).
      Specs: `tests/gym/logging.spec.ts` (per-exercise logger, Back Squat on today's "Legs" plan).
      Scenarios (done unless marked):
      - log a set: shows at once, one row in DB, survives reload
      - edit: tap a saved set, change reps, save; persists
      - delete: removes only that set; persists
      - history: last session's set shows as "Last" next to today's
      - mobile: on a short phone the Log button is pinned inside the viewport (mobile project only)
      - dropped response (hold after + drop): retry with the same client_id, exactly one set
      - todo: see "Not yet" in the lane entries below.
      - [x] weight choice (`tests/gym/weight-choice.spec.ts`): the go-up/stay switch on the setup card after an
            earned increase: default up, stay + flip back, logged with the set, gone after set 1, kept when
            leaving before set 1, absent without an earned increase.
      - [x] drop progression (`tests/gym/drop-progression.spec.ts`): a drop last session opens at the dropped
            weight with no increase and the weight sheet says why; a drop today shows "start at 90 next time".
            Not covered: sets loading in logged order after an edit (a HOT update keeps index order, so the
            unordered preload only misorders under a seq scan); the backend orders the preloads by id.
      - [x] tracking (`tests/gym/weight.spec.ts`, `steps.spec.ts`, `progression.spec.ts`,
            `tracking-load-errors.spec.ts`): weight save, Renpho import (+ rejected CSV), past day in
            Toronto late evening; steps save + overwrite/cancel/invalid; progression list per range, empty
            state, exercise switch; load errors + Retry. Weight and steps have no text history (canvas
            only), so specs read the slider, the overwrite prompt and the DB. Not yet: steps on a past day,
            save failure rollback, tab-separated/kg import.
      - [x] plans (`tests/gym/plans.spec.ts`): create + activate a program (/gym follows it), empty state,
            load error; routine create/rename/delete (+ Undo); add existing / "Create and add" / remove
            (+ Undo) exercises, rep rollover; day picker, move a taken day (confirm/cancel), drag (desktop);
            warmup/stretch items, cardio. Not yet: reorder exercises, rename program, write-failure
            rollbacks, offline, "Routine not found", "Couldn't load this routine".
      - [x] home (`tests/gym/home.spec.ts`): Today card states (fresh `@smoke`, in progress, complete),
            Start workout; plan switching (swap, rest day + Train anyway, swap away and back keeps the sets,
            switch failure rolls back); week strip status, streak, day navigation, week paging; load errors
            recover with Retry. Not yet: "Other plans…" dialog, Log freestyle, switching a past/future day,
            switch pending state, streak without a program.
      - [x] unilateral (`tests/gym/unilateral.spec.ts`): Edit dialog toggle, first-side dialog on Log Set 1
            (defaults to last session's; Esc logs nothing), switch clock (red, yellow at 0:15, green at 0:25),
            second side keeps the clock, pairs persist with sides, bilateral history counts for both sides,
            uneven weights carry per side, edit moves a side or deletes it. Not yet: finish with one side logged.
      - [x] session (`tests/gym/session.spec.ts`, `cardio-mobility.spec.ts`, helpers `tests/gym/logger.ts`):
            session list, Up next, finish + summary, weight/plates setup, notes (with and before the first
            set), persistent save failure + Retry, offline log, delete while saving, edit a saved set's
            weight, planned cardio, warmup/stretching checklists. `logging.spec.ts` repaired after the
            "Log Set N" rename. The logger keeps a draft in sessionStorage, so persistence checks clear it
            before reloading (`freshReload`). Not yet: add/remove exercise from the session, `?offset=`
            days, extra sets beyond the plan, Finish while a set didn't save.
- [x] **B3. Diet**: Quick log in `tests/diet/quick-log.spec.ts` (log `@smoke`, edit, delete). Edits of a
      logged meal (quick log and `/diet/log` Update) repoint the day log in one DB transaction, so the old
      create-then-delete risk is gone (fault-tested in `logmeal-edit.spec.ts`).
      - [x] home (`tests/diet/home-logging.spec.ts`, `home-day.spec.ts`, helpers `home-helpers.ts`): log a
            planned meal, double tap (one request), dropped response (de-duped by day + meal), log fault rolls
            back + Retry, delete a catalogue meal (foods and saved meal survive), delete a still-saving row,
            totals move and restore; day navigation via the week strip (Log hidden off today), week strip labels,
            average and paging; water preset + entries dialog delete; empty states; day load error + Retry;
            `test.fail` for future day rows. Not yet: held water write, `ml` on the home page.
      - [x] logmeal (`tests/diet/logmeal.spec.ts`, `logmeal-edit.spec.ts`, `logmeal-saved-foods.spec.ts`,
            helpers `logmeal-helpers.ts`): catalogue meal (and `?offset=`), Log and save, start from a saved
            meal, Update a logged meal, Update failing on either statement (no loss, no duplicate, retry), Edit
            and log a planned meal, new saved meal + validation, create a food, duplicate food name (409 toast),
            groups, quantity 0 / halve / double, draft kept across reload and failures, cleared after a log.
            Not yet: composite foods, variant swap, Ungroup / Remove group, `?type=editsaved`, edit on another
            day, load errors on the meal and pickers, double submit, the create-mode draft.
      - [x] plan (`tests/diet/plan.spec.ts`, `plan-failures.spec.ts`, helpers `plan-helpers.ts`): add from a
            saved meal, armed add (desktop only), remove, move earlier/later, copy to other days, copy from the
            previous day, clear day, week paging + This week, past days editable (pinned), no-saved-meals empty
            state, load error + Retry, add failure rolls back + Retry. Not yet: saved-meals load error, copy
            failure, Move/Remove while saving, logged meals surviving Clear day.
      - [x] targets (`tests/diet/targets.spec.ts`, `water-presets.spec.ts`): edit calories + split and save (new
            plan snapshot, History "Current"), goal chip from the formula TDEE, today uses the new targets while
            past days keep the old plan, dirty badge + Reset, save failure stays dirty + retry, load error + Retry,
            activity level; water sizes add/edit/delete (+ on /diet), validation, daily goal (oz) and ml unit.
            Not yet: g-per-lb presets, lock toggles, "Fill to" carbs, observed maintenance, Last 7 days intake
            Retry, deleting a size that has water logs.
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
- [x] **B5. Media**: fake upstream for the lookup APIs (`infra/upstream.ts`, fixture `upstream`).
      Specs (`tests/media/lookup.spec.ts`): look up an anime, pick a match, form fills, save, persisted
      with `source`/`external_id`; upstream 500 shows "Lookup failed" and keeps the form.
      Not yet: TMDB / Open Library / iTunes fakes (add handlers to `Upstream`), manual add/edit/delete,
      list filters and status tabs, rating, progress, cover image (needs a fake image route).
- [x] **B6. Grocery** (`tests/grocery/list.spec.ts`; add-failure modes are A8's, not repeated).
      Scenarios: empty state; add several items, persisted after reload; check off (leaves the list,
      stays gone after reload, `completed_at` set); delete via API is gone from the UI (soft delete).
      Not yet: Paste list dialog (`/items/parse`, `/items/bulk`, de-dupe), suggestions dropdown,
      offline complete, blank-name input.
- [x] **B7. Settings and auth**: login UI (no bypass cookie), logout, session expiry (via clock).
      Specs in `tests/auth/login.spec.ts` (`@auth`, `authenticated: false`): unauthenticated visit
      redirects to sign-in with `redirect=` and returns there after login (+ reload keeps session,
      `@smoke`); wrong password shows an alert; logout (account menu) ends the session;
      expired session (server clock +366d) redirects to sign-in on direct load and on next full load.
      Not yet: settings profile (body profile form) save + reload, register flow, rate-limit lockout.
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

- [x] **C1. CI step**: `e2e` step in `.cicd.yaml` runs `e2e/ci.sh`: CI settings, run history in
      `~/.local/state/simple-track-e2e/ci-history.jsonl`, failed runs' reports kept (blocking since C2).
- [x] **C2. Promote to blocking**: `E2E_BLOCKING=1` on the step. Verified with 20 runs of `ci.sh`,
      each from a clean clone of the merged A+B+C commit: 20/20 green, 145 passed, 0 flaky, ~18s each.
- [x] **C3. `large` seed profile**: `seed.Large` (two years of history in every module, odd values),
      `test.use({ seed: 'large' })`, `POST /__e2e/seed/large`, `e2eseed -profile large`; `tests/scale/`
      renders every page against it. Timing budgets wait for the performance project.
- [x] **C4. Aria snapshots**: `tests/aria/` covers the main nav and the gym week strip, per project.
      Grow only for regions where losing an item would go unnoticed.
- [x] **C5. Failure summary reporter**: `test-results/failures.md`. Revisit its contents after agents
      have used it on real failures.
- [ ] **C6. Two devices, one backend** (later): a `device('desktop' | 'mobile')` fixture that opens a
      second browser context with the same cookies, clock, guards and hermetic routing as `page`
      (~20 lines in `fixtures/index.ts`), plus one worked example. Scenarios are one person on phone +
      laptop: both logging weight/steps for the same day (unique-per-date tables), editing a set the
      other device deleted, overlapping optimistic writes. Order the interleaving with holds (e.g.
      phone's save held `after` commit while the laptop saves), never by racing clicks. With
      multi-user, the fixture gains a `user` option for cross-user isolation tests.

## Findings (bugs the suite surfaced; not fixed here)

- Optimistic rows appear before their request is sent, so a reload right after the row shows can
  cancel the write (found as a mobile-only flake in the first golden-path spec). Not a realistic user
  path, but specs must wait before reloading: `waitForApi`, or the row's `aria-busy` clearing where
  the row is a `ListRow` (see B8).
- `POST /diet/foods` with an existing name returns **500**: the unique index `idx_foods_name` rejects it
  and `PostFood` maps every error to `apierr.Internal`. Decide the intended status (409?), then pin it.

- Gym reps slider (`NumberSlider`): `syncingScroll` is cleared after one animation frame, but a
  smooth scroll outlasts it, so mid-scroll events can overwrite the tapped value (a tap on 9 saved 11 in
  ~1 of 3 runs before the spec waited for scrolling to stop). Worth making the tap deterministic.
- Adding a transaction with "also update the balance" is now one atomic request; the spec in
  `tests/finances` asserts the deposit is rolled back when the balance write fails.
- Grocery has no delete control in the UI (only `DELETE /tracking/grocery/items/:id`), so the delete
  scenario arranges via API and asserts the UI. Decide whether a UI delete is intended.
- Grocery checkbox accessible name is the item name ("Milk"), not "Complete Milk" (the `label` prop is
  not exposed by `Checkbox` when the slot is used).
- ~~The UI has no logout control~~: fixed by the account menu (PR #87); the logout spec signs out
  through it.
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
- 2026-11-22: CI is non-blocking through `ci.sh` (j-cicd steps have no allow-failure option); the
  switch to blocking is a data-backed decision (`npm run ci:history`, 20 clean runs).
- 2026-11-23: Failure analysis reviewed by injecting three bugs (a silent non-save on grocery
  complete, a render crash on `/gym/logging`, a first-attempt flake). Results: the crash went
  uncaught (no spec opened that page) until smoke covered every param-free route; the non-save was
  caught but only diagnosable via `db.sql`, so failures now list the rows the test changed; bundle
  stack frames were unreadable, so the e2e build has source maps and the guard maps them to src.
- 2026-11-23: e2e step made blocking after 20/20 clean CI-equivalent runs (clean clones, `ci.sh`).
  If it turns red without a code reason, unset `E2E_BLOCKING` first, then fix the flake.
- 2026-11-22: Orphan pruning drops only databases whose recorded checkout path no longer exists. An
  earlier version also dropped databases without a recorded path and broke concurrent runs in other
  worktrees: never prune what can't be proven orphaned.
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
