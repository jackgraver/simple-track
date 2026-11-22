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

## 2026-11-19: ListRow picks its element from its props; actions are a sibling (P3)
`to` renders a RouterLink, an attached `@click` renders a `<button>`, otherwise a `<div>`. The desktop `#actions` slot sits beside that element inside the `<li>` (a button can't contain buttons), and the hover/selected background lives on the `<li>` so the actions area highlights with the row. Actions are `lg`-only and appear on hover or focus-within; on mobile the caller puts an `IconButton` in `#trailing` (the story hides it at `lg`). `selected` sets `aria-current`, not `aria-selected` (rows are not an ARIA listbox).

## 2026-11-19: ErrorState takes `onRetry` as a prop; the logging shell gets a retry (P4)
`ErrorState` shows Retry only when `@retry` is bound (declared as an `onRetry` prop, so the template can tell). The logging shell's error branch now keeps the back header, shows `ErrorState` and emits `retry`, which `exercise/[id].vue` wires to a new `refetch` returned by `useWorkoutStore` (a one-line addition to the store). The pending branch waits 150ms (`useDelayedFlag`) before showing skeletons, and the "Set —" placeholder is gone.

## 2026-11-19: Overlay on Reka Dialog with keyframe tokens; swipe on the handle only (P5)
`Overlay` = Reka `Dialog*` + our sheet/dialog presentation chosen by `matchMedia(min-width: 64rem)` (`useIsDesktop`). Enter/exit use keyframes tokens in `style.css` (`animate-*-in/out`) because Reka's Presence only waits for animations, not transitions. Swipe-to-close is a ~30-line pointer handler on the drag handle and header (not vaul-vue: one less dependency); it captures the pointer only after 4px of movement so the ✕ still clicks, and closes past 96px or a fast flick over 40px. The body is never a drag target, so scrolling is safe.
`DialogContainer` keeps the `dialogManager` API and holds the last dialog while the exit animation runs. Without a description, Overlay renders the title as an `sr-only` `DialogDescription` (Reka warns otherwise). `confirm({ tone: "danger" })` renders `danger-solid`; it is set on the 10 delete/remove confirms (not on archive, overwrite or "day already assigned"). Confirm bodies now render `message` as the dialog description, and the footer stacks full-width on mobile with the primary on top.
Custom dialog bodies are unchanged and still use legacy styling until their module slice (M2–M5).

## 2026-11-19: Toasts go to the top on mobile focus pages; `bottom-tabbar` renamed (P6)
The Log set button sits at the bottom of the gym logging page, exactly where a mobile toast above the tab bar lands. Of the options in the task (stack above the CTA, or confirm inline) neither is page-agnostic, so `App.vue` passes `route.meta.focus` to the container and focus pages show toasts at the top (`pt-safe`); every other mobile page keeps them above the tab bar, and desktop is bottom-right. The `.has-tab-bar` class and its rule in `style.css` are gone (`App.vue` passes `showNav` as a prop instead).
`toast.push(message, type)` is unchanged; `push({ message, tone, action, duration })` adds Undo (6s default with an action, 5s for errors, 3s otherwise), at most 3 open (the oldest is dismissed), Reka pauses on hover/focus and swipes right to dismiss. `Toaster.vue` lives in `shared/ui/` (the only place Reka may be imported) and takes the list as props; `ToastContainer.vue` is the thin adapter.
Bug found: the `bottom-tabbar` / `pb-tabbar` utilities were silently shadowed by Tailwind's own `bottom-<spacing>` / `pb-<spacing>` (the `--spacing-tabbar` token), giving 4rem instead of the intended offset. Renamed to `bottom-above-tabbar` / `pb-above-tabbar` (neither was used before).

## 2026-11-19: PageHeader for sub-pages, ModuleTitleBar stays the module nav (P7)
`ModuleTitleBar` is the module's sub-navigation (title, sibling links, measured overflow, day offset). `PageHeader` is a page's own title with optional meta and actions, used on sub-pages and detail pages that have no sibling links (water-presets is the first). They don't merge: a module root keeps `ModuleTitleBar`; its overflow becomes a `Menu` in M9. Below `lg` PageHeader's actions stack under the title rather than collapsing into a menu (no Menu primitive yet).
`Button` `:to` now takes a route location object, not only a string. The water-presets page was migrated whole (header, sections, form, list with loading/empty/error) as the first real consumer of every phase 4 primitive; its saved-size rows use trailing Edit/Delete `IconButton`s at every width because the hover-actions pattern has no mobile equivalent until `Menu` exists.

## 2026-11-19: Phase 5 ran as four parallel workers, merged into `proj/ui-charter`
Gym (M1, M5), diet (M2, M3), finances + grocery (M4, M7) and media/home/settings + shared (M6, M8, M9), each in its own worktree with stacked branches per slice, reviewed and merged by a lead session. Workers did not edit these docs (four branches would conflict); the lead folded their notes in. Shared code was owned by one worker per change (gym: NumberField, Menu, ListRow `#end`, Button fix; shared worker: Combobox, SearchList, ModuleTitleBar), and the others cherry-picked those commits.

## 2026-11-19: ListRow `#end` slot for per-row actions (M1)
A clickable ListRow renders a `<button>`/`<a>`, so an IconButton in `#trailing` nests interactive elements, and `#actions` is lg-only. `#end` is a sibling of the main element at every width. `#trailing` is for passive content. Supersedes the P3 advice ("on mobile the caller puts an IconButton in `#trailing`"), which only works on non-clickable rows.

## 2026-11-19: NumberField, Menu and Combobox on Reka (M1, M9)
NumberField: labelled steppers, 44px below lg, `inputmode=decimal`, select-all on focus, `unit`, works inside Field. Menu: `<Menu :items @select>`, a popover at every width (no Sheet). Combobox: action-style (emits `select`, clears), our own filtering (`ignore-filter`) so server search works; Reka 2.11 keeps the search term on `ComboboxInput`, not on the root. Popover on mobile too: neither caller needed a Sheet. `Dropdown.vue` was kept as a wrapper during the parallel work, then deleted once its two gym callers moved to Combobox.

## 2026-11-19: `box-border` on primitives that render links
App.vue's `a { all: unset }` resets `box-sizing` to content-box on anchors, so a `w-full` Button `:to` measured 408px in a 358px column. Button, Card and ListRow carry `box-border`. The rule goes with the global reset in phase 6.

## 2026-11-19: Radio patterns built on sr-only native radios
Choice controls that have no primitive (finance account presets, media star rating, the diet planning calendar) are a `<fieldset>` of styled `<label>`s around `sr-only` `<input type=radio>`, styled with `has-checked:`/`has-focus-visible:`. Native keyboard and one tab stop, and the global input CSS never applies, so there's no `!`. Rejected: `role=button` spans (10 tab stops on the stars) and raw buttons with `!` overrides.
Promote to a primitive if a fourth user appears.

## 2026-11-19: Too many options for SegmentedControl: Select below lg (M3, M6)
SegmentedControl is for 2–5 options that fit at 390. The media status picker (5–6 statuses) is a SegmentedControl from `lg` and a native Select below it. TDEE goals (5 long labels) are a Select. The media status filter (All + 6 + counts) is a horizontally scrolling row of ghost/secondary Buttons with `aria-pressed`.

## 2026-11-19: Behaviour changes made during phase 5
- Diet `PlannedMealsDisplay` dropped its viewport-height pagination (it hid meals on short desktop screens). All planned meals render.
- MealCard macros per food are always visible (they were hover-only). The dot pager became a SegmentedControl.
- MacroBar: label + percent, number, and a 6px token-coloured bar (it was white text over a fill, which can't be themed).
- Grocery: the inline ghost-text suggestion became a short suggestion list under the input (same keys: Up/Down, Enter, Tab/→ to fill, Esc). The ghost needed a transparent input over a mirror span and never worked on mobile. Check-off keeps the row in place until the strike-through finishes; before, the refetch removed it within ~100ms.
- Gym exercise step nav (Setup/Log/Rest) is a SegmentedControl above the tab bar. It used to sit under the tab bar on mobile, where it couldn't be reached.
- Multi-control rows (planned cardio, assign to day) use a `Section` heading plus `aria-label` per control, since `Field` injects a single id into every control inside it.
- Plans week grid keeps native drag and drop, on desktop only (as before).

## 2026-11-19: Bugs found in phase 5
Fixed: diet targets ignored the typed g/lb ratio on Apply; gym step nav unreachable on mobile; grocery strike-through never visible; media `bg-primary` buttons; devseed stored grams as meal-item `Amount` (it's servings, so "18000g Salmon"); the Button/Card/ListRow link overflow; grids with only `lg:grid-cols-*` overflowed at 390 (an implicit `auto` column grows to min-content, so add `grid-cols-1`). Not fixed: `useExerciseLoggingSession.ts` assigns `group.logged` on a readonly query proxy (Vue warns on every set log; fix with optimistic set logging).

## 2026-11-19: Phase 6: no global element CSS, no palette, no suppressions
App.vue's `<style>` is gone. `html`/`body`/`:focus-visible` base rules live in `style.css` `@layer base`; the status-bar cover is utilities (`h-safe-top`); the desktop content column is an `app-content` utility applied with `lg:` (it used to be unlayered CSS that beat every utility). Tailwind's preflight is the only reset, so every `!` workaround and the `box-border` link fix were deleted. `--color-*: initial` removes the palette, which made a "restrict legacy tokens" lint rule pointless: they no longer exist.
Raw `<button>/<input>/<select>/<textarea>` are lint errors outside `shared/ui/`. The 12 remaining uses are deliberate (sr-only radios, MacroSlider's range input, heatmap cells, slider ticks, the tab bar, the date header's text button and hidden date input, the mobile header's module switcher) and each has an inline `eslint-disable-next-line … -- <why>`. Inline disables, not the suppressions file: the reason sits next to the code, and the file stays empty, so any count in it means a regression.
Alternative: make primitives for the tab bar item and date header. Rejected for now: one user each.

## 2026-11-19: Owner review fixes after phase 6
- **Primary button is blue, not inverted neutral.** The `bg-fg` fill read as glaring white in dark and heavy black in light. New tokens `primary`/`on-primary` (white on a blue that is a step darker than `accent` in dark, so it passes 4.5:1). Supersedes "Neutral primary" above.
- **IconButton `secondary` is borderless**: no border or fill, background only on hover (owner's call; ghost but with a full-strength `fg` glyph).
- **SegmentedControl**: the selected segment uses `rounded-control-inner` (outer radius minus the 2px padding, so no track shows in the corners), a new `thumb` fill (lighter than the track in dark, white in light) and a tight `shadow-thumb`; the track gets a `line` border.
- **Select** takes an optional `group` per option and renders `<optgroup>`s. The gym plan switcher groups plans by program again (regression from M5).
- **Card `fill`**: flex column whose body takes the remaining height. diet/log uses it for the bento layout again at `lg` (meal card spans two rows with a scrolling item list; Add foods and Saved meals take half each). The grid height is `100dvh - 8rem`, min 36rem. The route sets meta `wide`, which adds `lg:app-content-wide` (content column 80rem, new token `spacing-content-wide`) so the three panels aren't squeezed into 56rem.
- **NumberSlider**: the centre line sits after the markers in the DOM (same z) so a lined-up marker hides under it, as `z-5` under `z-10` did before phase 5.

## 2026-11-20: /diet redesigned as week strip + "left today" (ui-page)
Brief: log what I ate, then see what's left; phone and desk equally. Owner picked a mix of two sketches: the /gym-style week strip (kcal vs target per day, Sun–Sat, same week start as /gym) on top of a summary card led by kcal left, with macros as "g left", water and fiber, and one-tap water from drink presets. Planned and Logged stay separate sections.
- Planned meals show a check and "Logged" when a day log has the same `meal_id` (logging a planned meal creates exactly that). `PlannedMeal.logged` isn't used: deleting the log never resets it.
- One-tap **Log** on a planned meal only shows for today: `POST /diet/meals/meal/log-planned` always logs to today's day. Other days keep "Edit and log" in the row menu.
- Meals are expandable rows (`pages/diet/home/MealRow.vue`), not full cards with every food. `MealCard`/`MealCardMacroDetails` deleted; the item list moved to `MealItemsList.vue`.
- Over target: warn colour, an icon and the word "over" only past 5% over (`isMeaningfullyOver`), so landing 20 kcal over doesn't look like a failure.
- Log meal (primary) and Quick log sit at the bottom on mobile and under the summary on desktop. The title bar now carries planning links (Plan meals, Targets, New saved meal); Log water moved onto the page. The Macros / Water & fiber SegmentedControl is gone, since everything fits in one card.

## 2026-11-20: /diet/edit-planned as a week board (ui-page)
Brief: fill the coming week so each day lands near its kcal and protein target; mostly at a desk, quick tweaks on a phone. The owner picked the week board over a strip plus one day, and over a 14-day agenda.
- One full-width row per day, Sun–Sat (same week as /gym and /diet). Desktop: day + kcal bar + P/C/F against targets | meal chips | Add and ⋯. Mobile: the same card stacked, with Add and ⋯ in the header. The month grid, the date-range form and saved-meal edit/delete are gone. Saved meals are now only added from here, and you edit them through /diet/log.
- Revised after owner feedback the same day: the first version had 7 narrow columns, past days folded behind "Show meals" and the add button sat under the meals, so it moved with every add. Now meals are always visible as chips (tap one to move it or remove it), the actions stay fixed in the header, add never disables itself, and a successful add gets no toast, since the chips and totals already show it.
- Status: "Nothing planned" badge on empty upcoming days, warn "N kcal over" past 105% of the target. The 90–105% band (`ON_TARGET_*`) feeds the header summary.
- Desktop tray: pick a saved meal (the chip turns primary), then each day's add button adds it in one click. Esc or ✕ drops the pick. On mobile, add opens a sheet that shows the day's total after each meal ("→ 1,331").
- Copy a day uses a new endpoint, `POST /diet/meals/planned/copy-day {from_offset, to_offsets}`. It clones each planned meal into its own Meal row and appends it after the target day's meals. Client-side copying via from-saved didn't work because planned meals don't always have a `saved_meal_id`.
