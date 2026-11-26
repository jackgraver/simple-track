# simple-track

Go API (`backend/`, gin + gorm, Postgres) and Vue 3 + Vite frontend (`frontend/`).

Production topology (Caddy on the VPS, nginx in the frontend image, `/api` routing, `TRUSTED_PROXIES`)
is in `docs/ARCHITECTURE.md`. Read it before touching `frontend/nginx.conf`, the compose files,
proxy/CORS/cookie settings or anything about client IPs, and update it when that changes.

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

Starting a stack is cheap and safe: do it whenever a change needs to be seen or exercised (UI work,
API behaviour, anything a type-check can't prove), preferably from your own worktree (own slot, ports
and DB). Skip it for docs, config or pure refactors that tests cover. Only `stop` a stack you started.

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
- **e2e:** `./bin/dev e2e` (args go to `playwright test`, e.g. `--grep @smoke` or `tests/gym`). It is
  self-contained (own Postgres container, DBs and backends), so it doesn't need `dev start`. How to
  run, write and debug specs: `e2e/README.md`.
- `dev api` uses the `DEV_AUTH_TOKEN` cookie bypass from `backend/.env` (`ALLOW_BYPASS` is
  forced on by `dev/env.sh`; the backend rejects both in production).
- A fresh worktree starts with an empty DB: the one-shot `migrate` service applies every
  migration before the backend starts, then the one-shot `devseed` step (`cmd/devseed`) adds
  sample gym/diet/tracking/media/money data. It does one check (any exercises?) and skips if
  data exists, so it runs on every start; `dev db:seed` runs it by hand. `dev db:drop-worktree` removes a worktree's DB.

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

## Long-running projects

UI overhaul, optimistic UI and performance work are tracked in `docs/projects/`. Before working on any
of them, read `docs/projects/README.md` (session protocol) and that project's README/TASKS. Update
TASKS/DECISIONS in the same PR as the code.

## UI / design system

Any frontend UI change follows `docs/design/` (start with `docs/design/README.md`; the `ui` skill and
`frontend/CLAUDE.md` summarise the rules). `npm run lint` in `frontend/` enforces them for new code
and runs in `dev check` and the j-cicd pipeline.

## Notes

- Backend config: `backend/.env` (created by `dev` from `.env.example` with random secrets).
  Env vars exported by `dev/env.sh` take precedence over it.
- `cmd/devuser` creates or resets the dev login and refuses to run in production.
- `dev db:reset` and `dev db:drop-worktree` are destructive to dev data.
- Slot numbers are reused after a worktree is deleted; a reassigned slot's leftover containers/volumes
  are purged automatically (`dev/purge-slot`). `dev db:gc [--dry-run]` removes any others with no live worktree.

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

## Deployment

Prod deployments go through j-cicd (config in `.cicd.yaml`): run `j-cicd guide` for how it
works, and `j-cicd run --plan` to preview what a deploy would do. Don't build images or SSH to
the VPS by hand. `docker-compose.vps.yml` is only the compose file j-cicd's VPS agent uses, not
a deploy path.

- Commit and push to `main` first. Pipelines build main's latest commit, not the working tree.
- Always ask the user before running a prod deploy unless they already told you to deploy.
- A run ends by checking `https://st.jackgraver.ca/version.json` for the deployed commit (the
  `stamp-version` step bakes it into the frontend image).
- Pre-deploy DB backups are in `~/backups` on the VPS; `rollback` never reverts the database.
