# UI principles

One page. If a design question isn't answered here, pick the plainest option and log it in
the project's DECISIONS.md. Tokens: [TOKENS.md](TOKENS.md). Components:
[PRIMITIVES.md](PRIMITIVES.md). Rules and checks: [RULES.md](RULES.md).

## 0. What this app is, and how it's used
simple-track is a personal tracker for gym, diet, body tracking, money, media and groceries. It
gets used every day in two very different situations:
- **On a phone, mid-task**: logging a set between lifts, a meal, water, weight. One hand, short
  attention, often in a gym. Speed and big targets matter more than information density.
- **At a desk, planning and reviewing**: planning meals, editing workout plans, checking finances
  and progression charts. Width, overview and keyboard flow matter more than big targets.

The feel: a calm, dark, precise tool, more instrument than dashboard. It should feel fast and
trustworthy, with no decoration for its own sake, and a bit of craft in the details (type,
spacing, a quick transition). The **gym logging flow** (`pages/gym/logging/exercise/`) is the
closest existing screen to the target on mobile. No desktop screen is there yet.

## 1. What we're aiming for
- **Clean and functional first.** Every element earns its place by helping someone log, read
  or decide something. Polish (motion, subtle depth) comes after that.
- **Calm.** Neutral surfaces, one accent, colour only when it means something (state, macro,
  money sign, chart series). No decorative gradients.
- **Fast to use.** The common action is one tap or click away and sits where your thumb or
  eye already is. Show numbers before you explain them.
- **Same thing, same look.** A primary button, a card, a list row, an empty state look the
  same in gym, diet, finances and media. Module identity comes from content, not chrome.

## 2. Density
- Desktop: **compact-comfortable**. 14px body (`text-sm`) is the default in lists, tables and
  forms. Rows are 40px. Use the width for columns or a side panel, not bigger padding.
- Mobile: **comfortable**. 16px body for primary content, every control at least 44px tall
  (`min-h-touch`), 16px page gutter.
- Gap scale: 1/2/3/4/6/8 (4–32px). Stick to `gap-2` inside a control group, `gap-4` between
  fields, `gap-6` between sections, `gap-8` between page regions.

## 3. Colour
- Surfaces step up in lightness as they rise: `bg` page, then `surface` card, then `surface-2`
  input or raised element, then `surface-3` hover or pressed. Borders (`line`) separate as much
  as shadows do. In dark mode borders carry the separation.
- **One accent** (`accent`) for focus rings, the selected or active state, progress and links.
  The **primary button is inverted neutral** (`fg` on `bg`), not the accent.
- Semantic colour is reserved for meaning: `success` (done, +progress), `warn` (stale, over
  budget), `danger` (destructive, error), `info` (neutral notices). Red never decorates.
- Domain colour comes from tokens only: macros (`macro-*`) and charts (`chart-1..5`).
- Dark is the default. Light must work for every screen: if a token doesn't flip, it's a bug.

## 4. Motion
- Subtle and quick: 100ms for hover and press, 150ms for state changes, 250ms for enter and
  exit (sheet, dialog, toast). Ease-out going in, ease-in going out.
- Animate opacity and transform only. Never animate layout on data load.
- Respect `prefers-reduced-motion`: transitions drop to ~0 and nothing slides.

## 5. Mobile vs desktop
The breakpoint is **`lg` (64rem)**, the same one the nav uses (tab bar below it, sidebar
above). `sm`/`md` may tweak a component, but layout switches at `lg`.

