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

## 2026-11-27: B3 owner kit (agent, lane B3)

Everything a C lane needs to convert a table. Reference conversion: `drink_size_presets`
(migration `0011_owner_kit.sql`, `tracking/water`, `water/presets_pg_test.go`). Copy those.

**Go API** (`internal/database`: `owner.go`, `notfound.go`, `lock.go`):
- Owned model: embed `database.Owner` (`UserID`, `json:"-"`, so it's never bound from or sent
  in JSON). Catalog model: embed `database.Creator` (`CreatedBy`, visible in responses).
- **Stamping is a gorm `BeforeCreate` hook on `Owner`/`Creator`**, the only code that sets
  `user_id`/`created_by`. It reads the user from the statement's context, so every create path
  (`Create`, batch, `Save` of a new row, `FirstOrCreate`) is covered without callers remembering
  a helper. A preset `UserID` that differs from the user is an error, never overwritten. With no
  user on the context (seeds, tools on the owner role) the caller must set it, or the create
  fails with `database.ErrNoIdentity`. Chosen over an explicit `Stamp(ctx, &row)` call because a
  forgotten call would only show up on Postgres (FK error) and pass on SQLite module tests. A
  model with its own `BeforeCreate` must call `x.Owner.BeforeCreate(tx)` (rules test). Not
  covered: `Create(map...)` and raw `INSERT`, which skip hooks; RLS `WITH CHECK` still refuses
  them.
- Reads: `database.From(ctx).Scopes(database.Owned(ctx))`. The column is qualified with the
  statement's table (safe in joins) and survives `Unscoped()`. **Preloads don't inherit
  scopes**: `Preload("Items", database.Owned(ctx))`. No user on ctx -> the query fails with
  `ErrNoIdentity`, never runs unfiltered.
- Not found or not yours: `database.IsNotFound(err)`, `database.Affected(res)` (update/delete by
  ID: `ErrNotFound` when 0 rows), and in handlers `apierr.NotFoundOrInternal(c, err)`: 404
  `{"error":"not found"}` or a logged 500. Same body for missing and someone else's.
