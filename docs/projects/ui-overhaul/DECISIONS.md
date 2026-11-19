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
