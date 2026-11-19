# Optimistic UI (Linear-style)

**Goal**: interactions feel instant. Reads render from cache and refresh in the background; writes apply locally first and reconcile with the server; failures roll back with clear UX.

**Done when**: all safe mutations are optimistic through one shared convention, navigation renders cached data instantly, and failures are handled consistently (rollback + retry toast).

**Non-goals (for now)**: a full local-first sync engine; offline persistence (stage 3, decide later).

## Current state (audited 2026-11-19)
- `@tanstack/vue-query` v5, one QueryClient in `main.ts`: `staleTime` 5min, `gcTime` 10min, `retry: 1`, `refetchOnWindowFocus: false`.
- Query modules per domain in `frontend/src/api/{diet,media,money,tracking,workout}/` (`api.ts` axios calls, `keys.ts`, `queries.ts` hooks, ~2.9k lines). 136 query/mutation call sites, 49 `useMutation`.
- **Zero `onMutate`: nothing is optimistic today.** Mutations are `onSuccess` -> `setQueryData` from the server response, or `invalidateQueries`. So every write waits for a round trip.
- Key factories exist for some domains (`dietKeys`, `trackingKeys`, workout keys); check consistency in the audit.
- `useAuth` talks to axios directly (reasonable, auth is not cached server state).
- axios client has a 10s timeout, no retry policy, no idempotency keys.
- IDs are server-generated integers (e.g. `mealId: number`), so optimistic **creates** need temp IDs unless the backend accepts client UUIDs. Decide in DECISIONS.
- Backend returns full updated aggregates for several mutations (e.g. diet day), which suits `setQueryData` reconciliation.

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
- Run the query-usage audit first; foundations before any per-module work.
