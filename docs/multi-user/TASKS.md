# Tasks

Lanes are sized for one agent each. **Depends on** must be merged into `proj/multi-user` first.
Branch/worktree per lane: `mu/<lane>-<name>` in `.worktrees/mu-<lane>-<name>` (see README).
Tick boxes in the PR that does the work.

Worktrees created so far: `mu-a1-cleanup`, `mu-a2-pool-role`, `mu-a3-audit`, `mu-b1-identity`
(all merged), `mu-b2-request-db`, `mu-h1-http-hardening`. Create the others when their
dependencies have merged, so they start from an up-to-date base.

Security findings from A3 live in [BACKLOG.md](BACKLOG.md); the `#n` references below point there.

```
A1 cleanup ─┐
A2 pool+role ┼─> B1 identity -> B2 per-request DB -> B3 owner kit ─┬─> C1 tracking ─┐
A3 audit ───┘            │                                       ├─> C2 diet      │
  (read-only, anytime)   └─> C6 accounts (after B1)               ├─> C3 workout   ├─> D1 lock-in
H1 HTTP hardening (any time; before C6's setup page ships) ──────────────────────────┘
                                                                  ├─> C4 money     │
                                                                  └─> C5 media ────┘
```

Parallel now: A1, A2, A3. Then B1–B3 one at a time (they touch every module). Then C1–C6 in
parallel. Then D1.

## Phase A: groundwork (parallel)

### A1 Remove dead DB code
Depends on: none
- [x] Delete `internal/generics` (nothing imports it).
- [x] Remove `internal/database/repository`: inline what's still used (diet repo, workout
      models/workoutlog repo, `utils/query_helper.go`) or replace with plain gorm.
- [x] Remove `internal/database/dumps/` if it's a leftover (confirm nothing references it).
- Done when: `go build ./... && go test ./...` pass and neither package exists.

### A2 Connection pool and app DB role
Depends on: none. Touches compose, j-cicd and prod: read `docs/ARCHITECTURE.md` and `j-cicd guide` first.
- [x] Explicit pool settings in `internal/database` (max open/idle, lifetimes), configurable by env.
      Today: one shared pool on `database/sql` defaults (unlimited open, 2 idle, no lifetime),
      so bursts open new connections and nothing caps them below Postgres' `max_connections`.
- [x] Replace `router.Run` with an `http.Server` that has read/write/idle timeouts and graceful
      shutdown (drain in-flight requests on SIGTERM, then close the pool). Today a deploy cuts
      requests off mid-transaction.
- [x] `simpletrack_app` login role: DML grants only, no ownership, no `BYPASSRLS`; default
      privileges so future tables are granted automatically. `statement_timeout` and
      `idle_in_transaction_session_timeout` on the role.
- [x] Server uses the app role; `migrate`, devseed, e2eseed, devuser keep the owner role. Wire
      both DSNs through `dev/env.sh`, e2e setup and `docker-compose.vps.yml` / `.cicd.yaml`.
      Role password is a secret, never in a migration.
- [x] Update `docs/ARCHITECTURE.md`.
- [x] Prod rollout plan: [ROLLOUT.md](ROLLOUT.md).
- Done when: dev stack and `./bin/dev e2e` run with the server on the app role; a prod rollout
  plan is written to `docs/multi-user/ROLLOUT.md` (used at D1; nothing deploys before then).

### A3 Security audit (read-only)
Depends on: none
- [x] Review the whole app (backend handlers, auth/cookies/JWT, CORS defaults, proxy/IP handling,
      error messages, input validation, file uploads/imports, frontend storage of data, deps).
      Seeds: default CORS list includes LAN IPs (`cmd/server/main.go`); `"Invalid token: " + err`
      echoes internal errors.
- [x] Write findings to `docs/multi-user/BACKLOG.md` with severity and a suggested lane. Don't fix
      things that a B/C lane will rewrite anyway; note it instead.
- Done when: BACKLOG.md exists and anything critical is flagged to the owner.
- Result: 22 findings, none critical; 2 High (C2/C3), several for B1, C5, C6 and a proposed new
  lane H1 (HTTP hardening). Lanes: check BACKLOG.md for entries tagged with your lane.

