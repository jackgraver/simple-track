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
- `@tanstack/vue-query` v5, one QueryClient in `main.ts`: `staleTime` 5min, `gcTime` 10min, `retry: 1`, `refetchOnWindowFocus: false`. No global error handlers.
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
- Next: **F1** (defaults + error policy), then F3/F4 (date keys) before any optimistic hook. G1+G2 (gym sets) is the first user-visible slice.
- `PostLogPlanned` ignores the day (500 off-today). Fixed in D1, and until then don't build optimistic log-planned on top of it.
- `crypto.randomUUID()` is undefined on plain-http LAN testing from a phone. Use `newClientId()` (F6).
- TanStack v5 mutation `scope` is per `useMutation` instance, not per call. See DESIGN.md §2 "Concurrency".
- Open questions for the owner are listed at the bottom of this README. Answers go into DECISIONS.md.

## Open questions (owner)
1. Unlogging a planned meal: should it go back to "planned" (reset `planned_meals.logged`)? Today it doesn't.
2. Logging a planned meal on a past or future day: fix the backend (D1, assumed) or hide the button off-today?
3. Gym failures: OK to keep failed sets visible with inline retry, and to start the rest timer on tap rather than on save?
4. Is the app used on two devices at once (phone + laptop)? If yes, keep the short hot staleTime. Version columns stay out until it hurts.
5. Is multi-user ever planned? The schema has no `user_id`. It doesn't change the ID decision but does change idempotency scoping.
6. OK to replace `POST /workout/exercises/log` (full replace) with per-set endpoints? This frontend is the only client.
7. OK to delete dead code: `EditCuesDialog.vue`, `useEditLoggedMeal`, `useDeleteGroceryItem`, `useAllExercises`?
