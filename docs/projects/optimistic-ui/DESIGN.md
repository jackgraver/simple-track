# Optimistic UI design (2026-11-19)

Sketches only. Nothing here exists in `src/` yet. Built on `@tanstack/vue-query` v5 (`^5.92`), which
provides `queryOptions`, `useMutationState` and mutation `scope` (serial execution per scope id).

## 1. Conventions this builds on

1. **One key factory per domain** in `api/<domain>/keys.ts`, one root each (`['diet']`, `['workout']`,
   `['tracking']`, `['money']`, `['media']`). Fold `liveworkout` into `workout`. Shape:
   `all → <entity>.all → <entity>.list(params) / .detail(id) / .day(date)`. No inline literals.
2. **Day-scoped keys use the absolute date** (`'2026-11-19'`), never the offset. Hooks may still accept an
   offset and convert it at the edge. The date is fixed at `mutate()` time and travels in the variables.
3. **`queryOptions` objects** (`dietQueries.day(date)`) are the single source for hooks, prefetch and
   `ensureQueryData`:
   ```ts
   export const dietQueries = {
     day: (date: string) => queryOptions({
       queryKey: dietKeys.days.day(date),
       queryFn: () => getDietDay(date),
       staleTime: STALE.hot,
     }),
   }
   export const useDietDay = (date: MaybeRefOrGetter<string>) =>
     useQuery(computed(() => dietQueries.day(toValue(date))))
   ```
4. **One endpoint, one key.** Readers that need a slice use `select`, not a second key. For example, gym
   index uses `workoutQueries.day(date)` with `select: r => r.day`.
5. **Mutations take everything they target as variables** (`{date, …}`), never through closures read at
   settle time.
6. **Rows created optimistically carry a `client_id`** (ID-STRATEGY.md). Lists are keyed in templates by
   `row.client_id ?? row.ID`.

## 2. `useOptimisticMutation`

Lives in `api/optimistic.ts`. Every optimistic hook uses it (README rule), and wait-for-server hooks use
the same error/pending conventions through `useServerMutation` (the same helper without `optimistic`).

### API

```ts
import type { QueryKey, MutationScope } from '@tanstack/vue-query'

/** One cache entry this mutation patches. */
export type CachePatch<TVars, TCache = any> = {
  queryKey: QueryKey
  /** Pure: return the next cache value. Called with undefined if not cached (return undefined to skip). */
  apply: (old: TCache, vars: TVars) => TCache
  /**
   * Pure inverse used when *other* mutations on the same key are still in flight, so a blind snapshot
   * restore would wipe their optimistic changes. Usually "remove/restore entity by client_id".
   * If omitted, the helper restores the snapshot when it is the only pending writer, otherwise invalidates.
   */
  revert?: (current: TCache, vars: TVars, snapshot: TCache) => TCache
}

export type OptimisticMutationOptions<TVars, TData> = {
  /** Stable key, e.g. workoutKeys.mutations.logSet(). Used for pending-state queries and isMutating checks. */
  mutationKey: QueryKey
  mutationFn: (vars: TVars) => Promise<TData>
  /** Cache entries to patch before the request. */
  optimistic: (vars: TVars) => CachePatch<TVars>[]
  /** Write the server's answer into the cache (preferred over invalidation). */
  reconcile?: (data: TData, vars: TVars, qc: QueryClient) => void
  /** Keys to invalidate after settle (only when no sibling mutation on the key is still pending). */
  invalidate?: (vars: TVars, data: TData | undefined) => QueryKey[]
  /** Serialise mutations that must not reorder, e.g. all writes to one exercise's sets. */
  scope?: (vars: TVars) => MutationScope   // { id: `sets:${date}:${exerciseId}` }
  /** 'rollback' (default): undo + toast with Retry. 'mark': keep the row, flag it failed, retry inline. */
  onFailure?: 'rollback' | 'mark'
  /** Human message for the toast. Defaults to the API error body via getApiErrorMessage. */
  errorMessage?: (err: unknown, vars: TVars) => string
  /** Retries for idempotent writes (client_id creates, upserts, PATCH). Default 0. */
  retry?: number
}

export function useOptimisticMutation<TVars, TData>(
  opts: OptimisticMutationOptions<TVars, TData>,
): UseMutationReturnType<TData, unknown, TVars, OptimisticContext>
```

### Lifecycle

