# Performance tasks

## Baseline
- [ ] Large-volume seed (e.g. 2 years of gym/diet/tracking history) behind a `dev db:seed-large` task
- [ ] Enable `pg_stat_statements` in dev Postgres; script to print top queries by total time
- [ ] Aggregate `benchmark.log` into per-route p50/p95; record in BASELINE.md
- [ ] Frontend: bundle visualizer report, Lighthouse mobile (throttled) for key routes, record

## Backend
- [ ] N+1 / Preload audit across modules
- [ ] EXPLAIN ANALYZE top 10 queries; add indexes via migrations
- [ ] Payload trimming and pagination for large lists
- [ ] Query-count assertions in tests for hot endpoints

## Frontend
- [ ] Route lazy-loading and heavy-lib splitting (Chart.js)
- [ ] Long-list virtualisation where needed
- [ ] Bundle-size budget check in CI

## Guards
- [ ] Document how to re-run baselines in this README
