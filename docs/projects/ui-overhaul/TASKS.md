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
- [x] **B2 decision** (2026-11-19: Reka UI + own Tailwind, see DECISIONS; `lab/` stays on `proj/ui-bakeoff`, not merged): the owner picks; DECISIONS entry; update PRIMITIVES.md (library-backed vs built), TOKENS.md names, and the phase 3–4 tasks below. 
- ~~Screenshot every route~~: dropped, the owner knows the UI (DECISIONS 2026-11-19)

## Phase 3: foundations (re-checked after B2: Reka UI)
- [x] **F1 token layer** (branch `proj/ui-tokens`, 2026-11-19; `darkModeSelector` now follows `[data-theme="dark"]` until F2 deletes Aura; macro bars take the token colours, not just fat): add the TOKENS.md `@theme` block + utilities next to the legacy aliases (unchanged values); `data-theme` bootstrap in `index.html` (default dark); remove the `.dark` wrapper in `App.vue` and PrimeVue `darkModeSelector`; apply the B1 contrast fixes to the light-theme status/soft values (TOKENS.md); `html` colour from tokens; global macro tokens (delete `pages/diet/macro-colors.css`, `SimpleMacros`/`MacroBar` use them); delete `tailwind.config.js`. *Done:* build passes; home, diet and gym logging look identical at 390/1280 apart from the unified fat colour; setting `localStorage["st:theme"]="light"` flips the canvas/text of the shell.
- [x] **F2 charts** (branch `proj/ui-charts`, 2026-11-19; `chartTheme()` is reactive via a MutationObserver on `data-theme`; stories in `shared/chart/Chart.stories.ts`): one chart wrapper (library's or `shared/chart/Chart.vue`) + `shared/chart/theme.ts` reading `--color-chart-*`; switch the 5 chart users; delete both `readChartTheme` copies; drop the Aura preset and uninstall `primevue` / `@primeuix/themes`; register only the chart.js parts in use; read tokens after `data-theme` is set. *Done:* the weight, steps, progression and finances charts render in both themes and re-colour on theme switch.
- [x] **F3 theme toggle** (branch `proj/ui-theme`, 2026-11-19; control lives on the Settings profile page for now): Settings control (Dark/Light/System), `theme-color` meta follows. Shell (SideNav, BottomTabBar, MobileHeader, DateAndBreadcrumbs) migrated to new tokens. *Done:* shell screenshots in both themes at 390/1280; shell files have no entries in `eslint-suppressions.json`.

- [x] **F4 Storybook in `frontend/`** (branch `proj/ui-storybook`, 2026-11-19; sample = `shared/ui/Toggle` on Reka Switch) (`@storybook/vue3-vite`, a11y addon, MCP addon, theme switcher, 390/1280 viewports) + `reka-ui` installed (pin 2.x; the lab used 2.11.0) and wrapped in `shared/ui/`. *Done:* `npm run storybook` works; one sample story; `no-restricted-imports` blocks `reka-ui` outside `shared/ui/`.

## Phase 4: shared components (`frontend/src/shared/ui/`, each with stories for all states, both themes)
- [x] **P1** (branch `proj/ui-p1-buttons`, 2026-11-19; stories `shared/ui/{Button,IconButton,Spinner}.stories.ts`, `Matrix` shows variant × size × state) `Button`, `IconButton`, `Spinner`. *Done:* every variant × size × state (hover/focus/disabled/loading) visible; keyboard focus ring visible; 44px at 390px.
- [x] **P2** (branch `proj/ui-p2-fields`, 2026-11-19; stories `shared/ui/Field.stories.ts` "UI/Fields/Showcase"; `shared/input/Input.vue` deleted, its 2 users moved to `~/shared/ui`; signin and logmeal files pruned from suppressions) `Field`, `Input`, `Textarea`, `Select`, `Toggle`, `Checkbox` (evolve `shared/input/Input.vue`, keep its import path working or update its 4 users). *Done:* showcase with invalid/disabled states; no iOS zoom (16px) below lg; signin page uses Field+Input.
- [x] **P3** (branch `proj/ui-p3-lists`, 2026-11-19; stories `Card`, `Badge` (Matrix), `UI/List` Showcase; `SegmentedControl` already existed from F3 with its stories) `Card`, `Badge`, `List`/`ListRow`, `SegmentedControl`. *Done:* showcase includes `ListRow pending` and the desktop hover actions / mobile trailing button.
- [x] **P4** (branch `proj/ui-p4-states`, 2026-11-19; stories `UI/States`; `useDelayedFlag` has a unit test; `LoggingPageShell` has no `<style>` and is out of the suppressions; checked on the real page with `/api` blocked, delayed and an unknown id) `EmptyState`, `ErrorState`, `Skeleton`/`SkeletonList`, `useDelayedFlag`. *Done:* `LoggingPageShell.vue` uses them (first real consumer).
- [x] **P5** (branch `proj/ui-p5-overlay`, 2026-11-19; stories `UI/Overlay` and `UI/DialogContainer (dialogManager)`; swipe is a pointer handler on the handle/header; checked: sheet 390 and dialog 1280 in both themes, Esc, scrim, ✕, focus restore, scroll lock, swipe close/snap-back, diet Log water and Quick log in the real app at both widths, confirm/custom resolve in the story; the other call sites share the same path but were not each opened by hand) `Overlay` (sheet < lg, dialog ≥ lg) and `DialogContainer.vue` rewritten on it, with the `dialogManager` API unchanged + `confirm({ tone: "danger" })`. *Done:* all 19 dialog call sites open correctly at 390 (sheet) and 1280 (dialog); on Reka `Dialog`; swipe-to-close on the sheet (pointer handler or vaul-vue, none of the candidates had it); focus trap/restore and Esc work; `DialogContainer` has no `<style>`.
- [x] **P6** (branch `proj/ui-p6-toast`, 2026-11-19; view in `shared/ui/Toaster.vue` + stories `UI/Toast` incl. Undo; `ToastContainer` feeds it; checked in the real app at 390/1280 both themes, bottom above the tab bar on /diet, top on the gym logging page, cap of 3) Toast on Reka `Toast`, re-skinned + `toast.push({ message, tone, action })`. *Done:* toasts sit above the tab bar on mobile and don't cover the gym "Log set" CTA (stack above it on focus pages, or confirm inline); Undo action in a story; the `.has-tab-bar .toast-container` rule is gone from `style.css`.
- [x] **P7** (branch `proj/ui-p7-header`, 2026-11-19; stories `UI/PageHeader and Section`; used on `diet/water-presets`, which is now fully on primitives and out of the suppressions; checked add/edit/delete(confirm)/empty/error at 390 and 1280, both themes) `PageHeader`, `Section`; decide the `ModuleTitleBar` relationship (DECISIONS). *Done:* used on one page.

## Phase 5: module migration (one per slice; pair with optimistic-ui where its slice is ready)
Done 2026-11-19: four parallel workers (gym, diet, finances+grocery, media/home/settings/shared), each slice stacked on its own branch and merged into `proj/ui-charter`. `eslint-suppressions.json` now only lists `App.vue` (phase 6). Optimistic-ui was not ready, so no slice paired with it. Error/empty states were checked in code, not by forcing failures in the browser.
Done-state for each: the module's files have no entries left in `eslint-suppressions.json`; every query view has
loading/empty/error; every mutation has pending + failure toast; before/after at 390 and 1280,
dark and light.
- [x] **M1 gym/logging** (branch `proj/ui-m1-gym-logging`, 2026-11-19; new primitives NumberField, Menu, ListRow `#end`; step nav no longer hidden under the tab bar) (mobile priority): `logging/index.vue`, `ExerciseListView`, `exercise/*` screens, `LoggedSetsList`, `NumberSlider` tokens, `WeightSetupDialog`
- [x] **M2 diet day + log meal** (branch `proj/ui-m2-diet-day`, 2026-11-19; MacroBar redesigned, log-meal rows on NumberField + Menu, CreateFoodDialog failure toasts): `diet/index`, `DietDayView`, `MealCard`, `PlannedMealsDisplay`, `logmeal/*`, `QuickLogDialog`, `CreateFoodDialog`, water dialogs
- [x] **M3 diet planning** (branch `proj/ui-m3-diet-planning`, 2026-11-19; edit-planned split into PlanCalendar/PlannedDayEditor/BulkRangePlanner; targets g/lb ratio bug fixed): `diet/edit-planned` (largest file, 51 violations), `diet/targets/*`, `diet/water-presets`
- [x] **M4 finances** (branch `proj/ui-m4-finances`, 2026-11-19; `ui` object gone from `format.ts`): retire `ui` in `finances/format.ts` for primitives; `index`, `account`, `rules`, dialogs, `RoomCard`
- [x] **M5 gym/plans + progression + weight + steps** (branch `proj/ui-m5-gym-plans`, 2026-11-19; desktop-heavy; `EditExerciseDialog`, `CreateExerciseForPlanDialog`)
- [x] **M6 media** (branch `proj/ui-m6-media`, 2026-11-19; StarRating is a radio group) (fixes the broken `bg-primary` buttons)
- [x] **M7 grocery** (branch `proj/ui-m7-grocery`, 2026-11-19; check-off strike-through now visible; ghost-text suggestion became a suggestion list)
- [x] **M8 home, settings, auth/signin** (branch `proj/ui-m8-home-settings`, 2026-11-19; also not-found)
- [x] **M9 shared leftovers** (branch `proj/ui-m9-shared`, 2026-11-19; Combobox primitive; `Dropdown.vue` and `shared/input/` deleted in `proj/ui-p5-gym-followups`): `SearchList` on Input+List, `Dropdown` → `Combobox`, `ModuleTitleBar` overflow → `Menu`

## Phase 6: enforcement + cleanup
- [ ] Add the planned lint rules (`docs/design/RULES.md` "Planned rules") as their conditions are met
- [ ] Remove the global `button`/`input`/`select`/`.delete-button`/`.confirm-button` CSS from `App.vue`
- [ ] `--color-*: initial;` in `@theme`; delete the legacy aliases and `:root` kebab vars
- [ ] Final pass: no UI suppressions left, delete MIGRATION.md, retro, update README
