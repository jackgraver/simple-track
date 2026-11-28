---
name: ui-page
description: Shape what a page in simple-track should be before building it - pin down its job, generate 2-3 distinct design directions as quick sketches, let the owner pick or mix, then build. Use when the user wants to design, redesign or rethink a page or screen (e.g. "/ui-page /gym", "what should the media page be", "redo the diet dashboard"), especially when they don't have a clear picture yet.
argument-hint: <route or page, e.g. /gym/logging>
---

# Shape a page, then build it

The owner often knows what a page is *for* but not what it should *look like*. Your job is to
supply the ideas and keep the choices cheap: make the brief explicit, sketch real alternatives, and
write Vue only once a direction is picked. Don't ask them to design. Offer options they can react to.

Target: `$ARGUMENTS` (a route or page name; if empty, ask which page).

## 0. Look before asking (no questions yet)
- Load the `ui` skill (design system: PRINCIPLES, RULES, primitives). Everything below must fit it.
- Find the page: `frontend/src/router/index.ts` → the component → its children and composables.
- List the **data that exists**: queries and API response shapes (`frontend/src/api/<module>/`),
  plus fields the UI ignores today (e.g. `max`, `assigned_days`). Hit the dev API (`./bin/dev api ...`)
  to see real values. Unused data is often the best idea source.
- Note what's weak today: dead space, unclear primary action, missing states, things that read as bugs.
- Check git history for an earlier version (`git log --oneline -- <file>`). Old designs hold ideas too.
- Screenshot the current page only if it helps you (the owner usually has dev open).

## 1. The brief (one round of questions, then write it down)
Ask **one** `AskUserQuestion` round (max 4 questions). Pre-fill options from what you found, so
the owner picks rather than invents. Mark your recommendation. Cover:
1. **Main job**: the one thing someone comes here to do or learn. Offer 3-4 candidates.
2. **Situation**: phone mid-task, desk planning/reviewing, or both. This sets density and layout (PRINCIPLES §0, §5).
3. **Add**: multi-select of ideas the page could show, from unused data and common patterns
   (streaks, comparisons with last time, what's next, trends, warnings). Say which ones need mocking.
4. **Keep or drop**: anything current the owner may be attached to.

Then write the brief in chat, ~6 lines:
```
Page: /gym/logging
Job: see what's next and what's done in today's session, start the next thing in one tap
Situation: phone, mid-workout, one hand
Primary action: start the next step
Also shows: progress, last-time numbers, gains so far
Out of scope: plan editing (lives in /gym/plans)
States: loading, error, empty (no plan), in progress, all done, past/future day
```
If the owner corrects it, fix it before going on. Everything after this is judged against it.

## 2. Two or three directions
Make them **structurally different**: a different idea of what leads the page, not the same layout
in different colours. Useful axes: what sits at the top (a hero card, a list, a calendar, a single
number), one screen vs sections, glanceable vs detailed, how the primary action is reached.

For each direction, give:
- **Name and one-line idea** ("Up-next hero: the page is one big next step, the list is secondary")
- **Mobile sketch** (390px) as an ASCII wireframe using real labels and real-looking numbers from
  the dev data, not lorem. ~25 lines max. One line on how desktop (≥ lg) differs.
- **Data**: what's real today vs what needs a mock or a new endpoint.
- **Trade-off**: what it's good at and what it gives up.

Present them with `AskUserQuestion` using `preview` (side-by-side wireframes), plus an option to
mix ("Other": e.g. "A's header with C's list"). If the owner wants to see it rendered, build a
throwaway HTML sketch with tokens instead of more ASCII, but keep it quick.

## 3. Converge and build
- Restate the final pick as the brief plus a short element list, top to bottom, mobile first.
- Build it with the `ui` skill's workflow: primitives, tokens, every state from the brief, mobile + desktop.
- Put shared logic (derived models, formatters) in a `domain/` file with vitest tests, not in templates.
- **Mocks**: put mock data in one clearly named file (`mockX.ts`), headed with a comment on what each
  item needs to become real. Mark mocked UI with a `Badge` saying "Preview". Never mix mock and real
  numbers in one widget without saying so.
- Prefer real data where it's cheap. Check whether the API already sends it before mocking.
- Verify (type-check, lint, tests). Take screenshots only where they help you check layout.
- Finish with: what changed, what's real vs mocked, choices the owner may want to flip
  (e.g. week start day), and data or backend issues you noticed along the way.

## Tips
- Name the page's job in its UI: an "Up next" overline, a "This week" header. Structure should
  explain itself.
- One accent and one primary button per screen. If two things compete for the top, the brief is
  unclear: go back to step 1.
- When the owner says "I don't know", give a recommendation and why, then move on.
