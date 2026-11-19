# UI rules for agents

Apply these to **any** change that touches `frontend/src/**/*.vue` or UI-facing `.ts`, in any
project. Background: [PRINCIPLES.md](PRINCIPLES.md), [TOKENS.md](TOKENS.md),
[PRIMITIVES.md](PRIMITIVES.md). Migration tracking lives in `docs/projects/ui-overhaul/`.

## Rules
1. **Tokens only.** No hex/rgb/hsl literals, no Tailwind palette classes (`zinc-*`, `green-*`,
   `white`…), no arbitrary colour, type, radius, shadow or z values (`text-[10px]`,
   `bg-[#1b1b1b]`, `z-[1000]`) outside `src/style.css`. If a token is missing, add it to
   `style.css` and TOKENS.md in the same PR and log it in the project DECISIONS.md.
2. **Primitives only.** Buttons, inputs, selects, toggles, cards, rows, badges, overlays, toasts
   and states come from `shared/ui/`. Don't add a local copy of something that exists there. If
   it's missing, build it to the PRIMITIVES.md spec with a Storybook story, or use tokens and
   leave `<!-- ui: replace with <X> -->`.
3. **No `<style>` blocks**, no new global CSS. Tailwind classes only.
4. **No `!` overrides** (`m-0!`, `shadow-none!`) in new code. Needing one means the global
   `App.vue` element CSS is in the way: use a primitive.
5. **States required.** Every query-backed view renders loading (Skeleton), empty (EmptyState),
   error (ErrorState + retry) and success. Every mutation trigger shows pending and reports
   failure with a toast.
6. **Mobile and desktop both work.** Layout switches at `lg` only. Mobile: no horizontal
   overflow at 360px, controls ≥ 44px, primary action in thumb reach, safe areas respected,
   inputs ≥ 16px. Desktop: uses the column, keyboard works (Tab, Enter, Esc).
7. **Both themes work.** Never use `dark:`. If light looks wrong, fix the token.
8. **A11y floor.** Icon-only buttons have `aria-label`, inputs have labels, focus is visible,
   colour is never the only signal.
9. **Never grow the suppressions.** Don't run `--suppress-all`/`--suppress-rule` to make new code
   pass. When you clean a file up, prune them (command below).
10. **One module per PR** for migrations: tokens, primitives, states and the optimistic-ui
    conversion together, so each page is touched once.

## Lint (enforced)
ESLint (`frontend/eslint.config.js`) runs in `npm run lint`, `./bin/dev check` and the j-cicd
pipeline (`frontend-lint` step, before tests). It covers:

| Rule | Catches |
|---|---|
| `better-tailwindcss/no-restricted-classes` | palette colours (`text-zinc-500`, `hover:bg-green-600`, `text-white/15`) and arbitrary colour/type/radius/shadow/z values |
| `better-tailwindcss/no-unknown-classes` | classes Tailwind doesn't know: typos, undefined tokens (`bg-primary`), and the custom CSS classes that `<style>` blocks need |
| `better-tailwindcss/no-conflicting-classes`, `no-duplicate-classes`, `no-deprecated-classes` | contradictory, repeated or v3-era classes |
| `vue/no-restricted-block` (`style`) | `<style>` blocks |
| `no-restricted-syntax`, `vue/no-restricted-syntax`, `vue/no-restricted-static-attribute` | hex/rgb/hsl in TS strings, template expressions and `style="…"` |

Existing violations (849 at setup, 2026-11-19) are recorded in `frontend/eslint-suppressions.json`.
ESLint only fails when a file has **more** violations of a rule than recorded. When that
happens it lists every violation of that rule in the file, not just the new one: diff against
`main` to find yours.

```sh
cd frontend
mise exec -- npm run lint                            # check
mise exec -- npm run lint -- --prune-suppressions    # after cleanup: shrink the suppressions file
git diff eslint-suppressions.json                    # counts may only go down or disappear
```

## Planned rules (add when the condition holds)
| When | Add |
|---|---|
| Token layer lands | Restrict legacy tokens: `bg-(mainBg\|firstBg\|secondBg\|thirdBg)`, `text-(textPrimary\|textSecondary\|cfRed)`, `--color-border`, `--color-cf-red` |
| Button/Input primitives land | `vue/no-restricted-html-elements` for raw `<button>`, `<input>`, `<select>`, `<textarea>` outside `shared/ui/` |
| Global element CSS removed | Restrict `!` important classes |
| No palette suppressions left | `--color-*: initial;` in `@theme`, so palette classes stop compiling |

## How to verify visually
Screenshot only what you changed. The owner knows the rest of the UI.
1. **Components:** check them in Storybook (`npm run storybook`, port `$STORYBOOK_PORT` or 6006; its MCP endpoint is `/mcp`): every state, both themes, 390 and 1280
   viewports. The story URL needs no backend or login.
2. **Pages:** `./bin/dev start` (never bare `./bin/dev`), then `./bin/dev env` for the frontend URL.
   Never hardcode ports. Log in with `dev`/`dev`. devseed provides data for every module.
3. Use the Playwright plugin (or Claude in Chrome) to open the page at **390×844** and
   **1280×800**, in dark and light, and check the console for errors.
4. States: empty (filter to nothing or use a fresh worktree DB), error (block `/api/*` in
   DevTools, or `./bin/dev stop` with the page open), pending (network throttling).
5. `mise exec -- npm run type-check`, `npm run lint` and `npx vitest run` in `frontend/` must pass.
6. Put before/after screenshots in the PR description. `./bin/dev stop` if you started the stack.
