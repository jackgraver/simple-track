# Baseline

Record date, data volume, command used, and numbers.

## 2026-11-23: /gym and sub-pages, frontend (main @ bc8049d)

**Setup**: local production copy (backend + frontend images from their Dockerfiles, Caddy with
`encode zstd gzip` and `/api/*` → backend, `tls internal`), devseed data (small volume). Headless
Chromium 390×844 mobile, CDP slow 4G (150 ms RTT, 1.6 Mbps down) + 4× CPU throttle, empty cache.
The live site was checked read-only and serves the same build with the same headers.

| Route (cold) | FCP | LCP | CLS | Long tasks | JS (KB, compressed) | API calls |
|---|---|---|---|---|---|---|
| /gym | 1408 | 1972 | 0.090 | 353 ms | 168 | 5 |
| /gym/plans | 1396 | 2208 | 0.001 | 0 | 181 | 3 |
| /gym/plans/2/days/1 | 1404 | 2312 | 0 | 0 | 193 | 3 |
| /gym/logging | 1396 | 2244 | 0 | 0 | 187 | 5 |
| /gym/logging/:id | 1392 | 1392 | 0 | 0 | 267 | 4 |
| /gym/logging/cardio | 1412 | 2640 | 0 | 0 | 267 | 4 |
| /gym/progression | 1408 | 2320 | 0.004 | 73 ms | 247 | 5 |
| /gym/weight | 1400 | 2336 | 0.001 | 173 ms | 240 | 4 |
| /gym/steps | 1400 | 2252 | 0.001 | 0 | 236 | 4 |

In-app navigation from /gym (same throttling), ms until network idle: plans 301, plan day 400,
logging 299, exercise 614, cardio 689.

Found: first navigation waits for `/auth/me` before any route chunk or data request (one RTT on
every app open); /gym's programs + activity start only after mount (the CLS: Today card pushed down
88 px then 28 px); every gym sub-page fetches the 52-week activity; the exercise page loads chart.js
(63 KB) for a sheet that is closed; `formatDateLong` built two Intl formatters per call, ~700 calls
per heatmap render (top CPU cost on /gym); sub-page navigations wait on 6-16 chunk downloads.
Not worth it: chart.js is already tree-shaken; axios is ~13 KB.
