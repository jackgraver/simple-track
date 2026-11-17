# simple-track

Go API (`backend/`, gin + gorm, Postgres) and Vue 3 + Vite frontend (`frontend/`).

## Dev environment (mise + process-compose)

Tooling is managed by [mise](https://mise.jdx.dev) via `mise.toml` (Node, air, process-compose,
pgweb; Go comes from the system). mise may not be activated in the shell, so call
`./bin/dev` by path. It re-runs itself under `mise exec` when needed. For other tools use
`mise exec -- <cmd>`. If mise refuses the config in a new worktree, run `mise trust`.

Every checkout gets a **slot**: the main checkout is slot 0 (default ports), and each git worktree gets
slot N with every port shifted by N*100, its own docker compose project (separate Postgres
container + volume) and its own process-compose socket. Several worktrees can run stacks at
the same time. **Never hardcode ports; ask `./bin/dev env`.**

| | slot 0 | slot N |
|---|---|---|
| frontend (vite, proxies `/api` -> backend) | 3000 | 3000+N*100 |
| backend (`/health`) | 8080 | 8080+N*100 |
| pgweb | 8081 | 8081+N*100 |
| postgres (`simpletracker_dev`, postgres/postgres) | 5432 | 5432+N*100 |

### Verifying changes as an agent

Never run bare `./bin/dev` (it opens an interactive TUI). Use the background commands:

```
./bin/dev start            # start in background, block until ready, print URLs (idempotent)
./bin/dev status           # process table
./bin/dev logs backend 50  # recent logs (backend, frontend, postgres, pgweb, devuser)
./bin/dev api /auth/me     # curl the backend as the dev user (extra args go to curl)
./bin/dev stop             # stop when done, if you started it
```

- The backend (air) and frontend (vite) hot-reload, so there's no need to restart after edits.
  Check `dev logs backend` for build errors after Go changes.
- Login: `dev` / `dev` (prefilled on the sign-in page in vite dev mode). The `devuser` step
  creates the account on every start. Browser/e2e flows can log in through
  `POST <frontend>/api/auth/login` with that account.
- `dev api` uses the `DEV_AUTH_TOKEN` cookie bypass from `backend/.env` (`ALLOW_BYPASS` is
  forced on by `dev/env.sh`; the backend rejects both in production).
- A fresh worktree starts with an empty DB: the one-shot `migrate` service applies every
  migration before the backend starts, and there's no sample data. `dev db:drop-worktree` removes a worktree's DB.

### Layout (plug-and-play)

- `bin/dev`: entrypoint (`./bin/dev help`). `dev/env.sh`: slot/port/env computation.
- `dev/services/*.yaml`: one process-compose file per service, merged at startup. Use
  `${DEV_*_PORT}` vars, never literal ports. Long-running services need a
  `readiness_probe`, and one-shot steps must not have one, because `dev start` waits for the
  probes to pass and for the one-shots to exit 0. Toggle locally with `dev off|on <svc>`.
- `dev/tasks/`: executable scripts become mise tasks (`dev/tasks/db/psql` -> `dev db:psql`),
  described by a `#MISE description="..."` header. Run them via `./bin/dev <task>` so the
  slot env applies.
- `mise.toml` `[tools]`: add CLI tools here rather than installing them globally.

process-compose always runs with `-U` (unix socket). Its default TCP API port is 8080, the
same as the backend.

## Notes

- Backend config: `backend/.env` (created by `dev` from `.env.example` with random secrets).
  Env vars exported by `dev/env.sh` take precedence over it.
- `cmd/devuser` creates or resets the dev login and refuses to run in production.
- `dev db:reset` and `dev db:drop-worktree` are destructive to dev data.

### Database migrations

The schema lives only in `backend/migrations/NNNN_name.sql` (no GORM AutoMigrate). Format and
rules are documented in `backend/internal/migrate/file.go`. The files are embedded into the
server, which refuses to start while any are pending, and into `cmd/migrate`, which applies them.

- `./bin/dev migrate new <name>` creates the next file. Fill in `-- migrate:up` and
  `-- migrate:down` (or `-- migrate:down irreversible`).
- `./bin/dev migrate up | down [n] | redo | status | check`.
- Never edit a migration that has been applied anywhere but your dev DB. Write a new one instead.
  In dev, `migrate redo` re-runs the latest one after you've edited it.
- Tests for `internal/migrate` need `TEST_DATABASE_URL` (`dev test` sets it). Module tests
  still build their schema with AutoMigrate on SQLite.
