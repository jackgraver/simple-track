# Multi-user and security hardening

Turn simple-track from a single-owner app into one that a few family members can use, with
each person's data locked to them, and harden security across the app while doing it.

- **Who**: family accounts only. No open sign-up. The owner is the admin and creates accounts.
- **Now**: every personal row belongs to one user. Nobody can read or change another user's data.
- **Later (not in scope, but don't block it)**: groups/households that share chosen data (diet
  plans, finances, grocery). See [Future: sharing](#future-sharing).

Files: this README (design + rules), [TASKS.md](TASKS.md) (lanes and order),
[DECISIONS.md](DECISIONS.md) (what was decided and why), [BACKLOG.md](BACKLOG.md) (security
findings not yet scheduled; created by lane A3).

## Branches and PRs

All project work lands on the base branch **`proj/multi-user`** (worktree
`.worktrees/multi-user`). It has one draft PR into `main` that is merged at the end (lane D1).
Nothing from this project reaches `main` or prod before that.

```
main <── proj/multi-user (draft PR, merged last)
              ├── mu/a1-cleanup     (PR -> proj/multi-user)
              ├── mu/b1-identity    (PR -> proj/multi-user)
              │      └── mu/b1-identity-2   (a lane split into stacked PRs targets the previous one)
              └── ...
```

- Lane branch `mu/<lane>-<name>`, worktree `.worktrees/mu-<lane>-<name>`, branched from
  `origin/proj/multi-user`, PR targets `proj/multi-user`. If the worktree already exists, use it.
- Before opening or updating a PR: `git fetch origin && git rebase origin/proj/multi-user`.
- **Base upkeep**: when `main` moves, merge `origin/main` into `proj/multi-user` (one session
  does it, not every lane). If `main` gained migrations, renumber the project's migrations to
  come after them (they've only been applied in dev DBs, so renaming is safe).

## Session protocol

1. Read this README, your lane in `TASKS.md`, and the tail of `DECISIONS.md`.
2. Set up or reuse the lane worktree (above). Run `mise trust` if mise refuses the config.
   Use `./bin/dev env` for your slot's ports; never hardcode them.
3. Respect the lane's **Depends on**: it must be merged into `proj/multi-user`.
4. In the same PR as the code: tick `TASKS.md`, append to `DECISIONS.md` for any choice
   a later lane should know about, update "Gotchas" below if you hit one.
5. Verify: `./bin/dev test` (sets `TEST_DATABASE_URL` so Postgres tests run), `./bin/dev check`,
   and `./bin/dev e2e` when behaviour changed. Start a stack (`./bin/dev start`) to exercise
   the API or UI; stop only stacks you started.
6. Commit body: `Project: multi-user`. Push the lane branch and open the PR into
   `proj/multi-user`; don't merge it unless the lane says so (see "Merging").
7. Never deploy. Prod gets this project once, at D1, with the owner.

## Merging

- **A1 and A3**: the agent merges its own PR into `proj/multi-user` once checks pass
  (`gh pr merge --squash`).
- **Every other lane**: the owner reviews and merges. Open the PR and stop.

## Where things stand (2026-11-27)

- No table has a `user_id`. `user_profiles` is a single row for the whole app.
- The JWT carries only `username`. `AuthMiddleware` sets `username` on the gin context; no user ID
  reaches services or queries.
- Repositories use a package-global connection (`conn()` -> `database.GetDB()`), and most
  service/repo functions take no `context.Context` (e.g. `CreateFood(food, relatedFoodID)`).
  The logged-in user is lost between the handler and the query.
- About 130 hand-written gorm query sites, no common choke point. `internal/generics` is unused;
  `internal/database/repository` is partly used.
- Global unique indexes that assume one user: `days.date`, `body_weight_logs.date`,
  `step_logs.date`, `day_logs(day_id, meal_id)`. `client_id` uniques are global too.
- `investment_account_types` mixes shared rules (limits, eligibility) with personal values
  (`opening_room`, `opening_room_year`). Exercises carry personal machine settings (migration 0003).
- The DB pool uses `database/sql` defaults: unlimited open connections, no lifetime, no
  statement or idle-in-transaction timeout. The app connects as the table owner.

## Design

### Identity

- JWT claims gain the user ID (`uid`). Tokens without it are rejected (users sign in again once).
- `AuthMiddleware` puts the user (ID, admin flag) on the request context. Code reads it with
  `auth.UserID(ctx)` / `auth.MustUser(ctx)`. A missing user is a 401, never a default user.
- The dev-token bypass resolves to the `dev` user's real ID.

### Per-request database handle

Every request that touches user data runs its queries through one handle that already knows the
user:

```
middleware: tx := db.Begin(); tx.Exec("SELECT set_config('app.user_id', ?, true)", uid)
            ctx = database.WithTx(ctx, tx)          // commit on success, rollback on error/panic
repos:      database.From(ctx).Where(...)...        // never the global
```

- `set_config(..., true)` is transaction-local, so the user can't leak to the next request that
  reuses the pooled connection. This is the reason for the per-request transaction.
- Begin lazily on first `From(ctx)` so requests that never hit the DB don't hold a connection.
- Never make external HTTP calls (media/food lookup) while holding the request transaction.
  Those endpoints fetch first, then write.
- Nested `db.Transaction(...)` inside a request becomes a savepoint; existing code keeps working.

The final shape is in `DECISIONS.md` ("B2 per-request database handle").

### Ownership

Three kinds of table:

| Kind | Tables | Rule |
|---|---|---|
| **Owned** | days, plans, meals, planned_meals, saved_meals, body_weight_logs, step_logs, water_logs, drink_size_presets, grocery_items, investment_accounts, workout_logs, workout_programs, workout_plans, user_profiles (1:1), plus new per-user tables (user_media, exercise settings, personal account-type room) | `user_id bigint NOT NULL REFERENCES users`, visible/writable only by that user |
| **Owned children** | day_logs, meal_items, saved_meal_items, logged_exercises, logged_sets, cardios, account_balances, investment_deposits, workout_plan_days, workout_plan_exercises | Also carry `user_id`, enforced equal to the parent's with a composite FK `(parent_id, user_id) -> parent(id, user_id)` (parent gets a `UNIQUE (id, user_id)`). Each query filters its own table without joining to the parent. |
| **Shared catalog** | foods, composite_foods, composite_food_items, exercises, investment_account_types (rules part), investment_contribution_rules, media_items (catalog part) | Everyone reads. Anyone signed in can add (`created_by` = them). Creator or admin can edit. **Only admin deletes.** |

Lanes that own a table confirm its kind before migrating it and record changes in DECISIONS.

- Per-date and per-name uniques on owned tables become `(user_id, ...)`. `client_id` uniques
  become `(user_id, client_id)` and idempotency lookups filter by user.
- Lookups by ID return **404**, not 403, for rows the user can't see.
- App code still filters explicitly (`database.Owned(ctx)` scope, inserts set `UserID` from ctx).
  RLS is the backstop, not a replacement: module tests run on SQLite where RLS doesn't exist,
  and explicit filters keep queries readable.

### Row-level security (RLS)

Postgres enforces the ownership rule on every query, so a query that forgets its filter still
sees only the caller's rows.

```sql
ALTER TABLE water_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY owner ON water_logs
  USING      (user_id = current_setting('app.user_id', true)::bigint)
  WITH CHECK (user_id = current_setting('app.user_id', true)::bigint);
```

- No `app.user_id` set -> the setting is NULL -> no rows match. Mistakes hide data, they don't leak it.
- Catalog tables: SELECT for all, INSERT with `created_by` = caller, UPDATE for creator or admin,
  DELETE for admin (`app.is_admin`, set by the middleware alongside `app.user_id`).
- **RLS doesn't apply to the table owner or superusers.** The server connects as a separate
  login role (`simpletrack_app`: no ownership, no `BYPASSRLS`, only DML grants). Migrations,
  devseed, e2eseed and devuser keep the owner role; seeds set `user_id` explicitly.
- A test fails if any table in `public` (except `schema_migrations` and `users`) lacks RLS.
- Cross-user tests must run on Postgres (`TEST_DATABASE_URL`, as `dev test` sets it), since
  SQLite has no RLS.

### Connection pool

Set explicitly (values are lane A2's call): max open, max idle, conn max lifetime and idle
time. On the app role: `statement_timeout` and `idle_in_transaction_session_timeout` (the
per-request transaction makes the latter matter). Keep the production `sslmode` check.

### Accounts

- `users.is_admin`. The owner's account is the admin (set by migration from the existing user).
- `users.token_version` (in the JWT, checked by the middleware) so a password reset or a disabled
  account (`users.disabled_at`) signs that user out everywhere. Today a JWT can't be revoked.
- `auth_events` audit log: logins (ok/failed), setup-link created/used, user created/disabled,
  password changed. Admin can view it.
- Admin creates a user with a username only (no password). The server returns a **one-time setup
  link** (`/setup/<token>`, valid 48 h). The admin texts it; no email anywhere.
- Opening the link lets the user set a password and signs them in. The same mechanism does
  password resets (admin generates a new link).
- `user_setup_tokens(user_id, token_hash, purpose, expires_at, used_at)`. Store only a hash;
  single use; creating a new link revokes older unused ones.
- A user with no password can't log in. Login protection (`login_protection.go`) also covers
  the setup endpoint.
- Passkeys (WebAuthn) are a later add-on; the setup link becomes how a first passkey is registered.

### Migrating existing data

Per owned table, in one migration with the code that writes the column:

1. add `user_id` nullable, 2. backfill to the admin user, 3. `SET NOT NULL`, FK, index,
4. swap unique indexes to per-user, 5. enable RLS + policy.

The server refuses to start with pending migrations, so code and schema always ship together.
The whole project reaches prod in one deploy at D1, so D1 rehearses every migration on a copy
of a prod backup first. Pre-deploy backups on the VPS cover a bad backfill.
**Don't create a second real account before D1.**

## Rules for every lane

- **Prod changes go in ROLLOUT.md.** Anything prod needs beyond the new images gets a step in
  `docs/multi-user/ROLLOUT.md` in the same PR: an env var in `.env.vps`, compose, Caddy, j-cicd
  or VPS agent config, one-off SQL, a pre-check, or a user-visible effect (e.g. everyone signed
  out). D1 deploys from that file alone, so if it isn't there it won't happen.
- **Security findings**: check `BACKLOG.md` for entries tagged with your lane, fix them, and tick
  them there in the same PR. If you defer one, say why in DECISIONS.
- Never use `database.GetDB()` / `conn()` in request code after B2. Use `database.From(ctx)`.
- Every new owned table: `user_id NOT NULL`, per-user uniques, RLS policy, two-user test.
- Two-user test per endpoint: Alice's data, Bob's token. Bob gets 404 or an empty list, never Alice's rows.
  Bob can't create a child under Alice's parent.
- Update `cmd/devseed`, `internal/e2e/seed` and `cmd/devuser` for any table you change.
- `docs/ARCHITECTURE.md` gets updated by any lane touching DB roles, compose or j-cicd.
- Security-sensitive code (auth, tokens, RLS policies, roles) gets a `/code-review` pass before
  the PR is marked ready.

### Migration numbers across parallel lanes

Versions must run 1..N with no gaps, so parallel lanes will pick the same number. Use the next
free number while working. Before merging: rebase on `origin/main`, renumber if taken, then
`./bin/dev db:drop-worktree` and restart so your worktree DB re-applies cleanly. Renaming is
fine because the migration was only ever applied in your own worktree DB. Merge one lane at a time.

## Future: sharing

Not scheduled. Keep it cheap by keeping visibility in one place (`database.Owned(ctx)` and the RLS
policy). Sharing later is: `groups`, `group_members(role)`, `group_invites(code, expires)`,
a nullable `group_id` on tables that should be shareable, and the rule becomes
`user_id = me OR group_id IN (my groups)`. Don't add `group_id` columns now.

## Gotchas

(Add as you hit them.)

- **The server is `simpletrack_app`, not the owner** (A2). It can't `TRUNCATE`, restart
  sequences, run DDL or write `schema_migrations`. Anything needing that belongs in a migration or
  a tool on `DATABASE_OWNER_URL` (`database.ConnectOwner`). The e2e build's `/__e2e/reset` already
  has its own owner pool. psql/pgweb as the owner bypass RLS, so to see what the app sees use
  `SET ROLE simpletrack_app` plus `set_config('app.user_id', ...)`.
- **Roles are cluster-wide, migrations per database** (A2). Tests and e2e migrate many databases on
  one server at once, so a migration that touches roles must be idempotent and survive a
  concurrent run in another database (see `0010_app_role.sql`). In tests, act as the app with
  `SET ROLE simpletrack_app` on a single `*sql.Conn`. Role settings like `statement_timeout`
  only apply at login, not after `SET ROLE`.
- **The app role has `statement_timeout = 15s` and `idle_in_transaction_session_timeout = 30s`**
  (A2). B2's per-request transaction must not stay open across slow work (external lookups,
  e2e holds): after 30 s idle in a transaction Postgres kills the session.
- **One failed statement aborts the request's transaction** (B2). On Postgres, every query after
  a failed one fails too. Code that recovers from a SQL error must run the failing statement in
  a savepoint (`database.CreateInSavepoint`, or its own `db.Transaction`). SQLite module tests and
  e2e fault injection don't show this; Postgres tests do.
- **Locks live until the request ends** (B2). A repository's `Transaction` inside a request is a
  savepoint, so its row and advisory locks are released at the request's commit, not when the
  block returns.
- **`dev test` points `TEST_DATABASE_URL` at the owner** (`DATABASE_OWNER_URL`), since the app role
  can't create databases.
- **FK checks bypass RLS** (B3). A plain `REFERENCES parent (id)` accepts another user's parent
  ID. Children and anything pointing at an owned row use a composite FK on `(id, user_id)`.
- **gorm soft delete is an UPDATE** (B3), so catalog "only admin deletes" needs the
  `catalog_guard` trigger, not just a DELETE policy.
- **Scopes don't reach preloads** (B3): `Preload("X", database.Owned(ctx))`. RLS hides the rest.
- **Postgres tests skip without `TEST_DATABASE_URL`** but fail with `TEST_DATABASE_REQUIRED=1`
  (CI). New Postgres tests use `internal/database/pgtest` or `twousers`, never their own skip.
- **RLS + `RETURNING`** (C6). An INSERT whose new row fails the table's SELECT policy errors if
  it has `RETURNING`, and gorm's `Create` always adds `RETURNING id`. For write-only tables
  (`auth_events`: anyone inserts, only the admin reads) insert with a raw `INSERT` instead.
- **Acting as the app role in tests** (C6). A gorm pool whose DSN has
  `options=-c role=simpletrack_app` runs every connection as the app role (the test superuser may
  assume it even though it's NOLOGIN), so whole HTTP flows run under RLS. See
  `auth/accounts_pg_test.go`.
- **Writes that must survive a 4xx/5xx** go through `database.Always` (C6). Success audit rows
  don't: they belong in the request transaction so they roll back with the change.
- **`Save` with `Select("*")` writes `user_id` too** (C1). Updating a row you built by hand (the
  Renpho import's replace-all-columns) would set a zero `user_id`; add `Omit("user_id")`.
  `CreateInSavepoint` omits associations now.
- **Parent plus children, no nested writes** (C2). `Create(&meal)` with `Items` filled writes the
  items through the association, which the rules test forbids (a nested food would be written
  too). Create the parent with `Omit(clause.Associations)`, then the items as their own batch with
  the parent's ID (`createMealTx`). Check any ID in the body that points at another table (a
  food, a recipe) is one the caller may use: FK checks bypass RLS.
