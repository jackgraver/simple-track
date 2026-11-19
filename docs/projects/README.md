# Long-running projects

| Project | Goal | Status |
|---|---|---|
| [ui-overhaul](ui-overhaul/README.md) | Consistent, clean, functional UI built on tokens and shared primitives | Reka UI picked; Phase 3 done (F1–F4), Phase 4 (P1–P7) done; next: M1 |
| [optimistic-ui](optimistic-ui/README.md) | Linear-style instant UI on top of TanStack Query | Not started (planning) |
| [performance](performance/README.md) | Measured backend and frontend optimisation, open-ended | Not started (planning) |

Repo files are the source of truth for status. Agents read them, and nothing else is needed.

## Session protocol (every session, every project)

1. **Start**: read the project's `README.md` (rules), `TASKS.md` (pick ONE slice) and the tail of `DECISIONS.md`.
2. **Branch**: one slice = one branch = one worktree, branched from `main` (`git worktree add .worktrees/<name> -b <type>/<name>`). Keep it small and merge soon. No long-lived project branches.
3. **Work** the slice. Keep `main` shippable: new code coexists with old, pages migrate one by one.
4. **End**: tick `TASKS.md`, append any decision to `DECISIONS.md`, and update the "Next up / gotchas" block in the project README. Do this in the same PR as the code.
5. Commit style: `type(scope): ...` as usual. Add the project tag in the body (`Project: ui-overhaul`) so `git log --grep` shows progress.

## Slice sizing

A slice should fit in one session and have a concrete done-state, e.g. "gym/exercise-logging uses new primitives + optimistic set logging, screenshots at 390px and 1280px attached". If it can't be stated like that, split it.

## Cross-project order

UI and optimistic work both touch every page. Do foundations first (tokens/primitives; query conventions), then migrate **per module as one vertical slice** so each page is touched once. Performance work is orthogonal and fills gaps.

Cadence: every ~5 sessions, one retro session. Prune TASKS, revisit DECISIONS, check READMEs still match reality.