- Identity: `database.CurrentIdentity(ctx)` (the request's, from `SetIdentity`), and
  `database.WithIdentity(ctx, id)` for SQLite module tests and tools only (rules test forbids it
  in request code). SQLite test: `ctx := database.WithIdentity(database.WithDB(bg, db), database.Identity{UserID: 1})`.
- Advisory locks (#2): `database.XactLock(ctx, tx, "workout_logs.date", date)` takes
  `pg_advisory_xact_lock(hashtext("u<uid>:workout_logs.date:<date>"))`; `database.LockKey(ctx,
  parts...)` builds the key. Two users never share a key (tested: Bob doesn't wait on Alice).
  Raw `pg_advisory` calls outside `internal/database` are allowlisted for C2 (`lockClientID`) and
  C3 (`CreateMinimalIfMissing`), which also need their existence checks scoped (`Unscoped()` too).
- Creates: `Omit(clause.Associations)` in the chain (#1). `database.CreateInSavepoint` (B2,
  `request.go`, which C6 is editing) doesn't omit them yet; values passed to it must not carry
  nested structs until it does.

**SQL** (migration 0011 adds `public.app_user_id()`, `public.app_is_admin()` and
`public.catalog_guard()`). The tested copy of every pattern is
`backend/internal/database/testdata/rls_patterns.sql` (run by `TestRLSPatterns` against legacy-
shaped tables); copy from there. In short, per table, in the same migration as the code:
- **Owned**: `ADD COLUMN user_id bigint`; backfill DO block (rows -> the one live admin, else
  `RAISE`, see 0011); `SET NOT NULL`; `FOREIGN KEY (user_id) REFERENCES users (id)` (no cascade:
  users are soft-deleted); index on `user_id`; per-user uniques (`(user_id, name)`,
  `(user_id, client_id) WHERE client_id IS NOT NULL`, `(user_id, date)`); `UNIQUE (id, user_id)`
  if anything points at it; `ENABLE ROW LEVEL SECURITY`; policy
  `owner USING (user_id = public.app_user_id()) WITH CHECK (user_id = public.app_user_id())`.
- **Owned child**: same column, backfilled from the parent (`UPDATE child c SET user_id =
  p.user_id FROM parent p WHERE p.id = c.parent_id`; nullable parents fall back to the admin),
  then the parent FK becomes composite `(parent_id, user_id) REFERENCES parent (id, user_id)`
  with the old `ON DELETE`. Index `(user_id, parent_id)`. Same `owner` policy. **Why composite:
  FK checks bypass RLS**, so with a plain FK Bob could attach a row to Alice's parent by ID.
  Also applies to non-child references such as `water_logs.preset_id` (C1:
  `(preset_id, user_id) REFERENCES drink_size_presets (id, user_id)`; 0011 added the key).
- **Catalog**: `created_by` like `user_id` (backfill to admin, NOT NULL, FK, index); policies
  `catalog_read FOR SELECT USING (app_user_id() IS NOT NULL)` (signed-in users only, so a missing
  identity still sees nothing), `catalog_add FOR INSERT WITH CHECK (created_by = app_user_id())`,
  `catalog_edit FOR UPDATE USING/WITH CHECK (created_by = app_user_id() OR app_is_admin())`,
  `catalog_delete FOR DELETE USING (app_is_admin())`, and
  `CREATE TRIGGER catalog_guard BEFORE UPDATE ... EXECUTE FUNCTION public.catalog_guard()`.
  The trigger exists because gorm's soft delete is an `UPDATE deleted_at`, which no DELETE policy
  sees: it refuses `deleted_at` changes by non-admins and any `created_by` change. It only acts
  where RLS does (`row_security_active`), so migrations and seeds on the owner role are free.
  Handlers still check admin before deleting so the user gets a clean 403/404, not a 500.
- `app_user_id()` is `NULLIF(current_setting('app.user_id', true), '')::bigint`: after a pooled
  session has served a request the setting reads `''`, which a bare cast would turn into an error
  (`TestAppUserIDAfterRequest`).

**Tests**:
- `internal/database/pgtest`: `Migrated(t, prefix)` (fresh DB copied from a template built once
  per migration set), `NewDatabase`, `Open` (owner), `OpenAsApp` (every session `SET ROLE
  simpletrack_app`, so RLS applies), `Admin`. Skips without `TEST_DATABASE_URL`; **fails** with
  `TEST_DATABASE_REQUIRED=1`. Every existing Postgres test now goes through it (migrate's own
  tests copy the rule, since pgtest imports migrate).
- `internal/database/twousers`: `h := twousers.New(t)` gives a migrated DB with `h.Admin`,
  `h.Alice`, `h.Bob`, `h.Owner` (bypasses RLS: arrange and inspect), `h.App` (app role);
  `h.Router(register)` serves routes through `database.Middleware` + `auth.AuthMiddleware` like
  the server, `api.Do(h.Bob, "GET", path, body)`, `res.Status(t, 200)`, `res.NotFound(t)`;
  `h.Run(h.Alice, func(ctx) error)` for service tests. Sets `JWT_SECRET` (no `t.Parallel`).
  Per endpoint: Bob sees an empty list, gets the same 404 as a missing ID for Alice's rows on
  GET/PUT/DELETE, can't create a child under (or point at) Alice's row, and a forged
  `user_id` in a body is ignored. Plus one RLS backstop test per module (unscoped query as Bob).
- `TestRLSCoverage` (`internal/database/ownerkit_pg_test.go`): every public table has RLS with
  a policy, except `schema_migrations` and `users` (exempt) and the `rlsPending` allowlist (each
  table tagged with its C lane; delete yours in your PR; D1 needs it empty). Also every
  `user_id`/`created_by` column is NOT NULL with an FK to users. **C6's new tables need RLS or an
  entry with a reason.**
- `TestOwnerKitRules` (`internal/database/rules_test.go`, #1/#2): no gin bind into a type that
  (transitively, by type name) carries `UserID`/`CreatedBy`/`Owner`/`Creator`, and every bind
  target's type must be visible (`var body T`); no writes of `UserID`/`CreatedBy`/`"user_id"`/
  `"created_by"` outside the hooks; no model embedding both `Owner` and `Creator` (neither hook
  would be promoted); no `Create`/`Save`/`FirstOrCreate` without
  `Omit(clause.Associations)`; no raw `pg_advisory`; no shadowing `BeforeCreate`. Allowlists are
  per-file counts of today's sites: a file may drop below its count (or vanish), never exceed it.
  C6's `user_setup_tokens.user_id` (set by an admin for another user) needs an allowlist entry.
- `TestNoRawErrorsIn500s` (`internal/utils/apierr/rules_test.go`, #6): same per-file counts
  for 500 responses containing `.Error()`. It doesn't catch 4xx responses that echo a database
  error (water logs, grocery do today): C lanes fix those by eye when moving to `apierr`.

**CI (#20 and Postgres tests)**: `backend-test` runs `dev/ci/backend-test.sh` (throwaway
`postgres:16-alpine` container, `TEST_DATABASE_REQUIRED=1`); new steps `backend-vulncheck`
(`govulncheck@v1.10.0`, pinned to the version S4 audited with) and `frontend-audit` (`npm audit --omit=dev --audit-level=high`).
Proved locally by running the step's exact command from `backend/` (full suite green, and
`TEST_DATABASE_REQUIRED=1` without Postgres fails). On the rebased branch (S4's x/net and x/text
bump) govulncheck reports 0 reachable vulnerabilities and `npm audit` 0 findings. govulncheck
judges the standard library by the build host's Go, not the image's: keep the image's Go minor
current.

**Seeds**: devseed gives seeded owned rows to the dev login (`devOwner`; devseed now runs after
devuser) and seeds two presets as the pattern; lanes pass `owner` to their module's seed
function. The e2e seed has no presets (specs create them through the API as `dev`, and expect
none at start), so it's unchanged.

**Not done here**: water logs, grocery and everything else in `rlsPending` (C lanes). Until C1
makes `water_logs` owned, every user still sees every water log, and another user's log shows
no preset (its preload is scoped with `Owned`). Nothing ships before D1, which needs C1. POST
`/water` moved to sentinel errors + `apierr` here, since the scoped preset lookup could
otherwise echo `ErrNoIdentity` or SQL text as a 400; the other water-log handlers stay for C1.

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

**Tables** (migration `0012_accounts.sql`):
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

**Frontend**: `/setup/:token` (public route: `meta.public` skips the auth guard, hides the nav,
and is never saved as the last route), Account menu → **Account** (change password, everyone)
and **Users** (admin: add someone, link overlay with Share/Copy, disable/enable, sign-in
activity). `isAdmin` comes from `/auth/me` (`user.is_admin`); the server re-checks on every
request, the client flag only hides the menu entry and guards the route (`meta.admin`).

**Merged after B3** (owner): migration renumbered to `0012_accounts.sql`. The auth tables are
identity tables, not owned data, so B3's rules get reasoned exceptions rather than the Owner hook:
`ownerFieldWriteAllowed` entries for the auth files (their `user_id` is the account acted on, often
not the caller) and `nullableOwnerColumn` for `auth_events.user_id` and
`user_setup_tokens.created_by`. `accounts_pg_test.go` now uses `pgtest`, so CI fails instead of
skipping it.

## 2026-11-27: Tracking is per user (lane C1)

Migration `0013_tracking_owned.sql`, one file for all five tables (they share one backfill and
one admin lookup, and one file renumbers once when C2-C5 merge first). Same shape as B3's
reference (`drink_size_presets`); tests in each package (`*_pg_test.go`) and
`internal/database/tracking_migration_pg_test.go`.

- **Owned, no catalogs**: `body_weight_logs`, `step_logs`, `water_logs`, `grocery_items`,
  `user_profiles`. Each: `user_id` backfilled to the admin (soft-deleted rows too; stops unless
  exactly one live admin, when there are rows to give away), NOT NULL, FK, `owner` RLS policy.
  Models embed `database.Owner`; every read is `Scopes(database.Owned(ctx))`, including the
  client_id lookups, the preset preload and the aggregates (`ParseTextList`, `ListSuggestions`,
  `GetMissedYesterday`).
- **Uniques**: `(user_id, date)` on weight and steps (replacing the global date index),
  `(user_id, client_id) WHERE client_id IS NOT NULL` on water and grocery. The same client_id
  from two users is two rows, and each user's replay stays idempotent. They are not partial on
  `deleted_at`, like the indexes they replace (a deleted row still answers its client_id replay).
- **`water_logs.preset_id`** is the composite FK `(preset_id, user_id) -> drink_size_presets
  (id, user_id)` that 0011 prepared. With a NULL preset the FK isn't checked (MATCH SIMPLE), so
  logs without a preset are unaffected. A test inserts a log for Bob with Alice's preset as the
  schema owner and the FK refuses it.
- **`user_profiles` is one live row per user**: unique `(user_id) WHERE deleted_at IS NULL`. The
  row is created on the first `PUT /tracking/profile`; `GET` still answers `{"profile": null}`
  until then, so a new account sees the onboarding state. The old table was meant to hold one
  row (the app read and updated the oldest). If prod somehow has several live rows, the
  migration soft-deletes all but the oldest rather than failing or dropping data (ROLLOUT 1.9
  checks the count first). The first-save race (two requests for a user with no row) is
  handled: the loser's insert fails in a savepoint and it updates the winner's row.
- **Upserts** (weight, steps) do the same: find the user's row for the date, else create in a
  savepoint, and on a lost race update the winner's row. They answered a raw unique-violation as
  a 400 before.
- **Renpho import** writes as the importing user: the create hook stamps each new row and the
  existing-day lookup is `Owned`, so importing the same days as Alice leaves her rows alone. The
  update path is `Select("*").Omit(clause.Associations, "user_id").Save(...)`: `Select("*")`
  writes every column, and without the `Omit` it would write the model's zero `user_id` over
  the row's. (Rewriting `UserID` by hand is what the rules test forbids, so `Omit` is the way.)
- **`database.CreateInSavepoint` now omits associations** (it was the one create path the B3
  rule couldn't see). `WaterLog.Preset` is the one nested struct it matters for.
- **Errors (BACKLOG #6)**: every tracking handler is on `apierr`. Client mistakes are
  `common.Invalid(msg)` (400 with our own message, answered by `common.Respond`), a missing or
  someone else's row is the shared 404, anything else is a logged 500 with a fixed message. The
  4xx paths that echoed database errors (water delete, grocery complete/delete, weight, steps and
  profile upserts) now answer 500 for database failures. Bind errors answer `invalid body`
  instead of the binder's text. Validation messages and the Renpho parser's row/column messages
  are unchanged. The request bodies were already DTOs (`postWeightBody`, `postWaterBody`,
  `itemBody`, `putProfileBody`, ...), none binds a model.
- **Allowlists**: `rlsPending` has no C1 tables; the tracking files are gone from
  `createWithAssociationsAllowed` and apierr's `rawInternalErrorAllowed` (all their creates and
  saves now carry `Omit(clause.Associations)`).
- **Seeds**: devseed's weight and step rows belong to the dev login (the `owner` it already
  passes); the e2e `large` profile looks the `dev` user up (`devUser`) for its weight, step and
  grocery rows. The base e2e seed has no tracking rows.
- **Not done**: no advisory locks anywhere in tracking (upserts rely on the unique indexes, as
  before), so nothing for B3's `XactLock`. `missed` has no table of its own (it counts weight and
  step logs), so it only needed `Owned` and `apierr`.

## 2026-11-27: C4 money (agent, lane C4; split approved by the owner)

Migration `0014_money_owned.sql` (renumbered from 0013 when merged after C1).

**What is shared and what is personal**
- Catalog (`created_by`, catalog policies, `catalog_guard`): `investment_account_types` (name,
  eligibility age, carry-forward cap, lifetime limit, withdrawals-restore) and
  `investment_contribution_rules` (the per-year limits). Anyone signed in adds, the creator or the
  admin edits (handler answers 403, not 404: the row is visible), only the admin deletes (403).
- Per user, new table `investment_contribution_room` (`user_id`, type, `contribution_start_year`,
  `opening_room_year`, `opening_room`; unique per user and type; plain FK to the catalog type).
  **Start year moved here too**, not just opening room: it is "the first year *you* could
  contribute" (e.g. when you opened an FHSA). The migration copies each type's three old
  columns into an admin-owned row and drops them. Birth year still lives on `users` (C1 moves
  it with the profile: update `currentUserBirthYear` in `money/controller/routes.go` then).
- Owned: `investment_accounts`; children `account_balances` and `investment_deposits` with
  composite FKs `(account_id, user_id)`. An account's type FK stays a plain FK to the catalog.
  Net worth, room flows and every total only read the caller's rows (`Owned`).

**API**
- `PUT /money/contribution-types/:id` takes the shared rules only (creator or admin, else 403).
  `PUT /money/contribution-types/:id/room` saves the caller's own figures (any signed-in user;
  all null clears them) and returns the type. `POST /contribution-types` and the preset POST still
  accept `start_year` / `opening_room*` and store them as the caller's room.
- `GET` types merges the caller's room figures into each type under the same JSON names
  (`start_year`, `opening_room_year`, `opening_room`) and adds `created_by` and `can_edit`.
  The frontend dialog saves rules and room separately and hides the rules (and the limit form)
  from users who can't edit; delete buttons show for the admin only.
- **Preset reuse**: creating a preset whose name already exists (another user added their TFSA
  first) reuses that type and only saves the caller's room, instead of a 400. A manual create
  with a taken name is still a 400. Two users both have their own accounts called "TFSA"
  (account names were never unique); both use the one shared TFSA type, each with their own room.
- Deleting a type is admin-only and refused (400) while an account of the admin's uses it; if
  another user's account uses it the FK stops it (also a 400, via 23503). Admin can't see other
  users' accounts, so the FK is the only check there. Limits: delete is admin-only (hard delete
  by policy); upsert is creator or admin.
- Limit rows: `investment_contribution_rules.created_by` is the type's creator at migration time
  and the caller afterwards. Their UPDATE policy also lets the **type's** creator edit, so a limit
  the admin added to Bob's type stays editable by Bob (plain catalog policy would not).
- `respondError` now uses `apierr`: 400 for validation, 403 for `ErrForbidden`, 404/500 via
  `NotFoundOrInternal`. A 500 shows "Internal server error" instead of the raw error, so the e2e
  atomic-deposit spec asserts that text (the raw error is in the backend log).

**Not done / for later**: two users creating the same new preset at the same instant would hit
the unique name index on the second (a 500, then retry works); not worth a lock. devseed and the
e2e `large` profile give the seeded accounts to the dev login (`seedOwner`). The browser
click-through was not possible in this session (the shared Playwright browser was busy); the
e2e finances specs and a before/after API comparison across `migrate down 1` + `up` (net worth,
accounts, types and room unchanged) stand in for it.

**Review (/code-review high, scoped to this lane) and what was deferred**: nothing important. Fixed:
a unique-index race on shared types or first room rows now answers 409, not 500. Deferred, for the
owner: (1) the creator can add a limit but only the admin can remove one (a typo'd year stays
until the admin removes it; deletes are low priority per the owner); (2) deleting a type cascades
away other users' room rows (the admin can't count them through RLS); (3) the dialog saves rules
then room as two requests, so a room validation error after rules saved shows "Failed to save"
though the rules persisted; (4) `created_by` (a user ID) is visible on catalog types, as the kit
specifies; (5) migration 0013's down is lossy with several users.

