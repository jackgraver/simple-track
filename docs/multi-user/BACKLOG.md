# Security backlog

Findings from the A3 audit (2026-11-27, read-only, `proj/multi-user` at `8115e1e`, after A1). Each entry
has a severity, where it is, what goes wrong, a suggested fix and the lane that should own it.
Lanes pick up their entries and tick them here in the PR that fixes them. D1 checks that every
**High** is done before the deploy.

Not repeated here because the plan already covers them: missing `user_id` scoping and RLS, the
global `conn()` handle, pool defaults and the owner DB role (A2), server timeouts and graceful
shutdown (A2), JWT revocation via `token_version`/`disabled_at` (B1), admin-only catalog deletes,
per-user `client_id` uniques. Where the plan covers only part of a problem, the missing part is
listed below.

Severity: **Critical** = exploitable in prod now for data or account takeover. **High** = must
be fixed before a second account exists (D1). **Medium** = real risk, fix during the project.
**Low** = hardening. **Info** = noted, no action needed unless it changes.

Nothing Critical was found. Prod today has one account, so most multi-user findings are latent.

## Summary

| # | Sev | Finding | Lane |
|---|---|---|---|
| 1 | High | Request bodies bind straight into gorm models (mass assignment, nested association writes) | C2 (+B3 rule) |
| 2 | High | Advisory-lock keys and "exists for date" checks are global, not per user | C2, C3 |
| 3 | Medium | Anyone can lock the owner out of login, from one IP, indefinitely | C6 |
| 4 | Medium | IPv6 address rotation defeats per-IP limits and can fill the limiter (global lockout) | C6 |
| 5 | Medium | TMDB API key can leak to users in lookup error messages | C5 |
| 6 | Medium | Raw internal errors (SQL, constraint names) returned to clients at ~160 sites | C1–C5, B3 (lint) |
| 7 | Medium | No security headers (CSP, frame-ancestors, nosniff, Referrer-Policy); API has none at all | new (H1 headers) |
| 8 | Medium | No request body size limit outside `/auth` (imports, JSON) | B2 or new (H1) |
| 9 | Medium | Frontend local storage isn't scoped to the user; logout leaves drafts, timers, last route | C6 (frontend) |
| 10 | Medium | Tokens without `exp` never expire; 1-year session lifetime | B1: **partly done** (exp/iat/alg enforced; lifetime is the owner's call) |
| 11 | Low | JWT signature compared with `!=` (not constant-time) | B1: **done** |
| 12 | Low | `"Invalid token: " + err` echoes verification details | B1: **done** |
| 13 | Low | CSRF relies on SameSite=Lax only; JSON endpoints accept `text/plain` bodies | B1: **done** (Sec-Fetch-Site) |
| 14 | Low | Usernames are case-sensitive unique (`Jack` vs `jack`); no password policy; `max=72` counts runes, bcrypt limit is bytes | C6 |
| 15 | Low | Shared-catalog cover URLs are arbitrary image URLs shown to every user (tracking pixel) | C5 |
| 16 | Low | `TRUSTED_PROXIES` depends on an unpinned Docker gateway IP; drift merges all clients into one IP | A2 |
| 17 | Low | API responses carry no `Cache-Control`; personal data may sit in browser/bfcache on shared devices | new (H1) |
| 18 | Low | Default CORS list (used when `CORS_ORIGINS` is unset) hard-codes home LAN IPs | new (H1) |
| 19 | Low | Request logs include full query strings; login logs record attempted usernames | C6 |
| 20 | Low | No Go vulnerability scan in CI | B3 |
| 21 | Info | Dev stack: backend listens on 0.0.0.0, `/benchmark` unauthenticated, dev/dev login on LAN | none |
| 22 | Info | Postgres `sslmode=disable` inside the compose network | none |

## Findings

### 1. High: request bodies bind into gorm models (mass assignment)

- **Where**: `backend/internal/core/diet/controller/meal_route.go:194` (`PostNewCompositeFood`
  binds `models.CompositeFood`), `:236` (`PostNewSavedMeal` binds `models.SavedMeal`), `:318`
  (`PutSavedMeal`); persisted by `diet/repository/repository.go:199` (`CompositeFoodCreate`),
  `:239` (`SavedMealCreate`) and `:279` (`SavedMealReplace`) with plain `Create`.
- **Failure**: the client controls every JSON-visible field of the model and its associations.
  Items carry a nested `food` object (`SavedMealItem.Food`, `CompositeFoodItem.Food`), and gorm's
  default `Create` saves associations: an item with `"food": {"name": "x", "calories": 1}` inserts
  a new row in the shared `foods` catalog, skipping `PostFood`'s validation and (after C2) the
  `created_by` stamp. `gorm.Model` fields (`CreatedAt`, `DeletedAt`) are bindable too. Once
  models gain `UserID`, a body with `"user_id": <other>` sets it unless every create path
  overwrites it. RLS `WITH CHECK` would reject a foreign `user_id`, but catalog `created_by` has
  only the policy the C lane writes.
- **Fix**: bind into request DTOs (as the rest of diet already does with `CreateFoodRequest`
  etc.) and copy the allowed fields. Create with `Omit(clause.Associations)` and reference foods
  by `food_id` only. B3's "stamp `UserID` on create" helper should be the only thing that sets
  `user_id`/`created_by`. Add a B3 rule: no `ShouldBindJSON` into a type that has a `UserID` or
  `CreatedBy` field (a grep test is enough).
- **Lane**: C2 for the diet handlers; B3 for the rule/test.

### 2. High: advisory locks and per-date checks are global

- **Where**: `diet/repository/repository.go:749` (`lockClientID` keys on
  `table + ".client_id:" + clientID`), `workout/repository/workoutlog_repository.go:39-50`
  (`CreateMinimalIfMissing` locks `workout_logs.date:<date>` and counts logs `WHERE date = ?`).
  Same pattern anywhere else a `pg_advisory_xact_lock` or a "does a row exist for this date /
  client_id" check runs (`repository.go:774`, `:865` use `Unscoped()` counts by `client_id`).
- **Failure**: the plan scopes `client_id` uniques and lookups by user, but the lock keys aren't
  mentioned. With global keys, two users serialize on each other (same date: every user's first
  set of the day contends on one lock). Worse, `CreateMinimalIfMissing`'s count runs under RLS
  only if B2 is in place; with an unscoped count, Bob's first workout of the day is skipped
  because Alice already has a log for that date (the function returns "exists" and leaves Bob's
  log without an ID). `Unscoped()` counts also see soft-deleted rows, so the scope helper must
  still apply there.
- **Fix**: include the user ID in every advisory-lock key (`fmt.Sprintf("%s:%d:%s", table, uid,
  key)`, or the two-int form `pg_advisory_xact_lock(uid, hashtext(key))`) and filter every
  existence check by user, including `Unscoped()` ones. Add a two-user test: Alice and Bob both
  start a workout on the same date.
- **Lane**: C2 (diet), C3 (workout); B3 should provide a `LockKey(ctx, ...)` helper.

### 3. Medium: login lockout of a known username (exploitable now)

- **Where**: `backend/internal/core/auth/controller/login_protection.go:27` (`MaxPerUsername`
  5 per 15 min), `:116-125` (`Allow` blocks on the username counter regardless of IP).
- **Failure**: 5 wrong passwords for `jack` from any one IP block every login for `jack` for 15
  minutes, including the owner's correct password from their own phone. Repeating once per
  window keeps the account locked indefinitely from a single IP, well under the per-IP limit (30)
  and the spray rule (5 distinct usernames). Existing sessions keep working (1-year cookie), so
  it's an availability problem, not a breach. With family accounts, every username is a target.
- **Fix**: don't let failures from one source lock out everyone. Options: key the strict counter
  on `(ip, username)` and keep a much looser username-wide counter with growing delays instead of
  a hard block; or exempt clients that hold a valid "known device" cookie set on a previous
  successful login. Log lockouts to `auth_events` (C6) so the admin can see them.
- **Lane**: C6 (owns login protection per the plan).

### 4. Medium: IPv6 rotation defeats per-IP limits

- **Where**: `login_protection.go:102`, `:157` (keys use the full client IP), `:134-145`
  (`maxLoginRateEntries` = 10000, then every login is rejected with reason `capacity`).
- **Failure**: anyone with an IPv6 /64 (most VPS and many home connections) has 2^64 addresses.
  Rotating them gives unlimited per-IP budgets and never trips the spray rule. Creating 10000
  entries (two per new IP+username pair) fills the limiter and rejects **all** logins for up to
  a window. Whether this works today depends on Caddy answering on IPv6 (an AAAA record for
  `st.jackgraver.ca`); worth checking.
- **Fix**: normalize IPv6 clients to their /64 before keying. Evict oldest entries instead of
  rejecting everyone when full, or reject only new unknown keys.
- **Lane**: C6.

### 5. Medium: TMDB API key can leak in lookup errors

- **Where**: `backend/internal/core/media/lookup/tmdb.go:24` (v3 key goes in the query string),
  `lookup.go:139-142` (`fmt.Errorf("%w: %v", ErrUpstream, err)` wraps the `*url.Error`, whose
  text includes the full request URL), `media/controller/media_route.go:179`
  (`respondLookupError` returns `err.Error()` to the client). `lookup.go:150-152` also echoes up
  to 300 bytes of the upstream error body.
- **Failure**: any network error or timeout on a TMDB call returns
  `lookup source request failed: Get "https://api.themoviedb.org/3/search/movie?api_key=…&query=…": context deadline exceeded`
  to the browser. Today only the owner sees it; with family accounts, every user can read the key
  (and it goes into the server log too). It doesn't apply if the configured key is a v4 bearer
  token (`eyJ…`), which goes in a header.
- **Fix**: return fixed messages to the client ("Lookup service unavailable") and log the detail
  with the query string redacted (`url.Error.URL` stripped, or log `req.URL.Redacted()`-style
  with `api_key` removed). Prefer the v4 bearer token in prod.
- **Lane**: C5.

### 6. Medium: raw internal errors returned to clients

- **Where**: about 160 handler sites return `gin.H{"error": err.Error()}` for 500s and
  service errors, e.g. `diet/controller/diet_log_route.go:17,34,51`,
  `tracking/water/routes.go:42,107`, `tracking/grocery/routes.go:34,128,147,166`,
  `tracking/weight/routes.go:27,74`, `workout/controller/workoutlog_route.go:16,36,70,117`,
  `workout/controller/workoutplan_route.go:27,102,140,146`, `media/controller/media_route.go:145`,
  `money/controller/routes.go:454` (`respondError` default), `auth/controller/auth_route.go:179`.
- **Failure**: Postgres errors reach the browser verbatim: table, column and constraint names,
  SQL fragments, and after RLS, policy-violation text. With other users on the system this
  becomes a way to probe whether rows exist (a unique violation on a shared catalog name, a FK
  error that differs between "missing" and "not yours"). The plan's 404-not-403 rule is
  undermined if the error text differs.
- **Fix**: `internal/apierr` already exists (`apierr.Internal` logs and returns a fixed message;
  diet's `meal_route.go` uses it). Each C lane moves its module's handlers to `apierr` and keeps
  `err.Error()` only for errors it created itself (validation sentinels like `ErrInvalid`).
  B3 adds a grep test that fails on new `"error": err.Error()` in `StatusInternalServerError`
  responses.
- **Lane**: C1 (tracking), C2 (diet), C3 (workout), C4 (money), C5 (media); B3 for the test.

### 7. Medium: no security headers

- **Where**: `frontend/nginx.conf` (only HSTS and cache headers), `deploy/Caddyfile` (no `header`
  directive), backend (no headers middleware in `cmd/server/main.go`).
- **Failure**: no `Content-Security-Policy`, so any future XSS (or a compromised dependency)
  can exfiltrate data freely; no `frame-ancestors`/`X-Frame-Options`, so the app can be framed
  for clickjacking (delete buttons, the coming admin "create user / reset link" screen); no
  `X-Content-Type-Options: nosniff`; no `Referrer-Policy` (the coming `/setup/<token>` URL would
  be sent as a referrer to any external image or link on that page); no `Permissions-Policy`.
- **Fix**: set them once in Caddy for the whole site:
  `Content-Security-Policy` (start with `default-src 'self'; img-src 'self' https: data:;
  style-src 'self' 'unsafe-inline'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'`
  and tighten after checking the build output), `X-Content-Type-Options nosniff`,
  `Referrer-Policy no-referrer` (or `same-origin`), `Permissions-Policy` denying camera/mic/geo.
  Add `-Server` to drop the version banner. Ship with the Caddy j-cicd project; test on the
  local prod copy first.
- **Lane**: new (H1 "HTTP hardening": headers, body limits, cache headers; small, can run any
  time; must land before C6's setup page ships).

### 8. Medium: no request body size limit outside `/auth`

- **Where**: `auth/controller/auth_route.go:31,80` are the only `MaxBytesReader` calls.
  Everything else binds unbounded JSON, e.g. `tracking/weight/routes.go:61-77` reads a whole
  CSV string, `ParseRenphoCSV` (`weight/import.go:73`) `ReadAll`s it, and `ImportBodyWeights`
  (`:158`) upserts every row in one transaction. Caddy sets no `request_body max_size`.
- **Failure**: one signed-in user can post a multi-hundred-MB body and push the backend into
  OOM or hold a long transaction, taking the app down for everyone. Pre-auth requests are safe
  (the auth middleware aborts before the body is read).
- **Fix**: a global middleware with a sensible default (e.g. 1 MiB) and a larger explicit limit
  on the import route (e.g. 5 MiB, plus a cap on row count), and `request_body { max_size … }`
  under `handle_path /api/*` in Caddy as the outer guard.
- **Lane**: new (H1), or B2 since it adds request middleware anyway.

### 9. Medium: frontend storage isn't per-user

- **Where**: `frontend/src/composables/auth/useAuth.ts:92-103` (logout clears the query cache
  only); persisted state that survives logout: `utils/lastRoute.ts:33-43` (localStorage and a
  7-day cookie), `composables/useGlobalRestTimer.ts:32`,
  `pages/gym/logging/store/useWorkoutStore.ts:41` (hidden exercises per log),
  `pages/diet/logmeal/useLogMealDraftStorage.ts:168` and
  `pages/gym/logging/exercise/composables/useExerciseLoggingSession.ts:227` (sessionStorage
  drafts). The session-expired path (`main.ts:30-38`) also only clears the query cache.
- **Failure**: on a shared family tablet, Bob signs in after Alice signs out and gets Alice's
  half-written meal draft, her rest timer, and is resumed into her last route (e.g.
  `/finances/accounts/12`, which then 404s for him but reveals the path). Draft contents are
  Alice's data shown to Bob.
- **Fix**: namespace every stored key by user ID (from `/auth/me`) and, on logout or when the
  signed-in user differs from the stored owner, remove all `st:*` keys and the `st_last_route`
  cookie. One helper (`clearUserStorage()`) called from logout and the 401 handler.
- **Lane**: C6 (it owns the frontend auth/account flow); each C lane uses the namespaced helper
  for new storage.

### 10. Medium: tokens without `exp` never expire; 1-year sessions

- **Where**: `backend/internal/core/auth/jwt.go:86` (`if claims.Exp > 0 && …`: a token with no
  `exp` is valid forever), `auth/cookie.go:12` (default lifetime 365 days; prod sets no
  `AUTH_COOKIE_MAX_AGE_SEC`).
- **Failure**: any token minted before `exp` was added (or by a future bug) never expires.
  B1 rejects tokens without `uid`, which happens to kill old ones once, but `exp` should be
  required on its own. A stolen cookie is good for a year; `token_version` fixes revocation but
  only if someone notices.
- **Fix**: require `exp` (and `iat`), reject `exp - iat` larger than the configured max. Consider
  a shorter absolute lifetime (e.g. 30–90 days) with sliding re-issue on activity, so an idle
  stolen token dies on its own. Also verify the header's `alg` is `HS256`.
- **Lane**: B1.
- **Status (B1)**: `exp`, `iat` and `alg: HS256` are required and `exp - iat` can't exceed
  `AUTH_COOKIE_MAX_AGE_SEC`. Still open, for the owner: a shorter default lifetime (365 days
  today) and sliding re-issue. Both change how often family members sign in.

### 11. Low: non-constant-time JWT signature check

- **Where**: `auth/jwt.go:72` (`signature != expectedSignature`).
- **Failure**: string comparison leaks timing. Exploiting HMAC timing over the internet is
  impractical, but it's a one-line fix in code B1 is rewriting anyway.
- **Fix**: `hmac.Equal([]byte(signature), []byte(expectedSignature))`.
- **Lane**: B1.

### 12. Low: middleware echoes token errors

- **Where**: `auth/auth_middleware.go:29` (`"Invalid token: " + err.Error()`).
- **Failure**: tells a caller whether a forged token failed on format, signature or expiry, and
  will echo whatever B1's new checks return (e.g. user lookup errors). Already a seed in TASKS.
- **Fix**: fixed `"Authentication required"` for every failure; log the reason server-side.
- **Lane**: B1.

### 13. Low: CSRF protection is SameSite=Lax only

- **Where**: `auth/cookie.go:18-27` (Lax), every handler using `ShouldBindJSON` (gin decodes
  the body as JSON whatever the `Content-Type`).
- **Failure**: Lax stops cross-*site* POSTs, but any page on a sibling `*.jackgraver.ca`
  subdomain is same-site and its requests carry the cookie. A `text/plain` form POST (no CORS
  preflight) with a JSON body would be accepted. Risk depends on what else lives under
  `jackgraver.ca`; it grows if anything there ever serves user content.
- **Fix**: in the auth middleware, reject state-changing requests whose `Origin` (or
  `Sec-Fetch-Site`) isn't the app's origin, or require `Content-Type: application/json`, which
  forces a preflight cross-origin. Either is a few lines.
- **Lane**: B1 (it rewrites the middleware).

### 14. Low: username and password rules

- **Where**: `auth/models/user_model.go:8` (`uniqueIndex` on `username`, case-sensitive),
  `auth/repository/user_repository.go:5` (exact match lookup),
  `auth/controller/auth_route.go:74-75` (`max=72` counts characters; bcrypt's limit is 72
  bytes), no minimum length anywhere.
- **Failure**: the admin can create `Jack` next to `jack` (impersonation within the family; the
  rate limiter already treats them as one key). A 72-character non-ASCII password passes
  validation and then fails to hash (500) or is truncated depending on the bcrypt path. Users can
  set a one-character password on the setup page.
- **Fix**: C6's create-user and setup endpoints: lowercase usernames (or a unique index on
  `lower(username)`) with a restricted charset; password minimum (e.g. 10) and a byte-length
  check of 72.
- **Lane**: C6.

### 15. Low: arbitrary cover URLs in the shared media catalog

- **Where**: `media/service/media_service.go:162` (`cover_url` stored as given, no scheme/host
  check), rendered by `frontend/src/pages/media/index.vue:158-163` and `[id].vue:313-314`.
- **Failure**: once the media catalog is shared (C5), a cover URL set by one user loads in every
  other user's browser, so its host learns their IP and when they browse media. `referrerpolicy`
  is already `no-referrer`, which helps.
- **Fix**: accept only `https:` and, for the catalog part, only URLs that came from a lookup
  source's image host (TMDB, Open Library, AniList, iTunes); a free-form URL stays on the user's
  own `user_media` row. A CSP `img-src` list (finding 7) enforces the same thing in the browser.
- **Lane**: C5.

### 16. Low: `TRUSTED_PROXIES` relies on an unpinned gateway IP

- **Where**: `docker-compose.vps.yml` `networks.app-network` (no `ipam`), documented in
  `docs/ARCHITECTURE.md` "Client IPs and `TRUSTED_PROXIES`".
- **Failure**: if Docker recreates the network with a different subnet, Caddy's requests arrive
  from an untrusted address and every client shares one IP. Then one person trying 5 usernames
  trips the spray block and locks **everyone** out for 24 h (`LOGIN_SPRAY_BLOCK_SEC`).
- **Fix**: pin `ipam.config.subnet`/`gateway` for `app-network` and set `TRUSTED_PROXIES` to
  match, or have the backend fail its startup check in production when the peer of the first
  request isn't trusted. A2 is already editing the compose file.
- **Lane**: A2.

### 17. Low: API responses aren't marked uncacheable

- **Where**: only `/auth/*` sets `Cache-Control: no-store` (`auth_route.go:196-199`).
- **Failure**: personal data responses have no caching directive. Browsers won't usually cache
  them without validators, but the back/forward cache and any intermediary may keep them; on a
  shared device that's someone else's data.
- **Fix**: `Cache-Control: no-store` (or `private, no-cache`) on every `/api` response, in the
  backend or in Caddy's `/api` handler.
- **Lane**: new (H1).

### 18. Low: default CORS origins include home LAN IPs

- **Where**: `backend/cmd/server/main.go:44`.
- **Failure**: production must set `CORS_ORIGINS` (`production_config.go:27`), so prod is safe.
  Any non-production deployment that forgets it allows credentials from two fixed LAN addresses.
  `ExposeHeaders: Set-Cookie` (`:53`) does nothing (browsers never expose it) and suggests the
  cookie is meant to be script-visible. `dev/env.sh` sets `CORS_ORIGINS` itself, so the default
  is unused in dev too.
- **Fix**: default to no extra origins (same-origin only) and drop `ExposeHeaders`.
- **Lane**: new (H1).

### 19. Low: logs carry search terms and attempted usernames

- **Where**: `cmd/server/main.go:38` (`gin.Default()` logs every request path with its query
  string, e.g. `/media?q=…`, `/tracking/grocery/suggestions?q=…`),
  `login_protection.go:206-214` (logs the attempted username on every failure).
- **Failure**: users' searches end up in the server log. People sometimes type their password
  into the username field, which then gets logged in clear.
- **Fix**: log the route pattern (`c.FullPath()`) instead of the raw query; for failed logins
  log the username only if it exists in `users`, otherwise a hash or `unknown`. `auth_events`
  (C6) should follow the same rule.
- **Lane**: C6.

### 20. Low: no Go vulnerability scan

- **Where**: `.cicd.yaml` runs `go vet` and tests only. `npm audit` is clean today (0 findings,
  2026-11-27). `govulncheck` wasn't run during this audit (the tool fetch was blocked in this
  session), so Go module status is unchecked. Direct deps are recent (gin 1.11.0, x/crypto
  0.52.0, pgx 5.9.2, gorm 1.31.0).
- **Fix**: add a `govulncheck ./...` step (and `npm audit --omit=dev --audit-level=high`) to the
  pipeline next to `backend-test`. Run it once by hand now.
- **Lane**: B3 (already editing the CI test step).

### 21. Info: dev stack exposure

`dev/env.sh:71` binds the backend to `0.0.0.0`, vite runs with `host: true`, the dev account is
`dev`/`dev`, and `GET /benchmark` (`internal/utils/benchmarker.go:108`) is registered without
auth outside production. Anyone on the home LAN can read dev data and route timings. Dev data
only, and the bypass cookie is random. Fine as long as dev DBs never hold a prod copy; if the
D1 rehearsal restores a prod backup into a dev slot, bind that stack to localhost or stop it
when done.

### 22. Info: Postgres without TLS on the compose network

`docker-compose.vps.yml` sets `sslmode=disable` with `DATABASE_ALLOW_INSECURE=true`. Postgres
publishes no port and is reachable only on the bridge network, so this is acceptable. A2's app
role should keep the same reachability.
