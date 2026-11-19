# UI overhaul tasks

One task = one slice = one branch/PR. Each has a done-state. Every slice that touches UI also
meets `docs/design/RULES.md`: lint green with suppressions pruned, `vue-tsc` + `vitest` pass,
screenshots of the changed pages or stories at 390 and 1280, in dark and light (light once F1 lands).

## Phase 1: charter
- [x] Code audit (tokens, palette usage, primitives, states, PrimeVue): README "Current state", MIGRATION.md
- [x] PRINCIPLES.md, TOKENS.md, PRIMITIVES.md, RULES.md (now in `docs/design/`)
- [x] Decide the canonical Tailwind config (`style.css`); PrimeVue moved to the bake-off
- [x] Owner answers the open questions (defaults accepted 2026-11-19; PrimeVue → bake-off)
- [x] Setup: ESLint (vue + better-tailwindcss) with suppressions replaces the `ui-guard` test; lint in `dev check` + j-cicd; docs moved to `docs/design/`; `frontend/CLAUDE.md`, `ui` skill, root CLAUDE.md pointer
- [ ] Owner: install the Playwright and impeccable plugins

## Phase 2: library bake-off
- [x] **B1 lab app**: `lab/` (throwaway Vite + Vue + Tailwind v4 + Storybook, not part of the deploy) with one story set per candidate: PrimeVue Volt, PrimeVue styled + custom preset, Reka UI + own Tailwind. Each renders (a) mobile gym set logging at 390 (big number, stepper, full-width CTA, bottom sheet, toast with Undo) and (b) desktop finances at 1280 (table/list, cards, form dialog, badges, chart), with the TOKENS.md colours and a stock-look toggle, in dark and light. *Done:* screenshots of every candidate × screen × theme on one comparison page; notes on bundle size, a11y (Storybook a11y addon), effort to theme, and component gaps; optional `/impeccable critique` pass. *Result (2026-11-19):* branch `proj/ui-bakeoff`, page https://claude.ai/artifact/4buX5k9LpVSCJ36o8LEMux, summary in DECISIONS.
- [ ] **B2 decision**: the owner picks; DECISIONS entry; update PRIMITIVES.md (library-backed vs built), TOKENS.md names, and the phase 3–4 tasks below. Delete `lab/` or keep it out of the build.
- ~~Screenshot every route~~: dropped, the owner knows the UI (DECISIONS 2026-11-19)

