# gym home lane notes

Spec: `tests/gym/home.spec.ts` (all `@gym`; the fresh-day test is also `@smoke`). No new API helpers
were needed (`api.logSet` covers the arrangement); no app code changed.

## Scenarios

Tier 1, Today card:
1. done: fresh day: "Today · Thursday, Jan 15", heading Legs, "4 lifts", Start workout; Show exercises
   lists the 4 lifts in plan order, Back Squat "(up next)". `@smoke`.
2. done: in progress: a Back Squat set via API -> Continue workout, "In progress", "1 of 4", Back Squat
   "(done) 100 lb × 8", Romanian Deadlift up next.
3. done: complete: all 4 lifts logged -> Review workout, "All done", "4 of 4".
4. done: Start workout -> `/gym/logging`, h1 "Legs", all four lifts shown.

Tier 2, plan switching:
5. done: Change -> Push: toast "Workout plan updated", heading Push, "4 lifts · swapped from Legs";
   same after reload; the Thursday tile still reads "Legs, planned".
6. done: Make it a rest day: toast "Marked as a rest day", "Rest day" + "Next up: Push, Monday", same
   after reload; Train anyway -> Legs restores it (after reload too).
7. done: data loss: Back Squat set via API, Push, then "Legs (scheduled)" -> set still shown as done,
   "1 of 4", exactly one `logged_sets` row.
8. done: switch failure: fault `workout_logs` / `update` (times 1, the `UpdateWorkoutPlanID` statement)
   -> error toast, still Legs, still Legs after reload. Pins two oddities, see Bugs.

Tier 3, week strip + navigation:
9. done: Monday set via API -> tile aria-labels: Sunday/Tuesday/Friday/Saturday "rest day", Monday
   "Push, trained", Wednesday "Pull, missed", Thursday "Legs, planned" (+ `aria-current=date`); header
   "1 of 3 sessions · Push Pull Legs". No second aria snapshot: the label assertions already pin each
   tile, and `tests/aria` has the epoch snapshot.
10. done (two tests): sets on Wed 14, Mon 12, Thu 8, Wed 7 -> "4 in a row" (rest days in between
    don't break it; missed Mon 5 ends it). Without Mon 12 the streak is 1 and the badge is hidden
    (it only shows above 1), so "a missed day breaks it" is asserted as "no badge".
11. done: Monday tile -> `?offset=3`, "Monday · Jan 12", Push, "Log this workout"; survives reload.
    Friday -> `?offset=-1`, "Tomorrow · Friday, Jan 16", "Rest day", "Next up: Push, Monday". Next week
    -> Monday (`?offset=-4`) -> Push + "View plan".
12. done: Previous week -> `?offset=7`, heading "Jan 4 – Jan 10" (and Thursday Jan 8 "Legs, missed");
    Next week -> "This week".

Tier 4, errors:
13. done (two tests). Queries retry twice (`shouldRetryQuery`, 3 attempts with 1s + 2s backoff), so
    each fault has `times: 3` and Retry's 4th attempt succeeds. The error state waits up to 10s (the
    backoff runs on the browser clock, which flows naturally after `clock.install`).
    - week: fault `logged_sets` / `row`: the activity query (`DatesWithLoggedSets`) is the only raw
      statement on `logged_sets`. A `workout_programs` fault was avoided on purpose: `GetPlanByDay`
      also queries the active program when today's `workout_logs` row is created, so the hit could
      land on the day request instead. "Couldn't load your week" alert, the Today card still renders,
      Retry recovers (Monday "trained").
    - day: fault `workout_logs` / `query` (`LoadByDate`): "Couldn't load this day's workout", the week
      strip still renders, Retry recovers (heading Legs).

Adapted to the real UI: dates render as "Jan 15" (no period) under the suite's `en-CA` locale; the
step list items start with the step number ("1Back Squat"), except done steps, which show a check.

## Bugs found

- **SwitchPlan maps every error to 400.** `controller.SwitchPlan` returns `http.StatusBadRequest` for any
  service error, DB failures included, so a server fault looks like a client error (anything keyed on 5xx, such as
  monitoring or the e2e 5xx guard, misses it). Repro: fault `workout_logs`/`update`, switch
  plan. Pinned by "switch failure" (it allows the 400 console line). Fix: 404 for "workout plan not
  found", 500 (`apierr.Internal`) otherwise; then drop the `allowedErrors` there.
- **Raw server error text in the switch-failure toast.** `getApiErrorMessage` prefers the response's
  `error` field, and SwitchPlan puts `err.Error()` there, so the user sees the DB error
  ("e2e fault: update workout_logs"; in prod a Postgres message) instead of "Couldn't switch the plan".
  Pinned by the same test (asserts the exact text). Fix together with the status mapping (generic
  message for 5xx), then assert "Couldn’t switch the plan".

## Follow-ups

- **Existing `tests/gym/logging.spec.ts` is red on origin/main** (7 failures, both projects): commit
  ddf872f changed the primary button to "Log Set 1" (no "· N reps" echo), so `setReps` waits for
  `/· 8 reps$/` and the `/^Log set 1/` locators never match. Not this lane's spec; the logging lane or
  the lead should update it. The full `@gym` run: 38 passed, 7 failed (all logging.spec), 1 skipped.
- Not covered: "Other plans…" dialog (`SwitchWorkoutPlanDialog`, select a plan from another program),
  "Log freestyle" on a rest day, switching a past/future day's plan, the switch button's pending
  state (hold-based), streak without a program (7 idle days rule), Yesterday label (offset 1 = Wed).
- The week strip's activity query is shared with the heatmap, so a week fault presumably also puts the
  heatmap in its "Couldn't load activity" state (not asserted); Retry on the week refetches that one
  cache entry.
