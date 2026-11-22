# Decisions (append, newest last; 3-5 lines each)

## 2026-11-19: Stay on TanStack Query, not a sync engine
Cache + optimistic mutations give most of the "Linear feel" at a fraction of the cost. Revisit only after stage 3.

## 2026-11-19: Client ids: keep bigint PKs, add `client_id uuid UNIQUE` per table as needed
Index/insert cost is irrelevant at this scale (<100k rows/decade; at 10k rows the options differ by ~600 kB, see ID-STRATEGY.md).
The real problem is idempotency: a double tap already duplicates `logged_exercises`. `client_id` makes creates replayable and gives
pending rows a stable key. UUID PKs rejected (rewrites every FK/type/route for no benefit); temp ids alone and Idempotency-Key alone are insufficient.

## 2026-11-19: Retry and idempotency policy for writes
Mutations default to `retry: 0`. Only idempotent writes opt in (client_id creates, date upserts, PATCH/PUT). Deletes treat 404 as success.
Queries never retry 4xx. No Idempotency-Key middleware for now. Revisit only if a non-create write proves unsafe to replay.

## 2026-11-19: Day-scoped cache keys use absolute dates, and mutations carry their target in variables
Offset keys go wrong across midnight, and settle-time offset reads write to the wrong day when the user swipes mid-request.
Hooks may take an offset but convert it to `YYYY-MM-DD` at the edge. Prerequisite for optimistic writes and for persistence.

## 2026-11-19: Gym set failures stay visible ("mark"), everything else rolls back
Losing a set mid-workout is worse than showing a failed one, so sets/water stay with an inline retry. Other optimistic writes roll back with a Retry toast.
Pending state is derived from `useMutationState`, never stored in cache data. No spinners for optimistic actions.

## 2026-11-22: F1 QueryClient lives in `api/queryClient.ts`; error policy details
`createQueryClient(notify)` factory (testable) plus a `queryClient` singleton used by `main.ts` and `useAuth` (no injection needed, so logout works outside component setup). `isClientError` lives there, not in `httpError.ts`, to avoid conflicts with parallel slices.
Cache `onError` skips 401 (the interceptor already redirects to sign-in) and `meta.silent` / `meta.handled` opt out. `queryClient.clear()` runs on logout and on the 401 redirect so the next user never sees stale data.
The `STALE` tiers are exported but not yet applied per hook; the global default is 60 s, and tiering hooks happens in F3-F5 as keys are touched.

## 2026-11-22: Toast action API (F2)
The ui-overhaul toast store already carried `action` and `duration`; F2 only adds the 3rd arg `push(msg, tone, { action, duration })` alongside the object form.
Action shape stays `{ label, onClick }` (existing Toaster contract), so DESIGN.md's `run:` snippet should read `onClick:`. Default duration with an action is 6 s, matching Failure UX.
Dev-only Retry demo button lives in the Settings Appearance section (`import.meta.env.DEV`).

## 2026-11-22: F5 key factories (tracking, money, media, searchList)
Water logs moved to `['tracking','water','logs',date]` so invalidating `waterLogsAll` no longer hits presets. `moneyKeys` now lives in `api/money/keys.ts` (re-exported from `queries.ts`).
`SearchList` is keyed by route only: the `?exclude=` suffix is dropped (the backend honoured one id at most, and the component already filters client-side), so one cache entry serves every prefilter.
Media gets `mediaKeys` plus read hooks only; pages keep calling `api.ts` until M1. `SearchList`'s unused `queryKey` prop was left in place for F3 to remove when it folds exercises into `workoutKeys`.

