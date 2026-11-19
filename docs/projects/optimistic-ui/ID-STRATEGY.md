# ID strategy for optimistic creates (2026-11-19)

**Recommendation: (b) keep `bigint` PKs and add a nullable client-generated `client_id uuid` with a unique
index**, only on tables that get optimistic creates, one table per module slice. Combine it with the
natural-key fixes the audit found, chiefly unique `(workout_log_id, exercise_id)`. Index cost is not a
factor at this app's scale. **Correctness under retries and double taps is the factor.**

## The app's scale

- Single-tenant schema: no `user_id` anywhere (`backend/migrations/0001_baseline.sql`), one real user.
- Every PK is `bigint` from a sequence. FKs are `bigint`. GORM models use `uint` and the frontend uses
  `number`. Routes match ids with `(\d+)` (`frontend/src/router/index.ts:79,85,100,158`).
- Dev seed: 108 `logged_sets`, 36 `logged_exercises`, 18 `meals`. A realistic upper bound for the
  biggest table (`logged_sets`): ~6 sets × 6 exercises × 4 days × 52 weeks ≈ **7.5k rows/year**. Every
  table together stays **well under 100k rows for a decade**.

## Benchmark (dev Postgres 16.16, throwaway schema, script [`idbench.sql`](idbench.sql), run with `./bin/dev db:psql -v n=10000 -f - < idbench.sql`)

Table shaped like `logged_sets` plus a parent FK index. Five variants: (a) `bigint identity`, (b)
`bigint` + `client_id uuid UNIQUE`, (c) UUIDv7 PK bulk-inserted (many rows per ms), (d) UUIDv4 PK, (e)
UUIDv7 PK with 1 s between ids (models real traffic). The FK column has the same type as the PK.
PG16 has no `uuidv7()` (PG18 does), so v7 is emulated in SQL.

| rows | variant | heap | PK index | FK index | extra unique | **total** |
|---|---|---|---|---|---|---|
| 10k | a bigint | 752 kB | 240 kB | 128 kB | | **1.15 MB** |
| 10k | b bigint + client_id | 992 kB | 240 kB | 128 kB | 344 kB | **1.74 MB** |
| 10k | e uuid7 PK (spaced) | 912 kB | 328 kB | 144 kB | | **1.42 MB** |
| 10k | d uuid4 PK | 912 kB | 408 kB | 168 kB | | **1.52 MB** |
| 1M | a bigint | 73 MB | 21 MB | 11 MB | | **105 MB** |
| 1M | b bigint + client_id | 96 MB | 21 MB | 11 MB | 42 MB | **171 MB** (+63 %) |
| 1M | c uuid7 PK (bulk) | 89 MB | 49 MB | 16 MB | | **154 MB** |
| 1M | e uuid7 PK (spaced) | 89 MB | 30 MB | 12 MB | | **131 MB** (+25 %) |
| 1M | d uuid4 PK | 89 MB | 38 MB | 16 MB | | **143 MB** (+36 %) |

Insert time for 1M rows was 1.5 s for bigint against 2.9 to 3.8 s for the UUID variants. Most of that gap
is the SQL-level UUID generation function, not index maintenance, so treat it as an upper bound. In
the app the id comes from the client. Point lookups (1000 by PK) took 6 to 11 ms for every variant once
cached.

**What this means here:** at 10k rows the worst option costs **~600 kB** more than the best. The whole
database fits in Postgres's default 128 MB `shared_buffers` many times over under every variant. Index
size and insert cost **don't decide** this choice for this app. They only start to matter around 10⁷–10⁸
rows, which this app will not reach. (If it ever did, v7 beats v4 because it appends to the right edge
of the index. Bulk inserts sharing one millisecond lose part of that, as variant c shows.)

## Options

### (a) `bigint` PKs, temp client ids, reconcile from the response

The client inserts `{ID: -1, …}` (or a `tempId` string) and swaps in the server id on success. No
backend change.

What breaks:
- **Retries and double taps duplicate rows.** Nothing tells the server "this is the same create". This
  is the bug already reproduced for set logging (two `logged_exercises` from two taps).
- **Children of optimistic parents.** The first set of an exercise creates the `logged_exercise` (parent)
  and the set (child) together. A second set tapped before the first response has no parent id to send.
  You end up queueing the child behind the parent or sending the whole aggregate (today's full-replace
  design, which races).
- **Rapid sequential creates.** Each response must be matched to the right temp row. Without an id the
  server echoes back, matching falls back to values, and the current code
  (`mergeSavedExerciseIntoLoggedSets`) already gets this wrong for identical sets.
