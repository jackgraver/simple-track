# UI overhaul tasks

One task = one slice = one branch/PR. Each has a done-state. Every slice that touches UI also
meets RULES.md (guard green and baseline lowered, `vue-tsc` + `vitest` pass, screenshots of the
changed pages at 390 and 1280 in dark and light (light once F1 lands)).

## Phase 1: charter
- [x] Code audit (tokens, palette usage, primitives, states, PrimeVue): README "Current state", TOKENS.md
- [x] PRINCIPLES.md, TOKENS.md, PRIMITIVES.md, RULES.md
- [x] Decide canonical Tailwind config (`style.css`) and PrimeVue (drop): TOKENS.md, DECISIONS
- [x] `ui-guard` ratchet test + baseline (`frontend/src/ui-guard.test.ts`)
- [ ] Owner answers the open questions in README (accent, primary button, surfaces, theme toggle, sheets, desktop gym, PrimeVue)
- ~~Screenshot every route~~: dropped, the owner knows the UI (DECISIONS 2026-11-19)

## Phase 2: foundations
- [ ] **F1 token layer**: add the TOKENS.md `@theme` block + utilities next to the legacy aliases (unchanged values); `data-theme` bootstrap in `index.html` (default dark); remove the `.dark` wrapper in `App.vue` and PrimeVue `darkModeSelector`; `html` colour from tokens; global macro tokens (delete `pages/diet/macro-colors.css`, `SimpleMacros`/`MacroBar` use them); delete `tailwind.config.js`. *Done:* build passes; home, diet and gym logging look identical at 390/1280 apart from the unified fat colour; setting `localStorage["st:theme"]="light"` flips the canvas/text of the shell.
- [ ] **F2 charts**: `shared/chart/Chart.vue` + `theme.ts`; switch the 5 `primevue/chart` users; delete both `readChartTheme` copies; remove `primevue`, `@primeuix/themes` and the `app.use(PrimeVue)` block. *Done:* `grep -r primevue frontend/src` is empty; the weight, steps, progression and finances charts render in both themes and re-colour on theme switch.
- [ ] **F3 theme toggle**: Settings control (Dark/Light/System), `theme-color` meta follows. Shell (SideNav, BottomTabBar, MobileHeader, DateAndBreadcrumbs) migrated to new tokens. *Done:* shell screenshots in both themes at 390/1280; shell files at 0 in the guard baseline.

## Phase 3: primitives (`frontend/src/shared/ui/`, each shown in `/dev/ui` with all states, both themes)
- [ ] **P1** `/dev/ui` route (dev-only, behind `import.meta.env.DEV`) + `Button`, `IconButton`, `Spinner`. *Done:* every variant × size × state (hover/focus/disabled/loading) visible; keyboard focus ring visible; 44px at 390px.
- [ ] **P2** `Field`, `Input`, `Textarea`, `Select`, `Toggle`, `Checkbox` (evolve `shared/input/Input.vue`, keep its import path working or update its 4 users). *Done:* showcase with invalid/disabled states; no iOS zoom (16px) below lg; signin page uses Field+Input.
- [ ] **P3** `Card`, `Badge`, `List`/`ListRow`, `SegmentedControl`. *Done:* showcase includes `ListRow pending` and the desktop hover actions / mobile trailing button.
- [ ] **P4** `EmptyState`, `ErrorState`, `Skeleton`/`SkeletonList`, `useDelayedFlag`. *Done:* `LoggingPageShell.vue` uses them (first real consumer).
- [ ] **P5** `Overlay` (sheet < lg, dialog ≥ lg) and `DialogContainer.vue` rewritten on it, with the `dialogManager` API unchanged + `confirm({ tone: "danger" })`. *Done:* all 19 dialog call sites open correctly at 390 (sheet) and 1280 (dialog); focus trap/restore and Esc work; `DialogContainer` has no `<style>`.
- [ ] **P6** Toast re-skin + `toast.push({ message, tone, action })`. *Done:* toasts sit above the tab bar on mobile; Undo action demo in `/dev/ui`; the `.has-tab-bar .toast-container` rule is gone from `style.css`.
- [ ] **P7** `PageHeader`, `Section`; decide the `ModuleTitleBar` relationship (DECISIONS). *Done:* used on one page.

## Phase 4: module migration (one per slice; pair with optimistic-ui where its slice is ready)
Done-state for each: the module's files are at 0 `paletteClass`/`colorLiteral`/`styleBlock` in the
baseline (arbitrary values only for layout, with a reason); every query view has
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

## Phase 5: enforcement + cleanup
- [ ] Add `legacyToken` and `!important` metrics to `ui-guard` (RULES.md "Planned enforcement")
- [ ] Remove the global `button`/`input`/`select`/`.delete-button`/`.confirm-button` CSS from `App.vue`
- [ ] `--color-*: initial;` in `@theme`; delete the legacy aliases and `:root` kebab vars
- [ ] Add a CLAUDE.md line: "UI changes follow docs/projects/ui-overhaul/RULES.md"
- [ ] Final pass: guard baseline empty (or only justified layout arbitraries), retro, update README
