# Design tokens

Source of truth: **`frontend/src/style.css`** (Tailwind v4 `@theme`). Nothing else defines
colours, radii, shadows or type sizes. This file is the spec.

> **Status: landed in `frontend/src/style.css` (F1); the only colours since phase 6.** The legacy
> aliases (`bg-firstBg`, `text-textSecondary`…) are deleted and `--color-*: initial` removes
> Tailwind's palette, so `zinc-*`/`white` classes don't compile. Values below include the AA fixes from B1:
> status colours are darker in light and `danger` is lighter in dark so text on `*-soft` reaches
> 4.5:1 (checked by computation against `canvas`, `surface-2` and the soft tint).

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
| `fg-inverse` | text on `fg`- or `danger`-filled surfaces | `hsl(0 0% 7%)` | `hsl(0 0% 98%)` | (new) |
| `on-solid` | text on solid semantic/accent fill | `hsl(0 0% 100%)` | `hsl(0 0% 100%)` | `text-white` on coloured buttons |
| `primary` / `on-primary` | primary button fill / its text (white, 4.9:1 dark, 5.8:1 light) | `hsl(214 70% 48%)` | `hsl(217 85% 46%)` | (new, replaced `bg-fg` primary) |
| `thumb` | selected segment of a SegmentedControl | `hsl(0 0% 22%)` | `hsl(0 0% 100%)` | (new) |
| `accent` | focus ring, selection, active, links, progress | `hsl(212 92% 62%)` | `hsl(217 85% 46%)` | `focus-visible:ring-sky-500`, `--color-cf-red` as highlight |
| `accent-soft` | selected bg, accent chip | accent @ 16% | accent @ 12% (all `*-soft` use the same alphas) | (new) |
| `success` / `success-soft` | done, +progress, positive money | `hsl(142 65% 48%)` | `hsl(142 70% 27%)` | `green-*`, `emerald-400`, `#28a745`, `#55db2c`, `rgb(34,139,34)` |
| `warn` / `warn-soft` | stale, over target, caution | `hsl(38 92% 55%)` | `hsl(32 95% 31%)` | `amber-*`, `orange-400` |
| `rest-wait` / `rest-near` | gym rest clock before 2:00 ("not yet" → "close"); past 2:00 it uses `success` | `hsl(24 92% 60%)` / `hsl(50 95% 55%)` | `hsl(22 90% 40%)` / `hsl(42 100% 29%)` | (new) |
| `danger` / `danger-soft` | destructive, error, negative money | `hsl(0 80% 66%)` | `hsl(0 72% 44%)` | `cfRed`, `red-*`, `#ff2b2b`, `#dc3545`, `rgb(60,20,20)` |
| `info` / `info-soft` | neutral notice | `hsl(199 89% 60%)` | `hsl(200 90% 32%)` | `sky-*`, `#007bff` |
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

Contrast: `fg-muted` on `surface` is 8.0 (dark) / 6.2 (light). `fg-subtle` is 4.5 (dark) / 4.0
(light) on `surface`: hints and placeholders only, never content you have to read (so overlines
use `fg-muted`). Known gap: `on-solid` (white) on a solid `danger`/`success`/`warn` fill is below
4.5:1 in dark; P1 decides the solid-button treatment.

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

Patterns: overline = `text-xs font-semibold uppercase tracking-wide text-fg-muted` (used in
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
| `--spacing-content-wide` | 80rem | wide desktop column for multi-panel pages (route meta `wide` → `lg:app-content-wide`) |
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
| `shadow-thumb` | `0 1px 2px hsl(0 0% 0% / .5)` | `0 1px 2px / .12` + 1px hairline ring | SegmentedControl selected segment |
| `shadow-raised` | `0 4px 12px hsl(0 0% 0% / .45)` | `0 2px 8px hsl(0 0% 0% / .08)` | dropdown, popover, sticky action bar |
| `shadow-overlay` | `0 12px 32px hsl(0 0% 0% / .6)` | `0 12px 32px hsl(0 0% 0% / .16)` | dialog, sheet, toast |

### Motion
Durations come from bare utilities: `duration-100` (hover, press), `duration-150` (state
change, default), `duration-250` (enter and exit). Easing: `ease-out` to enter, `ease-in` to
exit (Tailwind defaults). CSS vars `--duration-fast|base|slow` exist for `<Transition>` CSS
inside primitives. A global reduced-motion rule clamps all transitions and animations.
Overlay animations are tokens in `style.css` (`animate-overlay-in/out`, `animate-sheet-in/out`; centred dialogs and their backdrop are deliberately instant: enter 250ms ease-out, exit 150ms ease-in). Reka waits for the exit animation before unmounting, so overlays use keyframes, not transitions.

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
- `pt-safe`, `pb-safe`, `px-safe`: `env(safe-area-inset-*)`. `pb-above-tabbar`: bottom padding that
  clears the tab bar + inset (replaces `pb-[4.5rem]`, `pb-[calc(5.5rem_+_env(...))]`).
- `bottom-above-tabbar`: `bottom` offset (not `bottom-tabbar`: Tailwind resolves that to the plain spacing token, 4rem, and ignores the utility) for fixed things above the tab bar (toasts, sticky actions).
- Layout breakpoint is `lg` (64rem) only. Replace the `@media (max-width: 767px)` blocks
  (`LoggingPageShell`, `PlannedMealsDisplay` matchMedia) with `lg` when touched.

## `style.css` shape
The block below is the spec; the file differs in two ways: `@theme static` (so every variable is
emitted even when no utility uses it, which JS such as the chart theme reads). Values in the table above win over the ones in this block.

```css
@import "tailwindcss";

/* Tokens only. Values flip per theme under [data-theme="light"]. */
@theme {
    --color-*: initial; /* no Tailwind palette */
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
    --color-primary: hsl(214 70% 48%);
    --color-on-primary: hsl(0 0% 100%);
    --color-thumb: hsl(0 0% 22%);
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
    --spacing-content-wide: 80rem;

    --radius-control: 0.375rem;
    --radius-control-inner: 0.25rem;
    --radius-card: 0.5rem;
    --radius-dialog: 0.75rem;

    --shadow-raised: 0 4px 12px hsl(0 0% 0% / 0.45);
    --shadow-thumb: 0 1px 2px hsl(0 0% 0% / 0.5);
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
    --color-primary: hsl(217 85% 46%);
    --color-thumb: hsl(0 0% 100%);
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
    --shadow-thumb: 0 1px 2px hsl(0 0% 0% / 0.12), 0 0 0 1px hsl(0 0% 0% / 0.06);
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
@utility pb-above-tabbar { padding-bottom: calc(var(--spacing-tabbar) + 0.5rem + env(safe-area-inset-bottom)); }
@utility bottom-above-tabbar { bottom: calc(var(--spacing-tabbar) + 1rem + env(safe-area-inset-bottom)); }

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
