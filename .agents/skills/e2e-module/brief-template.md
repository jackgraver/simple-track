# Shared brief: <module> e2e lanes (<N> parallel agents)

You are one of <N> agents, each writing one PR of Playwright e2e scenarios for the simple-track
<module> module (`/<module>` and its subpages). Repo: <repo path>. The other agents work at the same
time in other worktrees. Do only your lane.

## Setup
1. You are in a git worktree. Base your work on the latest main:
   `git fetch origin && git checkout -B e2e/<module>-<lane> origin/main` (local main may be stale).
   If mise refuses the config, run `mise trust`. Node comes from mise: use `mise exec -- <cmd>`.
2. Read first, fully: `e2e/README.md`, `e2e/TASKS.md` (the <module> entry + Findings), `e2e/BACKLOG.md`,
   the existing specs in `e2e/tests/<module>/`, `e2e/fixtures/api.ts`, `e2e/fixtures/index.ts`,
   `backend/internal/e2e/seed/seed.go`, the project `CLAUDE.md` and `frontend/CLAUDE.md`.
3. Seed facts: <the epoch date and weekday, the timezone, the seeded rows this module relies on, with IDs>.

## Rules
- Follow e2e/README.md exactly: real stack, no mocks, arrange via API, act + assert via UI, reload after
  every write, role-first selectors, web-first assertions, no waitForTimeout, faults/holds for failures.
- Tag every test `@<module>` (plus `@smoke` only where your lane says so).
- Queries retry twice on 5xx (3 attempts), so a load-error fault needs `times: 3`; skip the backoff with
  `page.clock.runFor` after a failed response lands. Find the real table/op for each fault (a gorm
  `Scan` or raw query is op `row`/`raw`).
- Run with a small worker count because several lanes run at once:
  `cd e2e && E2E_WORKERS=3 mise exec -- npx playwright test tests/<module>/<your files>` (desktop and mobile
  must both pass), then `E2E_WORKERS=3 mise exec -- npm run flake -- <your specs>` (10x, must pass), then
  `--grep @<module>` once to check you didn't break the existing tests. `mise exec -- npx tsc --noEmit` clean.
- **Never** run `npm run down`, drop/reset the shared `simple-track-e2e` Postgres container, or stop any
  dev stack. Other sessions use them. Only touch what your checkout owns.
- **Do not fix app bugs.** Keep the main test asserting something that holds before and after the fix,
  and pin the intended behaviour in a separate `test.fail()` with a comment pointing at your notes.
  No changes under `frontend/` or `backend/` (call out any test hook you truly need).
- If a scenario doesn't match the real UI, adapt it to what the app really does and say so in the notes.
  If something is untestable, note why and move on.
- API helpers: add them to `e2e/fixtures/api.ts`, appended at the END of the class under a comment
  `// <module> <lane>`. Spec-local helpers stay in the spec; helpers another lane could use go in a shared
  non-spec file under `tests/<module>/`.
- **Do not edit `e2e/TASKS.md` or `e2e/BACKLOG.md`.** Create `e2e/notes/<module>-<lane>.md` with:
  - `## Scenarios`: each one done / test.fail / skipped, with spec file names.
  - `## Bugs found`: what, repro, which test pins it, suggested fix.
  - `## Follow-ups`: uncovered scenarios, test-suite issues, product questions.
- Commit (`test(e2e): ...`), push the branch and open a PR against `main` with `gh pr create`. Title:
  `test(e2e): <module> <lane> scenarios`. Body: summary, scenarios, bugs found, run results. **Never merge.**

## Final report (your last message)
PR URL, scenario list with status, bugs found (short), anything you couldn't do and why, and the exact
commands you ran with their results. Keep it tight.