## 2026-11-22: F4 diet keys
`dietKeys.days.day` takes `YYYY-MM-DD`. `dietQueries` (`queryOptions`) resolves offset -> date at the edge, so key and request match. Day mutations carry `offset` in vars and
resolve the key date in `onMutate`, never at settle. All diet hooks live in `api/diet/queries.ts`, none under `pages/diet/`.
`useEditLoggedMeal` and `editLoggedMeal` were unreferenced, so deleted (owner Q7 still unanswered; only the diet part was touched). Unused offset-bound variants
(`useDeletePlannedMeal`, `useAddPlannedFromSaved`, `useReorderPlannedMeals`) removed. Month-summary keys still use month offset (unused by any query today).

## 2026-11-22: F3 workout keys (date keys, one key per endpoint)
`liveworkout` folded into `workoutKeys` (`['workout', ...]`, `days.day(date)`, `activity`, `exercises`). `workoutQueries` via `queryOptions`.
`api/workout/dates.ts` converts offset <-> local `YYYY-MM-DD`; the offset is recomputed from the date when a request is sent. Mutation hooks take no offset, vars carry `date`.
`/workout/logs/previous` has one key; `useWorkoutLogToday` is `select: r => r.day` on it. Local-time dates assume the backend's "today" matches the device timezone (unverified).
Deleted `useAllExercises`, `workouts.today` and `EditCuesDialog.vue`: all had zero references, so owner Q7 is moot for them (`useEditLoggedMeal`/`useDeleteGroceryItem` untouched, not my slice).
Not verified in a browser (no dev stack allowed): the "one request gym index -> logging" check is by construction (shared key), not by network tab.

## 2026-11-22: D1 diet backend fixes (migration 0005)
`POST /diet/meals/meal/log-planned` takes optional `day_id` (diet day row id; 0 = today, so the current frontend keeps working), is idempotent (marks logged + one day log, in a tx) and returns 404 when the day has no such planned meal. This fixes the off-today 500.
`DELETE /meal/logged` is idempotent: an already-missing log returns 200 with the full day (was 500).
**Unlog does NOT reset `planned_meals.logged`** (owner Q1 unanswered; resetting would make meals reappear in the planned list, a visible behaviour change). `DeleteLoggedMeal` documents this. Revisit when answered.
Migration `0005_diet_idempotency.sql`: partial unique `day_logs(day_id, meal_id) WHERE deleted_at IS NULL` (existing live duplicates soft-deleted first, oldest kept; soft-deleted rows are exempt so unlog + relog works), and nullable `client_id uuid` + partial unique index on `meals` and `planned_meals`.
`client_id` is optional on `POST /quick-log` (stored on the meal) and `POST /planned/from-saved` (stored on the planned meal); a replay with a known id creates nothing and returns the current full day. Not wired to `meal/new` or `logedited` yet (D2 can, if needed).
Numbering: G1 takes 0004, D1 takes 0005. `TestEmbeddedMigrationsParse` (no gaps) fails on this branch until G1 merges; it passes once both are in.

## 2026-11-22: F6 per-entity scope = promise chain in `api/optimistic.ts`, not TanStack `scope`
TanStack v5 `scope` is fixed per `useMutation` instance, and building mutations by hand per call loses the hook's reactive state. The helper's `scope(vars)` instead
chains requests per scope id (module-level `Map`): the slot is taken synchronously in `onMutate` (so order = call order, across hooks), `mutationFn` awaits
the predecessor, and the slot is released in `onSettled` (after all retries, success or failure). Patches still apply instantly. A request that never settles blocks its scope.

## 2026-11-22: F6 helper details that differ from or extend DESIGN.md §2
Siblings are found through `meta.optimisticKeys(vars)`, so they count from the moment they are pending (before `onMutate` resolves) and across hooks; keys
overlap when either prefixes the other. A mutation stops counting once it reaches `onSettled` (flag on its context), so two settling in one tick invalidate exactly once.
Rollback also counts failed `mark` mutations (their rows are still cached); settle-invalidation only counts pending ones. `apply` is skipped when the entry isn't cached.
`mark` mode: no toast, `gcTime: Infinity` so the failure outlives the 5 min default, `apply` must be idempotent (Retry re-runs it), and the hook returns `retry(vars)` /
`dismiss(vars)` which drop the failed mutation (matched by `client_id`). `reconcile` gets a 4th arg `{ siblingsPending(key) }`. `usePendingIds` returns a `Set`.
`useServerMutation` toasts without Retry (forms keep inputs); `errorMessage: false` leaves the error to the caller.

