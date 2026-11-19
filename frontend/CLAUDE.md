# frontend

Vue 3 + Vite + Tailwind v4. Node tooling via `mise exec -- npm run <script>` (`type-check`,
`lint`, `test:run`, `build`).

## UI rules (full version: docs/design/RULES.md)
Any change to a `.vue` file or UI-facing `.ts` follows the design system in `docs/design/`.
Read `docs/design/PRINCIPLES.md` once and `docs/design/RULES.md` before your first UI edit, or
use the `ui` skill.
- Tokens only: no hex/rgb/hsl, no Tailwind palette classes (`zinc-*`, `white`), no arbitrary
  colour/type/radius/shadow/z values. Theme lives in `src/style.css`.
- Shared components from `src/shared/ui/` only (once they exist). No `<style>` blocks, no `!` overrides.
- Every data view has loading, empty and error states. Every action shows pending and failure.
- Check mobile (390px) and desktop (1280px), dark and light. Layout switches at `lg` only.
- `npm run lint` must pass. Never grow `eslint-suppressions.json`; prune it after cleanups
  (`npm run lint -- --prune-suppressions`).