| | Mobile (< lg) | Desktop (≥ lg) |
|---|---|---|
| Who leads | **Gym logging**, quick logs (water, weight, meal) | Diet planning, finances, media, plans, progression |
| Primary action | Full-width button at the bottom of the content, reachable by thumb | Inline, top-right of the section or row |
| Overlays | **Sheet** from the bottom, drag/swipe or ✕ to close | Centred **Dialog**, Esc to close |
| Lists | Full-width rows, tap the row to open it, actions in a sheet or on swipe | Rows with hover actions; tables where columns help |
| Inputs | ≥16px font (stops iOS zoom), number pad via `inputmode` | Keyboard first: Enter submits, Esc cancels, tab order |
| Chrome | MobileHeader + BottomTabBar, safe areas respected | SideNav, content column ≤ 56rem centred |

Gym logging is the reference mobile flow: one task per screen, a big number, a big primary
button, previous-session context above it. Don't make it denser on desktop. Centre it at a
phone-like width instead.

## 6. States are part of the design
Every view that loads data shows all of these. A primitive exists for each one.

| State | Show | Never |
|---|---|---|
| Loading (first load) | `Skeleton` the shape of the content, after a ~150ms delay | A bare "Loading…" string, layout jump |
| Refetching | Nothing, or a quiet spinner in the header | Blank the existing data |
| Empty | `EmptyState`: one line of what this is + the action that fills it | "No data", an empty card |
| Error | Inline `ErrorState` with a short message + Retry. Toast only for failed *actions* | Raw `error.message` dumps, silent failure |
| Pending action | The button shows a spinner and disables itself. With optimistic-ui, the row renders now and dims (`opacity-60`) until confirmed | Double submits, a frozen UI |
| Offline | Banner "Offline – changes will sync" (once optimistic-ui lands); disable actions that can't queue | Pretending a save succeeded |

## 7. Accessibility basics
- Contrast: body text ≥ 4.5:1, large text and icons ≥ 3:1, in **both** themes.
- Visible focus on everything interactive (`focus-visible` ring in `accent`). Never strip
  `outline` without a replacement.
- Real elements: `<button>` for actions, `<a>`/RouterLink for navigation, `<label for>` on
  every input. Icon-only buttons need `aria-label`. Dialogs trap focus and restore it on close.
- Touch targets ≥ 44×44px on mobile. Don't rely on hover for anything essential (desktop-only
  hover actions need a keyboard or menu equivalent).
- Numbers that update use `tabular-nums`. Status is never shown by colour alone: add an icon or
  a word.

## 8. Do / don't (from this codebase)
| Don't (today) | Do |
|---|---|
| `class="text-zinc-500"`, `border-zinc-700` (500+ raw palette classes, e.g. `diet/edit-planned/index.vue`) | `text-fg-muted`, `border-line` |
| `bg-green-700` primary in `finances/format.ts` `ui.btnPrimary`, undefined `bg-primary` in `media/index.vue` (renders with no background) | `<Button variant="primary">` |
| `.confirm-btn { background: rgb(34,139,34) }` in `DialogContainer.vue`, `#28a745` in `ToastContainer.vue` | `Dialog`/`Toast` primitives using `success`/`danger` tokens |
| Global `button { padding:6px 12px; margin:4px 0; box-shadow }` in `App.vue`, then `m-0! p-0! shadow-none!` to undo it (BottomTabBar) | No global element styling beyond reset; primitives style themselves |
| `text-[10px]`, `text-[0.65rem]`, `text-[rgb(150,150,150)]` (NumberSlider, edit-planned) | `text-2xs`, `text-fg-muted` |
| Macro colours as `orange`/`red`/`green` scoped to `.diet-module`, with *different* fallbacks in `SimpleMacros.vue` | Global `--color-macro-*` tokens, same everywhere |
| `<div>Error: {{ error.message }}</div>` (`LoggingPageShell.vue`) | `<ErrorState :error @retry>` |
| Hard-coded `.dark` wrapper in `App.vue`; `.light` exists but nothing sets it | `data-theme` on `<html>`, set before first paint |
| Chart theme read from CSS vars in two copies (`trackingLogChart.ts`, `useProgressionCharts.ts`) | One `shared/chart/theme.ts` reading `--color-chart-*` |
