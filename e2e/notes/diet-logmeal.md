# Diet e2e: logmeal lane (`/diet/log`)

Specs: `tests/diet/logmeal.spec.ts`, `logmeal-edit.spec.ts`, `logmeal-saved-foods.spec.ts`; shared helpers in
`tests/diet/logmeal-helpers.ts`; arrange helpers appended to `fixtures/api.ts` under `// diet logmeal`.
All tagged `@diet`, desktop + mobile (16 tests, 32 runs).

## Scenarios

| # | Scenario | Status | Spec |
|---|---|---|---|
| 1 | Log a catalogue meal (search, 2 foods, amounts, "Log meal") -> /diet, right kcal after reload | done | `logmeal.spec.ts` |
| 1b | Extra: `?offset=2` logs to that day, today stays empty | done | `logmeal.spec.ts` |
| 2 | "Log and save meal" -> logged, and in the Saved meals picker | done | `logmeal.spec.ts` |
| 3 | Start from a saved meal, change an amount, log; saved meal unchanged | done | `logmeal.spec.ts` |
| 4 | Edit a logged catalogue meal ("Update"): one logged meal with the new amounts | done | `logmeal-edit.spec.ts` |
| 5 | Edit that fails midway: no loss, no duplicate, retry works | done, adapted (see below) | `logmeal-edit.spec.ts` (2 faults) |
| 6 | "Edit and log" a planned meal: logged copy, planned template + saved meal unchanged | done, adapted (see below) | `logmeal-edit.spec.ts` |
| 7 | New saved meal (`?type=create`), plus validation toast (nothing / no name / no foods) | done | `logmeal-saved-foods.spec.ts` |
| 8 | Create a food from search, fill the dialog: added to the meal, in the catalogue after reload | done | `logmeal-saved-foods.spec.ts` |
| 9 | Duplicate food name | done as a plain passing test, no `test.fail` (see below) | `logmeal-saved-foods.spec.ts` |
| 10 | Group two items, name the group, log: group survives in the logged meal | done | `logmeal.spec.ts` |
| 11 | Quantity 0 removes the item; Halve/Double amount update kcal | done | `logmeal.spec.ts` |
| 12 | Draft survives reload and is cleared after a log; log failure keeps the draft + error toast, retry logs once | done | `logmeal.spec.ts` |

Where the app differs from the brief:

- **5 (delete half fails).** `POST /diet/meals/meal/editlogged` is no longer create-then-delete: `EditLoggedMeal`
  creates the new meal and repoints the day's `day_logs` row in one DB transaction. So the test arms a fault on each
  statement (`meals` create, `day_logs` update), clicks Update, and asserts: error toast "Update Failed!", still on
  the edit page with the edit intact, the day still has exactly one logged meal with the old amounts, the `meals`
  table did not grow (rolled back), and a retry succeeds with one logged meal. The BACKLOG item "Diet meal edit is
  create-then-delete" looks obsolete for the logged-meal path.
- **6 ("planned row marked logged").** A logged planned meal leaves the Planned list (the API filters
  `logged = true`): the Planned region says "Nothing left to log" and the meal shows under Logged. The test asserts
  that, then reads the DB: `planned_meals.logged` is true and its own meal still has the planned amounts. The planned
  template is untouched because "Edit and log" logs a new copy (`logedited`) and flags the planned row.
- **9 (duplicate food = 500).** `PostFood` already maps `ErrFoodNameTaken` to **409** with "a food with this name
  already exists" (the BACKLOG note saying 500 is out of date). The dialog shows that as a toast, stays open and the
  Create button works again. The test pins this (and the retry with another name), so no `test.fail` was needed. The
  browser logs the 409 as a console error, hence `allowedErrors: [/409/]`. There is no inline field error, only the
  toast; whether that is enough is a product question.

## Bugs found

None that fail a test. Observations that are arguably bugs, not pinned (no stable "intended" behaviour to assert):

1. **Editing a logged meal leaves the old `meals` row behind.** `EditLoggedMeal` repoints `day_logs` at a new meal
   but never deletes the previous one (seen in `db`: 2 `meals` rows named "Dinner", 1 `day_logs`). Same for "Edit and
   log" copies. Harmless on screen, but orphan meals and meal_items accumulate with every edit. Fix: soft-delete the
   old meal in the same transaction (unless a planned meal or saved meal references it).
2. **Picking a saved meal keeps the time-of-day default name.** The page fills "Lunch" on load, and `setMeal` keeps
   any non-empty name, so tapping "Chicken and rice" yields a meal called "Lunch" with the saved meal's items. The
   spec renames the meal explicitly. Probably intended to adopt the saved meal's name when the name is still the
   untouched default.

## Follow-ups

Not covered (candidates for a later pass):

- Composite foods: group menu "Save as composite food", then adding a "Recipe" from the picker (expands to a named
  group with `composite_food_id`).
- Food variants: "Swap to <variant>" in the item menu.
- `?type=editsaved` (no UI entry point; product question already in the brief) and the missing-id error
  (`/diet/log?type=edit` without `id`).
- Select mode: Ungroup, group of fewer than two items, removing a whole group ("Remove group").
- Edit logged meal on another day (`?offset=` + `dayId`) and a stale `id` (meal already deleted).
- "Today after logging" macro bars/"kcal left" figures (the numbers animate, so a snapshot mid-count is unreliable;
  the spec only asserts the meal total).
- Load error on `/diet/log?type=edit&id=...` (ErrorState "Couldn't load this meal" + Retry) and on the two pickers
  ("Couldn't load the list").
- Double submit: the footer button is `loading` while submitting; a held response + second click was not tried.
- Draft for `?type=create` (separate sessionStorage key) is not covered.

Test-suite and a11y notes:

- The quantity input of a meal item (`NumberField`) is an unnamed `spinbutton`; only its wrapping `group` carries
  "<food> quantity". Specs use `getByRole('group', { name }).getByRole('spinbutton')`. Naming the input would let
  specs (and screen readers) use `getByRole('spinbutton', { name })`.
- On mobile the page header has its own "Log meal" button and `h1`, so `getByRole('button', { name: 'Log meal' })`
  matches two elements; the helpers scope to the meal card (`article`).
- Toasts are mirrored into a hidden live region; a modal dialog also hides the Notifications region from the a11y
  tree. `toast()` in `logmeal-helpers.ts` matches the visible text instead (string = exact, regex must be anchored).
  Worth a shared helper if more lanes need it.
- Summary kcal text has no space ("546kcal" in `textContent`), so specs match `/546\s*kcal/`.
- BACKLOG: drop "`POST /diet/foods` with a duplicate name returns 500" and rework the "edit is create-then-delete"
  item (see above); update TASKS B3.