## Phase 3: foundations (re-check after B2)
- [ ] **F1 token layer**: add the TOKENS.md `@theme` block + utilities next to the legacy aliases (unchanged values); `data-theme` bootstrap in `index.html` (default dark); remove the `.dark` wrapper in `App.vue` and PrimeVue `darkModeSelector`; `html` colour from tokens; global macro tokens (delete `pages/diet/macro-colors.css`, `SimpleMacros`/`MacroBar` use them); delete `tailwind.config.js`. *Done:* build passes; home, diet and gym logging look identical at 390/1280 apart from the unified fat colour; setting `localStorage["st:theme"]="light"` flips the canvas/text of the shell.
- [ ] **F2 charts**: one chart wrapper (library's or `shared/chart/Chart.vue`) + `shared/chart/theme.ts` reading `--color-chart-*`; switch the 5 chart users; delete both `readChartTheme` copies; drop the Aura preset. *Done:* the weight, steps, progression and finances charts render in both themes and re-colour on theme switch.
- [ ] **F3 theme toggle**: Settings control (Dark/Light/System), `theme-color` meta follows. Shell (SideNav, BottomTabBar, MobileHeader, DateAndBreadcrumbs) migrated to new tokens. *Done:* shell screenshots in both themes at 390/1280; shell files have no entries in `eslint-suppressions.json`.

- [ ] **F4 Storybook in `frontend/`** (`@storybook/vue3-vite`, a11y addon, MCP addon, theme switcher, 390/1280 viewports) + the chosen library installed and wrapped in `shared/ui/`. *Done:* `npm run storybook` works; one sample story; `no-restricted-imports` blocks the library outside `shared/ui/`.

## Phase 4: shared components (`frontend/src/shared/ui/`, each with stories for all states, both themes)
- [ ] **P1** `Button`, `IconButton`, `Spinner`. *Done:* every variant × size × state (hover/focus/disabled/loading) visible; keyboard focus ring visible; 44px at 390px.
- [ ] **P2** `Field`, `Input`, `Textarea`, `Select`, `Toggle`, `Checkbox` (evolve `shared/input/Input.vue`, keep its import path working or update its 4 users). *Done:* showcase with invalid/disabled states; no iOS zoom (16px) below lg; signin page uses Field+Input.
- [ ] **P3** `Card`, `Badge`, `List`/`ListRow`, `SegmentedControl`. *Done:* showcase includes `ListRow pending` and the desktop hover actions / mobile trailing button.
- [ ] **P4** `EmptyState`, `ErrorState`, `Skeleton`/`SkeletonList`, `useDelayedFlag`. *Done:* `LoggingPageShell.vue` uses them (first real consumer).
- [ ] **P5** `Overlay` (sheet < lg, dialog ≥ lg) and `DialogContainer.vue` rewritten on it, with the `dialogManager` API unchanged + `confirm({ tone: "danger" })`. *Done:* all 19 dialog call sites open correctly at 390 (sheet) and 1280 (dialog); focus trap/restore and Esc work; `DialogContainer` has no `<style>`.
- [ ] **P6** Toast re-skin + `toast.push({ message, tone, action })`. *Done:* toasts sit above the tab bar on mobile; Undo action in a story; the `.has-tab-bar .toast-container` rule is gone from `style.css`.
- [ ] **P7** `PageHeader`, `Section`; decide the `ModuleTitleBar` relationship (DECISIONS). *Done:* used on one page.

## Phase 5: module migration (one per slice; pair with optimistic-ui where its slice is ready)
Done-state for each: the module's files have no entries left in `eslint-suppressions.json`; every query view has
loading/empty/error; every mutation has pending + failure toast; before/after at 390 and 1280,
dark and light.
- [ ] **M1 gym/logging** (mobile priority): `logging/index.vue`, `ExerciseListView`, `exercise/*` screens, `LoggedSetsList`, `NumberSlider` tokens, `WeightSetupDialog`
- [ ] **M2 diet day + log meal**: `diet/index`, `DietDayView`, `MealCard`, `PlannedMealsDisplay`, `logmeal/*`, `QuickLogDialog`, `CreateFoodDialog`, water dialogs
- [ ] **M3 diet planning**: `diet/edit-planned` (largest file, 51 violations), `diet/targets/*`, `diet/water-presets`
- [ ] **M4 finances**: retire `ui` in `finances/format.ts` for primitives; `index`, `account`, `rules`, dialogs, `RoomCard`
- [ ] **M5 gym/plans + progression + weight + steps** (desktop-heavy; `EditExerciseDialog`, `CreateExerciseForPlanDialog`)
- [ ] **M6 media** (fixes the broken `bg-primary` buttons)
- [ ] **M7 grocery**
- [ ] **M8 home, settings, auth/signin**
- [ ] **M9 shared leftovers**: `SearchList` on Input+List, `Dropdown` → `Combobox`, `ModuleTitleBar` overflow → `Menu`

## Phase 6: enforcement + cleanup
- [ ] Add the planned lint rules (`docs/design/RULES.md` "Planned rules") as their conditions are met
- [ ] Remove the global `button`/`input`/`select`/`.delete-button`/`.confirm-button` CSS from `App.vue`
- [ ] `--color-*: initial;` in `@theme`; delete the legacy aliases and `:root` kebab vars
- [ ] Final pass: no UI suppressions left, delete MIGRATION.md, retro, update README
