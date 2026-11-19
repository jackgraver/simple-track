# Migration notes (ui-overhaul)

Project-only material moved out of `docs/design/TOKENS.md`: the audit snapshot, the old→new
class mapping and the migration plan. Delete this file when phase 5 is done.

## Audit snapshot (2026-11-19)
- `@theme` defines `--color-{mainBg,firstBg,secondBg,thirdBg,textPrimary,textSecondary,cfRed,mainBgTransparent}`
  as camelCase aliases of kebab vars in `:root`. `--color-border` exists but isn't in `@theme`, so
  code writes `border-(--color-border)`. `--color-cf-red-dark` is unused.
- `.light` overrides exist, but nothing applies them. `App.vue` hard-codes a `.dark` wrapper (for
  PrimeVue's `darkModeSelector`). `html { color: rgb(218,218,218) }` in `App.vue` ignores tokens.
- Usage: legacy tokens about 550 times (`text-textSecondary` 183, `text-textPrimary` 156,
  `bg-secondBg` 81…). Tailwind palette about 500 times (`text-zinc-500` 47, `text-zinc-100` 46,
  `border-zinc-600` 39…). Colour literals 166, arbitrary values 65, `<style>` blocks 23
  (now recorded in `frontend/eslint-suppressions.json`).
- Macro colours live in `pages/diet/macro-colors.css`, scoped to `.diet-module`, as CSS named
  colours (`orange`, `red`, `green`, `blue`). `SimpleMacros.vue` has different hex fallbacks.
- Chart colours are hard-coded hsl in `gym/progression/useProgressionCharts.ts` and
  `gym/trackingLogChart.ts` (two copies of `readChartTheme`).
- z-index: dialog 998, dropdown 1000, toast 9999, nav/status-bar 40, plus ad-hoc 5/10/20/30/50.
- Type: `text-sm` 269, `text-xs` 123, plus `text-[10px]`, `[11px]`, `[0.6rem]`, `[0.65rem]`,
  `[0.7rem]`, `[0.8rem]`, `[0.85rem]`, `[0.9rem]`, `[2rem]`.
- Radius: `rounded-md` 141, `rounded` 67, `rounded-lg` 30, `rounded-[5px]`; dialogs 10px.

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
1. **Token layer (phase 2, one slice, no visual change):** add the `docs/design/TOKENS.md` block **next to** the
   legacy `@theme` aliases, which keep their current values. Add theme bootstrap + `data-theme`,
   remove the `.dark` wrapper and PrimeVue `darkModeSelector`, move macro vars to global tokens
   (delete `macro-colors.css`, make `SimpleMacros`/`MacroBar` use `--color-macro-*`; this unifies
   the fat purple, the one visible change). Delete `tailwind.config.js`. Build + 390/1280
   screenshots of home, diet, gym logging to show nothing else moved.
2. **Primitives (phase 3)** use only new tokens. Storybook shows both themes side by side.
3. **Module slices (phase 4)** migrate files with the mapping table, swap inline markup for
   primitives and prune `frontend/eslint-suppressions.json` (`npm run lint -- --prune-suppressions`). The light theme is checked per module.
4. **Endgame (phase 5):** when the suppressions file has no `no-restricted-classes` entries left, add
   `--color-*: initial;` at the top of `@theme` (Tailwind's palette stops compiling), delete the
   legacy aliases and `:root` kebab vars, and remove the global `button/input/select` CSS in
   `App.vue`. Do that last: today ~170 raw `<button>`s rely on it and undo it with `m-0! p-0!`.

## Decision: duplicate `tailwind.config.js`
**Delete it.** Tailwind v4 only reads a JS config through an explicit `@config` directive, and
there is none. It's dead, and it's misleading too: it defines `border` and an undefined
`--color-gradient` that don't exist in the real theme. `style.css` `@theme` is canonical.
Alternative (keep it and `@config` it) rejected: two sources of truth and the v3 API.

## PrimeVue (removed in F2, 2026-11-19)
PrimeVue is only used for `primevue/chart` in 5 files plus the Aura plugin. It was a bake-off
candidate and lost to Reka UI (DECISIONS.md), so F2 replaces the chart with a chart.js wrapper
(`shared/chart/`, tokens via `theme.ts`) and uninstalls `primevue`, `@primeuix/themes` and the Aura
preset. The `.dark` coupling goes in F1.
