# Diet lane: plan (/diet/edit-planned)

Specs: `tests/diet/plan.spec.ts`, `tests/diet/plan-failures.spec.ts`. Shared helpers:
`tests/diet/plan-helpers.ts` (saved meals via API, day/chip locators), API helpers under `// diet plan`
in `fixtures/api.ts` (`createSavedMeal`, `addPlannedFromSaved`, `plannedMeals`).

## Scenarios

1. Add a saved meal to a day, shows on /diet for that day after reload: done (`plan.spec.ts`).
2. Armed add (pick in tray, tap several days, Esc drops the pick): done, desktop only (the tray is `lg`
   only; skipped on the mobile project) (`plan.spec.ts`).
3. Remove a planned meal; Move earlier / later persists order (incl. Move later disabled on the last
   meal): done (`plan.spec.ts`).
4. Copy to other days (Select empty days, Clear, picked days append after existing meals, source
   unchanged) and Copy from previous day (only on an empty day after a planned one): done (`plan.spec.ts`).
5. Clear day (Cancel keeps, confirm removes, other days untouched): done (`plan.spec.ts`).
6. Week paging / This week / URL `?offset=` survives reload / past empty days read "No plan": done
   (`plan.spec.ts`). Adapted: past days are **not** read-only (see Follow-ups); a second test pins that.
7. No saved meals: empty state (overlay on both projects, tray on desktop) and "New saved meal" link
   to `/diet/log?type=...&offset=-1`: done (`plan.spec.ts`).
8. Load error + Retry: done (`plan-failures.spec.ts`). The page runs 7 day queries; fault is `days`
   `query` x21 (7 x 3 attempts).
9. Write failure rolls back: done (`plan-failures.spec.ts`). Add-from-saved retries once (retry: 1), so
   the fault is `planned_meals` `create` x2; chip is optimistic, rolls back with a Retry toast; Retry saves
   exactly one planned meal.

No `test.fail()` needed.

## Bugs found

None.

## Follow-ups

- Product question: past days (before today) keep the Add button and the chip menus, so planning
  history is editable from here. Only the empty label ("No plan") differs. Pinned by "a past day can
  still be edited"; if it should be read-only, flip that test.
- Clear day issues one DELETE per meal (sequential per day), not one request. Fine today; a partial
  failure would leave a half-cleared day (each rolls back separately with its own toast). Not covered.
- Not covered: the saved-meals load error in the tray/overlay (separate query, `saved_meals`), copy
  failure toast ("Couldn't copy that day"), Move/Remove while a chip is still saving (guarded by
  `isTempId`), logged meals surviving Clear day.
- The copy overlay lists the 14 days after the source (never before today); copying from a past day
  therefore starts at today.
- The mobile project has no tray, so armed add is untestable there by design.