## 2026-11-27: C5 media (agent, lane C5)

`media_items` is the shared catalog; `user_media` holds everything personal. Migration
`0015_media_catalog.sql` (renumbered from 0013 when merged after C1 and C4), owner-approved column split ahead of the code.

- **Catalog (`media_items`)**: `type`, `title`, `creator` (director/author/...), `release_year`,
  `genres`, `cover_url`, `synopsis`, `length` (total pages/episodes/hours: a fact about the item),
  `source` + `external_id`, `created_by`. Catalog policies and `catalog_guard` as in B3. A partial
  unique index on `(source, external_id)` (live rows, both non-empty) means adding something that
  already exists reuses the row instead of duplicating it; hand-entered items have no external ID
  and are never deduplicated. NULL `source`/`external_id` were normalised to `''`.
- **Owned (`user_media`)**: `user_id`, `media_item_id`, `status`, `rating`, `review`, `favorite`,
  `progress`, `started_at`, `finished_at`, `times_consumed`. Unique `(user_id, media_item_id)`
  **where `deleted_at IS NULL`**, so a removed save can be saved again (gorm soft delete).
  `media_item_id` is a **plain FK** to the catalog: everyone may read the catalog, so the
  composite-FK reason (FK checks bypass RLS) doesn't apply. A soft-deleted catalog entry leaves
  people's saves in place but unlisted (the list joins live catalog rows).
