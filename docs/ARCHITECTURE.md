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

Current config:

```caddy
st.jackgraver.ca {
	encode zstd gzip
	handle_path /api/* {
		reverse_proxy 127.0.0.1:8080
	}
	handle {
		reverse_proxy 127.0.0.1:3000
	}
}
```

- **TLS:** automatic certificates for `st.jackgraver.ca`. The backend relies on HTTPS
  (`AUTH_COOKIE_SECURE=true`, and production CORS origins must be HTTPS).
- **Compression:** Caddy compresses the whole site (`encode zstd gzip`). nginx deliberately
  doesn't gzip (see below).
- **`/api` prefix is stripped** by `handle_path`. The backend's routes have no `/api` prefix
  (`/health`, `/auth/me`, ...), so the browser calls `/api/auth/me` and the backend sees `/auth/me`.
  The Vite dev proxy does the same rewrite, so dev and prod behave alike.
- **`/version.json`** goes to nginx and is served unauthenticated; j-cicd's post-deploy check reads it.
- No CDN sits in front of Caddy.

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

Local dev uses neither: Vite serves the frontend on :3000 and proxies `/api` to the backend.

## Backend

Go (gin + gorm) in `backend/`, image `ghcr.io/jackgraver/st-backend`. Listens on `:8080`
(`/health` for health checks). Config comes from env vars (see `docker-compose.vps.yml`);
production refuses insecure settings (non-HTTPS CORS origins, dev auth bypass).
The schema lives in `backend/migrations/` and the server refuses to start while any are pending.
The image also ships `/app/migrate`.

## Postgres

`postgres:16-alpine` with the `postgres_data` volume. Pre-deploy backups are in `~/backups` on the
VPS; rollbacks never revert the database.

## Build and deploy

`j-cicd` (`.cicd.yaml`) builds clean from `main`, runs vet/tests/lint/e2e, pushes both images to
GHCR, and has the VPS agent pull and restart the compose stack. It then checks
`https://st.jackgraver.ca/version.json` for the new commit.

## Stale artifact

`frontend/fly.toml` is a leftover from an old Fly.io deployment (`st-backend.fly.dev`). Production
does not use it; safe to delete.
