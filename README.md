# Simple Tracking
Couldn't find an app that worked for me, just wanted something simple that I could click as few buttons and it keeps track of all my numbers.
## Development

One-time setup: install [mise](https://mise.jdx.dev) (optionally `eval "$(mise activate bash)"`
in `~/.bashrc` so `dev` is on PATH), then run `mise trust` in the repo.

```
dev              # start postgres, backend (air), frontend (vite), pgweb in a TUI
dev start        # same, in the background; waits until ready (dev stop / status / logs <proc>)
dev env          # this checkout's ports and URLs
dev api /auth/me # curl the backend as the dev user
dev ls           # list services and tasks
dev off pgweb    # skip a service locally (dev on pgweb to re-enable)
dev db:psql      # any task in dev/tasks
```

Log in locally as **dev / dev**. The account is created automatically on start.

Each git worktree gets its own ports (+100 per worktree), database and process-compose instance,
so several can run at once. `dev env` shows the URLs.

- **Add a service:** drop a process-compose YAML in `dev/services/`. **Remove:** delete it, or `dev off <name>`.
- **Add a task:** drop an executable script in `dev/tasks/` (subfolders become `group:name`).
- **Add a tool:** add it under `[tools]` in `mise.toml`.

Main checkout URLs: frontend http://localhost:3000 · API http://localhost:8080 · pgweb http://localhost:8081