```ts
onMutate: async (vars) => {
  const patches = opts.optimistic(vars)
  // 1. Stop in-flight refetches from overwriting the optimistic value when they land.
  await Promise.all(patches.map(p => qc.cancelQueries({ queryKey: p.queryKey, exact: true })))
  // 2. Snapshot, 3. apply.
  const snapshots = patches.map(p => {
    const prev = qc.getQueryData(p.queryKey)
    qc.setQueryData(p.queryKey, (old: unknown) => old === undefined ? old : p.apply(old, vars))
    return { patch: p, prev }
  })
  return { snapshots }
},
onError: (err, vars, ctx) => {
  if (opts.onFailure === 'mark') { markFailed(vars); return }          // see "pending state"
  for (const { patch, prev } of ctx?.snapshots ?? []) {
    const others = siblingsPending(patch.queryKey)                        // isMutating on same key, minus self
    if (others === 0) qc.setQueryData(patch.queryKey, prev)              // safe: nobody else patched it
    else if (patch.revert) qc.setQueryData(patch.queryKey, (cur: any) => patch.revert!(cur, vars, prev))
    else void qc.invalidateQueries({ queryKey: patch.queryKey, exact: true })
  }
  toast.push(opts.errorMessage?.(err, vars) ?? getApiErrorMessage(err, 'Couldn’t save'), 'error', {
    action: { label: 'Retry', run: () => mutation.mutate(vars) },        // needs toast action support
    duration: 6000,
  })
},
onSuccess: (data, vars) => opts.reconcile?.(data, vars, qc),
onSettled: (data, _err, vars) => {
  for (const key of opts.invalidate?.(vars, data) ?? []) {
    // Refetching while siblings are pending would briefly drop their optimistic rows.
    if (siblingsPending(key) === 0) void qc.invalidateQueries({ queryKey: key })
  }
},
// scope: see "Concurrency" below; v5 scope is per useMutation instance, so per-entity scopes need care.
```

`siblingsPending(key)` = `qc.isMutating({ predicate: m => m.state.status === 'pending' &&
touches(m, key) })`. Here `touches` reads the keys stored on each mutation's context, so it works across
different hooks that patch the same entry.

### Concurrency and rapid taps

- **Different entities, same list** (log set 2 while set 1 is in flight): both apply immediately. Each
  `reconcile` replaces **only its own entity by `client_id`**, never the whole aggregate, while siblings
  are pending. When the last one settles, one `invalidate` (if any) aligns everything.
- **Same entity, ordered ops** (create set, then edit or delete it before the create returns): use
  `scope: { id: 'set:' + client_id }`. TanStack v5 runs same-scope mutations serially, so the edit is sent
  after the create resolves. The cache is already patched for both, so the UI doesn't wait.
  *Caveat:* v5 `scope` is fixed per `useMutation` instance, not per call. For per-entity scopes, call
  `qc.getMutationCache().build(qc, { …, scope })` in the helper's `mutate`, or chain on a per-key promise
  map (`api/optimistic.ts` keeps `Map<string, Promise<unknown>>`). Decide during the foundation slice.
- **Double tap on a create:** the button sends the same `client_id` only if the UI reuses it, and normally
  it doesn't (two taps = two sets is legitimate in the gym). The `client_id` makes *retries* (toast Retry,
  automatic `retry`, network replays) idempotent. That's the case the backend currently gets wrong.
- **Aggregate responses** (diet day, `LoggedExercise`): `reconcile` may `setQueryData` the whole aggregate
  only when `siblingsPending === 0`. Otherwise it merges the one entity and leaves the aggregate to the
  final invalidate.

### Failure UX

- Default `rollback`: the change visibly undoes, and a toast says what failed, with **Retry** (re-runs the
  same variables, hence the same `client_id`) for 6 s.
- `mark` (gym sets, water): the row stays and turns into an error state ("Not saved, tap to retry"). The
  user is mid-workout and losing an entered set is worse than showing a failed one. The rest timer still
  starts on tap, not on server success (today it waits, `useExerciseLoggingSession.ts:385-391`).
- 401 still goes through the existing interceptor (redirect to sign-in). There's no toast for 401.
- Global `MutationCache.onError` toasts only for mutations without `meta.handled` (the helper sets it),
  so there are no double toasts.

### Pending-state convention (no spinners)

- Pending is **derived, not stored in cache data**:
  ```ts
  export function usePendingIds(mutationKey: QueryKey) {
    return useMutationState({
      filters: { mutationKey, status: 'pending' },
      select: m => (m.state.variables as { client_id?: string }).client_id,
    })
  }
  ```
  Failed-and-marked rows use `status: 'error'` the same way.
- Visual: pending rows at ~60 % opacity, or a 4 px dot (token from ui-overhaul). Failed rows: danger tint
  plus a retry affordance. No spinners, no disabled buttons for optimistic actions.
