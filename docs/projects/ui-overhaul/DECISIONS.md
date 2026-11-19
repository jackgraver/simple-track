# Decisions (append, newest last; 3-5 lines each)

## 2026-11-19: Project scaffold
Repo-file tracking in `docs/projects/`. UI and optimistic work foundations first, then per-module vertical slices.

## 2026-11-19: Charter from code, not a screenshot audit
The owner knows the UI, so phase 1 became a code audit and a written charter (PRINCIPLES, TOKENS, PRIMITIVES, RULES).
Alternative: screenshot every route at 390/1280. Dropped as low-signal. Screenshots happen per slice, before and after.

## 2026-11-19: `style.css` `@theme` is canonical, delete `tailwind.config.js`
v4 ignores the JS config without `@config`, and it defines tokens that don't exist (`border`, `--color-gradient`).
Alternative: `@config` it. Rejected: two sources of truth and the v3 API.

## 2026-11-19: Semantic kebab tokens, added next to legacy and removed at the end
New names (`canvas`, `surface[-2/-3]`, `line`, `fg[-muted/-subtle]`, `accent`, `success/warn/danger/info[-soft]`, `macro-*`, `chart-*`) live next to the camelCase aliases, which keep their values, so nothing shifts until a module migrates. Themes flip values via `[data-theme]` on `<html>`, never `dark:`.
Alternatives: repoint the legacy names at once (an app-wide unreviewed restyle), or rename in place (one huge PR).

## 2026-11-19: Neutral primary, one accent (proposed blue), red means danger only
Primary button = inverted neutral (`fg` on `canvas`). Accent is for focus, selection and progress. `cfRed` stops being an accent.
Alternatives: red accent (clashes with destructive), green primary (finances today; clashes with success). Pending owner sign-off (README open question 1–2).

## 2026-11-19: Drop PrimeVue for a local chart.js wrapper
Only `primevue/chart` (5 files) is used. Aura adds runtime CSS and the `.dark` coupling.
Alternatives: keep the chart with `unstyled` (keeps a big dependency for 40 lines), or adopt PrimeVue as the primitive layer (fights the look, and the non-goals rule out a library migration).

## 2026-11-19: Own primitives in `shared/ui/`, native-first, keep existing APIs
Native `<select>`, `role=switch` toggle, one `Overlay` (sheet < lg, dialog ≥ lg). `dialogManager` and `toast.push` keep their signatures and get re-skinned, so their 19 and 162 call sites don't change.
Alternatives: a headless lib (Reka/Radix-Vue). Revisit only if focus-trap or popover positioning gets hard. A separate Sheet and Dialog would make every caller choose a presentation.

## 2026-11-19: Enforcement by a ratchet test, not ESLint
`frontend/src/ui-guard.test.ts` counts colour literals, palette classes, arbitrary values and `<style>` blocks per file, against `ui-guard.baseline.json`. It runs in vitest (and so j-cicd), can't fail current code, and the baseline only goes down. Endgame: `--color-*: initial` makes the palette uncompilable.
Alternative: ESLint + vue/tailwind plugins. There's no ESLint in the repo today, so it's a new toolchain for the same check.

## 2026-11-19: One layout breakpoint (`lg`)
Mobile/desktop switch at 64rem, matching the nav. `sm`/`md` only for in-component tweaks. The `max-width: 767px` media queries get replaced when their files are touched.

