# Starter prompt

Paste into a new session, started from the repo root (`~/Projects/simple-track`), and fill in the lane:

```
Multi-user project, lane <LANE> (e.g. B1). Read docs/multi-user/README.md, your lane in
docs/multi-user/TASKS.md and the tail of DECISIONS.md, from the base branch
(git fetch origin && git show origin/proj/multi-user:docs/multi-user/README.md, or read them in
.worktrees/multi-user after pulling). Then follow the README's session protocol: work in the
lane worktree .worktrees/mu-<lane>-<name> on branch mu/<lane>-<name> (create it from
origin/proj/multi-user if missing; if it exists, rebase it on origin/proj/multi-user first),
check the lane's dependencies have merged into proj/multi-user, do the work, verify with
./bin/dev test / check / e2e, update TASKS/DECISIONS/Gotchas in the same PR, and open the PR
into proj/multi-user. Don't merge it and don't deploy. If a design point in the README turns
out wrong or a decision is mine to make, stop and ask me.
```

For base upkeep (merge `main` into the base, renumber migrations), use:

```
Multi-user project: base upkeep. In .worktrees/multi-user, merge origin/main into
proj/multi-user as described under "Base upkeep" in docs/multi-user/README.md, renumber project
migrations if main added any, run ./bin/dev test, and push.
```