- **Backfill**: every `media_items` row (soft-deleted too) gets `created_by` = the one live admin
  and a `user_media` row for them with `id = media_items.id`, so `/media/:id` URLs and the frontend
  cache keep working; the sequence is advanced. Live duplicates of a (source, external_id) pair
  would break the new index: all but the oldest lose the pair (they stay as hand-entered rows,
  nothing is deleted). Down is lossy with several users (an item takes the lowest user's fields).
- **API** (`/media`): the flat item shape stays (catalog + personal fields side by side, `ID` is the
  `user_media` id) plus `media_item_id`, `created_by`, `can_edit_catalog`. `GET /media` and
  `/stats` are the caller's saves only. `POST /media` finds-or-creates the catalog row then saves it
  (201 new save, 200 if they already had it, which is returned untouched; an existing catalog row's
  fields are never overwritten by a later POST). `PATCH /media/:id` takes **personal fields only**;
  `DELETE /media/:id` removes the save, not the catalog entry. New: `GET /media/catalog?q=&type=`
  (with the caller's `saved_id`), `POST /media/catalog/:id/save`, `PATCH /media/catalog/:id`
  (creator or admin, else **403**; catalog rows are visible to all so 403 leaks nothing),
  `DELETE /media/catalog/:id` (admin only, else 403; soft delete). A save of someone else's, or a
  missing ID, is the shared 404 (`apierr.NotFoundOrInternal`).
- **Request DTOs**: `CatalogInput`, `PersonalInput` and `MediaInput` (both) have no `user_id` or
  `created_by`, so forged values are never read. `Omit(clause.Associations)` on every create/save;
  find-or-create and save run in a savepoint (`db.Transaction`) so losing a race to a unique index
  doesn't poison the request transaction, then re-read.
- **Name clash**: `database.Creator` is embedded in `MediaItem`, so the director/author column's Go
  field is `Author` (`json:"creator" gorm:"column:creator"`). Response DTOs name the creator's ID
  `AddedBy` (json `created_by`) so the rules test's "writes `CreatedBy`" grep sees no false positive.
- **Reads of `user_media`** all go through `saves(ctx)` (`database.Owned` + a join to live catalog
  rows); the catalog search reads `user_media` for its `saved_id`s with `Owned` too, not a join.
- **#15 covers**: `lookup.Service.CoverAllowed` (https + TMDB / Open Library / AniList / Apple
  image hosts, plus the service's configured image bases for e2e's fake upstream) is checked on
  catalog create and edit (400), and stored covers failing it are blanked in responses. No per-user
  cover column: BACKLOG #15 suggested one, but nothing needs it yet.
- **Frontend**: the add page lists matching shared catalog entries ("Add to my list", or "On your
  list"); the detail page edits catalog fields (disabled otherwise, with a note) through
  `PATCH /media/catalog/:id` and personal ones through `PATCH /media/:id`; "Delete" became "Remove
  from my list". No UI for the admin's catalog delete yet (API only).
- **Tests**: `controller/media_pg_test.go` (twousers: lists/stats/get/patch/delete per endpoint,
  same catalog item saved independently, catalog permissions, forged `user_id`/`created_by`, cover
  validation, RLS backstop as Bob incl. `catalog_guard`); SQLite service tests rewritten for the
  split; e2e specs for the catalog save flow and the cover rule. `media_items` left `rlsPending`;
  the media files' allowlist entries are gone from the rules tests.
- **Seeds**: devseed creates the catalog entries and the dev login's saves; the e2e `large` seed
  does the same for the `dev` user (the base e2e seed has no media).
- **Known limits (from the C5 review)**:
  - *First writer wins on an external ID.* A signed-in user can add a catalog row for a
    `(source, external_id)` with wrong metadata; later people adding that title reuse it and only
    its creator or an admin can fix it. Family-only accounts make this acceptable for now. If it
    bites: fill the metadata server-side from `lookup.Details` when a source is given, and add an
    admin UI for the catalog delete (API only today).
  - *Legacy covers are hidden, not cleaned.* A stored cover outside the allowlist stays in the
    database but is blanked in responses (ROLLOUT pre-check 1.8 lists them). The edit form only
    sends catalog fields the user changed, so rating an item never wipes the hidden cover.
  - Updates use explicit column lists (`Select(...).Updates` + `database.Affected`), not `Save`,
    because gorm's `Save` falls back to an upsert when a concurrent soft delete leaves 0 rows.
  - A legacy row with half a `(source, external_id)` pair stays editable; the "go together" rule
    only applies when an edit sets one of them.

## 2026-11-27: C2 diet (agent, lane C2; shape approved by the owner)

Migration `0016_diet_owned.sql` (after C1, C4, C5; C3 renumbered to 0017). One
file for all eleven tables, tests in `core/diet/diet_pg_test.go` (twousers) and
`internal/database/diet_migration_pg_test.go` (legacy-shaped data).

**What is shared and what is personal**
- **Catalog: `foods` only** (`created_by`, catalog policies, `catalog_guard`). Anyone signed in
  adds, the creator or an admin edits (403 for others: the row is visible, so 403 leaks nothing),
  only an admin deletes (soft delete; 403 otherwise). New endpoints `PUT /diet/foods/:id` and
  `DELETE /diet/foods/:id`; the frontend has no UI for them yet (like C5's catalog delete).
  Food names stay globally unique (a taken name is a 409, as before).
- **Recipes are personal** (owner's decision): `composite_foods` and its items are owned, name
  unique per user (`(user_id, name)`, a repeat is a 409). A recipe is "my pico: 50 g tomatoes,
  50 g onion", not something to share.
- **Owned**: `plans`, `days`, `meals`, `planned_meals`, `saved_meals`, `composite_foods`; owned
  children `day_logs`, `meal_items`, `saved_meal_items`, `composite_food_items`.
- **Nothing personal lived on a food row** (no favourites, last-used or custom servings in the
  schema or the frontend), so no per-user food table. A meal item is `food_id` (plain FK to the
  catalog: everyone may read it, so the composite-FK reason doesn't apply) on the user's own meal
  (composite FK `(meal_id, user_id)`).
- **Quick-entry foods** are the hidden food row behind a quick-logged meal: effectively part of
  that logged meal. They stay catalog rows (`created_by` = the logger) so the item FK shape is the
  same for every item, but the catalog read policy hides a `quick_entry` food from everyone except
  its creator (not even the admin), so one person's quick-logged calories can't be read by ID.
  The picker already filtered them out.

**Schema**
- Composite FKs (FK checks bypass RLS): every child to its parent as in the kit, plus
  `days.plan_id -> plans`, `day_logs.meal_id -> meals`, `planned_meals.meal_id -> meals`,
  `planned_meals.saved_meal_id -> saved_meals` and both `composite_food_id` columns ->
  `composite_foods`, all on `(id, user_id)`. The last three had no FK at all before; they're added
  `NOT VALID` and validated in the migration only if the data is clean (a dangling legacy ID
  logs a notice and leaves the key `NOT VALID`, which still guards new rows; ROLLOUT 1.10 lists
  them). Nothing is changed or dropped.
- Uniques per user: `days (user_id, date)`, `meals (user_id, client_id)`, `planned_meals (user_id,
  client_id)` (both partial on `client_id IS NOT NULL`, as before), `composite_foods (user_id,
  name)`. `uq_day_logs_day_meal` is unchanged (`day_id` already implies the user).
- The SQLite module tests no longer have a unique index on `days.date` (the model tag moved to
  the migration), so the lost-race fallback in `findOrCreateDietDayForCalendarDate` is only
  exercised on Postgres.
- Down is lossy only if two users collide on the restored global uniques.

**Code**
- Owned models embed `database.Owner`, `Food` embeds `database.Creator`. Every read of an owned
  table is `owned(ctx)` / `ownedTx(ctx, tx)` (`database.Owned`), including the `Unscoped()`
  client_id counts, updates and deletes by ID (`database.Affected`) and every preload level
  (`preloadMeal`). `CalculateTotals`'s raw SQL filters `dl.user_id`.
- **Nested writes are gone** (BACKLOG #1): requests bind into DTOs (`mealBody`, `savedMealBody`,
  `compositeFoodBody`, `foodBody`), never models, and every create is
  `Omit(clause.Associations)`; a meal and its lines are written as a meal, then its items
  (`createMealTx`, `insertMealItems`). A nested `food` object in a body is ignored (tested).
- Before writing items the repository checks each `food_id` is a live food the caller can see
  and each `composite_food_id` is the caller's own recipe, else 400 (an FK check would accept
  another user's ID). A deleted food stays visible on the meals that used it (`Food` preloads
  are `Unscoped`).
- **Locks** (BACKLOG #2): `lockClientID` is `database.XactLock(ctx, tx, table+".client_id", id)`;
  the raw `pg_advisory` call and its allowlist entry are gone. The "does this client_id exist"
  counts (`Unscoped()` included) are `Owned`, so Alice's replay isn't Bob's and the same
  `client_id` from both makes two rows. Tests: both users quick-log and plan with the same
  `client_id` (each replay idempotent) and Bob's quick log doesn't wait while Alice holds the
  lock. Day lookups and the find-or-create-day path are `Owned` too, so two users on the same
  date each get a day.
- **Variant links** (the picker's "ground beef 90/10, 80/20" shortcut): linking never touches
  meals. Joining an existing group writes only the new food. *Starting* a group writes
  `variant_group_id` on the related food, which is an edit of a catalog row, so only its creator
  or an admin may; anyone else gets a 403 and nothing is created. Group IDs are still
  `MAX(variant_group_id)+1` over the whole catalog, so two people starting groups at the same
  instant could share one (not worth a lock; the groups would merge, no data is lost).
- **Errors** (BACKLOG #6): diet handlers are all on `apierr` (`respondErr`: our sentinels,
  the shared 404, a logged 500). `GET /diet/meals/meal/:id` answered 500 for a missing meal; it
  is a 404 now. `replace_meal_id` mismatches, reorder mistakes and unknown foods are 400s with
  our own text. The `diet` files are out of `createWithAssociationsAllowed`, the raw-advisory
  list and apierr's raw-error list; `foods` and the other ten tables are out of `rlsPending`.
- **Plans**: a new user gets their own "Default" plan on first use (as the first user did).
  Targets are `PUT /diet/plans/plan/:id` on your own plan only. An onboarding to set targets at
  account creation is planned (owner).
- **Seeds**: devseed and the e2e seeds give every diet row to the dev login
  (`seedDiet(tx, owner)` / `devUser`), foods `created_by` the same.
- **Frontend**: no change. Its request bodies already carry only IDs and amounts (extra fields
  like a nested `food` are now ignored by the DTOs), and responses keep their shapes (foods
  gain `created_by`).
- **Review (/code-review high, scoped to this lane)**. Fixed: item checks accept a food the admin
  soft-deleted (old meals, saved meals and copied days keep working; the picker hides it), and
  `MAX(variant_group_id)` counts deleted rows so a group ID isn't reused. Known, left as is:
  (1) a food's creator editing its macros changes everyone's past totals, since meal items
  reference the shared row (copy-on-write is the fix if it matters); (2) starting a variant group
  on a food you didn't add is a 403, so non-admins can't group the migrated (admin-owned) foods
  unless the group exists; (3) pre-existing: `POST /meal/new` with a repeated `client_id` is a 500
  not a replay, and meal/saved-meal/recipe bodies don't reject negative amounts (only `foodBody`
  validates).

## 2026-11-27: C3 workout (agent, lane C3)

Migration `0017_workout_owned.sql` (after C2's 0016; renumbered when C3 merged second).

**What is shared and what is personal**
- Catalog (`created_by`, catalog policies, `catalog_guard`): `exercises` (name, rep rollover, cues,
  load type, unilateral). Anyone signed in adds, the creator or the admin edits (the handler answers
  **403**, the row is visible), only the admin deletes. `exercises.name` stays globally unique, so
  adding a name someone already added is a 409.
- Per user, new table `exercise_settings` (`user_id`, `exercise_id`, `machine_settings` jsonb;
  unique `(user_id, exercise_id)`). `exercise_id` is a **plain FK** to the catalog (everyone reads
  it, so the composite-FK reason doesn't apply, as in C5). The migration copies each exercise's
  0003 `machine_settings` into an admin-owned row and drops the column. Nothing else on an exercise
  row was personal (`rep_rollover` is arguably a preference; left shared). Notes already live on
  `logged_exercises`.
- Owned: `workout_programs`, `workout_plans`, `workout_logs`; children `workout_plan_days`,
  `workout_plan_exercises` (no `id`: composite PKs stay), `logged_exercises`, `logged_sets`,
  `cardios`. Composite FKs `(parent_id, user_id)` everywhere a row points at an owned row,
  including `workout_logs.workout_plan_id` and `workout_plans.workout_program_id`. Plain FKs to the
  catalog from `workout_plan_exercises.exercise_id` and `logged_exercises.exercise_id` (the plan
  exercise one never existed and still doesn't).
- Uniques per user: `idx_workout_logs_user_date` on `(user_id, date) WHERE deleted_at IS NULL`
  (new: the date had no unique index) and `idx_logged_sets_client_id` on `(user_id, client_id)`
  (was global). `CreateSet`'s `ON CONFLICT` targets `(user_id, client_id)`.

**Duplicate dates (prod)**: prod had one pair of empty logs for 2026-04-03 (ids 124/125, same
plan, no exercises, sets or cardio). Step 0 of the migration keeps the oldest live log per date
and soft-deletes extras that are empty; an extra holding exercises, cardio or mobility ticks
**raises** instead of merging (ROLLOUT 1.11 has the query to re-check). Soft-delete, not hard:
nothing is lost and the partial index ignores it.

**Fresh databases and new accounts**: the baseline inserted a placeholder "Default" program that
nobody can own on an empty database, so the migration deletes it when there is no admin and no
other workout data (prod keeps it, as the admin's). A new user therefore has **no program**; the
plans page already has a "No programs yet" state and the first program created becomes the active
one (`CreateWorkoutProgram`), so no lazily created default. The migrate test now expects 0 active
programs on a fresh database.

**API**
- Exercise JSON gains `created_by` and `can_edit`, and `machine_settings` is the **caller's own**
  (filled by `DecorateExercises` in one query wherever exercises leave the module: library list,
  plan exercises, a day's logged and planned exercises, a logged exercise). Month views and the
  previous/max logs don't decorate (they never show settings).
- New `PUT /workout/exercises/:id/settings` (`{"machine_settings":[{"label","value"}]}`): any
  signed-in user, for any catalog exercise, saves their own; an empty list clears them; blank rows
  dropped, labels required, 20 rows and 60 characters max. 404 for a missing exercise.
- New `DELETE /workout/exercises/:id`: admin only (403 otherwise), 409 while the admin's own plan or
  log uses it, soft delete. **API only, no UI.** Other users' use is invisible to the admin (RLS),
  so deleting an exercise someone else still logs hides it from their old logs (preloads skip
  soft-deleted rows). Fine at family scale; the fix would be a `SECURITY DEFINER` usage count.
- Logging against a soft-deleted exercise still works (old sessions), as before.
- Everything that names an owned row by ID answers the shared 404 body (`not found`) for missing
  and someone else's: plan, program, set, logged exercise, workout log. `RemoveExerciseFromPlan`
  checks the plan first, so "exercise not in plan" (`ErrNotInPlan`, 404) only ever means *your* plan
  lacks it.
- `POST /workout/exercises/log` (the legacy full replace the notes save still uses) now binds a DTO
  (`logExerciseRequest`): no `user_id`, no nested exercise, no `client_id`; the log and exercise it
  names must be the caller's / exist, else 404 before any write.
- Handlers use `respondError` (`controller/errors.go`) instead of echoing errors: validation errors
  are `workoutrepo.ValidationError` (400 with the message), the rest map as in the BACKLOG #6
  note. The old "400 for everything" handlers (create program/plan, rename, activate, cardio,
  mobility, switch plan) now return 500 for real failures; the e2e "switch failure" spec accepts the
  fixed text.

**Code shape**: models embed `database.Owner` (`Creator` on `Exercise`); every read uses
`database.Owned(ctx)` including preloads, joins in `GetExerciseProgression`, previous/max logs and
the activity SQL (explicit `user_id = ?` on each joined table); creates `Omit(clause.Associations)`;
updates go through `database.Affected`. `SaveCardio` and the mobility updates write only their
columns (not `Save`). `CreateMinimalIfMissing` uses `database.XactLock` (BACKLOG #2).

**Frontend**: the exercise edit dialog (routine page) has a "Your machine settings" list (label /
value rows, add and remove); the shared fields are disabled with a note unless `can_edit`. Saving
sends the exercise update only when the user can edit it, then the settings. The logging screens
already read `machine_settings`, which is now the caller's.

**Tests**: `controller/workout_pg_test.go` (twousers: every endpoint family, same-date start, no
waiting on Alice's lock, forged `user_id`/`created_by`, catalog permissions with the admin,
per-user machine settings, composite-FK refusals through the owner role, RLS backstop as Bob per
table); `services/migration_0017_pg_test.go` (backfill on legacy-shaped data, the duplicate pair,
the stop on a duplicate with data, the fresh database); `pg_sets_test.go` adapted (an admin user,
`created_by`/`user_id` on raw inserts). SQLite module tests act as user 1 through `testutil.Ctx`
(`SetupTestDB` returns a db whose statements carry the identity). `twousers.Harness.Engine` is new:
the same stack as `Router` for modules that register on the engine (`workout.RegisterRoutes`). e2e:
a spec for editing machine settings. Seeds: devseed and the e2e seed give exercises to
`created_by` = the dev login and everything else `user_id` = dev; devseed also seeds the dev
login's machine settings on Lat Pulldown.

**Not done / for later**
- No UI for the admin's exercise delete (API only).
- First writer wins on an exercise name, like C5's external IDs: only the creator or the admin can
  fix a wrong name or cue on a shared exercise.
- `workout_plan_exercises.exercise_id` has no FK (never had one).
- Migration 0017's down is lossy with several users (lowest user's machine settings win, the
  global `client_id` index fails if two users share one); the empty duplicate log stays deleted.
