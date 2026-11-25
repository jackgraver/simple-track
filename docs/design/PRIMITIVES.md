# Primitives

> **Status: spec; library chosen (2026-11-19): Reka UI, headless, styled with our Tailwind
> tokens.** Reka supplies behaviour only (focus, ARIA, keyboard, portals, positioning). Every
> primitive below is our own component in `shared/ui/` with our markup and classes, and wraps a
> Reka part only where the table in "Library-backed" says so. All built (phases 3–5, 2026-11-19);
> Storybook (`npm run storybook` in `frontend/`) is the reference for every state.

The minimal shared component set. Location: **`frontend/src/shared/ui/`**, one component per
file, PascalCase, exported from `shared/ui/index.ts`. Pages import from `shared/ui/` only, never
from the library directly. They use tokens only (TOKENS.md), have no `<style>` blocks, and each
has **Storybook stories** for every state in both themes.

Shared conventions:
- `size` is `sm | md | lg`, default `md`. **`md` is touch-sized below `lg` and compact at `lg`
  and up**: `min-h-touch lg:min-h-row` (44 → 40px). `sm` is desktop-dense (32px) and only for
  toolbars and table rows on desktop. Don't use `sm` in mobile-first flows.
- Every interactive primitive has `disabled`, a visible `focus-visible` ring (`ring-2 ring-accent
  ring-offset-2 ring-offset-canvas`), and `hover:` only as an enhancement.
- Tones are always `neutral | accent | success | warn | danger | info`, mapped to tokens.
- Native elements first (`<button>`, `<input>`, `<select>`) where Reka doesn't add real value.
  Focus traps, popovers and comboboxes come from Reka (see "Library-backed").
- Only `shared/ui/` imports `reka-ui` (`no-restricted-imports`). Pages never see it.
- There is no global element CSS (removed in phase 6): Tailwind's preflight is the only reset, so
  a primitive's classes are all it gets. Never add `!` classes.

## Priority order

| # | Primitive | Phase | Why first |
|---|---|---|---|
| 1 | Button, IconButton, Spinner | 3a | 173 raw `<button>`s, five competing styles |
| 2 | Field + Input (+ Textarea) | 3a | 93 raw `<input>`s in 32 files, three inconsistent styles |
| 3 | Card, Badge | 3a | Most-repeated container and chip markup |
| 4 | EmptyState, ErrorState, Skeleton | 3a | States are mandatory (RULES.md). Pages can't comply without them |
| 5 | Dialog / Sheet (re-skin `DialogContainer`) | 3b | 19 call sites, hard-coded rgb/hex, not a sheet on mobile |
| 6 | Toast (re-skin `ToastContainer`) | 3b | 162 calls, Bootstrap hexes, optimistic-ui needs an Undo action |
| 7 | ListRow (+ List) | 3b | Every module's main view is a list |
| 8 | Select, Toggle, Checkbox | 3b | 11 files with raw `<select>`, 10 with raw checkboxes |
| 9 | SegmentedControl, PageHeader/Section, Chart | 3c | Smaller repetition. Chart moves chart theming onto tokens |
| 10 | Combobox (from `Dropdown.vue`), Menu | 3c | Only a few users; do when the owning module migrates |

## Library-backed (Reka UI 2.x) vs built by us

| Primitive | Built on | Notes |
|---|---|---|
| `Overlay` (Dialog/Sheet) | Reka `Dialog*` | Focus trap and restore, Esc, scroll lock, `aria-*`. The sheet presentation, drag handle and swipe-to-close are ours (no candidate had swipe; pointer handler or vaul-vue) |
| `Toast` | Reka `Toast*` | Provider/viewport/swipe-dismiss/Undo action. The `toast.push` API stays ours |
| `Toggle` | Reka `Switch*` | |
| `Checkbox` | Reka `Checkbox*` | Indeterminate support |
| `SegmentedControl` | Reka `ToggleGroup*` (single) | |
| `NumberField` (number with ± steppers) | Reka `NumberField*` | Gym plate counts, log-meal quantities. Labelled steppers come for free |
| `Combobox`, `Menu` | Reka `Combobox*`, `DropdownMenu*` | Popover at every width (no Sheet presentation yet) |
| `Select` | **native `<select>`** | The OS picker wins on mobile. Reka `Select` only if a module needs custom option rendering |
| `Button`, `IconButton`, `Spinner`, `Field`, `Input`, `Textarea`, `Card`, `Badge`, `List`/`ListRow`, `EmptyState`, `ErrorState`, `Skeleton`, `PageHeader`, `Section` | **ours** | Plain Vue + Tailwind tokens |
| `Chart` | **ours** | chart.js directly (PrimeVue chart is dropped) |