## 2026-11-22: M1 media on TanStack Query
Pages hold no server state: list/stats/detail are hooks, the detail page's form is a draft filled once per item (a background refetch never overwrites edits). Detail uses `initialData` from the newest cached list containing the item.
Only the quick **status** change is optimistic (`useQuickUpdateMedia`, scope `media:<id>`): patches `detail(id)` + every cached list containing the item, invalidates stats + lists on settle, and adopts the server-filled `started_at`/`finished_at`/`progress` into the form on success; on failure it rolls back and restores the form's status with an error toast.
The hook accepts `favorite` and `rating` too, but there is no quick UI for them: they stay form fields saved with Save (conservative; a heart toggle / inline stars on the list/detail are a UI decision for the ui-overhaul media slice).
Status is patched in place in lists even when a status filter is active (the row leaves on the settle refetch). Stats are not patched optimistically (old status isn't known to the patch fn), only refreshed on settle.
Create/save/delete use `useServerMutation` (toast on error, inputs kept). Save no longer re-fills the form from the response before navigating.
Not verified in a browser (no dev stack allowed): only vue-tsc, vitest (patch/revert functions) and eslint were run.

## 2026-11-22: T2 grocery add/complete optimistic (migration 0007)
Migration `0007_grocery_client_id.sql`: nullable `grocery_items.client_id uuid` + partial unique index. `POST /tracking/grocery/items` accepts `client_id`; a replay (also after the item was completed or deleted, lookup is unscoped) returns the existing row and creates nothing. Concurrent duplicate falls back to a lookup after the unique violation. Bulk add is unchanged. Go tests cover replay. `TestEmbeddedMigrationsParse` (no gaps) fails until 0004/0006 are merged.
Add: temp row `{ID: negative hash of client_id, client_id, name}` appended to `groceryItems`; `reconcile` swaps in the server row by `client_id`; rollback removes it by `client_id`; settle invalidates the list. `retry: 2` (idempotent by client_id). The list keys rows by `client_id ?? ID` so the row doesn't remount on the swap. A temp row can't be checked off until saved (no server id); unsaved rows are only that way for one round trip.
Complete: the row is removed from the cache at once (the page's existing `lastSeen`/`completingIds` code keeps it visible with the strike-through for 1.2 s). Revert re-inserts it at its old index. Settle invalidates list + `grocerySuggestionsAll` (suggestions come from completed items, so no optimistic patch). `retry: 2` (PATCH is idempotent).
Page no longer toasts itself or guards on `isPending`; rapid adds are allowed (each has its own `client_id`), the input clears immediately. Not verified in a browser (no dev stack allowed). `useDeleteGroceryItem` untouched (unused, owner Q7).

## 2026-11-22: G1 per-set endpoints (gym backend)
Migration `0004_logged_sets_safe_writes`: merges duplicate live `logged_exercises` (sets move to the lowest id, notes joined), partial unique `(workout_log_id, exercise_id) WHERE deleted_at IS NULL`, FK indexes, `logged_sets.client_id uuid` + unique index. Down drops indexes/column; the merge stays.
`POST /workout/sets {client_id, date|?offset, exercise_id, reps, weight, weight_setup}` upserts the parent via `ON CONFLICT`, then the set via `ON CONFLICT (client_id) DO NOTHING`; a replay returns the stored set (first write wins, `created:false`), a client_id reused for another day/exercise is 409. Response: `{exercise (full LoggedExercise, sets ordered by id), set, created, first_set_of_day}`.
`PATCH /workout/sets/:id` (only fields present) and `DELETE /workout/sets/:id` (200 with `deleted:false` on repeat, `exercise_removed` when the last set took the parent with it, as before). `:id` is the numeric id or the client_id.
`workout_logs.date` stays non-unique (legacy duplicates can't be merged safely blind); get-or-create of a day is serialized with `pg_advisory_xact_lock` per date instead. Legacy `POST /exercises/log` and `DELETE /exercises/sets/:id` are unchanged for the current frontend, except `type:"previous"` on an existing parent now becomes a full replace (no duplicate) and it ignores `client_id`.

## 2026-11-22: T1 tracking optimistic (migration 0006)
`water_logs.client_id uuid` + partial unique index (migration `0006_water_client_id.sql`). `POST /tracking/water` accepts optional `client_id` (400 if not a uuid); a replay returns the existing row, including one soft-deleted since (a retry must not resurrect a deleted entry), and a unique-violation race re-reads the winner. Delete stays 404-on-missing; the frontend treats 404 as success.
Weight/steps need no backend change (upsert by unique date is already idempotent); their optimistic rows have `ID: 0` until reconcile and are keyed by day (`scope` per date so same-day saves reach the server in order).
Water rows pending create get a negative deterministic temp `ID` from the `client_id`; the log dialog disables delete on them (no server id yet) rather than queueing a delete behind the create. Water add retries once (idempotent); weight/steps do not retry.
Failures use the default rollback mode with the helper's Retry toast (no inline mark mode: nothing to keep on screen). Success toasts ("Logged"/"Saved") now fire on tap, not on server ack. Save buttons lost their loading state because the write is instant; Overwrite confirm dialogs for weight/steps are kept.
`google/uuid` is now a direct dependency (was `// indirect`), used only to validate `client_id`.
Not verified in a browser or against real Postgres (no dev stack/docker allowed): migration SQL checked by reading only, Go tests use SQLite with the partial index created by hand.

## 2026-11-22: P1 gym plans hooks
All routine hooks use the shared helper and patch only `workoutKeys.programs`; every settle invalidates programs, `plans.all` and `days.all` (plan edits change today's workout). Requests per routine are serialised with `scope: routine:<planId>`, so rapid toggles and drags reach the server in order.
Assign mirrors the backend rule locally (the program's other routines lose that day). Revert restores only the affected routines' fields from the snapshot, so concurrent edits survive a failure.
Optimistic hooks now toast via the helper (with Retry); the pages no longer toast for them and dropped their per-day pending dimming and the "reordering" lock. Delete/restore routine, add exercise, create/update exercise stay `useServerMutation` with `errorMessage: false` (the pages and dialogs show errors in context). Create/update exercise invalidate exercises, searchList, programs, plans and days.
Conservative: no `client_id` on routine edits (all idempotent PUT/POST/DELETE by id). Moving a routine between days sends assign then unassign (same scope, ordered); a failure of the first leaves the second to run, as before.

## 2026-11-22: D2 diet optimistic
All five day mutations use `useOptimisticMutation`; the patch is one key, `dietKeys.days.day(date)`, via pure functions in `api/diet/dayPatches.ts` (vitest covers apply and revert). `reconcile` writes the server's full day only when no other write to that day is pending; month summaries are invalidated on settle for planned changes. No extra day GET after a write.
Day and `client_id` are stamped into the variables at call time (`dayRef(offset)`), so key, scope and request cannot drift after a swipe or midnight. Every diet day mutation shares scope `diet-day:<date>` so reorder (which validates the exact planned set) cannot overtake an add. `retry: 1` only on idempotent calls (D1 made log/unlog idempotent; quick-log/from-saved carry `client_id`).
Logging adds a `DayLog` and unlogging removes it; planned rows stay in the cache (the UI derives "logged" from `loggedMeals`), consistent with Q1 unanswered (server still leaves `planned_meals.logged` set).
Optimistic rows have negative temp ids; delete/unlog/reorder on them are ignored until the server row replaces them (conservative, avoids sending a temp id). Success toasts were dropped (UI already changed); failure = rollback + Retry toast from the helper. Quick-log dialog now closes immediately. Optimistic totals use food macros in the cache; any gap is corrected when the response lands.
Left pessimistic on purpose: copy-day, saved-meal CRUD, `createMeal`/`logEditedMeal`/`updateLoggedMeal` (logmeal page; navigation after submit).

## 2026-11-22: $1 finances error convention + partial optimistic deletes
`useMoneyMutation(fn, failureMessage)` now wraps `useServerMutation` (key `moneyKeys.mutations.write`, invalidate `moneyKeys.all`, toast = API message or `failureMessage`). The 6 finance files dropped their own `toast.push(getApiErrorMessage(...))` in `catch`; they still catch so dialogs stay open and success toasts/navigation are skipped. Success toasts unchanged.
Delete transaction / delete balance are `useOptimisticMutation` (`useDeleteTransaction`, `useDeleteBalance`, vars `{accountId, id}`): only the row list is patched (`transactions(id)` / `balances(id)`); account totals, gain, net worth and room are NOT patched (server-computed, conservative) and refresh when `moneyKeys.all` is invalidated on settle. Failure rolls back (or reverts just the row if a sibling patched the list) with an error toast + Retry. No `retry` (not marked idempotent: a repeat DELETE of a gone row may 404). Deleting a balance can change the "current balance" shown until settle; accepted.
Reinsert order on revert is date DESC then id DESC (matches the backend queries); the "pending" row dimming on those lists was removed since the row is gone at once.
Not verified in a browser (no dev stack allowed): vue-tsc, vitest (patch/revert + hook config), eslint only.

## 2026-11-22: N1 prefetch on intent
`route.meta.queries(to) => queryOptions[]` (typed in `router/index.ts`, logic in `router/prefetch.ts`) lists a route's primary queries: diet/gym day (from `?offset=`), grocery, finances (accounts + net worth), media (default-filter list + stats). Only top-level nav targets have it; child routes (logging, plans, targets) don't yet.
`router.beforeResolve` calls `warmRoute`: `ensureQueryData` for each query (cached data returns at once even if stale, the page's own `staleTime` handles the background refetch) raced against 150 ms (`NAV_BUDGET_MS`). On timeout navigation proceeds and the request keeps running, so the page joins the in-flight fetch. Errors and a throwing meta fn never block navigation.
`NavLink.vue` (`shared/nav`, wraps `router-link`; used by SideNav and the mobile group menu) and the bottom tab bar buttons call `prefetchRoute` (`prefetchQuery`, honours `staleTime`) on `pointerenter`, `touchstart` (passive) and focus. The tab bar prefetches the same target a tap would open (remembered path or home).
Added `trackingQueries.groceryItems`, `moneyQueries.accounts/netWorth`, `mediaQueries.list/stats` (queryOptions) so route meta and hooks share one definition; only grocery/net-worth hooks were switched to them. `useAccounts`/`useMediaList` keep their reactive key but the same key + fn. Media list key matches the page's defaults because undefined params drop out of the key hash.
Not verified in a browser (no dev stack allowed): vue-tsc, vitest (budget logic with fake timers) and eslint only. Skeleton-free switching still depends on pages using `isPending && !data` (N2).

## 2026-11-22: G2 gym set logging optimistic (frontend, G1 endpoints)
The exercise page has no local set list: rows are `deriveSetRows(cache sets, intents)` where intents come from `usePendingIds` (pending/failed per key) plus `useMutationState` variables of pending/failed creates/edits, so a failed row survives a refetch (F6 gotcha) and a pending delete stays hidden. Set key = `client_id`, or `id:<ID>` for pre-G1 sets; it is the mutations' `client_id` var and maps to `:id` of `/workout/sets/:id`.
Hooks in `api/workout/queries.ts` (`useLogSet`/`useUpdateSet`/`useDeleteSet`/`useSaveExerciseNotes`), pure patches in `setsCache.ts`. All four share scope `sets:<date>:<exercise_id>`, so an edit/delete of a set still being created, and the notes save, wait for earlier writes of that exercise. Reconcile always merges the response `exercise` by key (keeps unsaved rows, local fields with an edit in flight, hides keys being deleted), never a blind replace.
Create/edit: `mark` + `retry: 2` (inline "Didn't save" with Retry and Delete). Delete: rollback + Retry toast, `retry: 2`; `exercise_removed` drops `group.logged` (and an unplanned group). Activity is invalidated only on `first_set_of_day` (create) or `exercise_removed` (delete; a day can only go inactive then). The day key itself is never invalidated by set writes.
Rest timer starts on tap (owner Q3 unanswered; per DESIGN.md). Finish is blocked while any set is failed (toast: retry or delete first), because failed writes live only in memory. Deleting a failed create dismisses it and still sends an idempotent DELETE by client_id (the create may have reached the server).
Notes: no notes endpoint exists, so `useSaveExerciseNotes` uses the legacy full replace (`type: "logged"`) built at send time from the cache's saved sets (`ID > 0`), inside the exercise scope. Risk: a set logged on another device and not yet in this cache would be deleted by it. A `PATCH` notes endpoint would remove that; left for G3/owner. Notes are saved on change (sheet), after a set, and at finish, only when they differ from the cached ones.
The draft is initialised once per (date, exercise), not on every cache change; the draft storage key moved to `gym-draft:v2:date:<date>:exercise:<id>`. Removed `useLogExercise`, `useDeleteLoggedSet`, `exerciseLoggingPayload.ts` (value-matching merge). Not verified in a browser (no dev stack allowed).

## 2026-11-22: G3 gym rest of logging
Patches are pure functions in `api/workout/dayCache.ts` on the day key (vitest in `dayCache.test.ts`, hook tests in `dayMutations.test.ts`). `useUpsertCardio` and `useUpsertMobilityPre/Post` are optimistic with `retry: 2` (upserts keyed by workout log). A tick sends the full `checked` list, so each slot has its own scope (`mobility:<date>:<slot>`) and cardio has `cardio:<date>`; the response is written only when no other write to the day is pending. Mobility with no logged row yet starts from the planned routine. Rollback with siblings restores just that slot/cardio field from the snapshot (a failed tick can undo a later successful one's display until the next refetch: accepted, rare).
Add/remove exercise share the set scope `sets:<date>:<exercise_id>`, so a first set can't overtake the add and a remove waits for pending sets. Add needs the catalog `Exercise` for the name (the list view passes it with the id); without it nothing is patched and the settle refetch shows it. The saved row replaces the temporary one (ID 0), but `previous`/`max` are server-computed, so add and remove still invalidate the day once on settle: the one GET G3 keeps. Remove keeps its old semantics: a planned exercise stays in the list (minus its log), an ad-hoc one disappears; an exercise with nothing logged is only hidden locally with no request (it used to 404 and hide). A 404 from the server counts as success.
Switch plan stays pessimistic (`useServerMutation`, it rebuilds planned exercises) and writes the full-day response with `setQueryData`: no GET. If another write to that day is pending, the day is only marked stale (`refetchType: 'none'`) so optimistic rows aren't wiped. Activity is invalidated on settle for all of these (cheap, separate key).
UI: store functions return `void`; success toasts dropped (the UI already changed), failures come from the helper (Retry toast). Cardio "Finish" navigates back immediately; the "removing" row dimming and `removingId` were deleted. Not verified in a browser (no dev stack allowed): vue-tsc, vitest, eslint only.