- Wait-for-server forms keep a busy state on the submit button only (a ui-overhaul `Button` `loading`
  prop).

### Example: log a set

```ts
export function useLogSet() {
  const qc = useQueryClient()
  return useOptimisticMutation({
    mutationKey: workoutKeys.mutations.logSet(),
    mutationFn: (v: LogSetVars) => postSet(v),        // POST /workout/sets {client_id, date, exercise_id, …}
    optimistic: v => [{
      queryKey: workoutKeys.day(v.date),
      apply: (day: WorkoutDayView) => withGroup(day, v.exerciseId, g => ({
        ...g, logged: appendSet(g.logged ?? emptyLogged(v), { client_id: v.client_id, weight: v.weight,
          reps: v.reps, weight_setup: v.weight_setup, created_at: new Date().toISOString(), ID: 0 }),
      })),
      revert: (day: WorkoutDayView) => withGroup(day, v.exerciseId, g => removeSet(g, v.client_id)),
    }],
    reconcile: (saved, v) => qc.setQueryData(workoutKeys.day(v.date),
      (day: WorkoutDayView) => withGroup(day, v.exerciseId, g => mergeLoggedByClientId(g, saved))),
    invalidate: (v, saved) => isFirstLogOfDay(saved) ? [workoutKeys.activity.all] : [],
    onFailure: 'mark',
    retry: 2,                                          // idempotent thanks to client_id
  })
}
```

## 3. Recommended QueryClient defaults

```ts
const STALE = { hot: 30_000, default: 60_000, catalog: 10 * 60_000, static: Infinity } as const

new QueryClient({
  queryCache: new QueryCache({
    onError: (err, query) => {
      // Background refetch failed but we still have data: tell the user quietly. A cold-load error is
      // rendered by the page's error state instead.
      if (query.state.data !== undefined && !query.meta?.silent) toast.push('Couldn’t refresh', 'error')
    },
  }),
  mutationCache: new MutationCache({
    onError: (err, _v, _c, m) => { if (!m.meta?.handled) toast.push(getApiErrorMessage(err, 'Couldn’t save'), 'error') },
  }),
  defaultOptions: {
    queries: {
      staleTime: STALE.default,
      gcTime: 24 * 60 * 60_000,          // instant back-nav; prerequisite for stage-4 persistence
      retry: (n, err) => n < 2 && !isClientError(err),   // never retry 4xx
      refetchOnWindowFocus: true,        // PWA resume; gated by staleTime so it is cheap
      refetchOnReconnect: true,
    },
    mutations: { retry: 0 },             // idempotent hooks opt in
  },
})
```

| Data kind | staleTime | Examples |
|---|---|---|
| hot (today, actively edited) | 30 s | diet day, workout day, water(date), grocery items |
| default | 60 s | weight, steps, money lists, media list/detail |
| catalog | 10 min | exercises, foods, saved meals, plans, programs, presets |
| static | ∞ | money presets, lookup tables |

With optimistic patches and `reconcile`, a short hot staleTime costs little and keeps a second device
(laptop and phone) honest.

Also: `queryClient.clear()` on logout. Wrap axios errors once (`isClientError`, `getApiErrorMessage`).

## 4. Instant navigation (stage 3)

1. **Prefetch on intent.** `<NavLink>` (ui-overhaul primitive) calls
   `qc.prefetchQuery(route.meta.prefetch?.(params))` on `pointerenter` and `touchstart`. The nav targets
   are diet today, gym today, grocery, finances, media list.
2. **Route-level warmup.** `router.beforeResolve` runs `ensureQueryData` for the route's primary query
   (from `route.meta.queries`), with a short budget (~150 ms) so slow networks don't block navigation.
3. **Adjacent days.** The diet and gym day views prefetch `date ± 1` after their own data resolves, so day
   swipes are instant.
4. **List → detail seeding.** Media detail, plan detail and account detail use `initialData` from the
   list cache (`getQueryData(list).find(…)`) with `initialDataUpdatedAt` from the list, then refetch in
   the background.
5. **Skeletons only on true cold loads:** `isPending && data === undefined`. Never show a skeleton or
   spinner over cached data. A background refetch is invisible (or a 1 px progress bar at most).
6. **App start.** Prefetch the landing route's own queries in the first `beforeEach` (after redirects),
   so only a page that reads today's day fetches it. Day GETs create the row, so never prefetch a
   future day (DECISIONS "NFIX").
7. **(Stage 4)** `persistQueryClient` with IndexedDB, keyed by build hash, `maxAge` = `gcTime`. Absolute
   date keys (convention 2) are a prerequisite, or a persisted "offset 0" would be wrong the next day.