---

## Button
`<Button variant size :loading :disabled block :to type>`. Label in the default slot. Icons go
in `#icon-left` / `#icon-right` (lucide, `size-4`).
| Variant | Look | Use |
|---|---|---|
| `primary` | `bg-primary text-on-primary` (blue), hover a step dimmer | The one main action per view/dialog |
| `secondary` | `bg-surface-2 border border-line text-fg`, hover `bg-surface-3` | Everything else |
| `ghost` | transparent, `text-fg-muted`, hover `bg-surface-3 text-fg` | Toolbar, inline, "Cancel" |
| `danger` | `bg-danger-soft text-danger` | Destructive action in a list or form |
| `danger-solid` | `bg-danger text-on-solid` | Confirm button of a destructive confirm dialog only |

- States: hover, active (a colour step such as `active:bg-surface-3`, no scale transforms),
  focus-visible, disabled (`opacity-50 cursor-not-allowed`), **loading** (spinner replaces the
  left icon, the label stays so the width doesn't jump, `aria-busy`, clicks ignored).
- `type="button"` by default (not submit). `:to` renders a `RouterLink` with button styling.
- `block` = full width. **Mobile primary CTA** = `size="lg" block` at the bottom of the content.
  On desktop the same button sits inline and right-aligned.
- Replaces: `ui.btnPrimary/btnSecondary/btnDanger` (`pages/finances/format.ts`, 8 finance files);
  `.delete-button`/`.confirm-button` globals (`App.vue`); `.confirm-btn`/`.cancel-btn`
  (`composables/dialog/DialogContainer.vue`); `.finish-button bg-green-600`
  (`gym/logging/components/ExerciseListView.vue`); `rgb(40,80,40)` submit (`auth/signin/index.vue`);
  `bg-primary` buttons (`media/index.vue`, `media/[id].vue`), which are **broken today**
  because `primary` is not a token; the Back button in `LoggingPageShell.vue`.

## IconButton
`<IconButton :label variant="ghost|secondary|danger" size>` with an icon in the slot. `label` is
**required** (it becomes `aria-label` and `title`). The visible glyph is `size-5`, and the hit area
is `size-touch` below `lg` and `size-8`/`size-row` above.
Replaces: `rounded p-1 text-textSecondary hover:bg-secondBg hover:text-(--color-cf-red)` (3×),
the edit-planned month chevrons (`border-zinc-600 p-1.5`), `.close-btn` in `DialogContainer.vue`,
`.remove-button` in `ExerciseListView.vue`.

## Spinner
`<Spinner size="sm|md">`: lucide `Loader2` + `animate-spin text-fg-muted`, `role="status"`
with sr-only text. Replaces the ad-hoc `Loader animate-spin` in `Dropdown.vue` and elsewhere.

## Field + Input + Textarea
- `<Field :label :hint :error :required v-slot="{ id, describedBy }">` does label, hint and error
  layout (`gap-1.5`; label `text-xs font-medium text-fg-muted`; error `text-xs text-danger` with
  an icon; `aria-describedby`, `aria-invalid`).
- `<Input v-model :type :inputmode :placeholder :size :invalid>` with `#prefix` / `#suffix` slots
  for units ("kg", "reps", "$"). `bg-surface-2 border border-line-strong rounded-control`, focus
  `border-accent ring-1 ring-accent`. Font `text-base lg:text-sm` (no iOS zoom). Numbers use
  `inputmode="decimal"`, `tabular-nums` and select-all on focus.
- `<Textarea>` has the same skin with auto-grow.
- States: default, focus, invalid, disabled, read-only.
- Replaces: `shared/input/Input.vue` (evolve it: it has hard-coded rgb, a scoped style, and an
  error box styled like an alert); `ui.input`/`ui.label` (`finances/format.ts`); the input
  styles in `CardioLoggingView.vue`, `CreateFoodDialog.vue`, `grocery/index.vue`,
  `auth/signin/index.vue`; and the global `input, select {}` rule in `App.vue`.

## Select
`<Select v-model :options="[{label, value}]" :placeholder :size>` wraps a **native `<select>`**
in the Input skin with a chevron icon. Native means the OS picker on mobile, which is better
than a custom one there. Replaces raw `<select>` in 11 files (`gym/plans/*`,
`finances/components/*Dialog.vue`, `media/*`, `diet/water-presets`, `CreateFoodDialog.vue`).

## Toggle / Checkbox
- `<Toggle v-model :label>` is a switch (`role="switch"`, `aria-checked`), track `bg-surface-3`
  → `bg-accent`, 44px touch row on mobile. Use it for settings-like booleans that apply at once.
- `<Checkbox v-model :label :indeterminate>` is for selection or completion in lists. Grocery
  check-off uses it with strike-through in the row.
- Replaces raw checkboxes in `grocery/*`, `media/*`, `finances/account.vue`,
  `finances/components/{AccountForm,ContributionType}Dialog.vue`, `diet/edit-planned`,
  `LogMealItemRow.vue`, `MobilityLoggingView.vue`.

## SegmentedControl
`<SegmentedControl v-model :options>` for 2–5 mutually exclusive options. Container `bg-surface-2`,
selected `bg-surface text-fg shadow-raised`, `role="radiogroup"`. Replaces the media status picker
(`media/[id].vue` `form.status === s ? 'bg-primary' : 'bg-secondBg'`) and similar tab-ish button
rows in diet.

## Card
`<Card :padding="none|sm|md" :interactive :as :to>` with optional `#header` (title + actions row)
and `#footer`. `bg-surface border border-line rounded-card`. Padding is `p-3`/`p-4`, or `p-4 lg:p-5`
for md. `interactive` adds hover `bg-surface-3`, focus ring, and the link/button semantics.
Replaces: `ui.card` (`finances/format.ts`); `rounded-lg border border-(--color-border) bg-firstBg p-4`
(several); `rounded-lg border border-zinc-700 bg-zinc-900/40` (`diet/edit-planned`);
the `MealCard.vue` scoped styles (152 lines); `RoomCard.vue`.
Rule: no cards inside cards. Inside a card, use `List` dividers or a `Section` heading.
`class` lands on the outer element; the content sits in an inner padded div, so layout classes
(`flex flex-col gap-4`) go on your own wrapper inside the Card.

## ListRow (+ List)
`<List :inset>` is a vertical group with `divide-y divide-line`, optionally inside a Card.
`<ListRow :to :selected :pending :disabled @click>` has slots `#leading` (icon/avatar, 20–32px),
default (title, truncates), `#subtitle` (`text-xs text-fg-muted`), `#trailing` (value, Badge or
chevron).
- Height: `min-h-touch` on mobile, `min-h-row` at lg. Padding `px-3 py-2`/`px-4`.
- `pending` dims the row (`opacity-60`) and adds `aria-busy`. This is the hook for optimistic-ui.
- `#trailing` is inside the row's main element: passive content only (value, Badge, chevron).
- `#end` is a sibling of the main element, visible at every width: per-row actions (an
  `IconButton` or a `Menu` trigger). A clickable row (`to`/`@click`) must never hold a button in
  `#trailing` (nested interactive elements).
- `#actions` (desktop only) shows on hover/focus-within. Never hover-only: pair it with `#end` or a
  Menu for mobile.
- Replaces `.exercise-item` (`ExerciseListView.vue`), `SearchList.vue` item markup, grocery rows,
  finance account/transaction rows, media list rows, `WaterLogsDialog.vue`, `LoggedSetsList.vue`.

## Badge
`<Badge :tone :variant="soft|outline" size="sm|md">` gives `rounded-full px-2 py-0.5 text-xs
font-medium`. Soft is `bg-{tone}-soft text-{tone}`, neutral is `bg-surface-3 text-fg-muted`.
Replaces the gym auto-increase chip (`bg-green-950 text-green-400`, `ExerciseSetupScreen.vue`; the
undo button inside it becomes `Badge` + `IconButton` or a `Button size=sm variant=ghost`), finance
"stale" amber chips (`amber-950/50`, `border-amber-700/50`), `rounded-md bg-secondBg px-2 py-0.5 text-xs`.

## Dialog / Sheet (one primitive: `Overlay`)
`<Overlay v-model:open :title :presentation="auto|dialog|sheet" :size="sm|md|lg" :dismissible>`
with slots default (scrolling body) and `#footer` (actions, right-aligned on desktop, stacked
full-width on mobile, primary last/right).
- `auto` (default): **Sheet below `lg`** (bottom, `rounded-t-dialog`, drag handle, `pb-safe`,
  max-height 90dvh, slides up in 250ms), **centred Dialog at `lg` and up** (`rounded-dialog`,
  width sm 24rem / md 32rem / lg 42rem, fade+scale from .98).
- Behaviour: focus trap, focus restored to the trigger, `Esc`/scrim/✕ close (unless
  `dismissible=false`), body scroll lock, `aria-modal`, `aria-labelledby`, `z-overlay`, scrim
  `bg-overlay`.
- **Keep the `dialogManager` API** (`composables/dialog/useDialog.ts`: `confirm()`, `custom()`).
  `DialogContainer.vue` is rewritten to render `Overlay`, so the 10 `custom()` + 9 `confirm()`
  callers don't change. `wide` maps to `size="lg"`. `confirm()` gains `tone: "danger"`, which
  renders the confirm as `danger-solid`.
- Replaces: `DialogContainer.vue` scoped styles (`rgb(26,26,26)`, `#3d3d3d`, z 998, `.confirm-btn`
  green-for-confirm-even-when-deleting).

## Toast
Keep `toast.push(message, type)` (162 call sites) and add an options form
`toast.push({ message, tone, action?: { label, onClick }, duration? })`, which optimistic-ui needs
for Undo. Look: `bg-surface border border-line shadow-overlay rounded-card`, tone icon on the
left (check / alert / info), `text-sm text-fg`. Position bottom-right on desktop, bottom-centre
full-width above the tab bar on mobile (`bottom-above-tabbar`, replacing the `.has-tab-bar
.toast-container` rule in `style.css`). On focus pages (`route.meta.focus`, e.g. gym logging) mobile toasts go to the **top** so the bottom primary button stays clear. Container `aria-live="polite"`, errors `role="alert"`.
Enter 250ms slide+fade, at most 3 visible. Replaces `ToastContainer.vue` (`#28a745`, `#dc3545`,
`#007bff`, `gray`, z 9999).

## EmptyState
`<EmptyState :icon :title :description>` with an `#action` slot (usually one `Button`). It is
centred, `py-10`, icon `size-8 text-fg-subtle`, title `text-sm font-medium text-fg`, description
`text-sm text-fg-muted max-w-xs`. Inside a Card or List it renders without its own border.
Replaces "No balances recorded yet.", "No saved meals.", "No plan", `LoggingPageShell`'s
"Nothing to show.", and `rounded-lg border bg-firstBg p-6/p-8 text-center` blocks.

## ErrorState
`<ErrorState :error :title="Couldn't load …" @retry>` shows an inline `danger` icon, a
human-readable message (`error.message` goes in a `<details>`, never as the headline), and a Retry
`Button`. `compact` is a one-line variant for inside cards. Replaces `<div>Error: {{ error.message }}</div>`
(`LoggingPageShell.vue`) and `text-red-400` error lines (`diet/edit-planned`, finances).

## Skeleton
`<Skeleton shape="line|block|circle" :width :height>` (`bg-surface-3 animate-pulse rounded-control`)
and a `<SkeletonList :rows>` preset that matches ListRow height. Also `useDelayedFlag(isPending, 150)`
so fast loads never flash a skeleton. Replaces the 16 "Loading…" / 2 "Loading..." strings and the
"Set —" placeholder header in `LoggingPageShell.vue`.

## PageHeader / Section
- `<PageHeader :title>` with `#actions` and `#meta` slots: title `text-xl font-semibold
  tracking-tight`, actions right-aligned on desktop. On mobile, actions collapse into a primary
  Button below or an IconButton menu. `inline` keeps a single icon action (a ⋯ menu) beside the
  title at every width. Use alongside `ModuleTitleBar.vue` (module sub-nav), not
  instead of it; decide on a merge when diet/gym migrate.
- `<Section :title>` with `#actions`: `text-base font-semibold`, `gap-3`, `gap-6` between sections.

## Chart
`shared/chart/Chart.vue` (`type`, `data`, `options`) plus `shared/chart/theme.ts`
(`chartTheme()` reads `--color-chart-*`, `fg-muted`, `line`; re-renders on theme change).
Replaces `primevue/chart` in the 5 chart files and the two `readChartTheme()` copies. `shared/chart/Chart.vue` is a thin
chart.js wrapper that registers only line/bar parts (no `chart.js/auto`). `chartTheme()` in
`theme.ts` resolves the tokens to `rgb()` and is reactive: call it inside a `computed` and the
chart re-colours when `data-theme` changes (the bake-off caught a chart that sampled the previous
theme). Callers pass `type`, `data`, `options` and an accessible `label`.

## NumberField, Combobox, Menu
- `<NumberField v-model :label :step :min :max :decimals :unit :size="md|lg">`: Reka NumberField
  with `Decrease/Increase <label>` steppers (44px below lg), `inputmode=decimal`, tabular-nums,
  select-all on focus. Works inside `Field` (id/invalid by injection).
- `<Combobox>`: action-style search (Reka Combobox, our filtering via `ignore-filter`): emits
  `select` and clears; `search` is debounced 300ms; `has-more`/`loading`/`load-more` for paging.
  Replaced `shared/input/Dropdown.vue` (deleted). `shared/SearchList.vue` is built from Input +
  List + ListRow.
- `<Menu :items="[{ key, label, icon?, tone?: 'danger', disabled? }]" align @select="key">` with
  the trigger in the default slot: an overflow action menu (Reka DropdownMenu), 44px items below
  lg. `ModuleTitleBar` overflow uses it.
- `<DropdownMenu :align :side width="sm|md">` is the composable shell under `Menu` (Reka
  DropdownMenu). Trigger: `#trigger` (as-child, pass one focusable element such as an IconButton) or
  `#trigger-content` inside a built-in button sized with `trigger-class` and named by
  `trigger-label`. Content: `MenuAction` (`icon`, `tone`, `to` makes it a link, `@select`),
  `MenuLabel` (section overline, or `variant="header"`), `MenuSeparator`, `MenuRadioGroup`
  (`v-model`, options with icons, a row of tiles; picking one keeps the menu open). The account
  menu (`shared/nav/AccountMenu.vue`) uses all of them.

## Avatar
- `<Avatar :name :src? size="sm|md|lg" :decorative>`: Reka Avatar. The photo when `src` loads,
  otherwise a text-free gradient disc (`avatar-a` → `avatar-b`) with a white highlight and inner
  ring. Decorative (aria-hidden) by default because the name sits beside it; pass
  `:decorative="false"` when it stands alone.

## Patterns without a primitive
- **Radio cards / ratings / day pickers:** a `<fieldset>` of styled `<label>`s around `sr-only`
  native `<input type="radio">` (`has-checked:`, `has-focus-visible:`, `has-disabled:`). One tab
  stop, native arrow keys, and the global input CSS never shows, so no `!` is needed. Used by
  finances `PresetDialog`, media `StarRating`, diet `PlanCalendar`.


## Not primitives (on purpose)
Domain components stay in their module but are **built from** primitives and tokens:
`NumberSlider`, `MacroBar(s)`, `SimpleMacros`, `WorkoutActivityGraph`, `MealCard`, `RoomCard`,
`LoggingHeader`. Promote one to `shared/` only when a second module needs it.
