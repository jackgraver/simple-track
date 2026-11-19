# Performance (open-ended)

**Goal**: find and fix *measured* bottlenecks in the DB, API and frontend. Every change records before/after numbers in `LOG.md`.

**Rules**
- Measure first; no change without a baseline in `BASELINE.md`.
- Timebox each session (e.g. top 3 offenders, ~2h).
- Test against realistic volume, not the small devseed set (see tasks).
- Add a regression guard whenever one is cheap (bundle budget, query-count test).

## Current state (audited 2026-11-19)
- Backend: gin + gorm on Postgres. 29 `Preload(` call sites (check for N+1 and over-fetching). Schema only in `backend/migrations/` (single `0001_baseline.sql` so far), so index changes are new migrations.
- Per-request latency is logged as JSON lines to `backend/benchmark.log` (route, method, latency_ms). It is a usable starting point but dev data is tiny, so numbers are near zero.
- Frontend: Vite 8, Vue 3, Chart.js via PrimeVue, ~14.7k lines of Vue. No bundle analysis or budget yet. Query staleTime defaults are generous (5 min).
- No `pg_stat_statements`, no EXPLAIN review, no Lighthouse/Web Vitals baseline.

## Next up / gotchas
- Build the large-data seed and take baselines before touching anything.