- **Ordering.** Lists sorted by `ID` (`PlannedMealsDisplay.vue:79` tie-break) jump when a temp id is
  replaced. You need a separate sort key anyway (`created_at`, `display_order`).
- **Id-keyed UI state.** `:key="log.ID"` remounts the row on reconcile (flicker, lost focus), and routes
  like `/gym/logging/:id` can't point at a pending row.

### (b) `bigint` PK + `client_id uuid UNIQUE` (recommended)

The client generates a UUID per create and sends it. The server does `INSERT … ON CONFLICT (client_id) DO
NOTHING` and returns the row either way, which makes creates idempotent and safe to retry. Caches key
rows by `client_id` for the pending period and keep the server `ID` for URLs, FKs and existing code.
Children reference a pending parent by the parent's `client_id` in the request, and the server resolves
it (or the per-set API takes `exercise_id` + day, which needs a unique natural key, see below).

- Cost: ~+60 kB per 1k rows per table (heap + unique index, measured). It's additive and nullable: no backfill, no FK changes, no type
  changes in Go/TS, and routes are untouched.
- Rows created before the migration have `client_id NULL` (the unique index allows many NULLs). The
  frontend keys rows by `client_id ?? 'id:' + ID`.
- Add it only where creates are optimistic: `logged_sets`, `logged_exercises` (or a natural unique key
  instead), `water_logs`, `grocery_items`, `meals`/`day_logs` (quick log), `planned_meals`, and later
  `media_items` and money transactions.

### (c) UUIDv7 as PK

The cleanest model in theory: one id from birth, no reconcile step. Here it costs the most for no
benefit. Every PK and FK column changes type across ~30 tables, GORM models `uint` → `uuid.UUID`, every
TS `number` id → `string`, routes `(\d+)` → UUID patterns, URLs get longer, and it's +25 % storage
(irrelevant but non-zero). The migration rewrites every table and FK and must ship in one deploy with
the code. It's also irreversible in practice. **Reject.**

### (d) Idempotency-Key header only

The client sends `Idempotency-Key: <uuid>` per mutation. The server stores `(key, response)` for ~24 h
and replays it. That fixes retries and double submits generically, including non-create writes, without
touching domain tables. It doesn't solve temp-id reconciliation, children of pending parents, or stable
row keys, and it needs a key-store table, middleware and cleanup. **Good complement, insufficient alone.**
With (b) it's mostly redundant: creates are idempotent via `client_id`, and the remaining writes are
already idempotent (upserts, PATCH, PUT) or can treat 404-on-delete as success.

## Decision and migration plan

No migration is added in this session. Each step lands in the module slice that needs it (TASKS.md).

1. **Gym slice (first):**
   - `NNNN_logged_exercise_unique.sql`: merge duplicate `(workout_log_id, exercise_id)` rows (move their
     sets to the lowest id, delete the rest), then
     `CREATE UNIQUE INDEX … ON logged_exercises (workout_log_id, exercise_id) WHERE deleted_at IS NULL`.
     Also add the missing FK indexes `logged_sets(logged_exercise_id)` and
     `logged_exercises(workout_log_id)` (coordinate with the performance project). Down: drop the indexes.
     The merge is irreversible, so mark it `-- migrate:down irreversible` or keep the down to the index
     drops only.
   - `NNNN_logged_sets_client_id.sql`: `ALTER TABLE logged_sets ADD COLUMN client_id uuid;` +
     `CREATE UNIQUE INDEX logged_sets_client_id_key ON logged_sets (client_id);`. Down: drop both.
   - Backend: `POST /workout/sets` `{client_id, date|workout_log_id, exercise_id, weight, reps,
     weight_setup}` upserts the parent `logged_exercise` (unique key, so no duplicate), then
     `INSERT … ON CONFLICT (client_id) DO NOTHING RETURNING`, then returns the full `LoggedExercise`. Add
     `PATCH /workout/sets/:id`. Keep `/exercises/log` for notes and the legacy path until the old flow is gone.
2. **Diet slice:** `client_id` on `meals` (quick log) and `planned_meals` (from-saved). Unique
   `day_logs(day_id, meal_id)`. `log-planned` takes `day_id`.
3. **Tracking slice:** `client_id` on `water_logs` and `grocery_items`.
4. **Later (media, money):** only if those creates become optimistic.

Client generation: add one shared `newClientId()` in `frontend/src/api/ids.ts` (v7 layout from
`crypto.getRandomValues`), as part of the foundation slice. **Don't use `crypto.randomUUID()`.** It only exists in secure contexts, so it
is `undefined` when testing on a phone over `http://<LAN-IP>:<port>`. (`useEditableMeal.ts:169,245`
already has this latent problem.)
