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

## 2026-11-22: T1 tracking optimistic (migration 0006)
`water_logs.client_id uuid` + partial unique index (migration `0006_water_client_id.sql`). `POST /tracking/water` accepts optional `client_id` (400 if not a uuid); a replay returns the existing row, including one soft-deleted since (a retry must not resurrect a deleted entry), and a unique-violation race re-reads the winner. Delete stays 404-on-missing; the frontend treats 404 as success.
Weight/steps need no backend change (upsert by unique date is already idempotent); their optimistic rows have `ID: 0` until reconcile and are keyed by day (`scope` per date so same-day saves reach the server in order).
Water rows pending create get a negative deterministic temp `ID` from the `client_id`; the log dialog disables delete on them (no server id yet) rather than queueing a delete behind the create. Water add retries once (idempotent); weight/steps do not retry.
Failures use the default rollback mode with the helper's Retry toast (no inline mark mode: nothing to keep on screen). Success toasts ("Logged"/"Saved") now fire on tap, not on server ack. Save buttons lost their loading state because the write is instant; Overwrite confirm dialogs for weight/steps are kept.
`google/uuid` is now a direct dependency (was `// indirect`), used only to validate `client_id`.
Not verified in a browser or against real Postgres (no dev stack/docker allowed): migration SQL checked by reading only, Go tests use SQLite with the partial index created by hand.
