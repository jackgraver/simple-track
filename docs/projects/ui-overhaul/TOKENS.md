# Design tokens

Source of truth: **`frontend/src/style.css`** (Tailwind v4 `@theme`). Nothing else defines
colours, radii, shadows or type sizes. This file is the spec. The phase-2 slice "token layer"
pastes the block below into `style.css`.

## Current state (2026-11-19)
- `@theme` defines `--color-{mainBg,firstBg,secondBg,thirdBg,textPrimary,textSecondary,cfRed,mainBgTransparent}`
  as camelCase aliases of kebab vars in `:root`. `--color-border` exists but isn't in `@theme`, so
  code writes `border-(--color-border)`. `--color-cf-red-dark` is unused.
- `.light` overrides exist, but nothing applies them. `App.vue` hard-codes a `.dark` wrapper (for
  PrimeVue's `darkModeSelector`). `html { color: rgb(218,218,218) }` in `App.vue` ignores tokens.
- Usage: legacy tokens about 550 times (`text-textSecondary` 183, `text-textPrimary` 156,
  `bg-secondBg` 81…). Tailwind palette about 500 times (`text-zinc-500` 47, `text-zinc-100` 46,
  `border-zinc-600` 39…). Colour literals 166, arbitrary values 65, `<style>` blocks 23
  (snapshot in `frontend/src/ui-guard.baseline.json`).
- Macro colours live in `pages/diet/macro-colors.css`, scoped to `.diet-module`, as CSS named
  colours (`orange`, `red`, `green`, `blue`). `SimpleMacros.vue` has different hex fallbacks.
- Chart colours are hard-coded hsl in `gym/progression/useProgressionCharts.ts` and
  `gym/trackingLogChart.ts` (two copies of `readChartTheme`).
- z-index: dialog 998, dropdown 1000, toast 9999, nav/status-bar 40, plus ad-hoc 5/10/20/30/50.
- Type: `text-sm` 269, `text-xs` 123, plus `text-[10px]`, `[11px]`, `[0.6rem]`, `[0.65rem]`,
  `[0.7rem]`, `[0.8rem]`, `[0.85rem]`, `[0.9rem]`, `[2rem]`.
- Radius: `rounded-md` 141, `rounded` 67, `rounded-lg` 30, `rounded-[5px]`; dialogs 10px.

## Naming
kebab-case and semantic (role, not value). `bg-canvas`, `text-fg-muted`, `border-line`. No
camelCase, no hue names outside the domain sets (`macro-*`, `chart-*`). Themes swap **values**,
never class names. Don't use Tailwind's `dark:` variant: the tokens flip on their own.

## The token set

### Colour
| Token (utility suffix) | Role | Dark | Light | Replaces |
|---|---|---|---|---|
| `canvas` | page background | `hsl(0 0% 5%)` | `hsl(0 0% 97%)` | `mainBg`, `bg-zinc-950` |
| `surface` | card, panel, sidebar, dialog | `hsl(0 0% 9%)` | `hsl(0 0% 100%)` | `firstBg` (as card), `bg-zinc-900`, `rgb(26/27,…)` |
| `surface-2` | input, raised or secondary button | `hsl(0 0% 13%)` | `hsl(0 0% 95%)` | `secondBg`, `bg-zinc-800`, `#1b1b1b` |
| `surface-3` | hover / pressed / selected row | `hsl(0 0% 18%)` | `hsl(0 0% 91%)` | `thirdBg`, `bg-zinc-700`, `hover:bg-zinc-800` |
| `overlay` | scrim behind dialog/sheet | `hsl(0 0% 0% / 0.6)` | `hsl(0 0% 0% / 0.3)` | `mainBgTransparent`, `rgba(0,0,0,.5)` |
| `line` | default border/divider | `hsl(0 0% 20%)` | `hsl(0 0% 87%)` | `--color-border`, `border-zinc-700/800`, `#383838`, `#3d3d3d` |
| `line-strong` | input border, emphasised divider | `hsl(0 0% 28%)` | `hsl(0 0% 78%)` | `border-zinc-600/500` |
| `fg` | primary text, icons | `hsl(0 0% 95%)` | `hsl(0 0% 9%)` | `textPrimary`, `text-zinc-100/200`, `text-white` |
| `fg-muted` | secondary text, labels | `hsl(0 0% 68%)` | `hsl(0 0% 38%)` | `textSecondary`, `text-zinc-300/400`, `rgb(150,150,150)` |
| `fg-subtle` | placeholder, hint, disabled, overline | `hsl(0 0% 50%)` | `hsl(0 0% 50%)` | `text-zinc-500/600`, `rgb(120,120,120)` |
| `fg-inverse` | text on `fg`-filled (primary) button | `hsl(0 0% 7%)` | `hsl(0 0% 98%)` | (new) |
| `on-solid` | text on solid semantic/accent fill | `hsl(0 0% 100%)` | `hsl(0 0% 100%)` | `text-white` on coloured buttons |
| `accent` | focus ring, selection, active, links, progress | `hsl(212 92% 62%)` | `hsl(217 85% 50%)` | `focus-visible:ring-sky-500`, `--color-cf-red` as highlight |
| `accent-soft` | selected bg, accent chip | accent @ 16% | accent @ 12% | (new) |
| `success` / `success-soft` | done, +progress, positive money | `hsl(142 65% 48%)` | `hsl(142 70% 33%)` | `green-*`, `emerald-400`, `#28a745`, `#55db2c`, `rgb(34,139,34)` |
| `warn` / `warn-soft` | stale, over target, caution | `hsl(38 92% 55%)` | `hsl(32 95% 40%)` | `amber-*`, `orange-400` |
| `danger` / `danger-soft` | destructive, error, negative money | `hsl(0 80% 60%)` | `hsl(0 72% 48%)` | `cfRed`, `red-*`, `#ff2b2b`, `#dc3545`, `rgb(60,20,20)` |
| `info` / `info-soft` | neutral notice | `hsl(199 89% 60%)` | `hsl(200 90% 38%)` | `sky-*`, `#007bff` |
| `macro-calories` | | `#fb923c` | `#ea580c` | `orange`, `--macro-calories` |
| `macro-protein` | | `#60a5fa` | `#2563eb` | `#60a5fa` |
| `macro-carbs` | | `#f87171` | `#dc2626` | `red`, `#f87171` |
| `macro-fat` | | `#c084fc` | `#9333ea` | `#a855f7` / `#c084fc` (conflicting) |
| `macro-fiber` | | `#34d399` | `#059669` | `green`, `#34d399` |
| `macro-water` | | `#38bdf8` | `#0284c7` | `blue` |
| `chart-1..5` | series in order | blue, green, orange, violet, pink (below) | darker variants | hsl in `useProgressionCharts.ts` |
| `chart-grid`, `chart-text` | axes | = `line`, `fg-muted` | | `readChartTheme()` fallbacks |

`*-soft` = `color-mix(in oklab, var(--color-X) 16%, transparent)`: tinted backgrounds for
badges, banners and the gym "+2.5" chip. Text on a soft background uses the solid token
(`bg-success-soft text-success`).

Contrast check: `fg-muted` on `surface` is above 4.5:1 in both themes. `fg-subtle` is for
hints and placeholders only (about 3.5:1 in dark), never for content you have to read. Check
again with a contrast tool in the token slice and adjust lightness, not hue.

### Type
Font: Plus Jakarta Sans (already loaded in `index.html`). Weights: **400, 500, 600** only.
| Utility | Size/line | Use |
|---|---|---|
| `text-2xs` (new) | 11/14px | tab-bar labels, chart ticks, micro meta (replaces `[10px]`, `[11px]`, `[0.6–0.7rem]`) |
| `text-xs` | 12/16 | labels, overlines, badges, secondary meta |
| `text-sm` | 14/20 | **desktop body**, list rows, buttons, inputs ≥ lg |
| `text-base` | 16/24 | **mobile body**, inputs < lg (stops iOS zoom) |
| `text-lg` | 18/28 | section/card title, logging header |
| `text-xl` | 20/28 | page title (`font-semibold tracking-tight`) |
| `text-display` (new) | 32/36 | big logging numbers (replaces `text-[2rem]`, `text-3xl`) |

Patterns: overline = `text-xs font-semibold uppercase tracking-wide text-fg-subtle` (used in
gym logging today, keep it). Numbers = `tabular-nums`.

### Spacing and sizing
Keep Tailwind's 4px scale (`--spacing: 0.25rem`). Allowed steps: 0, 0.5, 1, 1.5, 2, 3, 4, 5, 6,
8, 10, 12, 16. Named sizes:
| Token | Value | Use |
|---|---|---|
| `--spacing-touch` | 2.75rem (44px) | `min-h-touch`, `size-touch`: every mobile control |
| `--spacing-row` | 2.5rem (40px) | desktop list row / control height |
| `--spacing-control-sm` | 2rem (32px) | compact desktop button/input |
| `--spacing-tabbar` | 4rem (64px) | BottomTabBar height (currently `h-16`) |
| `--spacing-sidenav` | 14rem | SideNav width (now `--nav-w` in `.app-content`) |
| `--spacing-content` | 56rem | desktop content column (now `--content-w`) |
| `--spacing-gutter` | 1rem / 2rem at lg | page side padding (`px-4 lg:px-8` today) |

### Radius
| Utility | Value | Use |
|---|---|---|
| `rounded-control` | 0.375rem | button, input, select, chip-like control (`rounded-md` today) |
| `rounded-card` | 0.5rem | card, list group, popover (`rounded-lg`) |
| `rounded-dialog` | 0.75rem | dialog, sheet top corners |
| `rounded-full` | | badge pill, toggle, avatar, progress |
Plain `rounded` (0.25rem) and `rounded-[5px]` go away.

### Elevation
Flat by default. Surfaces separate by colour step and `line`.
| Utility | Dark | Light | Use |
|---|---|---|---|
| `shadow-raised` | `0 4px 12px hsl(0 0% 0% / .45)` | `0 2px 8px hsl(0 0% 0% / .08)` | dropdown, popover, sticky action bar |
| `shadow-overlay` | `0 12px 32px hsl(0 0% 0% / .6)` | `0 12px 32px hsl(0 0% 0% / .16)` | dialog, sheet, toast |

### Motion
Durations come from bare utilities: `duration-100` (hover, press), `duration-150` (state
change, default), `duration-250` (enter and exit). Easing: `ease-out` to enter, `ease-in` to
exit (Tailwind defaults). CSS vars `--duration-fast|base|slow` exist for `<Transition>` CSS
inside primitives. A global reduced-motion rule clamps all transitions and animations.

### z-index (semantic utilities, nothing else)
| Utility | Value | Layer |
|---|---|---|
| `z-raised` | 10 | sticky headers, in-flow overlaps (NumberSlider markers) |
| `z-nav` | 40 | SideNav, BottomTabBar, status-bar cover |
| `z-popover` | 50 | Select/menu dropdowns outside dialogs |
| `z-overlay` | 60 | Dialog, Sheet + scrim |
| `z-toast` | 70 | Toasts (above dialogs) |
A popover inside a dialog stacks inside the dialog's context, so 50 < 60 isn't a problem.

### Safe areas and breakpoint
- `pt-safe`, `pb-safe`, `px-safe`: `env(safe-area-inset-*)`. `pb-tabbar`: bottom padding that
  clears the tab bar + inset (replaces `pb-[4.5rem]`, `pb-[calc(5.5rem_+_env(...))]`).
- `bottom-tabbar`: `bottom` offset for fixed things above the tab bar (toasts, sticky actions).
- Layout breakpoint is `lg` (64rem) only. Replace the `@media (max-width: 767px)` blocks
  (`LoggingPageShell`, `PlannedMealsDisplay` matchMedia) with `lg` when touched.

## Exact `style.css` shape (target)

```css
@import "tailwindcss";

/* Tokens only. Values flip per theme under [data-theme="light"]. */
@theme {
    /* Phase 5 adds `--color-*: initial;` here to delete Tailwind's palette. */
    --font-sans: "Plus Jakarta Sans", ui-sans-serif, system-ui, sans-serif,
        "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";

    --color-canvas: hsl(0 0% 5%);
    --color-surface: hsl(0 0% 9%);
    --color-surface-2: hsl(0 0% 13%);
    --color-surface-3: hsl(0 0% 18%);
    --color-overlay: hsl(0 0% 0% / 0.6);
    --color-line: hsl(0 0% 20%);
    --color-line-strong: hsl(0 0% 28%);
    --color-fg: hsl(0 0% 95%);
    --color-fg-muted: hsl(0 0% 68%);
    --color-fg-subtle: hsl(0 0% 50%);
    --color-fg-inverse: hsl(0 0% 7%);
    --color-on-solid: hsl(0 0% 100%);

    --color-accent: hsl(212 92% 62%);
    --color-accent-soft: color-mix(in oklab, var(--color-accent) 16%, transparent);
    --color-success: hsl(142 65% 48%);
    --color-success-soft: color-mix(in oklab, var(--color-success) 16%, transparent);
    --color-warn: hsl(38 92% 55%);
    --color-warn-soft: color-mix(in oklab, var(--color-warn) 16%, transparent);
    --color-danger: hsl(0 80% 60%);
    --color-danger-soft: color-mix(in oklab, var(--color-danger) 16%, transparent);
    --color-info: hsl(199 89% 60%);
    --color-info-soft: color-mix(in oklab, var(--color-info) 16%, transparent);

    --color-macro-calories: #fb923c;
    --color-macro-protein: #60a5fa;
    --color-macro-carbs: #f87171;
    --color-macro-fat: #c084fc;
    --color-macro-fiber: #34d399;
    --color-macro-water: #38bdf8;

    --color-chart-1: hsl(212 92% 62%);
    --color-chart-2: hsl(142 65% 48%);
    --color-chart-3: hsl(27 96% 61%);
    --color-chart-4: hsl(270 85% 72%);
    --color-chart-5: hsl(330 80% 66%);
    --color-chart-grid: var(--color-line);
    --color-chart-text: var(--color-fg-muted);

    --text-2xs: 0.6875rem;
    --text-2xs--line-height: 0.875rem;
    --text-display: 2rem;
    --text-display--line-height: 2.25rem;

    --spacing-touch: 2.75rem;
    --spacing-row: 2.5rem;
    --spacing-control-sm: 2rem;
    --spacing-tabbar: 4rem;
    --spacing-sidenav: 14rem;
    --spacing-content: 56rem;

    --radius-control: 0.375rem;
    --radius-card: 0.5rem;
    --radius-dialog: 0.75rem;

    --shadow-raised: 0 4px 12px hsl(0 0% 0% / 0.45);
    --shadow-overlay: 0 12px 32px hsl(0 0% 0% / 0.6);
}

:root {
    color-scheme: dark;
    --duration-fast: 100ms;
    --duration-base: 150ms;
    --duration-slow: 250ms;
}

:root[data-theme="light"] {
    color-scheme: light;
    --color-canvas: hsl(0 0% 97%);
    --color-surface: hsl(0 0% 100%);
    --color-surface-2: hsl(0 0% 95%);
    --color-surface-3: hsl(0 0% 91%);
    --color-overlay: hsl(0 0% 0% / 0.3);
    --color-line: hsl(0 0% 87%);
    --color-line-strong: hsl(0 0% 78%);
    --color-fg: hsl(0 0% 9%);
    --color-fg-muted: hsl(0 0% 38%);
    --color-fg-subtle: hsl(0 0% 50%);
    --color-fg-inverse: hsl(0 0% 98%);
    --color-accent: hsl(217 85% 50%);
    --color-accent-soft: color-mix(in oklab, var(--color-accent) 12%, transparent);
    --color-success: hsl(142 70% 33%);
    --color-warn: hsl(32 95% 40%);
    --color-danger: hsl(0 72% 48%);
    --color-info: hsl(200 90% 38%);
    --color-macro-calories: #ea580c;
    --color-macro-protein: #2563eb;
    --color-macro-carbs: #dc2626;
    --color-macro-fat: #9333ea;
    --color-macro-fiber: #059669;
    --color-macro-water: #0284c7;
    --color-chart-1: hsl(217 85% 50%);
    --color-chart-2: hsl(142 70% 33%);
    --color-chart-3: hsl(24 90% 48%);
    --color-chart-4: hsl(270 70% 55%);
    --color-chart-5: hsl(330 70% 50%);
    --shadow-raised: 0 2px 8px hsl(0 0% 0% / 0.08);
    --shadow-overlay: 0 12px 32px hsl(0 0% 0% / 0.16);
}

@utility z-raised { z-index: 10; }
@utility z-nav { z-index: 40; }
@utility z-popover { z-index: 50; }
@utility z-overlay { z-index: 60; }
@utility z-toast { z-index: 70; }

@utility pt-safe { padding-top: env(safe-area-inset-top); }
@utility pb-safe { padding-bottom: env(safe-area-inset-bottom); }
@utility px-safe {
    padding-left: env(safe-area-inset-left);
    padding-right: env(safe-area-inset-right);
}
@utility pb-tabbar { padding-bottom: calc(var(--spacing-tabbar) + 0.5rem + env(safe-area-inset-bottom)); }
@utility bottom-tabbar { bottom: calc(var(--spacing-tabbar) + 1rem + env(safe-area-inset-bottom)); }

@layer base {
    html { background: var(--color-canvas); color: var(--color-fg); }
    :focus-visible { outline: 2px solid var(--color-accent); outline-offset: 2px; }
}

@media (prefers-reduced-motion: reduce) {
    *, *::before, *::after {
        transition-duration: 1ms !important;
        animation-duration: 1ms !important;
        animation-iteration-count: 1 !important;
    }
}
```

Implementation notes for the token slice:
- **`*-soft` in light**: `color-mix(... var(--color-success) ...)` is computed on `<html>`, the
  same element the light block overrides, so the soft tokens follow the theme with no extra
  declarations. Check one in devtools anyway. Never set theme tokens on a descendant
  element: soft and chart aliases would keep the `<html>` value.
- **Theme switch**: a 5-line inline script in `index.html` sets `<html data-theme>` from
  `localStorage["st:theme"]` (`dark` | `light` | `system`, default `dark`; `system` follows
  `prefers-color-scheme`) **before first paint**, and updates `<meta name="theme-color">`.
  A Settings toggle writes it. Remove the `.dark` wrapper in `App.vue`.
- Charts read tokens through one helper, `shared/chart/theme.ts`
  (`getComputedStyle(document.documentElement)`), and re-render when `data-theme` changes
  (MutationObserver on `<html>`).
- Never use `--color-*` from JS except through that helper.

## Legacy mapping (for migrating a file)
Map by **role**, not by value: a card that was `bg-firstBg` becomes `bg-surface`, even though
`firstBg` (15%) is lighter.

| Old | New |
|---|---|
| `bg-mainBg`, `bg-zinc-950` | `bg-canvas` |
| `bg-firstBg` (card/panel), `bg-zinc-900`, `bg-zinc-900/40..60` | `bg-surface` |
| `bg-secondBg`, `bg-zinc-800` | `bg-surface-2` |
| `bg-thirdBg`, `bg-zinc-700`, `hover:bg-zinc-800`, `hover:bg-thirdBg` | `bg-surface-3` / `hover:bg-surface-3` |
| `border-(--color-border)`, `border-secondBg`, `border-zinc-700/800` | `border-line` |
| `border-thirdBg`, `border-zinc-600/500` | `border-line-strong` |
| `text-textPrimary`, `text-zinc-100/200`, `text-white` (on neutral) | `text-fg` |
| `text-textSecondary`, `text-zinc-300/400` | `text-fg-muted` |
| `text-zinc-500/600`, placeholders | `text-fg-subtle` |
| `text-cfRed`, `text-(--color-cf-red)`, `text-red-*`, `bg-red-950/*` | `text-danger`, `bg-danger-soft` |
| `bg-green-*`, `text-green-*`, `text-emerald-400` | `success` / `success-soft` (primary *buttons* → `Button variant="primary"`) |
| `amber-*`, `orange-400` (non-macro) | `warn` / `warn-soft` |
| `sky-*`, `ring-sky-500` | `accent` (focus/selection) or `info` (notice) |
| `rounded` | `rounded-control` |
| `rounded-md` on cards | `rounded-card` |
| `text-[10px]`…`text-[0.7rem]` | `text-2xs` |
| `text-[0.8rem]`…`text-[0.9rem]` | `text-sm` (or `text-xs`) |
| `text-[2rem]` | `text-display` |
| `z-1000`, `z-index: 998/9999` | `z-popover`, `z-overlay`, `z-toast` |

## Migration plan
1. **Token layer (phase 2, one slice, no visual change):** add the block above **next to** the
   legacy `@theme` aliases, which keep their current values. Add theme bootstrap + `data-theme`,
   remove the `.dark` wrapper and PrimeVue `darkModeSelector`, move macro vars to global tokens
   (delete `macro-colors.css`, make `SimpleMacros`/`MacroBar` use `--color-macro-*`; this unifies
   the fat purple, the one visible change). Delete `tailwind.config.js`. Build + 390/1280
   screenshots of home, diet, gym logging to show nothing else moved.
2. **Primitives (phase 3)** use only new tokens. `/dev/ui` shows both themes side by side.
3. **Module slices (phase 4)** migrate files with the mapping table, swap inline markup for
   primitives and lower `ui-guard.baseline.json`. The light theme is checked per module.
4. **Endgame (phase 5):** when the baseline has no `paletteClass` entries left, add
   `--color-*: initial;` at the top of `@theme` (Tailwind's palette stops compiling), delete the
   legacy aliases and `:root` kebab vars, and remove the global `button/input/select` CSS in
   `App.vue`. Do that last: today ~170 raw `<button>`s rely on it and undo it with `m-0! p-0!`.

## Decision: duplicate `tailwind.config.js`
**Delete it.** Tailwind v4 only reads a JS config through an explicit `@config` directive, and
there is none. It's dead, and it's misleading too: it defines `border` and an undefined
`--color-gradient` that don't exist in the real theme. `style.css` `@theme` is canonical.
Alternative (keep it and `@config` it) rejected: two sources of truth and the v3 API.

## Decision: PrimeVue
Actual usage: `primevue/chart` in 5 files (`finances/index.vue`, `finances/account.vue`,
`gym/weight/index.vue`, `gym/steps/index.vue`, `gym/progression/ExerciseProgressionChartPanel.vue`)
plus the plugin with the **Aura** preset in `main.ts`. No other PrimeVue component is used,
so Aura only adds runtime-injected CSS variables and a `.dark` coupling.
**Recommendation: drop PrimeVue.** Add `shared/chart/Chart.vue` (about 50 lines:
`type/data/options` props, lazy `import("chart.js/auto")`, `ResizeObserver`, destroy on
unmount, theme defaults from `shared/chart/theme.ts`), switch the 5 files, then remove `primevue`,
`@primeuix/themes` and the `app.use(PrimeVue)` block. `chart.js` stays.
Alternatives: (a) keep `primevue/chart` and drop only Aura (`unstyled: true`). Smaller change,
but we keep a big dependency for a 40-line wrapper. (b) Adopt PrimeVue components as the
primitive layer. Rejected: we'd fight Aura's look, and the non-goals rule out a library migration.
