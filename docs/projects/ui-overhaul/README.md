# UI overhaul

**Goal**: one consistent, clean, functional UI across mobile and desktop, with a little beauty. Gym logging is mobile-first; other modules are desktop or split.

**Done when**: every page uses the token set and shared primitives, no raw colors/spacing outside the token layer, every list/form has loading/empty/error states, and `/dev/ui` showcases all primitives.

**Non-goals**: new features, new brand, a component-library migration (decide in DECISIONS first).

## Charter (read before any UI work)
- [PRINCIPLES.md](PRINCIPLES.md): one page. Density, colour, motion, mobile vs desktop, states, a11y, do/don't.
- [TOKENS.md](TOKENS.md): the token set, the exact `@theme` block, legacy mapping, migration plan, config + PrimeVue decisions.
- [PRIMITIVES.md](PRIMITIVES.md): the `shared/ui/` set with props, variants and states, what each replaces, priority.
- [RULES.md](RULES.md): enforceable rules for **any** agent touching UI, the `ui-guard` ratchet, how to verify visually.

**Status (2026-11-19)**: charter drafted (phase 1 done, pending owner sign-off on the open
questions below). The `frontend/src/ui-guard.test.ts` ratchet is live: it fails CI when a file
gains colour literals, palette classes, arbitrary values or `<style>` blocks. No app restyling yet.

## Current state (audited 2026-11-19; details in TOKENS.md)
- Tailwind v4 `@theme` in `frontend/src/style.css` has 8 camelCase colour aliases (`mainBg`, `firstBg`, …, `cfRed`) and nothing for spacing, radius, type, motion or z. `--color-border` isn't in `@theme`, hence `border-(--color-border)`.
- `frontend/tailwind.config.js` is dead (v4 needs `@config`). **Decision: delete** in F1.
- PrimeVue: only `primevue/chart` (5 files) + the Aura plugin. **Decision: drop** for `shared/chart` in F2.
- Shared UI: `shared/input/{Input,Dropdown}.vue`, `SearchList`, `ModuleTitleBar`, nav, and the `dialogManager`/`toast` singletons. Finances has its own mini design system (`ui` in `pages/finances/format.ts`, green primary).
- Violations (guard baseline): 70 files, **503 palette classes** (mostly `zinc-*`), 166 colour literals, 65 arbitrary values, 23 `<style>` blocks. Worst: `diet/edit-planned` (51), `diet/water-presets` (36), `QuickLogDialog` (15), `TDEECalculator` (14), `ExerciseSetupScreen` (13).
- 173 raw `<button>`s and 93 raw `<input>`s. Global element CSS in `App.vue` styles them and gets undone with `m-0! p-0! shadow-none!`.
- States: ~16 bare "Loading…" strings, error UIs in only 8 files, empty states ad hoc.
- 83 `.vue` files. Largest: `diet/edit-planned/index.vue` (782), `gym/plans/[id].vue` (526), `gym/plans/index.vue` (525), `finances/account.vue` (461), `media/[id].vue` (432).

## Phases
1. **Charter** (done 2026-11-19): code audit, principles, tokens, primitives, rules, guard. The owner
   knows the UI, so the full screenshot audit was dropped (DECISIONS).
2. **Foundations**: token layer next to legacy, theme bootstrap + light toggle, macro/chart tokens,
   delete `tailwind.config.js`, replace PrimeVue with `shared/chart`.
3. **Primitives + `/dev/ui` showcase**, in PRIMITIVES.md priority order.
4. **Module migration**, one module per slice, paired with optimistic-ui where it's ready.
5. **Enforcement + cleanup**: remove the global element CSS in `App.vue`, add `--color-*: initial`,
   delete legacy aliases, add the legacy-token guard metric, point CLAUDE.md at RULES.md.

## Next up / gotchas
- **Next slice: F1 token layer** (TASKS.md). Get answers to the open questions first, or ship
  with the defaults proposed in TOKENS.md (they're isolated in `style.css` and cheap to change).
- `ui-guard.test.ts` runs in `vitest run` and in j-cicd. If it fails, use a token or primitive.
  After cleanup run `UI_GUARD_UPDATE=1 mise exec -- npx vitest run src/ui-guard.test.ts`; the baseline only goes down.
- The global `button`/`input`/`select` CSS in `App.vue` styles every raw element. Primitives must
  reset it themselves until phase 5 removes it, and removing it early restyles ~170 buttons.
- `bg-primary` (`media/index.vue`, `media/[id].vue`) is not a token, so those buttons render with no
  background today. They get fixed by Button in the media slice.
- `.light` in `style.css` is never applied, and `App.vue` hard-codes a `.dark` wrapper for PrimeVue.
- Macro colours only exist inside `.diet-module` (`pages/diet/macro-colors.css`). Outside diet, `SimpleMacros` falls back to different hexes.

## Open questions for the owner
1. **Accent hue.** Proposed: blue `hsl(212 92% 62%)` for focus, selection and progress only. The
   alternatives are keeping the current red (`cfRed`, which clashes with `danger`), teal, or violet.
2. **Primary button = inverted neutral** (light button on dark, dark on light), not accent-filled
   and not green as in finances today. OK?
3. **Surfaces get darker**: cards go to 9% (today `firstBg` is 15%, `secondBg` 25%), matching the
   newer zinc-based gym and diet screens. OK, or keep the lighter greys?
4. **Light theme**: a Settings toggle with Dark / Light / System, default Dark. Is that enough?
5. **Mobile overlays become bottom sheets** everywhere (`Overlay presentation="auto"`), with dialogs on desktop. OK?
6. **Desktop gym logging** stays a centred phone-width column rather than a denser desktop layout. OK?
7. **Drop PrimeVue** for a ~50-line chart.js wrapper. OK?
