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
  anything else (DB down, missing JWT_SECRET) is **503**, so an outage doesn't sign people out.
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
