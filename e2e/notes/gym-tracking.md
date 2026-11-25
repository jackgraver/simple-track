# gym tracking lane: weight, steps, progression

Specs: `tests/gym/weight.spec.ts`, `tests/gym/steps.spec.ts`, `tests/gym/progression.spec.ts`,
`tests/gym/tracking-load-errors.spec.ts`. All tagged `@gym`. No API helpers added (the existing
`saveWeight` and `logSet` were enough).

## Scenarios

Adapted to the real UI: weight and steps have **no list**, only canvas charts. What the pages expose as
text is the weight slider's value (prefilled with the selected day's log, else the latest one), the
weight metrics card (only filled by an import), and the overwrite prompt, which quotes the saved value
("You already logged 8,420 steps for Thursday January 15th, 2026"). Specs assert through those, plus
the DB for row counts and dates.

- **Weight: save today** (done, weight.spec.ts): pick 180.5 on the slider (prior log at 180), save,
  reload: the slider shows 180.5, the overwrite prompt quotes it, and the DB has the Jan 10 and Jan 15 rows.
- **Weight: import** (done): paste a 3-row Renpho CSV over one existing day: toast "Imported 2 new,
  1 updated", DB rows (weight + body fat, Toronto dates), and `?offset=1` / `?offset=2` show the
  imported weight and the "Body fat" metric.
- **Weight: import rejected** (done): a CSV without `Weight(lb)` shows the server's error, the dialog
  keeps the text, no rows are written.
- **Weight: import error shown once** (`test.fail`): see bugs.
- **Weight: delete** (skipped): there is no delete in the UI or the API (`/tracking/weight` has GET,
  POST and POST /import only).
- **Weight on a past day, Toronto TZ** (done): now = Jan 15 22:30 EST (already Jan 16 in UTC); saving on
  `?offset=1` stores `2026-01-14 00:00` Toronto time; after a reload that day prompts to overwrite;
  today has no log (Save goes straight through).
- **Steps: save + overwrite** (done, steps.spec.ts): save 8420, reload, save "9 000" -> the overwrite
  prompt quotes 8,420 -> Overwrite; after a reload the prompt quotes 9,000; the DB has exactly one row
  (9000, Jan 15).
- **Steps: cancelling the overwrite keeps the value** (done) and **an invalid count is refused** (done).
- **Progression** (done, progression.spec.ts): Back Squat sets on 5 dates via API (plus a 0-rep set,
  which the API excludes); the List view shows day groups newest first with "- Set n: WxR"; 1 mo /
  3 mo / 6 mo / All change the groups. The Graph view is only checked as "no empty-state text" (canvas).
- **Progression empty state** (done): Leg Curl without sets shows "No data in the last month", and
  "...last year" after switching range, in both views. Switching exercise swaps the data (done).
- **Progression "All" empty copy** (`test.fail`): see bugs.
- **Load errors + Retry** (done, tracking-load-errors.spec.ts): a DB fault with `times: 3` (first try +
  the 2 query retries from `shouldRetryQuery`; backoff skipped with `page.clock.runFor`) shows the
  ErrorState, and Retry loads the page. Weight/steps fault `query` on their table; progression faults
  `row` on `logged_exercises` (the query is a gorm `Scan`).

## Bugs found

- **Weight import error toast shows twice.** `ImportWeightDialog` toasts the error itself, and the
  global `MutationCache.onError` (`frontend/src/api/queryClient.ts`) toasts it again because
  `useImportWeightCsv` has no `meta: { handled: true }`. Repro: Import, paste
  `Date,Steps` / `2026.01.14,8000`, Import: two "missing Weight(lb) column" toasts. Pinned by
  weight.spec.ts "the import error is shown once" (`test.fail`). Fix: add `meta: { handled: true }` to
  the import mutation (and check other `useMutation`s whose callers toast errors themselves).
- **Progression empty-state copy for "All".** `rangeLabel` for `all` is "this exercise", so the title
  reads "No data in the this exercise" (and "No chart data in the this exercise"). Pinned by
  progression.spec.ts (`test.fail`). Fix: build the sentence per range (e.g. "No data for this exercise").

## Follow-ups

- **Pre-existing on main:** 7 tests in `tests/gym/logging.spec.ts` fail (both projects) because ddf872f
  changed the primary button to "Log Set 1" (no "· 8 reps" echo, capital S). Not in this lane; the
  logging lane or the lead should update `setReps` and the button regexes.
- Weight and steps have no text history: only canvases, and `Chart` gets no `label`, so the
  `role="img"` has no name. An accessible summary (the `label` prop, or a short "recent entries" list)
  would let specs and screen readers read the history.
- Weight has no delete control or endpoint. Intended?
- The steps form does not prefill the selected day's saved value (weight does). Intended?
- Weight save through the slider relies on a tick click plus waiting for the smooth scroll to settle
  (the same `NumberSlider` race as the reps slider in BACKLOG).
- Not covered: steps on a past day (same offset code path as weight), tab-separated import and the kg
  export error, weight/steps save failure rollback + Retry toast, the progression "(last time)" marker
  (only used where `lastTimeSets` is passed, not on this page).
