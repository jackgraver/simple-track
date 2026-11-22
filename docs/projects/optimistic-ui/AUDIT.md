# Query-layer audit (2026-11-19)

Scope: `frontend/src/api/**`, `main.ts`, every component or composable that touches TanStack Query or
the axios client, plus the backend handlers behind the hot paths. Paths are relative to
`frontend/src/` unless they start with `backend/`. Line numbers are from commit `6e9174d`.

## Verdict

**The owner's suspicion is right, but the problem is the conventions, not the setup.** The QueryClient
setup is fine (one client, sensible defaults). Usage is inconsistent, though, and that blocks optimistic UI:

1. **Two modules bypass TanStack Query entirely.** Media is fully imperative (`onMounted` + refs). Gym
   plan editing calls axios from components and fires `refetch()` by hand. That is ~20 write paths
   with no cache contract.
2. **Hot paths do two round trips per write.** Every gym set POST is followed by a refetch of the whole
   `/workout/logs/previous` aggregate, plus the activity heatmap. Three diet endpoints return the full
   day, but the client throws it away and refetches.
3. **Keys are relative day offsets, not dates**, and mutation hooks read the offset when the request
   *settles*, not when it was *sent*. Both write the wrong day's data into the cache. This is a latent bug
   today and a real one once writes are optimistic.
4. **Gym logging already has a hand-rolled optimistic layer**, a local `loggedSets` copy with
   pending/error status. It also mutates the query cache object in place. The backend's full-replace
   set API breaks under double taps (reproduced below).
5. **There is no global error policy.** No `QueryCache`/`MutationCache` handlers exist, errors are
   swallowed in places, toasts are ad hoc, and 4xx responses are retried.

None of this needs a rewrite. The fix is a key/hook convention plus a shared mutation helper (DESIGN.md),
applied module by module.

## Setup (`main.ts:11-20`, `api/client.ts`)

| Setting | Value | Assessment |
|---|---|---|
| `staleTime` | 5 min default; overrides 2 min (diet day, workout day, activity), 10 min (plans, programs, exercises), `Infinity` (money presets) | OK as a default. Overrides are inconsistent. See staleTime table below |
| `gcTime` | 10 min | Too short for instant back-navigation and future persistence. Recommend 24 h |
| `retry` | 1 (queries), default 0 (mutations) | Retries 4xx (404, 400) too. Should skip 4xx |
| `refetchOnWindowFocus` | `false` | Wrong for a mobile PWA. Reopening the app next morning shows yesterday's cache with no refetch until remount. Turn it on and let `staleTime` gate it |
| `QueryCache`/`MutationCache` `onError` | none | No global error toast. Each site handles errors differently, or not at all |
| axios `timeout` | 10 s, no retry, no idempotency keys | Fine for reads. Writes need idempotency before any automatic retry (ID-STRATEGY.md) |
| Logout | `composables/auth/useAuth.ts:90-99` never calls `queryClient.clear()` | Cached data survives logout. Low risk (single user) but trivial to fix |

## Key inventory

"W" means writers/invalidators. "SQD" is `setQueryData`, "inv" is `invalidateQueries`.

### diet: `api/diet/keys.ts` (`dietKeys`)

| Key | Factory | Readers | Writers / invalidators |
|---|---|---|---|
| `['diet','days','day',offset]` | `dietKeys.days.day` | `useDietLogsToday` (`api/diet/queries.ts:20`) in `DietDayView`, `targets`, `edit-planned`, `logmeal.vue`, `useDietDayMealHandlers` | SQD: `useLogPlannedMeal` :45, `useDeleteLoggedMeal` :61, `useEditLoggedMeal` :83. inv: `useDeletePlannedMeal` :98, `useAddPlannedFromSaved` :113, `useReorderPlannedMeals` :128. inv prefix `days.all`: every `logmeal/queries/useMealMutations.ts` mutation, `targets/index.vue:86`, `edit-planned/index.vue:426` |
| `['diet','days','month-planned-summary',m]` | `dietKeys.days.monthPlannedSummary` | inline `useQuery` in `pages/diet/edit-planned/index.vue:171` | inv prefix from diet day mutations |
| `['diet','meals','detail',id]` | `dietKeys.meals.detail` | `pages/diet/logmeal/queries/useMeal.ts` | **nothing.** Editing a logged meal creates a new meal ID (`backend/.../diet/repository/repository.go:841`), so the old key just goes stale |
| `['diet','saved-meals','list']` | `dietKeys.savedMeals.list` | inline `useQuery` + raw `apiClient` in `edit-planned/index.vue:247-252` | inv prefix `savedMeals.all` |
| `['diet','saved-meals','detail',id]` | `dietKeys.savedMeals.detail` | `logmeal/queries/useSavedMeal.ts` | inv: `useUpdateSavedMeal` |
| `['searchList', path]` | `shared/searchListKeys.ts` | `shared/SearchList.vue:41` (generic GET by URL) | inv `searchListKeys.all` from create food/composite/exercise, inv `path('/diet/meals/saved-meal/all')` |

