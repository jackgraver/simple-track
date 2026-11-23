# Log

Newest first. Format: date, change, metric before -> after, commit.

## 2026-11-23: /gym fast paths (branch perf/gym-audit)

Same setup as the 2026-11-23 baseline; A/B against main, 3 alternating runs each.
- Prefetch programs + 52-week activity with the gym day (route `queries`): /gym CLS 0.090 -> 0.001.
- Cached Intl formatters in `utils/dateUtil.ts`: /gym long tasks 353 -> ~190 ms (4x CPU);
  `formatDateLong` gone from the profile (was 173 ms self time throttled, 45 ms unthrottled).
- Week/heatmap queries only on the gym home (`GymHome.vue`): gym sub-pages make 2 fewer API calls
  on a cold load (e.g. /gym/plans 3 -> 2, /gym/logging 5 -> 3).
- Progression chart sheet loads on open (`defineAsyncComponent`): exercise page JS 267 -> 200 KB.
- Idle preload of logging / exercise / plans / plan-day chunks from the gym home: navigation from
  /gym to plans 301 -> 40 ms, plan day 400 -> 46, logging 299 -> 35, exercise 614 -> 14, cardio 689 -> 34.
  /gym itself downloads ~65 KB more JS after it has rendered (idle, skipped on Save-Data).
- `framework` chunk group (vue, vue-router, @tanstack, axios) in `vite.config.ts`: entry JS 125 -> 119 KB,
  stable across lazy-import changes (without it the async chart panel split the entry into 2 more files).
- Cost: FCP +25-35 ms on every route under 4x CPU throttle (download finishes at the same time; the
  extra is script startup). LCP +20-40 ms except /gym/logging/cardio (-300 ms).

