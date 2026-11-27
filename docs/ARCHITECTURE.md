# Architecture

simple-track is a Go API plus a Vue 3 SPA, deployed to a single VPS with Docker Compose.
For dev tooling see `CLAUDE.md` and `README.md`; for deploys see `j-cicd guide`.

## Production request flow

```
                         browser (https://st.jackgraver.ca)
                                      │
                                      ▼
                       Caddy  (host process on the VPS, TLS + compression)
                          │                               │
              /api/*  (prefix stripped)            everything else
                          ▼                               ▼
              backend  127.0.0.1:8080          frontend  127.0.0.1:3000
              Go, gin + gorm                   nginx serving the built SPA
                          │                    (static files only, no proxying)
                          ▼
                 postgres  (compose network only, not published)
```

Caddy is the only thing that routes. nginx does **not** sit in front of the backend and the
backend does not talk to nginx: Caddy sends `/api/*` to the backend and everything else to nginx.
The frontend container never touches Postgres; only the backend does.

All three services run in `docker-compose.vps.yml` on the `app-network` bridge. The backend and
frontend publish their ports on `127.0.0.1` only (for Caddy on the host); Postgres publishes
nothing and is reachable only as `postgres:5432` from the backend container.

## Caddy

Caddy runs on the VPS host as a systemd service (not in compose) and is the public entry point.
Its config is `deploy/Caddyfile`, shipped by its own j-cicd project (`.cicd-caddy.yaml`,
`j-cicd run simple-track-caddy`), separate from the app deploy because j-cicd projects install
files *or* run compose services, not both. A run validates and `caddy fmt`-checks the file, the VPS
agent installs it atomically at `/opt/simple-track/Caddyfile`, restarts `caddy.service` (a restart,
not a reload: connections drop for a second or two), and checks `https://st.jackgraver.ca/version.json`
still answers. On failure it restores the previous file and restarts again.

VPS wiring (one-time setup):

- `/opt/simple-track/` owned by the deploy user (the agent must be able to rename files into it);
  `/etc/caddy/Caddyfile` is a symlink to `/opt/simple-track/Caddyfile`.
- sudoers: exactly `jack ALL=(root) NOPASSWD: /usr/bin/systemctl restart caddy.service`.
- Agent config (`~/.config/j-cicd-agent.yaml` on the VPS) has a `simple-track-caddy` project with
  `files: [{name: Caddyfile, path: /opt/simple-track/Caddyfile, mode: "0644"}]`,
  `systemd: {unit: caddy.service}` and a `check` that curls the site.