### workout: `api/workout/keys.ts` (`liveworkoutKeys`, `workoutKeys`, two roots)

| Key | Factory | Readers | Writers / invalidators |
|---|---|---|---|
| `['liveworkout','workouts','day',o]` | `liveworkoutKeys.workouts.day` | `useWorkoutLogToday` in `pages/gym/index.vue` (GET `/workout/logs/previous`, keeps `.day` only) | inv by all 8 logging mutations + switch plan |
| `['liveworkout','workouts','previous',o]` | `liveworkoutKeys.workouts.previous` | `useWorkoutLogsPrevious` in `pages/gym/logging/store/useWorkoutStore.ts:39` (**same endpoint**) | same |
| `['liveworkout','workouts','activity',mode,w,d]` | `...activity` / `activityPrefix` | `useWorkoutActivity` in `WorkoutActivityGraph.vue` | inv by **every** set log, cardio and mobility save |
| `['liveworkout','workouts','today']` | `...today` | **unused** | none |
| `['liveworkout','exercises','all']` | `...exercises.allList` | `useAllExercises`, **unused** | none |
| `['liveworkout','exercises','paginated',q]` | `...exercises.paginated` | `useExercisesPaginated` in `ExerciseListView.vue` | **never invalidated**. A new exercise doesn't show up for 10 min |
| `['workout','exercises','all']` | `workoutKeys.exercisesAll` | `useWorkoutExercisesAllQuery` in `progression.vue` | none |
| `['workout','exercises','progression',id]` | **inline literal** (`pages/gym/progression/ExerciseProgressionChartPanel.vue:53`) | same file | **never invalidated** after logging sets |
| `['workout','plans','all']` | `workoutKeys.plans.list` | `useWorkoutPlansAll` in `SwitchWorkoutPlanDialog`, `plans/[id].vue` | inv by program mutations, `EditExerciseDialog`. Manual `refetch()` in `plans/[id].vue` and `plans/index.vue` after direct axios writes |
| `['workout','programs']` | `workoutKeys.programs` | `useWorkoutPrograms` | inv by program mutations. **Not** invalidated by plan edits in `plans/[id].vue` |

`GET /workout/exercises/all` is cached under **four** keys: `workoutKeys.exercisesAll`,
`liveworkoutKeys.exercises.allList`, `...paginated`, and `searchList` with various `?exclude=` paths.
`GET /workout/logs/previous` is cached under two. Going from gym → logging fetches the same payload twice.

### tracking: `api/tracking/keys.ts` (`trackingKeys`)

| Key | Readers | Writers / invalidators |
|---|---|---|
| `['tracking','profile']` | `useProfile` (targets, settings/profile) | inv `useSaveProfile` |
| `['tracking','weight']` | `useWeightLogs` (targets, gym/weight) | inv `useSaveWeight` |
| `['tracking','steps']` | `useStepLogs` | inv `useSaveSteps` |
| `['tracking','grocery','items']` | `useGroceryItems` | inv create, bulk-create, complete, delete |
| `['tracking','grocery','suggestions',q]` | `useGrocerySuggestions` | inv **literal** `['tracking','grocery','suggestions']` (`api/tracking/queries.ts:143`). No factory for the prefix |
| `['tracking','water',date]` | `useWaterLogs` (`WaterLogsDialog`, `DietDayView`) | inv `useSaveWater`/`useDeleteWater` by date. Preset update/delete inv **literal** `['tracking','water']` (:242, :255), which also hits `water.presets` |
| `['tracking','water','presets']` | `useDrinkSizePresets` | inv preset mutations. **Shares a prefix with per-date water logs**, so `['tracking','water']` invalidates both |

### money: `moneyKeys` lives in `api/money/queries.ts:5` (not `keys.ts`)

Keys: `accounts(includeArchived)`, `account(id)`, `balances(id)`, `transactions(id)`, `netWorth`,
`contributionTypes`, `presets`. All 14 money writes go through `useMoneyMutation` (`queries.ts:76`),
which invalidates `['money']`. That's deliberate: balances, net worth and room are server-computed. It's
also the best-structured module.

