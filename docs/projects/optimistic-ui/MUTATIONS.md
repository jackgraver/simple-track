# Mutation classification (2026-11-19)

Inventory: **42 `useMutation`** (6 diet `api/`, 12 workout, 13 tracking, 9 `pages/diet/logmeal/queries`,
1 `targets`, plus the `useMoneyMutation` factory), **14 `useMoneyMutation` call sites**, and **~20 writes
outside TanStack Query** (gym plans pages, plan dialogs, media). Paths are relative to `frontend/src/`.

**Class**
- **SO (safe-optimistic):** the client can compute the visible result exactly, and the server returns
  enough to reconcile.
- **P (partial):** apply the visible part optimistically (row appears, disappears or moves, or a toggle
  flips) and mark server-derived fields (totals, room, dates) as pending until the response.
- **W (wait):** the result is server-computed, the write has side effects or likely validation failures,
  it's destructive behind a confirm, or it's a form that navigates away. Keep it pessimistic but use the
  shared error/pending convention.

**Value** H/M/L (H = mobile, frequent, mid-task). **Effort** S/M/L, including any backend change.

## Ranked: do these first

| # | Mutation | Class | Cache keys touched | Optimistic computation | Rollback | Reconcile | Backend facts / needed | Value | Effort |
|---|---|---|---|---|---|---|---|---|---|
| 1 | Log set (`useLogExercise`, `api/workout/queries.ts:214`, driven by `addNextSet`/`finishLogging` in `useExerciseLoggingSession.ts`) | SO | `workouts.previous(date)` (group's `logged.sets`), `workouts.day(date)` (same data: merge keys) | Append set `{client_id, weight, reps, weight_setup, created_at: now}` to `group.logged.sets`. Create `group.logged` if absent (first set) | Remove that set by `client_id` (not by value) | Replace the group's `logged` with response `exercise`, matching sets by `client_id` | Returns full `LoggedExercise` (`exercises_route.go:93-101`). **Not idempotent**: full replace, duplicate `logged_exercises` under concurrency (AUDIT.md, reproduced). `updated_at` only, no version. Needs a per-set endpoint keyed by `client_id` + unique `(workout_log_id, exercise_id)`. Activity heatmap: invalidate only when this was the day's first log | **H** | M (BE + FE) |
| 2 | Edit set (same endpoint, `editSet` → `addNextSet`) | SO | as #1 | Patch set fields by id/client_id | Restore the previous fields | Response set | Same endpoint today. Becomes `PATCH /sets/:id` (idempotent) | H | S after #1 |
| 3 | Delete set (`useDeleteLoggedSet` :292) | SO | as #1 | Remove set | Re-insert at the original index | Nothing (keep removed). Invalidate on settle only if the group becomes empty | 404 if already gone (`exercises_route.go:172-189`): treat 404 as success. Pending sets (no server id yet) must queue the delete behind the create (scope, DESIGN.md) | H | S |
| 4 | Log planned meal (`useLogPlannedMeal`, `api/diet/queries.ts:39`) | P | `dietKeys.days.day(date)`; month summary (inv) | Move the planned meal to logged and add its macros to the totals (`meal.items` × food macros are in the cache) | Snapshot restore | SQD full day from response | Returns full day + totals. Mostly idempotent (exists-check, but no unique index on `day_logs(day_id, meal_id)`). **Bug: ignores day, 500 off-today** (`meal_route.go:405`), so the request must carry `day_id` | **H** | S (BE fix + FE) |
| 5 | Unlog/delete logged meal (`useDeleteLoggedMeal` :54) | P | `days.day(date)` | Remove from logged, subtract totals. Decide whether it returns to planned (AUDIT.md) | Snapshot | SQD full day | Returns full day. Not idempotent: second call → `First` fails → **500**. Should be 404/200 | H | S |
| 6 | Quick log (`useQuickLog`, `logmeal/queries/useMealMutations.ts:53`) | P | `days.day(date)`; month summary | Append a logged meal (temp/client id) with the entered macros. Totals are exact (single food, amount 1) | Snapshot | SQD full day (response is `DietLogsTodayResponse`, currently ignored → inv all days) | Creates food+meal+day_log in a tx (`repository.go:785`). Needs `client_id` for retry safety. Replace path (`replace_meal_id`) is also SO | H | S-M |
| 7 | Water log add / delete (`useSaveWater` :166, `useDeleteWater` :186, `api/tracking/queries.ts`) | SO | `trackingKeys.water(date)` (+ day view totals derive from it) | Prepend `{client_id, amount_oz, preset}` / remove by id | Snapshot | Replace temp row with the response `log` / no-op | Create returns `log`. Delete 404 on missing. Create not idempotent → `client_id` | H | S |
| 8 | Reorder planned meals (`useReorderPlannedMeals` :121) | SO | `days.day(date)` | Rewrite `display_order` by index | Snapshot | SQD full day (returned, currently ignored) | Validates the exact unlogged set: a concurrent add makes it 400. Rollback + toast is fine. Idempotent | M | S |
| 9 | Delete planned meal (`useDeletePlannedMeal` :91) | SO | `days.day(date)`, month summary | Remove row | Snapshot | SQD full day | Returns full day. Repeat → error. Make it idempotent | M | S |
| 10 | Add planned from saved (`useAddPlannedFromSaved` :106) | P | `days.day(date)`, month summary | Append a planned row with a temp id built from the cached saved meal (`savedMeals.list`) | Snapshot | SQD full day | Server copies the saved meal into a new Meal (new ids). Needs `client_id`. **Ordering:** list sorts by `display_order, ID` (`PlannedMealsDisplay.vue:79`). Give the temp row `display_order = max+1` and keep the temp id out of the sort | M | M |
| 11 | Grocery complete (`useCompleteGroceryItem` :136) | SO | `grocery.items`; suggestions (inv on settle) | Set `completed_at = now` / remove from the active list | Snapshot | Response `item` | Idempotent (`service.go:107`) | M | S |
| 12 | Grocery add (`useCreateGroceryItem` :110) | SO | `grocery.items` | Append `{client_id, name}` | Snapshot | Swap in the response `item` | Needs `client_id` | M | S |
| 13 | Add exercise to workout (`useAddExerciseToWorkout` :244) | P | `workouts.previous(date)` | Append an `ExerciseGroup` from the cached exercise catalog | Snapshot | Response `exercise` (no `previous`/`max`: refetch on settle) | 409 if already present. Previous/max are server-computed | M | S |
| 14 | Remove exercise from workout (`useRemoveExerciseFromWorkout` :269) | SO | `workouts.previous(date)` | Remove group (or hide, matching today's 404 behaviour) | Snapshot | none | Hard delete + cascade sets. 404 is already handled as "hide" | M | S |
| 15 | Cardio / mobility pre/post (`useUpsertCardio` :360, `useUpsertMobilityPre/Post` :310/335) | SO | `workouts.previous(date)` (`logged_cardio`, `logged_*_mobility`) | Set the field to the submitted value | Snapshot | Response `cardio`/`mobility`, patched in. Activity inv only if this was the first log | Upserts keyed by workout log, so idempotent | M | S |
| 16 | Weight / steps save (`useSaveWeight` :57, `useSaveSteps` :80) | SO | `trackingKeys.weight` / `steps` | Upsert by date in the list (sorted desc) | Snapshot | Response `log` | Server upsert by unique date, so idempotent | M | S |
| 17 | Media quick status / favorite / rating (`pages/media/[id].vue:225`) | P | new `mediaKeys.detail(id)`, `list(params)`, `stats` | Set status. Dates are server-filled ("backend fills in dates") | Snapshot | Response `item`. Stats inv | Returns `item`. PATCH, idempotent | M | M (module has no hooks yet) |
| 18 | Water presets create/update/delete (3 hooks, `api/tracking/queries.ts:208-258`) | SO | `water.presets` (+ water logs embed preset: inv on settle) | Straightforward list ops | Snapshot | Response `preset` | Delete 404 on missing | L | S |
| 19 | Grocery delete (`useDeleteGroceryItem` :149, **unused**) | SO | `grocery.items` | Remove | Snapshot | none | 404 on missing | L | S |
| 20 | Workout plan day assign/unassign, planned cardio/mobility, plan exercise add/remove/reorder (axios in `pages/gym/plans/[id].vue`, `plans/index.vue`, `AddExerciseDialog.vue`) | SO/P | `workoutKeys.plans.list`, `programs`, `workouts.*` (inv) | List ops on the plans list | Snapshot | inv plans + programs + today on settle | Move into hooks first. Reorder/assign are idempotent PUT/POST | L-M | M |

## Wait-for-server (keep pessimistic, adopt the shared pending/error convention)

| Mutation | Why wait |
|---|---|
| `useCreateMeal`, `useLogEditedMeal`, `useUpdateLoggedMeal` (`useMealMutations.ts:100-165`) | Multi-step form that navigates away on success (`afterMealLogNavigate`). Edit creates a **new** meal id (`repository.go:841`). Could later become "navigate immediately + optimistic row" (P), but that needs `client_id` on meals. Medium value, later |
| `useEditLoggedMeal` (`api/diet/queries.ts:69`) | Dead duplicate of `useUpdateLoggedMeal`. Delete it |
| `useCreateSavedMeal`, `useUpdateSavedMeal`, `useDeleteSavedMeal` | Library management with a dependents preview/force flow (`DELETE` returns a reference count first) |
| `useCreateFood`, `useCreateCompositeFood` | Unique-name validation (`idx_foods_name`, `idx_composite_foods_name`) fails often. Result feeds a picker |
| targets `saveMutation` (`pages/diet/targets/index.vue:73`) | Navigates away. Could patch `days.*.plan` macros optimistically later (L) |
| `useSaveProfile` | Rare, settings form |
| `useParseGroceryList` | Pure server computation (not a write) |
| `useCreateGroceryItems` (bulk) | Server dedupes and returns `skipped`. P possible later |
| `useCreateWorkoutProgram`, `useRenameWorkoutProgram`, `useCreateWorkoutPlan` | Rare. Rename could be SO trivially if wanted |
| `useActivateWorkoutProgram`, `useSwitchWorkoutPlan` | Rebuild today's planned exercises server-side. `switch-plan` **returns the full `WorkoutLogsPreviousResponse`**, so do SQD into `previous(date)` instead of invalidating (saves a round trip, still wait) |
| Exercise create/update/cues (axios in `CreateExerciseForPlanDialog`, `EditExerciseDialog`, `AddExerciseDialog`; `EditCuesDialog` unused) | Unique names. Move into hooks; invalidate all exercise keys (today `paginated` is never invalidated) |
| All 14 `useMoneyMutation` sites (accounts create/update/archive/delete, balances set/bulk/delete, transactions create/delete, contribution types/limits/presets, birth year) | Balances, net worth and contribution room are server-computed across accounts. Exception worth doing later: **delete transaction / delete balance** as P (remove row now, totals pending) |
| Media create / full save / delete (`pages/media/[id].vue`) | Forms that navigate. Delete is behind a confirm. Move into hooks first |
| Diet bulk "apply range" (`edit-planned/index.vue:395-428`) | N×M sequential server copies. Wrap it in one mutation (or a backend bulk endpoint) and keep it pessimistic |

## Backend checklist for optimistic endpoints

| Endpoint | Full aggregate returned? | Idempotent? | Version | Needed |
|---|---|---|---|---|
| `POST /workout/exercises/log` | yes (`LoggedExercise` incl. sets) | **no** (dup rows, full-replace races) | `updated_at` only | per-set create keyed by `client_id`, unique `(workout_log_id, exercise_id)`, `PATCH`/`DELETE /sets/:id` |
| `DELETE /workout/exercises/sets/:id` | no (`success`) | 404 on repeat | n/a | treat 404 as success |
| `POST /workout/exercises/add` / `DELETE …/remove` | add: `exercise` / remove: `success` | add 409 / remove 404 on repeat | n/a | OK as is |
| `POST /workout/logs/{cardio,mobility/*}` | entity | yes (upsert) | n/a | OK |
| `PATCH /workout/logs/switch-plan` | **full view** | yes | n/a | use response |
| `POST /diet/meals/meal/log-planned` | full day + totals | yes-ish (no unique index) | n/a | **accept `day_id`**, unique `day_logs(day_id, meal_id)` |
| `DELETE /diet/meals/meal/logged` | full day | **500 on repeat** | n/a | 404 (or 200) on missing, decide planned.logged reset |
| `POST /diet/meals/quick-log` | full day | no | n/a | `client_id` |
| `POST /diet/meals/planned/{from-saved,reorder}`, `DELETE /planned` | full day | from-saved no, reorder yes, delete error on repeat | n/a | `client_id` on from-saved |
| `POST /tracking/water`, `/tracking/grocery/items` | entity | no | n/a | `client_id` |
| `PATCH /tracking/grocery/items/:id/complete` | entity | yes | n/a | OK |
| `POST /tracking/{weight,steps}` | entity | yes (upsert by date) | n/a | OK |
| `PATCH /media/:id` | entity | yes | `UpdatedAt` | OK |

No table has a version column. With one user and one device at a time, last-write-wins on `updated_at`
is acceptable. Don't add optimistic-concurrency versions until multi-device editing is a real problem.

**No user scoping anywhere.** The schema has no `user_id` (`backend/migrations/0001_baseline.sql`), so it's
a single-tenant database. That simplifies IDs (ID-STRATEGY.md). It also means any authenticated account
can write any row, which is out of scope here but worth knowing.
