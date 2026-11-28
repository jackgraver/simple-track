# Decisions

Newest at the bottom. One entry per decision: what, why, and who/when.

## 2026-11-27: Project kickoff (owner)

- **Audience**: family only, no open sign-up. Owner is admin and creates accounts.
- **Onboarding**: admin-generated one-time setup links, sent by text. No email anywhere. Same
  mechanism for password resets. Passkeys maybe later.
- **Sharing/groups**: deferred. Design must not block it (see README "Future: sharing").
- **Shared catalogs**: foods, exercises, investment/contribution rules and the media catalog are
  global. Personal data on them (exercise machine settings, personal contribution room, media
  status/rating/notes) moves to per-user tables.
- **Catalog permissions**: anyone signed in can add, creator or admin can edit, only admin deletes.
- **Postgres RLS**: yes, as the backstop under explicit app-level filters. App connects as a
  non-owner role.
- **Per-request DB handle**: replace the global `conn()` pattern; each request's queries go
  through one transaction carrying `app.user_id`.
- **Connection pool**: harden explicitly (limits, lifetimes, timeouts).
- **Dead code**: delete `internal/generics` (unused experiment); untangle and remove what's left
  of `internal/database/repository`.
- **Docs location**: `docs/multi-user/`, not `docs/projects/` (that folder is being retired).
- **Branching**: all work stacks on `proj/multi-user`; one draft PR to `main`, merged and deployed
  once at D1 after a rehearsal on a prod backup copy. Lanes PR into the base.
- **Thoroughness**: prefer a solid foundation over shortcuts; skip something only when its
  effort clearly outweighs what it brings, and record why here.
- **Added for a solid foundation**: revocable sessions (`token_version`, `disabled_at`), an
  `auth_events` audit log, Postgres-backed tests running in CI, HTTP server timeouts and graceful
  shutdown, migration rehearsal on a prod backup copy.

## 2026-11-27: Who merges lane PRs (owner)

Agents merge A1 (dead-code cleanup) and A3 (audit doc) themselves. The owner reviews and merges
every other lane, since nothing is reviewed again before the single deploy at D1.

## 2026-11-27: A1 dead DB code removed

- `internal/generics`, `internal/database/repository` and `internal/database/dumps` (a one-off
  Python import script with hard-coded exercise IDs) are gone.
- Their last users are plain gorm on `conn()`: `dietrepo.DaysByDateRange` / `DayByIDGeneric`,
  `workoutrepo.GetByDateRange`. B2 swaps `conn()` for `database.From(ctx)` like everywhere else.
- `GET /diet/plans/plan/all` (the only `?page=&orderBy=&<filter>=` endpoint) uses `utils.List[T]`, a small
  plain-gorm helper next to `ParseQueryParams`. Field names still come only from a
  `QueryPolicy` allowlist. B3's `database.Owned(ctx)` scope goes on the `*gorm.DB` passed in.
- Models lost `GetID()` / `GetDate()` / `Preloads()` (they only served the repository
  interfaces). `TableName()` stays.
- One visible difference: `GET /diet/logs/day/:id` for a missing ID still answers 500, but the
  message is gorm's `record not found` instead of `entity not found`.

## 2026-11-27: A3 audit results (A3 agent)

- Findings are in `BACKLOG.md`, each tagged with an owning lane. Lanes tick their entries in the
  PR that fixes them; D1 checks every High is done.
- Proposed **new lane H1 "HTTP hardening"** (no dependencies, small): security headers/CSP in
  Caddy, request body limits (backend middleware + Caddy), `Cache-Control: no-store` on `/api`.
  Must land before C6 ships the `/setup/<token>` page (Referrer-Policy, frame-ancestors).
  Owner to confirm or fold into another lane.
- Gaps the original plan didn't cover, now assigned: advisory-lock keys must include the user
  (C2/C3), no binding request JSON into gorm models (C2 + a B3 check), login lockout per username
  and IPv6 keying (C6), require `exp` and check `Origin` in the middleware (B1), user-scoped
  frontend storage cleared on logout (C6).

## 2026-11-27: B1 user identity on the request (agent, lane B1)

- **JWT claims**: `uid` (users.id) and `ver` (users.token_version) join `username`. The middleware
  picks the user by `uid` only; `username` in the token is informational. A token without `uid`
  is rejected (every pre-B1 token), so everyone signs in once after D1.
- **Admin flag is not in the token.** The middleware reads `is_admin` from the users row, so
  promoting or demoting takes effect on the next request. "Admin flag source" = the row.