## 2026-11-19: ESLint with bulk suppressions replaces the vitest ratchet
`eslint-plugin-vue` + `eslint-plugin-better-tailwindcss`: restricted palette/arbitrary classes, unknown/conflicting/deprecated classes, no `<style>`, no colour literals. The 849 existing violations live in `eslint-suppressions.json`. Lint runs in `dev check` and j-cicd (`frontend-lint`). Rules get added as phases land (RULES.md "Planned rules").
Alternatives: oxlint (only lints `<script>` in `.vue`, no template support; the `oxlint-vue` wrapper is v0.3), or keep `ui-guard.test.ts` (a custom regex, not extensible, can't ban elements).

## 2026-11-19: Permanent design docs in `docs/design/`, discoverable without the project
PRINCIPLES/TOKENS/PRIMITIVES/RULES moved out of the project folder, which closes when the migration ends. Entry points: `frontend/CLAUDE.md` (auto-loaded under `frontend/`), the `ui` skill, a root CLAUDE.md pointer. Migration-only content moved to `MIGRATION.md`. PRINCIPLES gained §0 (what the app is, usage contexts, feel).
Alternative: keep everything in the project folder. Rejected: agents outside the project never read it.

## 2026-11-19: Storybook instead of a `/dev/ui` route
Stories are isolated per state, screenshot-able without the backend or login, come with the a11y addon, and have an MCP addon so agents can look components up.
Alternatives: a `/dev/ui` route (no tooling, but no isolation or addons), Histoire (Vue-native, smaller ecosystem).

## 2026-11-19: Component library allowed; bake-off before tokens
Supersedes the earlier "drop PrimeVue" call. Candidates: PrimeVue Volt (Tailwind v4, copy-in, on unstyled PrimeVue), PrimeVue styled mode with a custom preset, and Reka UI + own Tailwind. Ruled out: Nuxt UI and shadcn-vue (Vercel-owned/affiliated, the owner's preference; shadcn also looks generic), daisyUI (no behaviour), Vuetify/Quasar (Material, fights Tailwind), Naive UI/Element Plus (own theming, admin look). The token layer waits for the pick because names may follow the library.

## 2026-11-19: Owner accepted the charter defaults
Blue accent, inverted neutral primary button, darker surfaces, Dark/Light/System toggle (default dark), bottom sheets below `lg`, phone-width gym logging on desktop.

## 2026-11-19: Bake-off results (B1); the pick is the owner's (B2)
`lab/` built gym logging (390) and finances (1280) with PrimeVue Volt, PrimeVue 4 styled + a token preset, and Reka UI + own Tailwind, on the TOKENS.md values, both themes, stock toggle. Page: https://claude.ai/artifact/4buX5k9LpVSCJ36o8LEMux.
- Looks (impeccable critique, PRINCIPLES as the standard): Reka 28/40, Volt 24, styled 24. Once themed all three look close; Reka alone hits the desktop density (14px, 40px rows) without overrides. Aura's proportions show through styled; Volt's secondary buttons read as disabled after the remap.
- Effort: styled = one 125-line preset (plus a CSS layer-order fix). Volt = 112-line palette remap + hand edits to 4 of 18 copied files, and the copies have upstream class typos. Reka = 322 lines of our own components for this set (all PRIMITIVES are ours to build).
- Size (gzip, both screens): Reka +55 kB; styled +159 kB, Volt +172 kB, both + 69 kB lazy `chart.js/auto` from `primevue/chart`. Shared baseline 85 kB.
- Behaviour: all trap focus and close on Esc. PrimeVue Drawer drops focus to `<body>` and doesn't lock scroll by default; InputNumber's ± buttons have no accessible name. Reka restores focus, locks scroll and labels its steppers. None has swipe-to-close sheets.
- Licence: PrimeVue 5 moved to the PrimeUI licence (yearly Community key for individuals, a badge without one). The lab pinned 4.5.5 (MIT, frozen since Apr 2026).
Recommendation: Reka + own Tailwind (looks and behaviour match the charter; costs the most build time). Runner-up: PrimeVue 4 styled + preset (fastest to theme; licence and Aura density to manage). Not recommended: Volt.

## 2026-11-19: Reka UI + own Tailwind (B2); lab stays on `proj/ui-bakeoff`
The owner picked Reka UI. It is headless: behaviour (focus, ARIA, keyboard, portals) from the library, every look from our tokens and Tailwind. Supersedes "Own primitives in `shared/ui/`, native-first" only where that said to revisit a headless lib: Reka is now the behaviour layer for Overlay, Toast, Toggle, Checkbox, SegmentedControl, the number stepper, and later Combobox/Menu. Select stays native, and Button/Input/Card/Badge/List/states stay plain Vue. Table: PRIMITIVES.md "Library-backed".
- Why: matched the density and behaviour in the charter at the best critique score (28/40), +55 kB gzip against +159–172 kB for PrimeVue, no licence question, no stock look to fight or Aura proportions to override.
- Cost accepted: we build and own every primitive (322 lines for the bake-off set), including sheet swipe-to-close.
- Dropped: PrimeVue entirely. Charts move to a chart.js wrapper (F2), so `primevue`, `@primeuix/themes` and the Aura plugin go.
- Token names are unchanged (Reka has no naming convention). The B1 contrast fixes apply to values in F1.
- `lab/` is not merged to `main`. It stays on branch `proj/ui-bakeoff` with the comparison page https://claude.ai/artifact/4buX5k9LpVSCJ36o8LEMux as the record. `lab/scripts/contrast.mjs` is the source for the F1 contrast fixes.
Alternatives: PrimeVue 4 styled + preset (runner-up: fastest to theme, but a 159 kB runtime, the licence track and Aura density), Volt (not recommended).

## 2026-11-19: Light theme bridges the legacy aliases until phase 6
`:root[data-theme="light"]` repoints `mainBg/firstBg/secondBg/thirdBg/textPrimary/textSecondary/--color-border` at the light tokens, so unmigrated pages stay readable when the toggle is switched. Dark keeps the original legacy values (nothing shifts). Hard-coded palette classes (`zinc-*`, `amber-*`) in unmigrated files still render their dark colours in light; each module slice fixes its own. The bridge is deleted with the aliases in phase 6.
Alternative: ship the toggle with the legacy aliases unchanged (light would show white-on-white text in every unmigrated page).

## 2026-11-19: Solid danger button uses `text-fg-inverse`, not `on-solid` (P1)
White on the dark `danger` fill (hsl 0 80% 66%) is under 4.5:1. `fg-inverse` flips per theme (near-black in dark, near-white in light) and gives 5.9:1 on `danger` in dark and 5.8:1 in light, so no new token is needed. `on-solid` stays for the accent Toggle thumb. Primary (`bg-fg text-fg-inverse`) is unchanged.
Also: Button's `:to` renders a `RouterLink` only while not disabled or loading (an anchor can't be disabled); `loading` keeps the label, shows a spinner in place of the left icon, sets `aria-busy` and ignores clicks without removing focus. Sizes are 44px below `lg` for all three, then 32/40/44px (sm/md/lg) from `lg`.

## 2026-11-19: Field provides the control id by injection; signin tidied (P2)
`Field` provides `{ id, describedBy, invalid }` and `Input`/`Textarea`/`Select` inject it, so callers write `<Field label><Input v-model/></Field>` with no ids or aria wiring (PRIMITIVES showed a slot-prop form; injection is the same contract with less typing). All three also work without a Field when given `aria-label`.
`shared/input/Input.vue` was deleted rather than kept as a shim: it had two users. The signin page got real labels (it had empty ones and an empty login button label) and lost a stray always-visible email input and ~40 lines of commented-out markup. Registration mode is still unreachable (no UI toggles it): out of scope, left as is.
Checkbox tick uses `text-fg-inverse` on the accent fill (contrast in both themes), like danger-solid in P1.