### H1 HTTP hardening
Depends on: none. Can run alongside B2. Must merge before C6's `/setup/:token` page.
Touches Caddy: read `docs/ARCHITECTURE.md` and `j-cicd guide` first.
- [ ] #7 Security headers in `deploy/Caddyfile` for the whole site: CSP (start from the
      BACKLOG suggestion, tighten against the real build output), `frame-ancestors 'none'`,
      `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy`, drop the `Server` banner.
- [ ] #8 Request body limits: backend middleware with a small default and an explicit larger
      limit on the weight import route (plus a row cap); `request_body max_size` under `/api` in Caddy.
- [ ] #17 `Cache-Control: no-store` on every `/api` response.
- [ ] #18 No default CORS origins (same-origin only); drop `ExposeHeaders: Set-Cookie`.
- [ ] Verify the headers against a local prod-like stack (images + Caddy) and that the app,
      PWA and phone proxy still work. Add the Caddy deploy order to ROLLOUT 3.3a.
- Done when: a test covers body limits and `Cache-Control`; the CSP produces no console
  violations on any page.

## Phase B: foundation (sequential, one agent at a time)

### B1 User identity on the request
Depends on: none (can start alongside A)
- [x] `uid` (and admin flag source) in JWT claims; tokens without `uid` -> 401.
- [x] Middleware puts the user on `context.Context`; `auth.UserID(ctx)` / `auth.MustUser(ctx)`.
- [x] Dev-token bypass resolves the `dev` user's ID.
- [x] Migration: `users.is_admin`, `users.token_version`, `users.disabled_at`; owner account is admin.
- [x] `token_version` in the JWT, checked against the user row by the middleware (cache briefly
      if needed); disabled users are rejected.
- Done when: auth tests cover missing/old/revoked tokens and disabled users; `dev api /auth/me`
  returns the ID.

### B2 Per-request DB handle
Depends on: A1, A2, B1
- [x] `database.WithTx` / `database.From(ctx)`: lazy per-request transaction with
      `set_config('app.user_id', ..)` and `app.is_admin`; commit/rollback/panic handling.
- [ ] Thread `ctx` through every service and repo function; replace every `conn()` /
      `GetDB()` in request paths. Mechanical, no behaviour change. The handle carries the
      request context, so a disconnected or timed-out request cancels its query (today only the
      workout repos pass a context; diet, money, tracking and media queries can't be cancelled).
- [x] External lookups (media, food) don't run inside the request transaction.
- [x] Record the final shape in DECISIONS.
- Done when: no request-path code calls `GetDB()`/`conn()` (add a test or lint that greps for it);
  full test suite and e2e pass.

### B3 Owner kit
Depends on: B2
- [ ] `database.Owned(ctx)` gorm scope; helper to stamp `UserID` on create; 404 helper for
      not-found-or-not-yours.
- [ ] SQL pattern for owned tables, children (composite FK) and catalogs, as a worked example
      in DECISIONS that C lanes copy.
- [ ] Two-user Postgres test harness (Alice/Bob, tokens, DB) usable from any module's tests.
- [ ] RLS coverage test (fails for any table without RLS, with an allowlist C lanes shrink).
- [ ] Postgres-backed tests run in the j-cicd `backend-test` step. Today it runs `go test`
      without `TEST_DATABASE_URL`, so they're skipped. Give it a throwaway Postgres and make
      those tests fail instead of skip in CI (`CI=true` or similar).
- [ ] #1 Rule + grep test: no `ShouldBindJSON` into a type with a `UserID` or `CreatedBy` field;
      only the stamp helper sets `user_id` / `created_by`; creates use `Omit(clause.Associations)`.
- [ ] #2 `database.LockKey(ctx, ...)` helper that puts the user ID into every advisory-lock key.
- [ ] #6 Grep test: no new `"error": err.Error()` in 500 responses (allowlist the existing
      sites; C lanes shrink it by moving to `apierr`).
- [ ] #20 `govulncheck ./...` and `npm audit --omit=dev --audit-level=high` in the pipeline.
- Done when: one small table (suggest `drink_size_presets`) is fully converted as the reference.

## Phase C: modules (parallel)