- **token_version check: one primary-key lookup per request, no cache.** The middleware loads
  the user by `uid` on every authenticated request and rejects when the row is gone
  (soft-deleted too), `disabled_at` is set, or `ver != token_version`. For a family app this is
  a sub-millisecond indexed read; a cache would delay revocation and need invalidation across
  processes, for no measurable gain. Revisit only if profiling says so. B2 may move this
  lookup into the request transaction; `users` stays outside RLS (README), so it works there.
- **token_version starts at 1** (column default and `models.InitialTokenVersion`), so a token
  with no `ver` (decodes as 0) never matches. Revoking = `token_version = token_version + 1`
  (C6: password change/reset, disable/enable). Code that creates users sets it explicitly
  (`authrepo.CreateUser`, devuser, e2e seed) rather than relying on gorm reading the default back.
- **Rejections**: every invalid/revoked/disabled/unknown-user session is 401
  `{"error":"Invalid or expired session"}`; the reason is logged server-side, never echoed (fixes
  the "Invalid token: " + err echo flagged for A3). Token errors wrap `auth.ErrInvalidToken`;
  anything else (DB down, missing JWT_SECRET) is **503**, so the client can tell an outage from
  a dead session. The frontend doesn't use that yet: `resolveAuthSession` (useAuth.ts) treats
  any `/auth/me` failure, network errors included, as signed out (pre-existing; owner accepted
  it, see below).
- **Context from gin**: `router.ContextWithFallback = true` in `cmd/server`, so a context
  derived from a `*gin.Context` still carries the user. Prefer passing `c.Request.Context()`.
- **Login refuses disabled users** with the same "Invalid username or password" as a wrong
  password (checked after bcrypt, so timing matches). C6 owns the rest of the account flows.
- **Context API** (`internal/core/auth`): `auth.User{ID, Username, IsAdmin}`, `auth.WithUser`,
  `auth.CurrentUser`, `auth.UserID(ctx) (uint, bool)`, `auth.MustUser(ctx)` (panics: only for
  code behind the middleware). They accept a `*gin.Context` too (read via its request
  context). The middleware still sets the gin key `username` for handlers not yet moved
  (money's birth-year lookup); the old `timestamp` key is gone (nothing read it).
  **Import rule**: `internal/database` must not import `internal/core/auth` (auth imports
  database through its repo: cycle). B2's middleware should read the user with
  `auth.UserID(ctx)` and pass the ID into `database.WithTx`.
- **Dev bypass**: resolves `DEV_AUTH_USER` (default `dev`) by username on every request (no
  cache: e2e truncates and reseeds between tests, so IDs move). A missing or disabled dev user
  is a 401. devuser and the e2e seed create that user as an enabled admin; devuser also
  re-enables and re-admins it on every run.
- **Migration 0009 owner backfill**: with exactly one live user (prod today), it becomes admin;
  with none (fresh/e2e DBs) nothing happens; with more than one the migration **fails** rather
  than guess who the owner is. D1's rehearsal on a prod backup will surface it if prod ever
  gains a second account before then (README says not to).
- **A3 backlog items done in B1**: #11 constant-time signature compare, #12 no echoed reasons,
  #10 partly (tokens must carry `exp` and `iat`, `alg` must be HS256, and `exp - iat` must not
  exceed `AUTH_COOKIE_MAX_AGE_SEC`, so lowering that setting also ends older long-lived tokens).
  Lifetime: see the owner's answer below.
