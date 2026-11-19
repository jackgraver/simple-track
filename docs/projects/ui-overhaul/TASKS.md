# UI overhaul tasks

## Phase 1: audit
- [ ] Screenshot every route at 390px and 1280px (dev stack, `dev`/`dev` login) into `docs/projects/ui-overhaul/audit/`
- [ ] Inventory inconsistencies per page (buttons, inputs, cards, spacing, empty/loading/error states)
- [ ] Inventory repeated inline patterns that should become primitives
- [ ] Write principles page (1 page max) and agree it
- [ ] Decide: PrimeVue (charts only?) and canonical Tailwind config location

## Phase 2: foundations
- [ ] Token set: spacing, radius, type scale, motion, elevation, semantic colours (success/warn/danger)
- [ ] Remove duplicate `tailwind.config.js` or make it canonical
- [ ] Lint rule: no hex / raw palette colours in `.vue`

## Phase 3: primitives
- [ ] Button, IconButton, Input (extend existing), Select, Toggle, Card, ListRow, Badge, Sheet/Dialog, EmptyState, Skeleton, Toast style
- [ ] `/dev/ui` showcase route (dev only) rendering each primitive and its states

## Phase 4: module migration (pair with optimistic-ui slice where possible)
- [ ] gym/logging (mobile priority)
- [ ] gym/plans, gym/progression
- [ ] diet
- [ ] tracking (water, weight, steps)
- [ ] finances
- [ ] media
- [ ] grocery
- [ ] home, settings, auth/signin

## Phase 5: enforcement
- [ ] Delete dead CSS and `<style>` blocks, final audit, update CLAUDE.md rules
