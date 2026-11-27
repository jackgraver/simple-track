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