Every C lane: migrations (backfill to admin, per-user uniques, RLS), scoped queries, two-user
tests for every endpoint, devseed/e2eseed updates, remove its tables from the RLS allowlist,
move its handlers to `apierr` (#6) and remove them from that allowlist, and fix the BACKLOG
entries tagged with the lane.

### C1 Tracking
Depends on: B3
- [ ] body_weight_logs (incl. Renpho import), step_logs, water_logs (client_id), drink_size_presets
      (if not done in B3), grocery_items (client_id), user_profiles -> 1:1 with user, missed.
- [ ] BACKLOG #6.

### C2 Diet
Depends on: B3
- [ ] Catalog: foods, composite_foods, composite_food_items (`created_by`, catalog policies).
- [ ] Owned: days, day_logs, plans, meals, meal_items, planned_meals, saved_meals, saved_meal_items.
- [ ] client_id idempotency lookups scoped by user.
- [ ] BACKLOG #1 (High): request DTOs instead of binding models in meal_route.go; no nested
      association writes. #2 (High): advisory-lock keys and existence checks per user,
      including `Unscoped()` counts. #6.

### C3 Workout
Depends on: B3
- [ ] Catalog: exercises. Move machine settings (0003) to a per-user table.
- [ ] Owned: workout_programs, workout_plans, plan days/exercises, workout_logs, logged_exercises,
      logged_sets, cardios.
- [ ] BACKLOG #2 (High): `CreateMinimalIfMissing` lock key and date check per user; two-user
      test where Alice and Bob both start a workout on the same date. #6.

### C4 Money
Depends on: B3
- [ ] Split investment_account_types: shared rules vs per-user opening room. Contribution rules
      are catalog.
- [ ] Owned: investment_accounts, account_balances, investment_deposits. Net worth per user.
- [ ] BACKLOG #6.

### C5 Media
Depends on: B3
- [ ] Split media_items into a shared catalog and per-user `user_media` (status, rating, notes, dates).
- [ ] Frontend follows the API change.
- [ ] BACKLOG #5 (TMDB key in error messages), #6, #15 (catalog cover URLs).

### C6 Accounts and setup links
Depends on: B1
- [ ] `user_setup_tokens` table; admin endpoints: create user, generate setup/reset link, list users,
      disable/enable user (bumps `token_version`).
- [ ] `auth_events` audit log written by login, setup, reset, create/disable; admin view.
- [ ] Signed-in users can change their own password (bumps `token_version`, keeps the current session).
- [ ] Setup endpoint: validate token, set password, sign in; login protection applies.
- [ ] Frontend: admin section in settings (create user, copy link), `/setup/:token` page.
      Follow `docs/design/` and the `ui` skill.
- [ ] BACKLOG #3 (login lockout of a known username), #4 (IPv6 /64 keying, limiter eviction),
      #9 (per-user frontend storage, cleared on logout and 401), #14 (username/password rules),
      #19 (logs: route patterns, no attempted usernames).
- Done when: e2e covers create user -> open link -> set password -> sign in -> link no longer works.

## Phase D: lock-in

### D1 Verify, rehearse, ship
Depends on: all A, B and C lanes
- [ ] RLS allowlist is empty; every endpoint has a two-user test; CI runs them.
- [ ] BACKLOG: every High done; every Medium done or deferred with a reason in DECISIONS.
- [ ] #16 Pin `app-network`'s subnet/gateway in compose and confirm `TRUSTED_PROXIES` (ROLLOUT 3.2).
- [ ] ROLLOUT.md is complete: every lane's prod-side change has a step, and the rehearsal ran it.
- [ ] Rehearse: restore a copy of a recent prod backup locally, run every migration as the
      deploy would, start the server as the app role, click through each module as the owner.
      Check the backfill assigned every row and nothing disappeared (row counts before/after).
- [ ] Prod rollout plan from A2 (app role, secrets, compose/j-cicd) checked against the rehearsal.
- [ ] Merge `origin/main` into the base, renumber migrations if needed, mark the base PR ready.
- [ ] With the owner: merge to `main`, deploy via j-cicd, verify, then the owner creates the
      first family account. ARCHITECTURE.md current.
