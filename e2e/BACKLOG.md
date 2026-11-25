# e2e backlog (temporary)

Scratch list of bugs and follow-ups the B-scenarios surfaced, to address once the e2e suite is better
established. Delete items as they are fixed or turned into TASKS.md entries; delete this file when empty.
Details and reasoning live in the Findings section of [TASKS.md](TASKS.md).

## App bugs (found by specs, not fixed)

- [ ] **Gym reps slider races its smooth scroll.** Looks fixed in `NumberSlider` (sync ends on reaching the
      target / `scrollend`); 390/390 flake runs passed. Drop this and the `scrollSettled` wait in
      `tests/gym/logger.ts` after a few more green runs.
- [ ] **Diet: viewing /diet creates day rows for future days.** `DietWeekStrip` loads all 7 days and
      `GET /diet/logs/today` creates a missing day with the plan in force now, so opening /diet on a Thursday
      pins Fri and Sat to today's plan (`prefetchDays.ts` skips future days for this reason). Fix: don't fetch
      future days in the strip, or don't create rows for future dates on GET; then drop the `test.fail` in
      `tests/diet/home-day.spec.ts`.
- [ ] **Diet: editing a logged meal leaves the old `meals` row behind.** `EditLoggedMeal` (and "Edit and log")
      repoint `day_logs` at a new meal and never delete the previous one, so orphan meals and items pile up.
      Fix: soft-delete the old meal in the same transaction unless a planned or saved meal references it;
      then assert the count in `tests/diet/logmeal-edit.spec.ts`.
- [ ] **Diet: picking a saved meal keeps the default name** ("Lunch"), because `setMeal` keeps any non-empty
      name. Fix: adopt the saved meal's name while the name is still the untouched default; then drop the
      explicit rename in `tests/diet/logmeal.spec.ts`.
- [ ] **Diet: "Save goal" on water sizes toasts twice.** The goal input saves on `change` (blur) and the button
      saves again. Fix: one path; then tighten the toast checks in `tests/diet/water-presets.spec.ts`.
- [ ] **Diet: `log-planned` drops `client_id`.** `useLogPlannedMeal` makes one but `api.logPlannedMeal` doesn't
      send it; idempotency rests on the server's (day, meal) check. Send it or remove it.
- [ ] **Gym: unplanned cardio can't be saved.** `/gym/logging/cardio` opens without planned cardio, but the
      client sends no `type` and the server answers 400 "cardio type is required when the plan has no planned
      cardio"; the optimistic "done" rolls back. Only reachable by URL today. Fix: default the type, or don't
      offer the route without a plan; then drop the `test.fail` in `tests/gym/cardio-mobility.spec.ts`.
- [ ] **Gym: switching a day's plan returns 400 with the raw DB error.** `controller.SwitchPlan` maps every
      error to 400 and puts `err.Error()` in the body, so the toast shows a Postgres message instead of
      "Couldn’t switch the plan", and 5xx monitoring misses it. Fix: 404 for a missing plan, 500 otherwise;
      then drop the `test.fail` and `allowedErrors` in `tests/gym/home.spec.ts`.
- [ ] **Gym: cardio minutes input on the routine page is 0px wide** (desktop and mobile). The `NumberField`
      in a `w-36` box leaves no room for the input, so only +/- work. Fix: widen the box or drop the
      steppers/unit; then drop the `test.fail` in `tests/gym/plans.spec.ts` and type the value instead.
- [ ] **Gym: weight import error toast shows twice.** `ImportWeightDialog` toasts the error and the global
      `MutationCache.onError` toasts it again (`useImportWeightCsv` has no `meta: { handled: true }`). Fix
      that (and audit other mutations whose callers toast), then drop the `test.fail` in `weight.spec.ts`.
- [ ] **Gym: progression empty state for "All" reads "No data in the this exercise".** Build the sentence per
      range; then drop the `test.fail` in `progression.spec.ts`.

## Product decisions needed

- [x] **Logout button.** Added as "Sign out" in the account menu (PR #87); the auth spec uses it.
- [ ] **Grocery delete control.** Only the API can delete; the spec arranges via API. Intended?
- [ ] **No edit-transaction** in finances. Intended?
- [ ] **Unplanned cardio/mobility.** The session list has no way to add cardio or mobility that the plan
      doesn't have. Intended?
- [ ] **No way to delete a workout program** (UI or API). Intended?
- [ ] **Activating a program creates today's day with the old program's plan** (`ActivateWorkoutProgram`
      calls `GetOrCreateToday` first). Should today switch to the new program?
- [ ] **Cardio minutes stepper:** from empty, the first "+" gives 0, not 5. Intended?
- [ ] **Weight has no delete** (UI or API). Intended?
- [ ] **The steps form doesn't prefill the day's saved value** (weight does). Intended?

- [ ] **Diet: deleting a logged planned meal doesn't return it to Planned** (`planned_meals.logged` stays
      true), so it disappears from both lists. Intended? (optimistic-ui Q1)
- [ ] **Diet: saved meals can't be edited or deleted from the UI** (`?type=editsaved` and the delete endpoint
      exist, nothing links to them). Intended?
- [ ] **Diet: past days stay editable on Plan meals** (add, remove, reorder). Pinned by "a past day can still
      be edited" in `tests/diet/plan.spec.ts`. Should they be read-only?
- [ ] **Diet: every targets save adds a plan row**, even several saves on one day (History piles up). Should a
      same-day save replace today's snapshot?
- [ ] **Diet: one-tap Log for planned meals is today only.** Intended for past days too?
- [ ] **Diet: water goal and unit live in localStorage**, so another device starts at 64 oz. Intended?

## Test-suite follow-ups

- [ ] Extend `aria-busy` on pending rows (grocery and gym sets have it via `ListRow`) to diet and finances, so
      those specs can stop using `waitForApi` before reloading.
- [ ] `Checkbox.vue`: the slot replaces the `label` prop, so the accessible name is "Milk", not
      "Complete Milk". Fix the component or accept it.
- [ ] Diet meal item quantity `spinbutton` (`NumberField` in `LogMealItemRow`) has no accessible name; only its
      wrapping group does. Name the input.
- [ ] Media fakes for TMDB, OpenLibrary and iTunes (only AniList is faked); cover images are not faked
      (hermetic routing blocks them).
- [ ] Uncovered scenarios: settings profile form, register, rate-limit lockout; grocery paste dialog,
      suggestions and offline complete; gym and diet "Not yet" lists under B2/B3 in TASKS.md.
- [ ] Rollback-with-Retry is only covered by the grocery-add example; add it for finances (gym and diet have it now).
- [ ] Cleanup bug seen once: "missing template DB" errors from an `e2e-baseline-f8` run; a rerun passed.
- [ ] Weight/steps charts have no accessible name or text summary (`Chart` gets no `label`), so specs and
      screen readers can't read the history. Add a label or a short "recent entries" list.
