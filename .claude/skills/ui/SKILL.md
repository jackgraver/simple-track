---
name: ui
description: simple-track's design system and UI workflow. Use before building or changing any page, component, layout or styling in frontend/ (Vue templates, Tailwind classes, colours, spacing, dialogs, forms, lists, charts), and when reviewing UI changes or checking them visually.
---

# simple-track UI work

## 1. Load the design system (once per session)
Read these, in order:
1. `docs/design/PRINCIPLES.md`: what the app is, the feel, mobile vs desktop, states, a11y.
2. `docs/design/RULES.md`: the hard rules, the interim note, lint and verification.
3. As needed: `docs/design/TOKENS.md` (any colour/size/radius/z/motion choice) and
   `docs/design/PRIMITIVES.md` (any shared component).

If the work is part of the ui-overhaul project, also follow `docs/projects/README.md`
(session protocol) and read `docs/projects/ui-overhaul/{README,TASKS,DECISIONS}.md`.

## 2. Before writing markup
- Look in `frontend/src/shared/ui/` (and Storybook: `npm run storybook` in `frontend/`) for a component that already
  does the job. Use it. Don't copy its markup.
- Decide which states the view needs: loading, empty, error, pending, offline.
- Decide the mobile (< lg) and desktop (≥ lg) behaviour (PRINCIPLES.md §5): sheet or dialog,
  where the primary action sits, list vs table.

## 3. While writing
- Tokens only, no `<style>`, no `!` overrides, no `dark:` (RULES.md).
- Run `mise exec -- npm run lint` in `frontend/` often. A failure in a file with suppressed
  violations lists all of them: `git diff` to find the one you added.

## 4. Verify (don't skip)
1. `cd frontend && mise exec -- npm run type-check && mise exec -- npm run lint && mise exec -- npx vitest run`
2. Visual check of **only what you changed**: Storybook story, or the page via
   `./bin/dev start` + `./bin/dev env` (never bare `./bin/dev`, never hardcode ports, login
   `dev`/`dev`). Use the Playwright plugin's browser at 390×844 and 1280×800, dark and light.
   Check the console.
3. Check the states you added (RULES.md "How to verify visually").
4. Stop the stack if you started it. Report what you checked and attach screenshots.

## 5. When the rules don't cover something
Pick the plainest option consistent with PRINCIPLES.md, then record it in
`docs/projects/ui-overhaul/DECISIONS.md` while the overhaul is running. If it's a new token,
add it to TOKENS.md and `style.css` in the same change.
