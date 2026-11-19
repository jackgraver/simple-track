# UI overhaul

**Goal**: one consistent, clean, functional UI across mobile and desktop, with a little beauty. Gym logging is mobile-first; other modules are desktop or split.

**Done when**: every page uses the token set and shared components, no raw colours or styles outside the token layer (`eslint-suppressions.json` has no UI entries left), every list/form has loading/empty/error states, and Storybook shows every shared component in both themes.

**Non-goals**: new features, a new brand. A component library is now allowed, chosen by a bake-off (see Phases).

## Where things live
- **Design system (permanent):** `docs/design/`: [PRINCIPLES](../../design/PRINCIPLES.md), [RULES](../../design/RULES.md), [TOKENS](../../design/TOKENS.md), [PRIMITIVES](../../design/PRIMITIVES.md).
- **Agent entry points:** `frontend/CLAUDE.md` (auto-loaded under `frontend/`), the `ui` skill (`.claude/skills/ui/`), and a pointer in the root `CLAUDE.md`.
- **This project:** README (status), TASKS, DECISIONS, [MIGRATION.md](MIGRATION.md) (audit snapshot, old→new class mapping, migration plan).
- **Enforcement:** ESLint (`frontend/eslint.config.js`) with existing violations in `frontend/eslint-suppressions.json`. Runs in `npm run lint`, `dev check` and j-cicd.

## Status (2026-11-19)
- Charter done and agreed. The owner accepted the proposed defaults: blue accent, inverted neutral primary button, darker surfaces, Dark/Light/System toggle, bottom sheets on mobile, phone-width gym logging on desktop.
- Setup done: ESLint replaces the earlier vitest ratchet. Docs moved to `docs/design/`. `frontend/CLAUDE.md` and the `ui` skill added. The owner is installing the Playwright and impeccable plugins.
- **Not started**: no tokens, components or pages changed yet.

## Current state (audited 2026-11-19; details in MIGRATION.md)
- Tailwind v4 `@theme` in `frontend/src/style.css` has 8 camelCase colour aliases and nothing for spacing, radius, type, motion or z. `tailwind.config.js` is dead (delete in the token slice).
- PrimeVue: only `primevue/chart` (5 files) + the Aura plugin. Its future depends on the bake-off.
- Lint baseline: 849 suppressed violations in 80 files: 552 restricted classes (mostly `zinc-*`), 185 unknown classes (mostly custom CSS classes for `<style>` blocks), 66 deprecated v3-era classes, 23 `<style>` blocks, 23 colour literals in TS.
- 173 raw `<button>`s and 93 raw `<input>`s, styled by global element CSS in `App.vue`.
- States: ~16 bare "Loading…" strings, error UIs in only 8 files.

## Phases
1. **Charter + setup** (done 2026-11-19).
2. **Library bake-off**: a throwaway `lab/` app with Storybook. Candidates are PrimeVue Volt, PrimeVue styled mode with our own preset, and Reka UI with our own Tailwind styling. Each builds the same two screens (mobile gym set logging, desktop finances) in both themes with our tokens. The owner picks one. Vercel-owned options (Nuxt UI, shadcn-vue) and daisyUI (no behaviour) were ruled out.
3. **Foundations**: token layer (names adapted to the chosen library), Storybook in `frontend/`, library install + `shared/ui/` wrapper pattern, theme toggle, charts on tokens, delete `tailwind.config.js`.
4. **Shared components** in PRIMITIVES.md priority order, each with stories.
5. **Module migration**, one module per slice, paired with optimistic-ui where it's ready.
6. **Enforcement + cleanup**: planned lint rules (RULES.md), remove the global element CSS, `--color-*: initial`, delete the legacy aliases.

## Next up / gotchas
- **Next slice: B1 bake-off** (TASKS.md). Don't start the token layer before the library is picked, because token names may follow the library's convention.
- A lint failure in a file with suppressed violations lists *all* of that rule's violations in the file. Diff against `main` to find the new one. Never run `--suppress-all` to get past it.
- The global `button`/`input`/`select` CSS in `App.vue` styles every raw element. Removing it early restyles ~170 buttons.
- Real bugs the lint found: `bg-primary` (media pages, no background) and `md:grid-none` (not a Tailwind class). Fix them in their module slices.
- `.light` in `style.css` is never applied; `App.vue` hard-codes a `.dark` wrapper for PrimeVue.
- Macro colours only exist inside `.diet-module` (`pages/diet/macro-colors.css`).

## Open questions for the owner
- After the bake-off: which library, and is its stock look a starting point or something to restyle fully?
