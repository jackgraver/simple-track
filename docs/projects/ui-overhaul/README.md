# UI overhaul

**Goal**: one consistent, clean, functional UI across mobile and desktop, with a little beauty. Gym logging is mobile-first; other modules are desktop or split.

**Done when**: every page uses the token set and shared primitives, no raw colors/spacing outside the token layer, every list/form has loading/empty/error states, and `/dev/ui` showcases all primitives.

**Non-goals**: new features, new brand, a component-library migration (decide in DECISIONS first).

## Principles (draft, refine in the audit)
- Clean and functional first; polish second. Dense but readable, one accent colour, subtle motion.
- Touch targets and safe areas are first-class. Desktop is not just a stretched phone layout.
- States are part of the design: loading, empty, error, offline, pending.
- Tokens, not literals. Primitives, not one-off markup.

## Current state (audited 2026-11-19)
- Tailwind v4 with a `@theme` block in `frontend/src/style.css`. Tokens exist for 5 surface/text colours, border and one accent (`cfRed`); dark default, `.light` override. No spacing, radius, type or motion tokens.
- A stale `frontend/tailwind.config.js` duplicates the colour tokens (v4 ignores it unless loaded). Decide which is canonical.
- Token names mix camelCase (`mainBg`) and kebab (`--color-main-bg`) via an alias layer.
- PrimeVue is installed with the Aura preset but only `primevue/chart` and `primevue/config` are used. Aura styling is effectively unused. Decide: keep PrimeVue for charts only, or drop the theme.
- Shared primitives are sparse: `shared/input/{Input,Dropdown}.vue`, `shared/nav/*`, a few list/title components. No Button, Sheet/Dialog, Card, ListRow, Badge, Toggle. Pages likely rebuild these inline.
- 83 `.vue` files, 14.7k lines. Largest: `diet/edit-planned/index.vue` (782), `gym/plans/[id].vue` (526), `gym/plans/index.vue` (525), `finances/account.vue` (461), `media/[id].vue` (432).
- ~93 uses of raw palette colours or arbitrary hex in templates (`bg-red-500`, `[#...]`), ~52 hex literals in total; 23 files with `<style>` blocks.
- Mobile nav (BottomTabBar, MobileHeader) and SideNav exist and are relatively recent.

## Phases
1. **Audit**: screenshots of every page at 390px and 1280px, inconsistency list (spacing, colours, buttons, states). Output to `audit/` + TASKS.
2. **Foundations**: token set (colour, spacing, radius, type, motion), consolidate tailwind config, lint rule banning literals.
3. **Primitives + `/dev/ui` showcase**.
4. **Module migration**, one module per slice (coordinate with optimistic-ui).
5. **Enforcement + cleanup**: remove dead styles, PrimeVue decision, docs.

## Next up / gotchas
- Run the audit (phase 1) first. Nothing else until the inconsistency list exists.
