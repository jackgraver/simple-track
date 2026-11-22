# UI overhaul

**Goal**: one consistent, clean, functional UI across mobile and desktop, with a little beauty. Gym logging is mobile-first; other modules are desktop or split.

**Done when**: every page uses the token set and shared components, no raw colours or styles outside the token layer (`eslint-suppressions.json` has no UI entries left), every list/form has loading/empty/error states, and Storybook shows every shared component in both themes. **Met 2026-11-19** (phases 1–6 done on `proj/ui-charter`; merging to `main` is the owner's call).

**Non-goals**: new features, a new brand. A component library is now allowed, chosen by a bake-off (see Phases).

## Where things live
- **Design system (permanent):** `docs/design/`: [PRINCIPLES](../../design/PRINCIPLES.md), [RULES](../../design/RULES.md), [TOKENS](../../design/TOKENS.md), [PRIMITIVES](../../design/PRIMITIVES.md). This is what new UI work reads; this project folder is history.
- **Agent entry points:** `frontend/CLAUDE.md` (auto-loaded under `frontend/`), the `ui` skill (`.claude/skills/ui/`), and a pointer in the root `CLAUDE.md`.
- **This project:** README (status + retro), TASKS, DECISIONS.
- **Enforcement:** ESLint (`frontend/eslint.config.js`), runs in `npm run lint`, `dev check` and j-cicd. `eslint-suppressions.json` is empty.

## Status (2026-11-19): done
- Phases 1–6 are done. Every page is on tokens and `shared/ui/` primitives (Reka UI underneath), both themes work, and every primitive has stories in Storybook.
- All of it is on `proj/ui-charter` (the phase 6 slices are stacked on it: `proj/ui-p6-element-css` → `-tokens` → `-lint` → `-final`). **None of it is merged to `main` or pushed.** Merging is the owner's call; it's a large diff, so deploy it on its own.
- `lab/` and the bake-off stay on `proj/ui-bakeoff`, unmerged (comparison page https://claude.ai/artifact/4buX5k9LpVSCJ36o8LEMux).

## Phases
1. **Charter + setup** (done).
2. **Library bake-off**: Reka UI + own Tailwind picked over PrimeVue Volt and PrimeVue styled (done).
3. **Foundations**: tokens, Storybook, Reka in `shared/ui/`, theme toggle, charts on tokens (done).
4. **Shared components**, each with stories (done).
5. **Module migration**: M1–M9, run as four parallel workers (done).
6. **Enforcement + cleanup**: lint rules, no global element CSS, `--color-*: initial`, no legacy aliases (done).

## Open follow-ups (outside this project)
- Not verified in a browser: forced error and empty states (they're in the code), Combobox load-more paging, grocery and finances with no data, signin register mode (no UI reaches it).
- `useExerciseLoggingSession.ts` writes to a readonly query proxy on every set log (Vue warning). Fix with optimistic set logging (optimistic-ui project).
- Candidate primitives if a second user appears: a radio-card group (PresetDialog, StarRating, PlanCalendar share the sr-only-radio pattern), a Progress bar (RoomCard), a Banner/tinted Card (finance stale warnings), a Sheet presentation for Menu/Combobox on mobile.
- `MobileHeader`'s module switcher is a hand-rolled popover; move it to `Menu`.
- Dev: `POST /api/auth/login` is rate-limited, so screenshot scripts get 429s. Use a cookie `auth_token=<DEV_AUTH_TOKEN>` instead.

## Retro (2026-11-19)
- **What worked:** the bake-off before tokens (the library decided naming and effort, and Reka's density matched the charter). ESLint with bulk suppressions as a ratchet: 849 violations to 0 without blocking anyone. Building primitives with a real consumer per slice (P4 logging shell, P7 water-presets) caught API gaps early. Pixel-diffing every route before and after the cleanup made deleting the global CSS safe.
- **Parallel module migration** (four workers, one lead): fast, and module boundaries kept conflicts to `eslint-suppressions.json` and `shared/ui/index.ts`, both trivial to resolve. What made it work: one owner per shared change (others cherry-pick its commit), workers writing doc updates to notes files instead of the shared docs, and the lead reviewing screenshots before merging. The review still caught real issues (nested buttons, `role=button` stars, `!` in a page, devseed amounts), so keep it.
- **What cost time:** the global element CSS in `App.vue`. It forced `!` overrides into every primitive, hid a content-box bug on links (`a { all: unset }`), and made Storybook differ from the app. Next time, remove global element styling first, before building primitives.
- **Gaps:** error/empty states were mostly checked in code, not by forcing failures in a browser. A small Playwright "force each state" helper would close that for future work.
