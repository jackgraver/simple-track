# Performance tasks

## Baseline
- [ ] Large-volume seed (e.g. 2 years of gym/diet/tracking history) behind a `dev db:seed-large` task
- [ ] Enable `pg_stat_statements` in dev Postgres; script to print top queries by total time
- [ ] Aggregate `benchmark.log` into per-route p50/p95; record in BASELINE.md
- [ ] Frontend: bundle visualizer report, Lighthouse mobile (throttled) for key routes, record
  - [x] /gym and sub-pages, throttled mobile, via local production copy (2026-11-23, BASELINE.md)
  - [ ] Other modules (diet, finances, media)

## Backend
- [ ] N+1 / Preload audit across modules
- [ ] EXPLAIN ANALYZE top 10 queries; add indexes via migrations
- [ ] Payload trimming and pagination for large lists
- [ ] Query-count assertions in tests for hot endpoints

## Frontend
- [ ] Route lazy-loading and heavy-lib splitting (Chart.js)
  - [x] Gym: chart panel on demand, idle preload of gym sub-page chunks, framework chunk group (2026-11-23)
- [ ] Start route chunk + data loads in parallel with the first `/auth/me` (render still gated on auth)
- [ ] Long-list virtualisation where needed
- [ ] Bundle-size budget check in CI

## Guards
- [ ] Document how to re-run baselines in this README
