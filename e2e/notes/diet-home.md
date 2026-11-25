# Diet lane: home (`/diet` day view)

Specs: `tests/diet/home-logging.spec.ts`, `tests/diet/home-day.spec.ts`. Shared helpers:
`tests/diet/home-helpers.ts` (saved-meal fixtures `BOWL`/`BREAKFAST`, `planMeal`, locators). API helpers
in `fixtures/api.ts` under `// diet home`. Planned meals and saved meals are arranged through the API
(saved meal, then `planned/from-saved`); the UI does the logging.

## Scenarios

1. Log a planned meal: done (`home-logging`, "Log moves the meal to Logged..."). Moves to Logged, "Nothing
   left to log", after reload it is logged once and the heading is "Planned" (no "N to go"); DB has one
   `day_logs` row.
2. Double tap with a hold: done (`home-logging`). `dblclick` with `POST /meal/log-planned` held: one request
   is sent, one row after release and reload. In practice the second tap cannot land: the planned row leaves
   the list on the first tap, so the test pins "one request, one row" rather than a de-dupe on the server.
3. Dropped response: done (`home-logging`), adapted. The log-planned request carries **no `client_id`**
   (`useLogPlannedMeal` generates one but `api.logPlannedMeal` never sends it). The retry is de-duplicated
   by (day, meal) on the server (`LogPlannedMeal`, row lock + count). Test: hold `after`, drop, the
   automatic retry, reload: one logged row, one `day_logs` row.
4. Log fails (DB fault): done (`home-logging`). Fault `day_logs`/`create` x2 (first try + the mutation's one
   retry): the row rolls back, "Couldn't log that meal" + Retry toast, planned row and Log button return;
   Retry persists it.
5. Delete a logged catalogue meal: done (`home-logging`). Gone after reload; in the DB the 8 foods, the saved
   meal and its 2 items survive. The planned meal is not brought back (see Follow-ups).
6. Delete a still-saving row: done (`home-logging`). Log held, row shows `aria-busy`, delete confirmed (queued
   behind the log), release, reload: no `day_logs` row left.
7. Totals: done (`home-logging`). Logging two meals moves "kcal left" (2,500 to 2,205 to 1,956) and the
   protein/carbs/fat "g left" lines; deleting one restores them, checked after reload too.
8. Day navigation: done (`home-day`). Quick log on yesterday via the week strip: after reload it is on
   `?offset=1`, not today (also checked through the API). The planned "Log" button is absent on
   `?offset=1` and present today.
9. Week strip: done (`home-day`). Tile labels carry kcal ("500 of 2,500 kcal", "nothing logged"), the
   summary reads "avg 600 / 2,500 kcal · 2 days logged", Previous/Next week page (title "Jan 4 – Jan 10",
   `?offset=7`, back to "This week"); an empty week says "Nothing logged this week".
10. Water: done (`home-day`). Preset button logs, reload, the entries dialog lists it, delete from the dialog
    persists (reload, "No water logged").
11. Empty states: done (`home-day`). "Nothing planned", "Nothing logged yet", "Add drink sizes" (and it goes
    to `/diet/water-presets`), "Nothing left to log" after everything is logged.
12. Load error: done (`home-day`). The fault is `days`/`query` x21 = 7 days x 3 attempts: the strip loads
    all 7 days, so every attempt of every day fails and the fault is used up exactly. "Couldn't load this
    day" (and the strip's "Couldn't load this week") appear; the day's Retry recovers, then the strip's.
- Extra: `test.fail` "viewing the week creates no rows for future days" (bug 1).

## Bugs found

1. **The week strip creates diet day rows for future days.** `GET /diet/logs/today?offset=N` creates the
   day row with the plan in force now (`findOrCreateDietDayForCalendarDate`), and `DietWeekStrip` reads all
   7 days, so opening `/diet` on a Thursday creates Fri and Sat. `prefetchDays.ts` deliberately skips
   future days for this reason ("an early fetch would pin tomorrow's plan before the user changes their
   program"). Repro: open `/diet`, `SELECT date FROM days` has Jan 16 and 17. Pinned by `test.fail` in
   `home-day.spec.ts` ("viewing the week creates no rows for future days"). Fix: the strip should not
   fetch future days (show "–"), or the GET should not create rows for future dates; then drop the
   `test.fail`.

## Follow-ups

- Deleting a logged planned meal does not put it back in Planned (`planned_meals.logged` stays true; the
  server and the optimistic patch agree). After "log then delete" the day shows "Nothing left to log"
  /"Nothing planned" and the meal is gone from both lists. Documented in the repo as an open owner
  question (optimistic-ui Q1); the specs only assert the Logged side. Product decision.
- `log-planned` does not send `client_id` although the frontend generates one; idempotency rests on
  (day, meal). Harmless today, but a dead field.
- Not covered: "Edit and log" on a planned row, editing a logged catalogue meal (logmeal lane), the custom
  water dialog (covered by the smoke spec), water `ml` display unit, a preset logged while its request is
  held, month-planned-summary.
- Two Log buttons (header and bottom bar) exist in the DOM for Quick log / Log meal; one is `display:none`
  per breakpoint, so role selectors work, but the two layouts are only covered by the mobile project.
