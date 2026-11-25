---
name: e2e-module
description: Lock down a simple-track module (e.g. /diet, /finances, /media) with Playwright e2e scenarios - research the pages, propose tiered flows for the owner to approve, split them into lanes, run one agent per lane in parallel worktrees, review, fold findings into e2e/TASKS.md and e2e/BACKLOG.md, merge, and report the bugs found. Use when the user wants e2e coverage for a module or its subpages ("e2e the diet pages", "/e2e-module /finances").
argument-hint: <module route, e.g. /diet>
---

# e2e a module: propose, fan out, review, merge

Target: `$ARGUMENTS` (a top-level route such as `/diet`; if empty, ask which module).

The owner decides **what** gets tested (they approve the flows) and **which bugs get fixed**. They don't
want to review test code. You own the rest: research, lanes, agents, review, merging, the bug write-up.

## 0. Ground truth first (read-only, no questions yet)
- `git fetch origin`. Work from `origin/main`; the local `main` is often behind.
- Read `e2e/README.md` (principles, fixtures), `e2e/TASKS.md` (this module's B-entry, Findings),
  `e2e/BACKLOG.md`, `backend/internal/e2e/seed/seed.go` (what the seed gives you and the epoch date) and
  the module's existing specs under `e2e/tests/<module>/`.
- Map the pages: `frontend/src/router/index.ts` → each route's component → `frontend/src/pages/<module>/`
  → its queries and mutations (`frontend/src/api/<module>/`) → the backend handlers for the writes.
  Note labels, ARIA names, empty and error states, and where the data-loss risks are (writes that span
  several requests, replace-by-id, optimistic rows, offline queues).
- **Run the module's existing specs on origin/main** (`./bin/dev e2e --grep @<module>`). If they're red,
  that becomes part of a lane's job; say so in the proposal.
- `gh pr list`: note open PRs that touch the module's UI. They will break new selectors; mention them.

## 1. Propose flows (one message, then wait)
Write a tiered list for the owner to agree with or change. No code yet.
- Group by page or area. Inside each, rank: golden paths and data-loss risks first, then navigation
  and state, then error states. 30-50 well-chosen tests per module beat hundreds of shallow ones.
- One line per scenario: what's arranged, what's done in the UI, what's asserted (after a reload).
- Say what you're leaving out and why (mock/preview sections, canvas-only details, things that need
  data the seed and API can't produce).
- If the module has subpages, propose **lanes**: one per area that can be tested independently
  (for gym: home, session/logging, plans, tracking). Each lane becomes one PR and one agent.
- Give a recommendation for scope and order. Ask nothing else.

## 2. Fan out (after the owner agrees)
1. Copy `brief-template.md` (next to this file) to your scratchpad as the shared brief. Fill in the
   module, the seed facts and anything module-specific you learned in step 0.
2. Spawn **one agent per lane** with `isolation: "worktree"` and `run_in_background: true`. Each
   prompt is short: "Read and follow <brief path>. Your lane is `<lane>` (branch `e2e/<module>-<lane>`,
   notes `e2e/notes/<module>-<lane>.md`)", then the lane's pages/files and the approved scenario list,
   word for word.
3. Lanes branch off `origin/main` independently (parallel agents can't stack). To keep merges easy the
   brief forbids agents from editing `TASKS.md`/`BACKLOG.md`: they write a per-lane notes file and
   append API helpers at the end of `fixtures/api.ts` under a `// <module> <lane>` comment.
4. While they run, relay cross-lane findings with `SendMessage` (for example "the existing spec X is red
   on main; it's in your lane, fix it"). Report each finished lane to the owner in a few lines: PR,
   scenario count, bugs found.

## 3. Review (after every lane has reported)
For each PR, read the spec diff (not every line; look for these):
- **Bugs pinned as passing behaviour.** A test that asserts the bug's output (raw error text, wrong
  copy) must instead assert something that also holds after the fix, plus a separate `test.fail` that
  states the intended behaviour.
- Violations of e2e/README principles: CSS selectors, `waitForTimeout`, arranging through the UI or
  `db` when the API could do it, a write without a reload check, mocks (`route.fulfill`).
- Helpers duplicated across lanes (extract them into a shared non-spec file, e.g.
  `tests/<module>/<helpers>.ts`, when merging).
- Things that hide real behaviour (for example clearing storage before a reload). Keep them only with
  a stated reason in a comment.
Fix what you find in the lane's worktree and re-run that spec.

Then fold each lane's notes into the docs, **on its own branch**, and delete the notes file:
- `TASKS.md`: one `- [x] <lane> (<spec files>): <what's covered>. Not yet: <gaps>.` entry under the
  module's B-entry. Update the module's old todo line.
- `BACKLOG.md`: app bugs under "App bugs" as `**<Module>: <symptom>.** <cause>. Fix: <fix>; then drop
  the test.fail in <spec>.`; product questions under "Product decisions needed"; test-suite items under
  "Test-suite follow-ups". Update or drop items the lanes found fixed.

## 4. Merge
Ask the owner once whether they want to look at the PRs or merge straight away (for e2e-only PRs they
may say merge). Then, one PR at a time, in an order where anything that fixes main goes first:
1. `gh pr merge <n> --merge`.
2. In the next lane's worktree: `git fetch origin && git merge origin/main` (a merge, not a rebase, so
   no force-push is needed). `TASKS.md`/`BACKLOG.md` conflicts are both-sides additions: keep both.
   Re-run `--grep @<module>`, then push.
3. GitHub may report "merge conflicts" for a few seconds after a push. Wait briefly and retry
   `gh pr merge` before investigating.

## 5. Report
End with a short bug rundown for the owner, grouped as:
- **App bugs** (one line each: symptom, cause, suggested fix, the spec whose `test.fail` to drop).
- **Product questions** (missing controls, odd defaults: "intended?").
- **Test-suite follow-ups** (only the ones worth their attention).
Point them at `e2e/BACKLOG.md`. Don't fix app bugs until they sign off.

Clean up: remove the lane worktrees once their PRs are merged (`git worktree remove`), and only the
ones you created.
