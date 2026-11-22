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
