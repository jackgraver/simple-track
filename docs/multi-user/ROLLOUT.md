# Production rollout (lane D1)

Exact steps to take the multi-user project to production. Written by lane A2 (app DB role,
pool, HTTP server); later lanes add their parts in the marked places. D1 checks the whole thing
against the rehearsal on a copy of a prod backup before anything here runs on the VPS.

Background: `docs/ARCHITECTURE.md` ("Database roles", "Build and deploy") and `j-cicd guide`.
Commands below run on the VPS as `jack` in `~/simple-track` unless they say otherwise.
Throughout, `psql_owner` means:

```sh
docker compose -f docker-compose.vps.yml --env-file .env.vps exec -T postgres \
  psql -U simpletracker -d simpletracker
```

(Use the real `POSTGRES_USER` / `POSTGRES_DB` from `.env.vps` if they aren't the defaults, and
the compose file path from the agent config if it isn't `~/simple-track/docker-compose.vps.yml`.)

## What changes in production

| | Before | After |
|---|---|---|
| Server's DB login | `simpletracker` (owner, superuser) | `simpletrack_app` (DML only, timeouts) |
| `DATABASE_URL_PRODUCTION` | owner DSN | app DSN, password `APP_DB_PASSWORD` |
| `DATABASE_OWNER_URL` | (none) | owner DSN, used by `/app/migrate` in pre_swap |
| `.env.vps` | | new `APP_DB_PASSWORD` (required: compose refuses to start without it) |
| Backend stop | SIGKILL after 10 s, requests cut off | graceful drain, `stop_grace_period: 30s` |
| Pool | unlimited open, 2 idle | 20 open, 10 idle, 30 min lifetime (env-overridable) |

The new server **refuses to start in production as a superuser, BYPASSRLS role or table owner**.
So the new image with the old compose file fails its health check and j-cicd rolls it back:
the compose file and `.env.vps` must be updated before the deploy, not after.

## 1. Pre-checks (read-only, any time before D1)

1. Agent config on the VPS (`~/.config/j-cicd-agent.yaml`, project `simple-track`): note
   `compose_file`, `env_file` and confirm `pre_swap` is
   `{service: backend, command: ["/app/migrate", "up"]}`. Migrations and the role's password
   depend on pre_swap running in the backend service.
2. Every table is owned by the role migrations run as (`ALTER DEFAULT PRIVILEGES` and the grants
   only cover that role's objects):
   ```sql
   SELECT tableowner, count(*) FROM pg_tables WHERE schemaname = 'public' GROUP BY 1;
   ```
   Expect one row, `simpletracker`. Anything else: `ALTER TABLE ... OWNER TO simpletracker` first.
3. The owner can create roles: `SELECT rolsuper, rolcreaterole FROM pg_roles WHERE rolname = current_user;`
   (the postgres image makes `POSTGRES_USER` a superuser).
4. No leftover role: `SELECT rolname, rolcanlogin FROM pg_roles WHERE rolname = 'simpletrack_app';`
   should return nothing. If it exists, the migration strips any superuser/BYPASSRLS/CREATEROLE
   and role memberships from it, and `migrate up` resets a wrong password, so it's safe, but
   find out where it came from. One that has a password but is NOLOGIN makes `migrate up` stop
   (it won't re-enable a role someone disabled): drop it or `ALTER ROLE ... LOGIN` first.
5. Headroom: `SHOW max_connections;` (default 100). The server uses at most 20
   (`DB_MAX_OPEN_CONNS`), migrate a few, so anything ≥ 40 is fine.
6. Exactly one live account (migration 0009 makes it admin and **fails** with more than one):
   `SELECT count(*) FROM users WHERE deleted_at IS NULL;` must be 1.
7. (B3) Every owned-table migration backfills existing rows to the one admin and **fails** if
   there isn't exactly one live admin (0011 `drink_size_presets` first; C lanes copy it). After
   0009 that's the account from check 6, so nothing to do beyond it; the rehearsal proves it.

8. (C1) `user_profiles` should hold at most one live row (the app only ever kept one):
   `SELECT count(*) FROM user_profiles WHERE deleted_at IS NULL;` expect 0 or 1. With more, the
   migration keeps the oldest live row and soft-deletes the rest (nothing is dropped), so note
   which one you expect to survive.

## 2. Rehearsal (D1, locally, before touching the VPS)

On a copy of the latest prod backup, with the D1 build of the backend image:

1. Restore the dump into a local `postgres:16-alpine` (owner role named like prod's).
2. Run `/app/migrate up` from the new image with `APP_ENV=production`,
   `DATABASE_ALLOW_INSECURE=true`, `DATABASE_OWNER_URL=<owner DSN>` and
   `DATABASE_URL_PRODUCTION=postgresql://simpletrack_app:<pw>@.../<db>?sslmode=disable`.
   Expect every pending migration applied, then
   `migrate: set login password for simpletrack_app from the app DSN`.
3. Start `/app/server` from the same image with the same env (plus the usual `JWT_SECRET`,
   `CORS_ORIGINS` etc.). Expect no `database role:` error; `/health` 200.
4. `SELECT usename, count(*) FROM pg_stat_activity GROUP BY 1;` shows the server as
   `simpletrack_app`. As that role: `SHOW statement_timeout;` gives `15s`.
5. Click through every module as the owner (other lanes' checks: backfill row counts etc.).
6. Send the server SIGTERM during a request and confirm `shutting down: draining requests` then
   `shutdown complete` in its log.

## 3. Deploy day

Do 3.1–3.3 back to back. Nothing else may deploy in between: any backend recreate with the new
compose file before the D1 migrations ran would try to log in as a role that doesn't exist yet.

### 3.1 Secret

```sh
cd ~/simple-track
cp .env.vps .env.vps.pre-d1
printf 'APP_DB_PASSWORD=%s\n' "$(openssl rand -hex 24)" >> .env.vps
chmod 600 .env.vps
```

Hex only: the password goes into a URL without escaping. It lives only in `.env.vps` (and the
SCRAM verifier Postgres derives from it). Keep the owner's `POSTGRES_PASSWORD` as it is.

### 3.2 Compose file

The agent uses the VPS's own copy of the compose file (`.cicd.yaml` has no
`deploy.compose_file`), so copy the D1 version of `docker-compose.vps.yml` from `main` to the
path in the agent config, keeping the old one:

```sh
cp docker-compose.vps.yml docker-compose.vps.yml.pre-d1
# then copy the file from the merged main commit, e.g. from the build host:
#   scp docker-compose.vps.yml st-vps:~/simple-track/docker-compose.vps.yml
docker compose -f docker-compose.vps.yml --env-file .env.vps config >/dev/null && echo ok
```

`config` must succeed (it fails on a missing `APP_DB_PASSWORD`). Don't `up` anything by hand.

If D1 also takes BACKLOG #16 (pin `app-network`'s subnet so `TRUSTED_PROXIES` can't drift), the
`ipam` change lands in this same file and compose will want to recreate the network. Plan that
recreate inside the deploy window and re-check `TRUSTED_PROXIES` afterwards.

### 3.3 Deploy

From the build host, after `main` has the merged project:

```sh
j-cicd run simple-track --plan      # confirm the commit
j-cicd deploy simple-track --brief
```

What happens on the VPS: pull → `pre_deploy` backup → `pre_swap` runs the new image's
`/app/migrate up` as the owner (applies every project migration, creates `simpletrack_app`,
grants, timeouts, then sets its LOGIN password from `DATABASE_URL_PRODUCTION`) → backend recreated,
connects as `simpletrack_app` → health check → verify `/version.json`.

### 3.3a Lane additions

Each lane adds what its changes need here (README: "Prod changes go in ROLLOUT.md").

- **B1 (identity)**: every existing session is rejected (pre-B1 tokens carry no `uid`), so the
  owner signs in again once after the deploy. `AUTH_COOKIE_MAX_AGE_SEC` comes with the compose
  file (3.2), default 365 days; no `.env.vps` change.
- **H1 (HTTP hardening)**: the backend part (body limits, `Cache-Control: no-store`, no default
  CORS origins) ships in the backend image with 3.3; nothing to configure (`APP_ORIGIN` already
  sets `CORS_ORIGINS`). The Caddy part is a separate deploy **after 3.4**: see 3.5.
- **C6 (accounts)**: migration 0012 runs in pre_swap with the rest (usernames become unique
  case-insensitively, an empty email becomes NULL, two new tables `user_setup_tokens` and
  `auth_events`). It relies on pre-check 1.6 (one live account). The `/setup/` nginx location
  ships in the frontend image. Nothing to configure. The first family accounts come after the
  deploy: see 3.6.

- **B2 (per-request DB handle)**: nothing to do. In 3.4, also check the backend log has no
  `database: commit` or `database: rollback` errors (a failed commit answers 500).
- **S2 (frontend storage)**: nothing to configure. On first load after the deploy each browser
  drops its old un-namespaced drafts, rest timer and resume route (ownership unknown), so a
  half-written draft or an in-progress rest timer is lost once.

- **B3 (owner kit)**: nothing to configure. Effects to know:
  - The pipeline has three new blocking steps before images are built: `backend-test` now
    starts a throwaway Postgres container (`dev/ci/backend-test.sh`, needs Docker on the build
    host, like `e2e`), `backend-vulncheck` (`govulncheck`) and `frontend-audit` (`npm audit
    --omit=dev --audit-level=high`). The last two fetch from the network (Go proxy and vuln DB,
    npm registry). **D1 checks before the deploy**: run `j-cicd check simple-track --ref
    proj/multi-user --brief` (steps only, no deploy) and confirm all three pass on the build
    host; a new advisory published since the lane merged also fails the run, so fix or decide
    before deploy day, not during it.
  - In 3.4, the water page's drink sizes are the owner's (0011 backfilled them).

- **C1 (tracking)**: migration 0013 runs in pre_swap with the rest and makes `body_weight_logs`,
  `step_logs`, `water_logs`, `grocery_items` and `user_profiles` owned (every row goes to the one
  admin, RLS on, `water_logs.preset_id` becomes a composite FK to the owner's presets). Nothing to
  configure; relies on pre-check 1.6. It **stops** if rows exist and there isn't exactly one live
  admin. **Row counts for the D1 rehearsal**: run before and after `migrate up`, the numbers must
  match, and after it `owned` must equal `total` for every table:
  ```sql
  -- psql_owner (bypasses RLS, so it counts soft-deleted rows too)
  SELECT 'body_weight_logs' AS tbl, count(*) AS total FROM body_weight_logs
  UNION ALL SELECT 'step_logs',     count(*) FROM step_logs
  UNION ALL SELECT 'water_logs',    count(*) FROM water_logs
  UNION ALL SELECT 'grocery_items', count(*) FROM grocery_items
  UNION ALL SELECT 'user_profiles', count(*) FROM user_profiles;
  -- after the migration, rows given to the admin (expect owned = total on every line):
  SELECT 'body_weight_logs' AS tbl, count(*) AS owned FROM body_weight_logs WHERE user_id = (SELECT id FROM users WHERE is_admin AND deleted_at IS NULL)
  UNION ALL SELECT 'step_logs',     count(*) FROM step_logs     WHERE user_id = (SELECT id FROM users WHERE is_admin AND deleted_at IS NULL)
  UNION ALL SELECT 'water_logs',    count(*) FROM water_logs    WHERE user_id = (SELECT id FROM users WHERE is_admin AND deleted_at IS NULL)
  UNION ALL SELECT 'grocery_items', count(*) FROM grocery_items WHERE user_id = (SELECT id FROM users WHERE is_admin AND deleted_at IS NULL)
  UNION ALL SELECT 'user_profiles', count(*) FROM user_profiles WHERE user_id = (SELECT id FROM users WHERE is_admin AND deleted_at IS NULL);
  ```
  In 3.4: the weight, steps, water, grocery and profile pages show the owner's data as before
  (a new account starts with an empty profile and no history). Error responses from these
  endpoints are now fixed messages; database errors are only in the backend log.
- **C4 (money)**: migration 0014 runs in pre_swap with the rest. It moves each account type's
  start year and opening room to a new per-user table (`investment_contribution_room`, the
  admin's row), drops those three columns from `investment_account_types`, and gives every
  account, balance, deposit, type and limit row to the admin. Nothing to configure. **D1
  rehearsal check** (run on the restored prod copy as `psql_owner`, before and after `migrate up`;
  the numbers must match exactly):
  ```sql
  -- row counts (all five tables, deleted rows included)
  SELECT 'accounts' t, count(*) FROM investment_accounts UNION ALL
  SELECT 'balances', count(*) FROM account_balances UNION ALL
  SELECT 'deposits', count(*) FROM investment_deposits UNION ALL
  SELECT 'types', count(*) FROM investment_account_types UNION ALL
  SELECT 'limits', count(*) FROM investment_contribution_rules;
  -- net worth today: latest balance per live, unarchived account; liabilities subtract
  SELECT round(sum(CASE WHEN a.kind IN ('credit_card','loan') THEN -b.balance ELSE b.balance END), 2) AS net_worth
  FROM investment_accounts a
  JOIN LATERAL (SELECT balance FROM account_balances
                WHERE account_id = a.id AND deleted_at IS NULL ORDER BY date DESC LIMIT 1) b ON true
  WHERE a.deleted_at IS NULL AND a.archived_at IS NULL;
  ```
  Before `migrate up` also note the room settings the types carry
  (`SELECT name, contribution_start_year, opening_room_year, opening_room FROM investment_account_types ORDER BY id;`);
  after it, the same rows must be in
  `SELECT t.name, r.contribution_start_year, r.opening_room_year, r.opening_room FROM investment_contribution_room r JOIN investment_account_types t ON t.id = r.investment_account_type_id ORDER BY t.id;`
  (types with all three NULL have no row, and `types` and `limits` counts are unchanged). Then in
  the app as the owner: `/finances` net worth and `/finances/rules` room figures equal what they
  were before the rehearsal. In 3.4, the same two pages are the check on prod. The migration
  fails (changing nothing) unless there is exactly one live admin.
  Effects to know: other family accounts see the owner's contribution types (TFSA, ...) and their
  limits, but never the owner's accounts, balances or room figures. They can't edit the owner's
  types; only the admin deletes types or limits.

<!-- Later lanes: add anything else that must happen in this deploy here. -->

### 3.4 Verify

```sql
-- psql_owner
SELECT usename, count(*) FROM pg_stat_activity WHERE datname = current_database() GROUP BY 1;
  -- backend connections are simpletrack_app; only your psql is simpletracker
SELECT rolname, rolcanlogin, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = 'simpletrack_app';
  -- t, f, f
SELECT setconfig FROM pg_db_role_setting WHERE setrole = 'simpletrack_app'::regrole;
  -- {statement_timeout=15s,idle_in_transaction_session_timeout=30s}
SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 3;
```

```sh
docker compose -f docker-compose.vps.yml --env-file .env.vps logs backend --since 10m | grep -i "database role\|pool:\|listening"
# expect "postgres: connected (pool: max open 20, ...)" and "server listening on [::]:8080", no "database role"
```

Then sign in on the site and use each module once.

### 3.5 Caddy (lane H1)

`deploy/Caddyfile` now sets the security headers and `request_body max_size 6MB` on `/api`. Only
the separate `simple-track-caddy` j-cicd project deploys it (`.cicd-caddy.yaml`, builds `main`).
Do this only once 3.4 has passed, so a problem in either deploy is easy to attribute. Neither
side depends on the other: the CSP's script hash also matches the pre-D1 `index.html`, so either
can be rolled back alone.

1. `j-cicd run simple-track-caddy --plan`: the revision must be the `main` commit 3.3 deployed.
   Then `j-cicd deploy simple-track-caddy --brief`. The agent validates, installs, restarts
   `caddy.service` (connections drop for a second or two) and checks the site; on a failed
   check it restores the previous Caddyfile by itself.
2. Headers: `curl -sI https://st.jackgraver.ca/ | grep -iE 'content-security|x-frame|nosniff|referrer|permissions|^server'`
   shows the five headers and no `server:` line. `curl -sI https://st.jackgraver.ca/setup/x | grep -i referrer`
   shows exactly one line, `no-referrer` (C6). `curl -sI https://st.jackgraver.ca/api/health`
   (404 for HEAD is fine) carries them too, plus `cache-control: no-store`.
3. Body limit: `head -c 7000000 /dev/zero | curl -s -w ' %{http_code}\n' -X POST --data-binary @- https://st.jackgraver.ca/api/health`
   prints just ` 413` with an empty body: that's Caddy. A JSON body
   `{"error":"request body too large"}` means the request reached the backend, so Caddy's limit
   isn't in effect.
4. Open the site on the phone (installed PWA) and a desktop browser with devtools: sign in,
   visit each module, and check the console has no `Content-Security-Policy` errors.

If something is blocked: `j-cicd rollback simple-track-caddy` puts the old Caddyfile back (no
headers at all), then fix the CSP and redeploy.

### 3.6 First family accounts (lane C6, owner, after 3.4 and 3.5)

1. Sign in as the owner. Account menu → **Account**: if the current password is shorter than 10
   characters, change it (new passwords must be 10+; changing it signs out other devices).
2. Account menu → **Users** → add each person by username (lowercase). Each gets a one-time link
   (`https://st.jackgraver.ca/setup/...`, 48 h, single use). Text it with **Share** or **Copy
   link**; never email it. A new link for the same person cancels the previous one.
3. Check on the Users page that each person shows "Last sign-in" once they've used their link,
   and that **Sign-in activity** lists `Account created`, `Link created`, `Password set with link`.
4. Optional: `docker compose -f docker-compose.vps.yml --env-file .env.vps logs frontend backend | grep -c /setup/`
   should be 0 (the token is never logged).

Lost password or phone later: Users → the person's menu → **Password reset link** (signs them out
everywhere once used). Leaving the family app: **Disable and sign out** (reversible with Enable).

## 4. If it goes wrong

- **pre_swap fails** (migration error): nothing was changed except what the failed migration's
  own transaction rolled back. Each migration runs in a transaction. Fix forward; to back out the
  prep, restore `.env.vps.pre-d1` and `docker-compose.vps.yml.pre-d1`.
- **Health check fails** (e.g. `database role:` or login errors in the backend log): j-cicd
  rolls the images back automatically, but the old image now meets the new compose file and
  would log in as `simpletrack_app` against a schema it doesn't know. Restore the two `.pre-d1`
  files, then `j-cicd rollback simple-track` (or redeploy the previous release) so the old
  server runs as the owner again. The migrations stay applied. If the old code can't run on the
  new schema, restore the `pre_deploy` backup from `~/backups` (rollback never touches the DB).
- **Rolling back after D1 to a pre-D1 image** always needs the `.pre-d1` compose file first.
  The old `/app/migrate` reads `DATABASE_URL_PRODUCTION` as its only DSN, and as
  `simpletrack_app` it can't create or lock `schema_migrations`, so its pre_swap fails.

## 5. Later operations

- **Rotate the app password:** edit `APP_DB_PASSWORD` in `.env.vps`, then
  `docker compose -f docker-compose.vps.yml --env-file .env.vps run --rm --no-deps backend /app/migrate up`
  (prints `set login password`) and immediately
  `docker compose -f docker-compose.vps.yml --env-file .env.vps up -d --no-deps backend`.
  Open connections survive the change, but new ones fail until the restart, so do both together.
  A normal backend deploy does the same: pre_swap, then recreate.
- **Tune the pool:** set `DB_MAX_OPEN_CONNS` etc. in `.env.vps` and recreate the backend.
- **Cut the server off from the DB in an emergency:** `ALTER ROLE simpletrack_app NOLOGIN;` plus
  `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE usename = 'simpletrack_app';`.
  Later deploys' `migrate up` then fail on purpose until you run `ALTER ROLE simpletrack_app LOGIN;`.
- **Ad-hoc SQL:** `psql_owner` sees everything. Once RLS is on, connecting as `simpletrack_app`
  shows only rows matching `app.user_id`, and none without it.
- **Possible follow-up:** `deploy.compose_file: docker-compose.vps.yml` in `.cicd.yaml` would ship
  the compose file with each release and restore it on rollback, removing steps 3.2 and the
  rollback caveats above. It changes how every deploy works, so it's the owner's call
  (DECISIONS.md).
