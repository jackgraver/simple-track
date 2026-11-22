# Optimistic UI tasks

One slice = one branch = one session. Each slice states its done-state. Order: foundations (F) first, then
module slices paired with the ui-overhaul phase-4 slice for the same module, then navigation (N).
References: AUDIT.md (findings), MUTATIONS.md (per-mutation plan), ID-STRATEGY.md, DESIGN.md.

## Stage 0: audit (done 2026-11-19)
- [x] Key inventory, invalidators, misuse list → AUDIT.md
- [x] Classify every mutation + backend facts → MUTATIONS.md
- [x] Decide: client ids → ID-STRATEGY.md, DECISIONS.md (bigint + `client_id`)
- [x] Decide: retry/idempotency → DECISIONS.md (retry only idempotent writes; `client_id` makes creates idempotent)
- [x] Helper / defaults / navigation design → DESIGN.md

## Stage 1: foundation (no visible behaviour change)
- [ ] **F1 QueryClient defaults + error policy.** DESIGN.md §3 defaults (`gcTime` 24 h, focus refetch on,
  no 4xx retry), `QueryCache`/`MutationCache` `onError`, `isClientError`, `queryClient.clear()` on
  logout. *Done:* `main.ts` uses them, `vue-tsc` + tests pass, manual check that a 404 detail page makes
  one request.
- [ ] **F2 Toast action + duration.** `toast.push(msg, type, { action, duration })`, styled via the
  ui-overhaul Toast primitive if it exists (otherwise minimal, restyled later). *Done:* a Retry button
  works in a dev-only demo.
- [ ] **F3 Workout keys.** Fold `liveworkout` into `workoutKeys`. Absolute-date day keys.
  `workoutQueries` via `queryOptions`. One key per endpoint (`select` for gym index). Mutation vars carry
  `date`. Delete `useAllExercises`, `workouts.today` and `EditCuesDialog` (if the owner agrees). *Done:* gym
  index → logging makes one `/workout/logs/previous` request (network tab), tests pass.
- [x] **F4 Diet keys.** The same for `dietKeys` (date keys, `dietQueries`). Move `pages/diet/logmeal/queries/*`
  and the inline queries/mutations in `edit-planned`/`targets` into `api/diet/`. Delete `useEditLoggedMeal`
  and the duplicate `editLoggedMeal` API function. *Done:* no `useQuery`/`useMutation` imports under
  `pages/diet/`, and swiping days mid-request no longer writes to the wrong day (manual check with a
  throttled network).
- [ ] **F5 Tracking/money/media/searchList keys.** Prefix factories (no literals), `moneyKeys` → `keys.ts`,
  new `mediaKeys` + `api/media/queries.ts` (no behaviour change yet). `SearchList` keyed by route only,
  with client-side exclusion. *Done:* `grep -rn "queryKey: \[" frontend/src` is empty.
- [ ] **F6 `useOptimisticMutation`, `useServerMutation`, `usePendingIds`, `newClientId`** in
  `api/optimistic.ts` / `api/ids.ts` per DESIGN.md §2, including the per-entity scope decision. *Done:*
  vitest covers apply/rollback, revert with a sibling pending, reconcile merge, settle-invalidate
  suppression, `mark` mode.

## Stage 2: modules (pair with the ui-overhaul phase-4 slice of the same module)
- [ ] **G1 Gym backend: safe set writes.** Migration: merge duplicate `logged_exercises`, unique
  `(workout_log_id, exercise_id)`, FK indexes, `logged_sets.client_id` + unique index. `POST
  /workout/sets` (idempotent by `client_id`, upserts the parent), `PATCH /workout/sets/:id`, `DELETE`
  treating repeats as success. *Done:* Go tests for idempotent replay and concurrent first-set
  (no duplicate parent), `dev migrate check` clean.
- [ ] **G2 Gym set logging optimistic** (pairs with ui-overhaul gym/logging). Replace the
  `loggedSets` local copy with cache + `usePendingIds`. Log/edit/delete sets via the helper, `mark` on
  failure, rest timer starts on tap, activity invalidated only on the day's first log. *Done:* with
  "Slow 3G" throttling, a set appears instantly; airplane mode → set marked failed → retry succeeds with
  no duplicate; 390 px screenshots.
- [ ] **G3 Gym rest of logging:** cardio, mobility, add/remove exercise optimistic, switch-plan uses the
  response (SQD). *Done:* each action updates instantly under throttling, and no follow-up GET appears
  in the network tab.
- [ ] **D1 Diet backend fixes:** `log-planned` takes `day_id` (fixes the 500 off-today), unlog
  idempotent + planned reset (per the owner's answer), unique `day_logs(day_id, meal_id)`, `client_id` on
  `meals` and `planned_meals`. *Done:* Go tests, plus logging a planned meal on yesterday works.
- [ ] **D2 Diet optimistic** (pairs with ui-overhaul diet): log/unlog planned, quick log, reorder,
  delete planned, add-from-saved. Totals recomputed client-side, reconciled from the full-day response.
  *Done:* throttled checks, no extra GET after each write, 390 px screenshots.
- [ ] **T1 Tracking:** water add/delete (`client_id`), weight/steps upsert optimistic. *Done:* as above.
- [ ] **T2 Grocery:** add (`client_id`), complete optimistic, suggestions invalidated on settle.
- [ ] **P1 Gym plans into hooks:** remove axios from `pages/gym/plans/**`. Assign/unassign/reorder
  optimistic. Plan edits invalidate programs + today's workout. Exercise create/update invalidates every
  exercise key.
- [ ] **M1 Media on TanStack Query:** list/detail/stats hooks, list → detail `initialData`, optimistic
  status/favorite/rating. Create/save/delete stay `useServerMutation`.
- [ ] **$1 Finances:** adopt `useServerMutation` error convention. Delete transaction/balance as partial
  optimistic (row removed now, totals pending).

## Stage 3: instant navigation
- [ ] **N1 Prefetch on intent:** `route.meta.queries` + NavLink `pointerenter`/`touchstart` prefetch +
  `beforeResolve` `ensureQueryData` with a 150 ms budget. *Done:* main nav switches render without a
  skeleton after hovering or touching (screen recording).
- [ ] **N2 Adjacent days + cold-only skeletons:** prefetch date ± 1 in diet/gym. Skeleton only when
  `isPending && !data` across pages. App-start prefetch of today's diet + gym.

## Stage 4 (deferred)
- [ ] Cache persistence (IndexedDB) and offline mutation queue. Needs F3/F4 date keys and `client_id` everywhere.
