# Optimistic UI (Linear-style)

**Goal**: interactions feel instant. Reads render from cache and refresh in the background; writes apply locally first and reconcile with the server; failures roll back with clear UX.

**Done when**: all safe mutations are optimistic through one shared convention, navigation renders cached data instantly, and failures are handled consistently (rollback + retry toast).

**Non-goals (for now)**: a full local-first sync engine; offline persistence (stage 3, decide later).

## Docs
- [AUDIT.md](AUDIT.md): query-layer audit, key inventory, findings with file:line (verdict at the top)
- [MUTATIONS.md](MUTATIONS.md): every mutation classified (safe-optimistic / partial / wait), ranked, backend checklist
- [ID-STRATEGY.md](ID-STRATEGY.md): bigint vs UUID benchmark, recommendation, migration plan
- [DESIGN.md](DESIGN.md): `useOptimisticMutation` API, QueryClient defaults, prefetch plan

## Current state (audited 2026-11-19, see AUDIT.md)
- `@tanstack/vue-query` v5, one QueryClient in `api/queryClient.ts` (since F1; DESIGN.md §3 defaults and global error toasts).
- 42 `useMutation` + 14 `useMoneyMutation` sites + ~20 writes outside TanStack (gym plans pages, media). **Zero `onMutate`.**
- Media and gym plan editing bypass TanStack Query (axios in components). Diet keeps hooks under `pages/diet/logmeal/queries`.
- Day keys use relative offsets, and mutations read the offset at settle time (wrong-day writes).
- Gym set logging has its own local optimistic copy, mutates the query cache in place, and the backend full-replace
  endpoint duplicates rows on a double tap (reproduced).
- IDs: bigint everywhere, single-tenant schema. Decision: keep bigint and add `client_id uuid UNIQUE` per table (DECISIONS.md).

## Staging
1. **Foundation**: audit query usage vs. conventions, fix setup (defaults, key factories, error handling, mutation wrapper with optimistic + rollback + toast helpers).
2. **Optimistic writes** per module, classified safe/unsafe in the audit. Start with gym set logging and diet log/unlog.
3. **Instant navigation**: prefetch on hover/touch, route-level cache warmup, skeletons only on true cold loads.
4. **(Later, optional)** IndexedDB cache persistence + mutation queue.

## Rules
- All server state goes through TanStack Query hooks in `api/<domain>/queries.ts`. No axios in components.
- Optimistic mutations use the shared helper (to be created), never hand-rolled `onMutate` per hook.
- Every optimistic mutation defines: cache keys touched, rollback, and `onSettled` invalidation.

## Next up / gotchas
- F1 done: `api/queryClient.ts` (defaults, `isClientError`, `STALE` tiers, cache `onError` toasts; cleared on logout/401). Hooks should pick `STALE.hot/catalog/static` explicitly where the 60 s default is wrong.
- F2 done: `toast.push(msg, tone, { action: { label, onClick }, duration })` (key is `onClick`, not `run`). Dev demo: Settings > Appearance.
- F5 done (tracking/money/media/searchList keys). Remaining literal `queryKey: [` sites, if any, belong to F3/F4. `api/media/queries.ts` has read hooks only; M1 migrates the pages.
- F4 (diet date keys, hooks in `api/diet/queries.ts`) done; mutations take `offset` in vars.
- F3 (workout keys, date keys via `api/workout/dates.ts`) is done; mutation vars carry `date`, offset is derived at send time. Gym hooks no longer take an offset.
- F6 done: `api/optimistic.ts` (`useOptimisticMutation`, `useServerMutation`, `usePendingIds`, `siblingCount`) and `api/ids.ts` (`newClientId`). Pass `scope: v => ({ id: 'set:' + v.client_id })` for same-entity ordering (promise chain, not TanStack scope).
- Gotcha (G2): a background refetch/invalidate of a key drops failed `mark` rows from the cache (the server never got them), while `usePendingIds(key, 'error')` still lists them. G2 must re-add failed rows (e.g. from `useMutationState` variables) or keep them out of refetch-owned data.
- D1 done (migration 0005): `log-planned` takes `day_id` (omit = today), log/unlog are idempotent, `client_id` accepted on quick-log and planned/from-saved. Unlog still leaves `planned_meals.logged` set (Q1). D2 should send `day_id` and `client_id`. Migration 0004 belongs to G1: the migrate no-gaps test fails until it is merged.
- `crypto.randomUUID()` is undefined on plain-http LAN testing from a phone. Use `newClientId()` from `api/ids.ts`.
- TanStack v5 mutation `scope` is per `useMutation` instance, not per call. Use the helper's `scope(vars)` option instead (DECISIONS.md, F6).
- T1 done (migration 0006): `POST /tracking/water` takes `client_id` (idempotent). Water add/delete and weight/steps save use `useOptimisticMutation` with pure patch fns in `api/tracking/patches.ts` (vitest). Callers `mutate()` instead of awaiting; the helper toasts failures with Retry. T2 uses 0007 and can reuse `GroceryItem.client_id` the same way.
- Next: G1 (in progress, migration 0004), then G2 (first user-visible slice), D2, T2.
- Open questions for the owner are listed at the bottom of this README. Answers go into DECISIONS.md.

## Open questions (owner)
1. Unlogging a planned meal: should it go back to "planned" (reset `planned_meals.logged`)? Today it doesn't.
2. Logging a planned meal on a past or future day: fix the backend (D1, assumed) or hide the button off-today?
3. Gym failures: OK to keep failed sets visible with inline retry, and to start the rest timer on tap rather than on save?
4. Is the app used on two devices at once (phone + laptop)? If yes, keep the short hot staleTime. Version columns stay out until it hurts.
5. Is multi-user ever planned? The schema has no `user_id`. It doesn't change the ID decision but does change idempotency scoping.
6. OK to replace `POST /workout/exercises/log` (full replace) with per-set endpoints? This frontend is the only client.
7. OK to delete dead code: `EditCuesDialog.vue`, `useEditLoggedMeal`, `useDeleteGroceryItem`, `useAllExercises`?