### media: no keys, no hooks

`pages/media/index.vue:42-79` and `pages/media/[id].vue:191-251` call `listMedia`, `getMediaStats`,
`getMedia`, `createMedia`, `updateMedia` and `deleteMedia` directly into refs, with a hand-rolled
`requestId` race guard. There's no cache at all, so every visit is a cold load.

## Findings

Severity: **H** blocks optimistic work or is a correctness bug, **M** wasted work or an inconsistency,
**L** cleanup.

### Correctness

- **H: wrong-day cache writes.** Mutation hooks capture `offset` as a getter and read it in `onSuccess`,
  not at `mutate` time: `api/diet/queries.ts:46,62,84,99,114,129` and `api/workout/queries.ts:166,232,257,…`.
  Say you log or unlog on day A and swipe to day B before the response arrives. `setQueryData` then writes
  **day A's response into day B's key**. Fix: pass the day as a mutation variable and use `vars` in callbacks.
- **H: offset-relative keys.** `dietKeys.days.day(offset)` and `liveworkoutKeys.workouts.*(offset)` mean
  "today" by offset. Leave the app open past midnight and the `0` key holds yesterday. With
  `refetchOnWindowFocus: false` it is shown as today until a remount, and an optimistic write would target
  it. Water already keys by `YYYY-MM-DD` (`trackingKeys.water(date)`). Diet and workout should too.
- **H: `group.logged = savedExercise`** (`pages/gym/logging/exercise/composables/useExerciseLoggingSession.ts:304`)
  mutates an object inside the query cache directly. Observers aren't notified, structural sharing breaks,
  and the next refetch silently overwrites it.
- **H: logging a set is a full replace and isn't safe under concurrency.** The client resends every set
  each time (`exerciseLoggingPayload.ts:59-126`). The backend creates every `ID:0` set and deletes any
  set missing from the payload (`backend/internal/core/workout/repository/logged_exercise_repository.go:16-79`).
  `type:"previous"` always inserts a new `logged_exercises` row (`controller/exercises_route.go:69-78`),
  and nothing enforces uniqueness on `(workout_log_id, exercise_id)`. **Reproduced on the dev stack:** two
  concurrent first-set POSTs for exercise 9 created `logged_exercises` 37 (2 sets) and 38 (1 set). The
  read model shows only one of them, so 2 sets silently vanished from the UI. On mobile mid-workout a
  double tap or a retry after a slow network does exactly this.
- **H: set reconciliation matches by value.** `mergeSavedExerciseIntoLoggedSets`
  (`exerciseLoggingPayload.ts:129-174`) pairs pending and saved sets by `(weight, reps, setup)`. Two
  identical pending sets (the common "3×10 @ 135") leave the second one stuck in `pending` forever,
  because `find` always returns the first match.
- **H: logging a planned meal on any day other than today fails.** `PostLogPlanned` ignores the day and
  uses `ZerodTime(0)` (`backend/.../diet/controller/meal_route.go:405`). The UI offers the button on every
  day (`PlannedMealsDisplay.vue:173`). **Reproduced:** a planned meal on offset −1 → `POST
  /diet/meals/meal/log-planned` → HTTP 500. On success the client would also write today's day into the
  other day's key.
- **M: unlogging doesn't un-plan.** `DeleteLoggedMeal` (`repository.go:614`) deletes the `day_logs` row
  but leaves `planned_meals.logged = true`. Log/unlog therefore isn't symmetric, which matters for an
  optimistic toggle. Needs a decision: should unlog return the meal to "planned"?
- **M: plan edits don't refresh dependents.** `plans/[id].vue` writes (planned cardio, mobility, exercise
  add/remove/reorder, day assignment) only `refetch()` the plans list. `workoutKeys.programs` and today's
  `liveworkout` day (built from the plan) stay stale for up to 2 to 10 min.

### Wasted round trips / over-invalidation

- **H: each gym write = POST + full aggregate GET (+ heatmap GET).** All 8 logging mutations invalidate
  `previous`, `day` and `activityPrefix` (`api/workout/queries.ts:231-240`, repeated six times). The log
  endpoint already returns the saved `LoggedExercise`, enough to patch `previous` in place. Activity only
  changes on the day's first log.
- **M: diet endpoints return the full day, the client refetches anyway.** `useDeletePlannedMeal`,
  `useAddPlannedFromSaved` and `useReorderPlannedMeals` (`api/diet/queries.ts:91-134`) invalidate even though
  the backend returns `DietLogsTodayResponse` (`meal_route.go:552-653`). The `logmeal` mutations ignore it
  too and invalidate `['diet','days']` wholesale.
