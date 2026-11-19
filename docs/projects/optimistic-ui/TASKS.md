# Optimistic UI tasks

## Stage 0: audit
- [ ] List every query key and who invalidates it; flag inconsistent/duplicate keys, over-invalidation, missing keys
- [ ] Classify every mutation: safe-optimistic / wait-for-server (server-computed results, side effects)
- [ ] Find misuse: queries fetched in `onMounted`, local copies of server state, components bypassing hooks, missing `enabled`, unstable keys
- [ ] Decide: client-generated UUIDs vs temp integer IDs for creates (touches backend + migrations)
- [ ] Decide: retry policy and idempotency for writes

## Stage 1: foundation
- [ ] Review QueryClient defaults (staleTime per data kind, mutation retry, global error handler)
- [ ] `useOptimisticMutation` helper (snapshot, apply, rollback, settle-invalidate, toast with retry)
- [ ] Pending-state UX convention (subtle, no spinners)
- [ ] Key factory convention applied to every domain

## Stage 2: modules (pair with ui-overhaul slice)
- [ ] gym set logging / workout session
- [ ] diet: log/unlog/edit meal, planned-meal reorder
- [ ] tracking: water, weight, steps
- [ ] grocery
- [ ] finances
- [ ] media

## Stage 3: instant navigation
- [ ] Prefetch on hover / touch-start for main nav and list -> detail
- [ ] Cold-load-only skeletons

## Stage 4 (deferred)
- [ ] Cache persistence (IndexedDB) and offline mutation queue
