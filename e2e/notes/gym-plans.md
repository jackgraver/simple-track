# gym lane: plans (/gym/plans and routine pages)

Spec: `tests/gym/plans.spec.ts` (`@gym`), 21 tests per project. No API helpers were added: every
arrangement is the seed itself, except the empty state (DB delete, since no API deletes programs).

## Scenarios

All done unless marked. "Persists" = asserted again after `page.reload()` once the write's response landed.

1. **Programs**
   - Create a program (Program actions > New program), add a routine for Friday from the day picker
     ("New routine for Friday"), Make active. After a reload the new program shows first with the Active
     badge, Push Pull Legs shows "Make active" and still has its routines, DB has exactly one active
     program, and the /gym week strip shows Fri = Upper and Mon/Wed = rest day.
     Adapted: a new program with no routines is never picked by /gym (`useGymWeek` takes the first active
     program *with plans*), so the test schedules a routine before activating.
2. **Routines**: create (New routine > NameDialog; in the library after a reload, "No days yet"),
   rename (Routine actions > Rename; exercises and Thursday kept), delete (Routine actions > Delete
   routine; back on the schedule with "Deleted Legs"; after a reload only Legs is gone, Push/Pull keep
   their days; DB: only Legs has `deleted_at`, all 12 `workout_plan_exercises` and 12 exercises remain).
   Extra: delete + Undo restores the routine with its Thursday and its four exercises in order.
3. **Routine exercises**: add an existing exercise (search, pick; appended), "Create and add" a new one
   (exactly one `exercises` row), remove one (the other three keep their order; DB: 3 links on Legs, 11
   overall, the exercise stays in the library), remove + Undo (back in its old place after a reload),
   edit rep rollover 12 -> 15 via Edit exercise (reopened after a reload).
4. **Scheduling**
   - Day picker: Friday > Pull (row says "also Wed"); Fri and Wed both Pull after a reload; /gym strip
     shows both.
   - Day picker: Monday > Rest day clears the day; Push stays in the library with its 4 exercises.
   - takenBy: lives on the routine page, not in the DayPicker (the picker shows "also Mon" for a
     routine's other days instead). Legs page shows "Monday (Push)"; clicking it asks "Move Monday to
     Legs?"; confirming moves the day (Mon + Thu = Legs on the schedule and in /gym's strip, Push
     unscheduled but not deleted). Cancelling leaves Monday with Push after a reload.
   - Drag (desktop only, `dragTo` works with the HTML5 DnD here): library chip "Legs" onto Saturday
     (assign, Thursday kept); Wednesday card onto Tuesday (moves: assign then unassign). Skipped on
     mobile: the day columns only render at `lg`.
5. **Warmup / stretch / cardio**: add two warmup moves and a stretch, reload, remove one warmup move;
   the other warmup move, the stretch and the lifts are untouched. Cardio type + minutes save on
   focus-out and persist.
6. **Empty / error states**: "No programs yet" (arranged by deleting the program/plan rows in the DB)
   then "Create a program" creates it, and it is Active after a reload (first program). Load error:
   DB fault on `workout_programs` query x3 (first try + 2 auto-retries, backoff skipped with
   `page.clock.runFor`) shows "Couldn't load your programs"; Retry loads the schedule.

- test.fail: "cardio minutes input is wide enough to type in" (bug below).

## Bugs found

- **Cardio minutes input on the routine page is 0px wide** (desktop and mobile). In `[id].vue` the
  `NumberField` sits in a `w-36` box; the two 44px steppers, the gaps, padding and the "min" unit leave
  the `<input role="spinbutton">` with no width, so Playwright reports it "not visible" and a user can't
  tap or type into it, only use +/-. Repro: open /gym/plans/1/days/3 and look at the Cardio row (the box
  only shows "min"). Pinned by `test.fail` in `plans.spec.ts` ("cardio minutes input is wide enough to
  type in"). Fix: widen the box (e.g. `w-44`/auto) or drop the steppers/unit at that size; then remove the
  `test.fail` and let the cardio test type the value instead of stepping.
- Minor: from an empty cardio minutes field the first "+" sets 0 (the minimum), not 5, so it takes four
  taps to reach 15. Possibly intended (reka NumberField behaviour); product call.

## Follow-ups

- **`tests/gym/logging.spec.ts` is broken on origin/main** (not caused by this lane): ddf872f changed the
  logger's primary button from "Log set 1 · 8 reps" to "Log Set 1" (no reps echo, capital S), so
  `setReps` and the `/^Log set 1/` / `/^Save set 1 · 8 reps$/` selectors fail: log, edit and
  dropped-response on both projects plus the mobile pinned-bar test (7 failures; reproduced on a clean
  origin/main checkout before this lane changed anything). Needs the gym-logging lane (or lead) to update
  the selectors and find another way to wait for the slider value.
- Not covered: reorder exercises (Move up/down menu and drag), rename program, program select persistence
  (`?program=`), failure paths for writes (assign/remove/add rollback with Retry toast via faults),
  offline queueing ("Waiting for connection"), "Routine not found" / "Invalid routine link" states,
  "Couldn't load this routine" on the routine page, DayPicker "Open <routine>" footer link.
- Product questions: there is no way to delete a program in the UI or API; `ActivateWorkoutProgram`
  calls `GetOrCreateToday` first, so activating creates today's day row with the old program's plan
  (not asserted here; worth deciding whether today should switch to the new program).
