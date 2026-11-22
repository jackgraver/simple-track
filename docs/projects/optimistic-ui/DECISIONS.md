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

## 2026-11-22: F4 diet keys
`dietKeys.days.day` takes `YYYY-MM-DD`. `dietQueries` (`queryOptions`) resolves offset -> date at the edge, so key and request match. Day mutations carry `offset` in vars and
resolve the key date in `onMutate`, never at settle. All diet hooks live in `api/diet/queries.ts`, none under `pages/diet/`.
`useEditLoggedMeal` and `editLoggedMeal` were unreferenced, so deleted (owner Q7 still unanswered; only the diet part was touched). Unused offset-bound variants
(`useDeletePlannedMeal`, `useAddPlannedFromSaved`, `useReorderPlannedMeals`) removed. Month-summary keys still use month offset (unused by any query today).
