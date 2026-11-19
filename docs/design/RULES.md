# UI rules for agents

Apply these to **any** change that touches `frontend/src/**/*.vue` or UI-facing `.ts`, in any
project, not only ui-overhaul. Background: [PRINCIPLES.md](PRINCIPLES.md),
[TOKENS.md](TOKENS.md), [PRIMITIVES.md](PRIMITIVES.md).

## Rules
1. **Tokens only.** No hex/rgb/hsl literals, no Tailwind palette classes (`zinc-*`, `green-*`,
   `white`…), no arbitrary colour, size, radius or z values (`text-[10px]`, `bg-[#1b1b1b]`,
   `z-1000`) outside `src/style.css`. Use the TOKENS.md utilities. If a token is missing,
   add it to `style.css` and TOKENS.md in the same PR and log it in DECISIONS.md. Don't inline it.
2. **Primitives only.** Use `shared/ui/*` for buttons, inputs, selects, toggles, cards, rows,
   badges, overlays, toasts and states. If a primitive doesn't exist yet, either build it
   (to PRIMITIVES.md spec, in `/dev/ui`) or write the markup with tokens and leave a
   `<!-- ui: replace with <X> -->` comment. Don't add another local copy of an existing primitive.
3. **No new `<style>` blocks** and no new global CSS. Tailwind classes only. Animations go
   through `<Transition>` classes or tokens in `style.css`.
4. **No `!` overrides** (`m-0!`, `shadow-none!`) in new code. If you need one, the global
   `App.vue` element CSS is in the way, so use a primitive.
5. **States required.** Every query-backed view renders loading (`Skeleton`), empty
   (`EmptyState`), error (`ErrorState` + retry) and success. Every mutation trigger shows pending
   (`Button :loading` or row `pending`) and reports failure with a toast.
6. **Mobile and desktop both work.** Check below and above `lg`. Mobile: nothing overflows
   horizontally at 360px, controls ≥ 44px, primary action reachable, safe areas respected, inputs
   ≥ 16px. Desktop: content uses the column, keyboard works (Tab, Enter, Esc).
7. **Both themes work.** Never use `dark:`. If something looks wrong in light, the fix is a
   token, not a per-component override.
8. **A11y floor.** Icon-only buttons have `label`/`aria-label`, inputs have labels, focus is
   visible, colour is never the only signal.
9. **Lower the ratchet.** If your change removes violations, lower
   `frontend/src/ui-guard.baseline.json` (command below). Never raise it.
10. **One module per PR** for migrations. Touch a page once: tokens, primitives, states and the
    optimistic-ui conversion together where that project is ready.

## Mechanical enforcement (in place)
`frontend/src/ui-guard.test.ts` runs in `vitest run`, so it runs locally, in `dev test` and in
the j-cicd deploy pipeline (`frontend-test` step). For every `.vue`/`.ts` file in `src/`
(except tests), it counts:

| Metric | Matches |
|---|---|
| `colorLiteral` | `#abc`, `#1b1b1b`, `rgb(`, `rgba(`, `hsl(`, `hsla(` |
| `paletteClass` | `bg-/text-/border-/ring-/…` + Tailwind palette hue (`zinc-500`, `white`, `amber-700/50`) |
| `arbitraryValue` | `text-[…]`, `bg-[…]`, `border-[…]`, `rounded-[…]`, `shadow-[…]`, `z-[…]`, … |
| `styleBlock` | `<style` |

It **fails only if a file's count goes above its entry in `ui-guard.baseline.json`**. New files
start at 0, so it can't fail on current code. Baseline at creation (2026-11-19): 70 files, 166
colour literals, 503 palette classes, 65 arbitrary values, 23 style blocks.

```sh
cd frontend
mise exec -- npx vitest run src/ui-guard.test.ts                    # check
UI_GUARD_UPDATE=1 mise exec -- npx vitest run src/ui-guard.test.ts  # rewrite baseline after cleanup
git diff src/ui-guard.baseline.json                                 # must only go DOWN
```

Reviewers reject a baseline diff with any increase. Known false positives (a hex in a comment, a
`rgb(` in a chart helper) get fixed by moving the value into a token, not by raising the count.

## Planned enforcement (specified, not yet added: add when the condition holds)
- **Legacy token metric**: once the token layer lands, add `legacyToken`
  (`\b(?:bg|text|border)-(?:mainBg|firstBg|secondBg|thirdBg|textPrimary|textSecondary|cfRed)\b`,
  `--color-border`, `--color-cf-red`) to the guard with a fresh baseline.
- **`!important` metric**: add once primitives replace the global `button` CSS (`\b[\w:-]+!(?=[\s"'])`).
- **Palette removal**: when `paletteClass` is 0 everywhere, add `--color-*: initial;` to `@theme`
  in `style.css`. Raw palette classes then stop compiling, which makes the rule a build fact.
- **ESLint** (`eslint-plugin-vue` + `vue/no-restricted-class`, `vue/no-restricted-syntax` for
  `<style>`, or `eslint-plugin-tailwindcss`): not added. The repo has no ESLint today, and the
  ratchet test covers the same ground without a new toolchain. Revisit if editor-time feedback is wanted.

## How to verify visually
The owner knows the UI. Screenshot only what you changed, not every page.
1. `./bin/dev start` (never bare `./bin/dev`), then `./bin/dev env` for the frontend URL. Never
   hardcode ports. Log in with `dev`/`dev` (or `POST <frontend>/api/auth/login`). The devseed
   data covers gym, diet, tracking, media and money.
2. For each page you changed, check **390×844** (mobile) and **1280×800** (desktop), in **dark and
   light** (`localStorage["st:theme"]="light"` once the token layer lands). Use whatever browser
   tooling the session has (Claude in Chrome, the built-in browser, or an ad-hoc
   `mise exec -- npx playwright screenshot`; don't add it as a dependency).
3. Check the states: empty (a fresh worktree DB or filter to nothing), error (block `/api/*` in DevTools
   request blocking, or `./bin/dev stop` with the page open), pending (DevTools network throttling). For new
   primitives, `/dev/ui` is the place to check all states and both themes.
4. `mise exec -- npx vue-tsc --noEmit` and `mise exec -- npx vitest run` in `frontend/` must pass.
5. Put screenshots in the PR description or `docs/projects/ui-overhaul/audit/<slice>/` (only for
   before/after of a migrated module). `./bin/dev stop` if you started the stack.