Config: `deploy/Caddyfile` (not copied here, so it can't drift). In outline: `encode zstd gzip`,
a site-wide `header` block, `handle_path /api/*` with a `request_body` limit -> backend, and
`handle` -> frontend. Until the multi-user project's D1 deploy, production still runs the
pre-H1 file (no `header` or `request_body`).

- **TLS:** automatic certificates for `st.jackgraver.ca`. The backend relies on HTTPS
  (`AUTH_COOKIE_SECURE=true`, and production CORS origins must be HTTPS).
- **Compression:** Caddy compresses the whole site (`encode zstd gzip`). nginx deliberately
  doesn't gzip (see below).
- **`/api` prefix is stripped** by `handle_path`. The backend's routes have no `/api` prefix
  (`/health`, `/auth/me`, ...), so the browser calls `/api/auth/me` and the backend sees `/auth/me`.
  The Vite dev proxy does the same rewrite, so dev and prod behave alike.
- **`/version.json`** goes to nginx and is served unauthenticated; j-cicd's post-deploy check reads it.
- No CDN sits in front of Caddy.
- **Security headers** (whole site, set only here; nginx adds just HSTS and cache headers):
  - `Content-Security-Policy`: scripts and everything else from `'self'` only. The one inline
    script in `frontend/index.html` (theme before first paint) is allowed by its `sha256` hash, so
    editing it means updating the hash in the Caddyfile and deploying Caddy first;
    `frontend/src/csp.test.ts` fails when they differ. `style-src 'unsafe-inline'` because Vue,
    reka-ui and motion set style attributes. `img-src https:` for media cover images,
    `data:` for Vite-inlined fonts. `frame-ancestors 'none'`.
  - `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Permissions-Policy` denying
    camera, microphone, geolocation, payment and USB, and `-Server` (no version banner).
  - `Referrer-Policy: same-origin`: no referrer to other sites (so C6's `/setup/<token>` URL
    can't leak), the full one within the site. If this is ever tightened to `no-referrer`, note
    that browsers then send `Origin: null` on native form POSTs (not on `fetch`/XHR), which the
    backend's CORS middleware rejects with 403. The SPA submits every form through axios today.
- **Request body limit:** `request_body max_size 6MB` on `/api` is the outer guard. The backend
  enforces the real per-route limits (below) and Caddy's must stay at or above the largest.

### Client IPs and `TRUSTED_PROXIES`

Caddy adds `X-Forwarded-For`, and the backend only trusts it from addresses listed in
`TRUSTED_PROXIES` (`backend/cmd/server/main.go`). Caddy runs on the host and the backend in a
container, so the backend sees every request coming from the compose network's gateway. With
`TRUSTED_PROXIES` empty, per-IP login rate limiting and the spray block treated all clients as one IP.

Fixed 2026-11-26: `TRUSTED_PROXIES=172.18.0.1` (the `simple-track_app-network` gateway) is set in
`~/simple-track/.env.vps` on the VPS, which the j-cicd agent uses for the compose env. Verify with
`docker inspect simple-track-backend-1 | grep TRUSTED`; the backend logs a warning at startup if it
is empty in production.

Caveat: the gateway is assigned by Docker, not pinned. If the network is ever recreated
(`docker compose down`, Docker upgrade) the subnet may change and the setting silently stops
matching. Re-check with
`docker network inspect simple-track_app-network --format '{{range .IPAM.Config}}{{.Gateway}}{{end}}'`.
Pinning the subnet under `networks.app-network.ipam` in `docker-compose.vps.yml` would remove this risk.

## nginx

nginx is **not** just a build tool: it is the runtime web server inside the frontend image.

- `frontend/Dockerfile` is multi-stage: a Node stage runs `npm run build` (with
  `VITE_API_BASE=/api`), then the `dist/` output is copied into `nginxinc/nginx-unprivileged`,
  which listens on **port 3000**.
- Config is `frontend/nginx.conf`. It has no proxying, only static serving and cache headers:
  - SPA fallback: unknown paths return `index.html`.
  - `/index.html`: `Cache-Control: no-cache`, so browsers pick up a new build's hashed assets.
  - `/assets/*`: cached for a year, `immutable`; a missing chunk is a real 404, not the HTML shell.
  - `/version.json`: `no-store`. Written by the j-cicd `stamp-version` step with the commit hash.
  - `Strict-Transport-Security` on every response. nginx only inherits `add_header` into a
    location that sets none of its own, so the locations repeat it.
- No gzip in nginx; Caddy compresses.
- `/setup/<token>` (one-time account setup links) has `access_log off`: the token is a credential
  until used. Caddy keeps no access log, and the backend logs route patterns only
  (`cmd/server/requestlog.go`), so the token isn't logged anywhere. The API takes it in request
  bodies, never in URLs.

Local dev uses neither: Vite serves the frontend on :3000 and proxies `/api` to the backend.

## Backend

Go (gin + gorm) in `backend/`, image `ghcr.io/jackgraver/st-backend`. Listens on `:8080`
(`/health` for health checks). Config comes from env vars (see `docker-compose.vps.yml`);
production refuses insecure settings (non-HTTPS CORS origins, dev auth bypass).
The schema lives in `backend/migrations/` and the server refuses to start while any are pending.
The image also ships `/app/migrate`, which j-cicd's VPS agent runs as the `pre_swap` step
(`docker compose run --rm --no-deps backend /app/migrate up`, in the new image, before the
container is replaced).

- **HTTP server** (`cmd/server/serve.go`): read-header 5 s, read 30 s, write 45 s, idle 120 s.
  On SIGTERM it stops accepting connections, lets in-flight requests finish for up to 25 s,
  then closes the DB pool. `stop_grace_period: 30s` on the backend service gives it that time
  (Docker's default is 10 s).
- **Hardening middleware** (`cmd/server/hardening.go`, first in the chain): every response gets
  `Cache-Control: no-store`; request bodies are capped at 1 MiB, with larger limits per route
  in `bodyLimits` (5 MiB for the Renpho weight import, which also caps rows). Over the limit:
  413 when `Content-Length` says so, otherwise the read fails and binding returns 400.
- **CORS** (`corsMiddleware`, same file) allows only the origins in `CORS_ORIGINS` (required in
  production: `APP_ORIGIN`). There's no default: unset means same-origin only, and requests with
  any other `Origin` get 403. The SPA always calls the API on its own origin (Caddy in prod, the
  Vite proxy in dev and for `dev phone`).
- **Connection pool** (`internal/database/pool.go`): max 20 open, 10 idle, 30 min lifetime,
  5 min idle time. Override with `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`,
  `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`.

## Postgres

`postgres:16-alpine` with the `postgres_data` volume. Pre-deploy backups are in `~/backups` on the
VPS; rollbacks never revert the database.

### Database roles

Two roles, each with its own connection string in the backend container's environment:

| Role | Env var | Used by | Can |
|---|---|---|---|
| owner (`POSTGRES_USER`, `simpletracker`; a superuser in the postgres image) | `DATABASE_OWNER_URL` | `/app/migrate` (pre_swap); in dev also devseed, devuser, e2eseed, pgweb, `dev test` | everything; owns every table |
| `simpletrack_app` | `DATABASE_URL_PRODUCTION` (`DATABASE_URL` in dev/e2e) | the server | `SELECT/INSERT/UPDATE/DELETE` on tables, `USAGE/SELECT` on sequences, read-only `schema_migrations`; no DDL, no `TRUNCATE`, no ownership, no `BYPASSRLS` |

- The app role exists so row-level security (multi-user project) applies to the server: RLS
  doesn't apply to table owners or superusers. In production the server **refuses to start**
  if its role is a superuser, has `BYPASSRLS` or owns a table (`database.CheckAppRole`); in
  development it logs a warning.
- Migration `0010_app_role.sql` creates the role `NOLOGIN` with no password, grants, and
  `ALTER DEFAULT PRIVILEGES` so tables later migrations create are granted automatically. It also
  sets `statement_timeout = 15s` and `idle_in_transaction_session_timeout = 30s` on the role
  (cluster-wide; the owner has neither, so migrations aren't limited).
- **Password:** `APP_DB_PASSWORD` in `~/simple-track/.env.vps`, interpolated into
  `DATABASE_URL_PRODUCTION`. `migrate up` checks whether that DSN can log in and, if not, sets
  `LOGIN` and the password on the role (sent as a SCRAM verifier, never in a migration). So
  rotating it is: change `APP_DB_PASSWORD`, run a backend deploy (or `migrate up` by hand), and
  restart the backend. Use hex (`openssl rand -hex 24`): it goes into a URL unescaped.
- The owner's credentials are still in the backend container's environment, because the
  agent's `pre_swap` runs in the backend service. Code running in the server could read them;
  the split protects against query bugs, not a compromised server process. See
  `docs/multi-user/DECISIONS.md`.
- Dev (`dev/env.sh`) and the e2e suite (`e2e/infra/postgres.ts`) use the same split, with fixed
  local passwords. The e2e build's `/__e2e/reset` uses its own owner connection, since
  `TRUNCATE ... RESTART IDENTITY` needs ownership.
- The root `docker-compose.yml` `backend` service still connects as the `postgres` superuser.
  It runs with `APP_ENV=development`, so it only gets the warning. A local production-mode copy
  must use an app-role DSN.

First rollout (multi-user project, lane D1): `docs/multi-user/ROLLOUT.md`.

## Build and deploy

`j-cicd` (`.cicd.yaml`) builds clean from `main`, runs vet/tests/lint/e2e, pushes both images to
GHCR, and has the VPS agent pull and restart the compose stack. `.cicd.yaml` sets no
`deploy.compose_file`, so the agent uses the VPS's own copy of `docker-compose.vps.yml` (the path
in its config): **changes to that file reach production only when copied there by hand**. It then checks
`https://st.jackgraver.ca/version.json` for the new commit.

## Stale artifact

`frontend/fly.toml` is a leftover from an old Fly.io deployment (`st-backend.fly.dev`). Production
does not use it; safe to delete.