- **CSRF (#13)**: AuthMiddleware refuses POST/PUT/PATCH/DELETE with `Sec-Fetch-Site` other
  than `same-origin` or `none` (403). Chosen over an `Origin` check (would need the public
  origin behind Caddy) and over forcing `Content-Type: application/json` (would break multipart
  uploads/imports). The SPA always calls `/api` same-origin (Caddy in prod, vite in dev, the
  phone proxy). Requests without the header (curl, `dev api`, Playwright's API client, very old
  browsers) fall back to SameSite=Lax. `/auth/login` and `/auth/logout` aren't behind the
  middleware and aren't covered.

## 2026-11-27: B1 review answers (owner)

- **Session lifetime stays 365 days**, no sliding re-issue. It must be easy to change:
  `AUTH_COOKIE_MAX_AGE_SEC` (seconds) is passed through `docker-compose.vps.yml` with a
  365-day default. Lowering it ends existing sessions issued for longer (the token's
  `exp - iat` is checked against the current setting), so no stale long sessions survive.
- **DB outage signing people out is accepted** for now (frontend treats a failed `/auth/me` as
  signed out). Nice-to-have, unscheduled: keep the session and retry on 5xx/network errors.
## 2026-11-27: App DB role, pool and HTTP server (lane A2)

**Two DSNs.** The server reads the existing `DATABASE_URL` / `DATABASE_URL_PRODUCTION` /
`DATABASE_URL_DEVELOPMENT`, which now name `simpletrack_app`. A new `DATABASE_OWNER_URL` is the
schema owner for `cmd/migrate`, devseed, devuser, e2eseed, the e2e build's `/__e2e` reset/seed,
pgweb and `dev test`. `database.ConnectApp()` / `ConnectOwner()` replace `ConnectToPostgres()` so
every caller picks one explicitly. Without `DATABASE_OWNER_URL` the tools fall back to the app
DSN with a logged warning (single-role setups): migrate then fails at its first DDL, DML-only
tools still work, and the e2e server refuses to start (its reset needs the owner).

**Role bootstrap: migration for structure, `migrate up` for the secret.**
- `0010_app_role.sql` creates `simpletrack_app` NOLOGIN with no password, strips any
  superuser/BYPASSRLS/CREATEROLE/CREATEDB/REPLICATION and role memberships from a pre-existing one, sets the role's
  timeouts, grants DML on all tables and USAGE/SELECT on sequences, revokes writes on
  `schema_migrations`, and sets `ALTER DEFAULT PRIVILEGES` for the owner so later migrations
  need no GRANTs. Its down step revokes the grants but keeps the role, since roles are
  cluster-wide and other databases may use it.
- `migrate up` (after a full migrate) calls `database.EnsureAppLogin`: if the app DSN names
  `simpletrack_app` and can't log in because of a wrong password, or because the role is still
  NOLOGIN with no password (as 0010 leaves it), it runs `ALTER ROLE ... LOGIN PASSWORD` with a
  client-computed SCRAM-SHA-256 verifier, so the plain password never reaches Postgres or its
  logs. In prod that happens in j-cicd's `pre_swap` with no extra manual step, and rotating the
  password is just changing `APP_DB_PASSWORD` and running `migrate up`. Checking first keeps
  normal deploys no-ops. A concurrent run that changed the role first counts as success. A role
  with a password that's NOLOGIN was switched off on purpose, so it is refused rather than
  re-enabled. A LOGIN role that still can't connect (pg_hba) is an error.
- Rejected: a manual `psql` step in prod (dev and e2e would each need their own bootstrap, and it
  isn't repeatable); putting the password in a migration (forbidden, and it would be the same
  everywhere).
- Roles are cluster-wide but migrations run per database, often in parallel (go test packages, e2e
  templates). The cluster-wide statements only run when something is missing and tolerate a
  concurrent creator (`duplicate_object`/`unique_violation`, and "tuple concurrently updated" for
  role settings). Checked with 3 × 8 concurrent `migrate up` runs on fresh clusters.

**Server guard.** At startup the server checks its own role (`database.CheckAppRole`). Superuser,
BYPASSRLS, owning a public table or inheriting its owner's role (RLS counts that as the owner),
or membership in a superuser/BYPASSRLS role is fatal in production and a warning elsewhere. The root
`docker-compose.yml` backend still uses `postgres` in development mode, and a local
production-mode copy has to use an app DSN.

**Known gap: the owner DSN is in the backend container's env.** j-cicd's `pre_swap` runs only
in a tracked service, so migrate has to run in the `backend` service and gets the same
environment as the server. Code execution inside the server could read the owner password. The
role split guards against query bugs (missing filters under RLS), not a compromised process.
Closing the gap needs j-cicd to run pre_swap in a separate compose service. Not worth it now.

**Pool** (`internal/database/pool.go`, env-overridable `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`,
`DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`; invalid values fail startup):
- max open **20**: one process, a few family users, and Postgres' default `max_connections`
  of 100 shared with migrate and psql. With B2's per-request transaction, a connection is held
  for a whole request, so 20 means 20 concurrent DB-touching requests; more queue rather than
  overload Postgres.
- max idle **10**: bursts reuse warm connections, while quiet periods hold few.
- lifetime **30 min**, idle time **5 min**: recycle connections slowly enough to be free, and
  release server-side memory after a burst.
- `/__e2e` owner pool: 2 open.

**Role timeouts:** `statement_timeout = 15s` (app queries are small single-user reads/writes; the
largest is the Renpho CSV import; 15 s catches runaways with a wide margin);
`idle_in_transaction_session_timeout = 30s` (frees locks held by a request transaction that was
abandoned). Both on the role, not the owner, so migrations are unaffected. Not added:
`lock_timeout`, since statement_timeout already bounds lock waits.

**HTTP server** (`cmd/server/serve.go`): read-header 5 s, read 30 s, write 45 s (above the 15 s
statement timeout plus the media lookups' 10 s client timeout), idle 120 s. SIGTERM/SIGINT:
stop accepting, drain for up to 25 s, then force-close and close the DB pool.
`stop_grace_period: 30s` on the prod backend service.

**`.cicd.yaml` unchanged:** it holds no DSNs (the e2e step uses `e2e/infra`, which now wires both).
`deploy.compose_file` (ship the compose file with each release) would simplify ROLLOUT's
manual compose-file copy and rollbacks. It changes every deploy, so it's left to the owner.

Follow-ups (owner, 2026-11-27): the owner-DSN gap and `deploy.compose_file` are j-cicd/deploy
config work after D1, tracked as BACKLOG #23 and #24. BACKLOG #16 (pin the network subnet) is
deferred to D1, since it needs a network recreate on the VPS.

## 2026-11-27: B2 per-request database handle (agent, lane B2)

Lands as three stacked PRs: the handle, middleware, auth, tracking, money and media; then diet;
then workout plus removing the global connection. Until the last one merges, diet and workout
still query the global pool outside the request transaction.

**API** (`internal/database/request.go`, `middleware.go`):
- `database.Middleware(pool)` gives every request a lazy transaction. `cmd/server` installs it
  after `e2e.Install` (so an e2e hold never waits inside an open transaction) and before every
  feature route.
- `database.From(ctx) *gorm.DB` is the only way request code reaches the database. The first
  call begins the transaction (`BeginTx` with the request context, so a client that goes away
  rolls it back) and sets `app.user_id` / `app.is_admin` with `set_config(..., true)`
  (transaction-local, so nothing leaks to the next user of the pooled connection). Each call
  returns a fresh session on that transaction. It accepts a `*gin.Context` too, and **panics**
  when the context has no handle (a route outside the middleware, or a test without `WithDB`).
- `database.SetIdentity(ctx, database.Identity{UserID, IsAdmin})`: AuthMiddleware calls it after
  resolving the user (applied at begin, or immediately if the transaction is already open).
  `internal/database` still doesn't import `internal/core/auth`.
- `database.WithDB(ctx, db)`: `From` returns `db` itself, no transaction, no identity. For tools,
  tests and the session lookup below; the guard test forbids it elsewhere in request code.
- `database.WithTx(ctx, pool) (ctx, *Request)` with `Commit` / `Rollback` / `Begun`, for code that
  needs a request-shaped transaction outside gin (the Postgres concurrency tests use it).
- `database.CreateInSavepoint(db, value)` and `database.Optional(ctx, fn)`, see "failed
  statements" below.
- `GetDB` / `SetDB`, the global connection and every package's `conn()` are gone.

**When the transaction ends.** Just before the first byte of the response is written: status
< 400 commits, 4xx/5xx rolls back. A handler that writes nothing ends it when it returns, and a
panic rolls it back. Committing before the response means a failed commit becomes a 500
(`{"error":"Could not save changes"}`) instead of a success the client (and optimistic UI) would
trust. That 500 drops the handler's `Set-Cookie`/`Location`/cache headers. A `From` call after
the response gets `ErrRequestDone` (its handle can't reach the pool, not even to begin a
transaction); nothing does that today.
**Behaviour change:** a request that answers 4xx/5xx no longer leaves the writes it made before
failing (e.g. `POST /diet/meals/meal/new` creating the meal, then failing to log it). Every
response the client sees as an error now means "nothing changed".

**Session lookup stays outside the transaction.** `auth.AuthMiddleware(pool)` (and
`auth.RegisterRoutes(router, pool)`) look the user up on the pool through `WithDB`, then call
`SetIdentity`. B1 left moving it in open; kept out so the transaction stays lazy: a route that
never queries (media lookups, `/auth/logout`) holds no connection. Cost: a DB-touching request
checks out a connection twice, one after the other (never two at once). `users` has no RLS.
`/auth/login` and `/auth/register` run entirely on the pool (`auth.OutsideRequestTx`), so a
burst of login attempts doesn't pin connections while bcrypt runs; they need no identity.

**External lookups.** The media lookup routes never call `From`, so no connection or transaction
is held while TMDB/OpenLibrary/AniList/iTunes answer (`media/controller/media_route_test.go`
checks the pool has nothing in use during the upstream call). The food lookup isn't on this
branch (`feat/food-api`); when it lands, fetch before the handler's first `From`. If an endpoint
ever needs the database both before and after an external call, add an explicit
"commit and start a new transaction" helper rather than holding the transaction across it (the
app role's `idle_in_transaction_session_timeout` is 30 s).

**Nested transactions are savepoints.** Existing `From(ctx).Transaction(...)` blocks keep
working; their rollback undoes only their own work. Locks they take (row locks,
`pg_advisory_xact_lock`) are now held **until the request ends**, not until the block returns.
Still correct; the Postgres concurrency tests (diet replays, workout sets/days) now run each call
in its own request transaction (`inRequest` helpers) to prove it. C2/C3 keep that in mind when
re-keying the advisory locks.

**Failed statements abort the Postgres transaction.** Anything after a failed SQL statement in
the same transaction fails with "current transaction is aborted". Code that recovers from a
statement error (insert, unique violation, read the winner's row) must run the statement in a
savepoint: `database.CreateInSavepoint(db, &row)` or its own `db.Transaction(...)`. Converted:
diet day find-or-create (both copies), grocery `CreateItem`, water `CreateWaterLog`. Best-effort
reads whose errors are ignored go through `database.Optional(ctx, fn)` (a savepoint inside a
request): diet `EnrichFoodVariants` and `CalculateTotals`, workout previous/max logs in
`GetPreviousWorkoutView`, and `CreateWorkoutProgram`'s active-program check. (Not-found is a gorm
error, not a SQL one, so plain `First` + `ErrRecordNotFound` checks are fine.) Errors
injected by the e2e fault hooks never reach Postgres, so e2e can't catch this; review for it.

**Conventions for later lanes:**
- Services and repositories take `ctx context.Context` first and call `database.From(ctx)` (once
  at the top as `db := database.From(ctx)`, or per query). Helpers that run inside a
  `Transaction` callback keep taking `tx *gorm.DB`. Handlers pass `c.Request.Context()`.
- Handlers and route registration hold no `*gorm.DB` (`tracking.NewHandler()`,
  `money.NewHandler()`, `media.NewHandler()`, `RegisterXRoutes(group)`).
- Module tests on SQLite: `ctx := database.WithDB(context.Background(), db)`; route tests add
  `r.Use(database.Middleware(db))`.
- `internal/database/guard_test.go` fails on `GetDB`/`SetDB`/`conn()`, a `*gorm.DB` struct field,
  opening a pool, or `WithDB` in `internal/core` / `internal/utils` (testutil excluded; the one
  allowed `WithDB` is the session lookup).
- Every DB-touching request holds one pooled connection from its first query to its response.
  The pool's 20 (A2) is still ample.

Not done here: money's birth-year lookup still goes by the `username` gin key (C4). No BACKLOG
entry was tagged B2 (#8 moved to H1). Prod needs nothing beyond the new image (ROLLOUT 3.3a).

## 2026-11-27: HTTP hardening (lane H1)

**Headers live in Caddy, CSP included.** One place for the whole site, as BACKLOG #7 suggested.
The cost: the CSP allows the inline theme script in `frontend/index.html` by hash, and Caddy
deploys separately from the frontend image. `frontend/src/csp.test.ts` (runs in the pipeline's
`frontend-test`) fails when the hash and the script differ, and whoever edits the script deploys
Caddy first (list both hashes during the switch if old builds must keep working). Rejected:
moving the script to a file (a render-blocking request on every load) and setting the CSP in
nginx (every `location` would have to repeat it; ARCHITECTURE.md "nginx").

**CSP**: `default-src 'self'`; `script-src 'self'` + the hash; `style-src 'unsafe-inline'`
(Vue/reka-ui/motion style attributes); `img-src 'self' data: https:` (external media covers,
until C5 decides on #15); `font-src 'self' data:`; `object-src 'none'`; `base-uri`,
`form-action 'self'`; `frame-ancestors 'none'`. The build has no `eval`/`new Function`.
Lanes adding an external fetch, iframe, worker or inline script must extend it, and check the
console on a local prod copy (ARCHITECTURE.md "Caddy").

**`Referrer-Policy: same-origin`**: keeps C6's `/setup/<token>` URL (and every other path) from
reaching other sites, and costs nothing, since the SPA and API share one origin. `no-referrer`
would also work: tested in Chromium, it changes `Origin` to `null` only on native form POSTs,
not on `fetch`/XHR, and every form in the app uses `@submit.prevent` + axios. A native form POST
under `no-referrer` would get 403 from the CORS middleware, so check that before tightening it.

**Body limits** (`backend/cmd/server/hardening.go`, kept out of the request middleware B2
rewrites): one global middleware, first in the chain, with a 1 MiB default and per-route
overrides keyed by `METHOD /gin/path` (`TestBodyLimitRoutesExist` catches a renamed route).
Only the Renpho import gets more: 5 MiB, and at most `weight.MaxImportRows` = 10,000 days after
same-day rows collapse (that's what the transaction writes; a tab-separated export of 10,000
rows is under 2 MB as JSON). Caddy's `request_body max_size 6MB` is the outer guard. A route that needs
a larger body adds a `bodyLimits` entry and, above 6 MB, raises Caddy's limit too.

**`Cache-Control: no-store` on every backend response**, set by the same middleware (in the
backend rather than Caddy so dev and e2e behave like prod). Nothing in the API wants caching.

**CORS**: no default origin list. With `CORS_ORIGINS` unset the middleware still runs but allows
no origin, so same-origin requests pass and any other `Origin` (including `null`) gets 403. That
keeps cross-site POSTs off the `/auth` routes, which AuthMiddleware's Sec-Fetch-Site check
doesn't cover. Production still requires `CORS_ORIGINS` (`APP_ORIGIN`); `dev/env.sh` and e2e
set their own.

## 2026-11-27: Writes that survive a rollback: `database.Always` (owner, built in C6)

Since B2, everything a request writes is rolled back when the response is 4xx/5xx. Some writes
must survive that, above all audit rows for rejected attempts ("setup link expired", "disabled
user tried to sign in"). Rather than going around the request transaction, the database package
gets one explicit, greppable way to say so:

- `database.Always(ctx, fn func(db *gorm.DB) error)` registers `fn` on the request. The
  middleware runs it **after** the request transaction has committed or rolled back (and after a
  panic), in its own short transaction with the same identity (`app.user_id`, `app.is_admin`), so
  RLS applies as usual. Several calls run in order.
- After, not alongside: no second connection while the request holds one, and no waiting on
  locks the request itself holds.
- Limits: `fn` can't rely on rows the request wrote (they may have been rolled back), gets no
  result back to the handler, and its failure is logged without changing the response (an
  audit write must never turn a success into an error).
- With no request transaction (`OutsideRequestTx`, tools, tests) it runs immediately.
- Use it only for writes that must outlive the request's outcome (audit/security events). Every
  use is visible with `grep -rn 'database.Always'`; the B2 guard test keeps forbidding the other
  ways around the request transaction.
- C6 builds it (first user) with tests: runs after commit, after rollback and after a panic;
  sees the identity; failure doesn't change the response; a rejected setup attempt still leaves
  its `auth_events` row.

## 2026-11-27: S lanes (owner)

Small BACKLOG fixes that sit outside B2's and H1's files run as **S lanes** (TASKS "S lanes"):
one Sonnet subagent per lane, each in its own worktree from `origin/proj/multi-user`. An Opus
session reviews each diff as a security reviewer, fixes it up on the same branch, runs
`dev test` / `dev check`, ticks BACKLOG and TASKS, and squash-merges the PR into
`proj/multi-user` once checks pass, one at a time. The owner doesn't review these PRs
individually. C5 and C6 skip BACKLOG entries an S lane has ticked.

## 2026-11-27: S1 login protection (S lane, reviewed by Opus)

`auth/controller/login_protection.go`, BACKLOG #3, #4, #19 (login part). Env names unchanged.

- **Source** = IPv4 address, or IPv6 /64 (IPv4-mapped unmapped). Used for the per-IP limit
  (`LOGIN_RATE_LIMIT_MAX_IP`) and the 24 h credential-spray block, so hosts sharing a /64 share
  them. Intended: per-address keys let one /64 rotate past both.
- **Strict counter per (source, username)**: `LOGIN_RATE_LIMIT_MAX_USERNAME` attempts per window,
  then 429. One source can no longer lock a username for everyone.
- **Username-wide counter is a serialized backoff, not a block.** After 2×`MAX_USERNAME` attempts
  in a window, each attempt reserves the next slot (250 ms doubling, capped at 5 s per step) and
  waits for it, so parallel sources still get at most one password check per step. If the queue
  is over 10 s long the attempt gets 429 `username_backoff`. One source can't reach the
  threshold (its pair counter stops it first). A plain delay without the queue was rejected in
  review: it doesn't limit parallel guessing.
- **Trusted sources**: a successful login trusts that (source, username) for 30 days (in memory,
  max 1000), and it skips the username-wide queue. A distributed attack can still slow or 429 the
  owner from a new network, not from one they've signed in from before. A restart forgets trust.
- **Success resets only that source's pair counter**, not the username-wide backoff.
- **Full limiter** (100,000 entries): evicts lowest-count, then oldest entries; blocking entries,
  active spray blocks and the current request's keys last/never. No more global `capacity` 429.
- **A client that disconnects during the backoff** is dropped without a password check or log.
- **Logs**: failed attempts log the username only if it exists in `users` (one extra lookup per
  logged failure), otherwise `unknown`. C6's `auth_events` should follow the same rule.
- C6 still owns: lockout rows in `auth_events`, login protection on the setup endpoint.
## 2026-11-27: Accounts and setup links (lane C6)

**`database.Always` as built** (`internal/database/always.go`), per the entry above, plus:
- The middleware runs the queued functions in its deferred cleanup, after `Rollback` (a no-op once
  settled), so the request's connection is already back in the pool. That is still inside the
  handler chain: net/http finishes the response only after it, so a rejection reaches the client
  a few ms later (at most the 5 s limit, if the pool is exhausted). A goroutine would answer
  sooner but escape graceful shutdown and make tests racy; not worth it for a handful of rows. `Request.RunAlways()` is the
  hook; other `WithTx` owners call it after `Commit`/`Rollback`.
- The context is `context.WithoutCancel(request ctx)` with a 5 s limit: a client that hangs up
  doesn't lose the audit row, and a stuck database doesn't hold the handler goroutine forever.
- One transaction for all of a request's functions, **each in its own savepoint**: one that fails
  (a SQL error aborts a Postgres transaction) or panics loses only its own write. Errors and
  panics are logged (`database: always METHOD /route: ...`) and never reach the response.
- `fn` gets a `*gorm.DB`, not a context: `From(ctx)` inside it would be the finished request
  handle (`ErrRequestDone`, tested).
- Without a request transaction (`WithDB`: login, tools, tests) it runs immediately, still in a
  transaction, and still only logs failures.

**Which audit writes go through `Always`.** Only rejections (failed login, rate-limited login,
dead setup link, wrong current password): `services.RecordRejection`. Successful actions
(user created, link created, disabled/enabled, password set or changed) are written **inside the
request transaction** with `services.RecordEvent`, so the event commits or rolls back with the
change it describes. Through `Always` a success event would survive its own rollback (e.g. a
failed commit turning a password change into a 500) and record something that didn't happen.
This narrows "auth_events written through database.Always" to the writes that need it, as the
`Always` entry itself says ("only for writes that must outlive the request's outcome").
`/auth/login` runs outside the request transaction, so its events go straight to the pool.

**Tables** (migration `0011_accounts.sql`):
- `user_setup_tokens(user_id, token_hash, purpose setup|reset, created_by, created_at,
  expires_at, used_at, revoked_at)`. Token: 32 random bytes, base64url (43 chars), stored as
  SHA-256 hex (a fast hash is enough for 256 bits of entropy). 48 h, single use. Creating a link
  revokes the user's unused ones; disabling a user revokes them too. Purpose is derived: `setup`
  if the user has no password, else `reset`.
- `auth_events(created_at, event, user_id, actor_id, client_ip, detail)`. Events and details are
  constants in `auth/models/account_models.go`. `detail` never holds an attempted username or a
  password (BACKLOG #19): a failed login records the account if the name exists, else
  `unknown_user` with no user. Login rate-limit events are written once per window (when the
  limiter says `LogRejection`), so a flood doesn't fill the table.

**RLS (C6's call)**:
- `auth_events`: INSERT for anyone (`WITH CHECK (true)`: failed logins and setup attempts have no
  identity), SELECT only with `app.is_admin = 'true'`, no UPDATE/DELETE policy and those
  privileges (plus TRUNCATE) revoked from `simpletrack_app`: append-only for the app. Inserts are
  raw `INSERT` without `RETURNING`, since RETURNING needs the row to pass the SELECT policy (a
  gorm `Create` would fail for non-admins).
- `user_setup_tokens`: everything for the admin; otherwise a row is visible (and updatable) only
  when `app.setup_token_hash` equals its hash. The setup endpoints set that with
  `set_config(..., true)` before looking the token up, so only a caller holding the token sees
  its row, and nobody can list tokens. Bob can't mint one for himself (RLS violation, tested).
- `users` stays without RLS (session lookup and login need it before there is an identity).

**Endpoints**: admin `GET/POST /admin/users`, `POST /admin/users/:id/setup-link|disable|enable`,
`GET /admin/auth-events?before=&limit=` (AuthMiddleware + `auth.RequireAdmin`, 403 otherwise);
public `POST /auth/setup/inspect` and `POST /auth/setup`; signed in `POST /auth/password`.
- Tokens travel in request **bodies**, never in API URLs. The page URL `/setup/<token>` is the
  only place it appears, and it is logged nowhere:
  - nginx serves that location with `access_log off` (via `rewrite ... break`: a `try_files`
    fallback would finish in `location = /index.html` and log the request line there);
  - the page's own asset requests would carry the URL as `Referer`, so Caddy sends
    `Referrer-Policy: no-referrer` under `/setup/*` (H1's `same-origin` everywhere else, as two
    exclusive matchers) **and** nginx's log format drops `$http_referer` (covers the time before
    the Caddy deploy, and any later Caddy change). Found in /code-review.
  - Caddy keeps no access log; the backend logs route patterns only.
- The setup endpoint validates the password and runs bcrypt **before** its first query, so the
  request transaction (and the token's `FOR UPDATE` lock) is held only for the short write. It
  bumps `token_version` (a reset signs out everywhere) and sets a fresh cookie.
- Setup attempts and password changes are rate-limited with S1's `LoginProtection`, but a
  **separate instance** from sign-in, so they never spend a shared home IP's sign-in budget.
  Setup counts against the "username" `setup-link` (per source: `MAX_USERNAME` per window; all
  sources: S1's username-wide backoff, which the handler waits out like login does), password
  changes against `password-change:<user id>`. Rejections are audited as `login_rate_limited`
  with detail `setup_link:<reason>` / `password_change:<reason>`. `inspect` isn't limited
  (256-bit tokens can't be guessed, and reloading the page shouldn't burn attempts).
- A username taken by a soft-deleted user (the unique index covers every row) or by a
  concurrent create is a 409, not a 500: the insert runs in a savepoint and the unique
  violation on `idx_users_username_lower` maps to `ErrUsernameExists`.
- Disable/enable are no-ops when the account is already in that state (no sign-out, no event).
- Password change hashes the new password before its first query, so the request transaction
  holds its connection for one bcrypt (checking the current one), not two.
- Wrong current password answers **400**, not 401 (the client treats 401 as signed out).
- Disabling yourself is refused (400) so the admin can't lock the only admin out.
- Disabled user: login fails as before, a link can't be created (409) and a live one stops
  working.

**Username and password rules (BACKLOG #14)**: new usernames are trimmed, lowercased and must
match `^[a-z0-9][a-z0-9._-]{1,31}$`. Uniqueness is on `lower(username)` (unique index replaces
the case-sensitive one), and login matches case-insensitively, so an existing mixed-case name
(the owner's, maybe) keeps working and can't be duplicated in another case. Existing names aren't
rewritten. Passwords being set: at least 10 characters, at most 72 bytes (bcrypt's limit). Login applies
neither: an existing password may be shorter, or longer than 72 bytes if it was set when the
binding counted characters (bcrypt compares the first 72 bytes, as it hashed them). An account without a password can't sign in; it answers
like a wrong password, in the same time (dummy bcrypt compare).
- `users.email` unique index now ignores NULL/empty (accounts are username-only; two `''` emails
  used to collide). `authrepo.CreateUser` stores NULL for an empty email.

**Request logs (BACKLOG #19, request part)**: `gin.Default()`'s logger is replaced by
`cmd/server/requestlog.go`, which logs the route pattern (`/media/:id`), never the raw path or
query string; unmatched paths log as `(no route)`. The login-log part of #19 is S1's
(`login_protection.go`), and auth_events already follows it.

**S lanes** (merged before this PR; C6 rebased on them): #3, #4 and the login-log part of #19 are
S1's, #9 is S2's. C6 adds #3's audit part: lockouts and backoff rejections are
`login_rate_limited` rows (once per window per key, like S1's log line). The setup page signs a
possibly different user in on a shared device: `markAuthenticated(name, id, admin)` passes the
id to S2's `setUserStorageOwner`, which clears the previous user's storage, and the page clears
the query cache.