- **M: invalidation isn't awaited.** `onSuccess` fires `invalidateQueries` without returning it, so
  `mutateAsync` resolves, the success toast shows, and the UI updates a round trip later.
- **L: `useMoneyMutation` invalidates `['money']`.** Acceptable (server-computed). Leave it.

### Convention breaks

- **H: axios in components.** `pages/gym/plans/[id].vue:73,93,128,230,246,276`,
  `pages/gym/plans/index.vue:188,207`, `plans/components/AddExerciseDialog.vue:20,37`,
  `CreateExerciseForPlanDialog.vue:30`, `EditExerciseDialog.vue:43`, all of `pages/media/*`, and
  `ExerciseProgressionChartPanel.vue:61`, `edit-planned/index.vue:250` (inside `queryFn`).
  `edit-planned/index.vue:405-411` (bulk apply) calls `addPlannedMealFromSaved` in a loop outside any mutation.
- **M: hooks outside `api/<domain>/queries.ts`.** `pages/diet/logmeal/queries/*` (9 mutations, 2
  queries), `pages/diet/targets/index.vue:73` (inline `useMutation`), `pages/diet/edit-planned/index.vue:171,247`
  (inline `useQuery`).
- **M: inconsistent key factories.** There are two workout roots (`liveworkout`, `workout`). `moneyKeys`
  is in `queries.ts`. Media has none. There are inline literals (progression panel, `tracking` prefixes)
  and the `searchList` path-keyed side cache.
- **M: unstable/fragmented keys.** `SearchList.vue` keys by URL including `?exclude=` ids (:30-45), so
  every distinct exclusion list is a new cache entry and a new fetch. The component already filters
  client-side (:53-60), so server exclusion is redundant. Its `queryKey` prop is never passed by any
  caller (dead feature).
- **M: local copies of server state.** `loggedSets` in `useExerciseLoggingSession.ts:108` (a parallel
  optimistic store, which should become the cache). The media pages hold `items`/`stats`/form in refs.
  The plan form refs in `plans/[id].vue:54-68` are fine (they're drafts).
- **L: dead code.** `useEditLoggedMeal` (duplicates `useUpdateLoggedMeal`, and `api/diet/api.ts` has both
  `editLoggedMeal` and `updateLoggedMeal` for the same endpoint), `useDeleteGroceryItem`,
  `useAllExercises`, `liveworkoutKeys.workouts.today`, `pages/gym/logging/exercise/dialog/EditCuesDialog.vue`
  (no importer).

### Error handling

- **H: no global policy.** There are no `QueryCache.onError`/`MutationCache.onError` handlers.
  `useWorkoutStore.logExercise` swallows the error and returns `null` (`useWorkoutStore.ts:151-165`).
  `useDietDayMealHandlers` shows "Planned Meal Log Failed!" and drops the server message. Money and
  tracking leave it to each dialog. The toast has no action slot, so a "Retry" button isn't possible yet
  (`composables/toast/useToast.ts`).
- **M: query errors render inconsistently.** Some pages show `error.message` raw (progression panel :74),
  some show nothing.
- **M: `retry: 1` retries 4xx.** A 404 on a deleted detail page costs an extra request and ~1 s.

### `enabled` / staleTime

| Query | staleTime | Note |
|---|---|---|
| diet day, workout day/previous, activity, month summary | 2 min | Hot data. Fine once writes patch the cache |
| default (profile, weight, steps, water, grocery, money, meal/saved-meal detail) | 5 min | `useMeal`/`useSavedMeal` restate 5 min redundantly |
| plans, programs, exercises | 10 min | Plan edits don't invalidate dependents (see above) |
| money presets | ∞ | Static. Correct |
| progression | 5 min | Never invalidated after logging |

`enabled` is mostly right. `useMeal`/`useSavedMeal` key on `detail(0)` while disabled (harmless).
`useWaterLogs` has no guard for an empty date, so it fetches "today" implicitly. That's minor.

## What's good (keep)

- Domain folders with `api.ts` / `keys.ts` / `queries.ts` exist for 4 of 5 domains.
- Diet mutations that `setQueryData` from a full-day server response are the right reconcile pattern.
- `useMoneyMutation` is a clean module-level policy.
- `useWorkoutLogToday` and the diet hooks take `MaybeRefOrGetter` and build options in a `computed`.
- Weight, steps and profile writes are server-side upserts keyed by date/singleton, so they're idempotent.
