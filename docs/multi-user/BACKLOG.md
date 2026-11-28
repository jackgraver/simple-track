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
| 1 | High | Request bodies bind straight into gorm models (mass assignment, nested association writes) | C2; B3 rule: **done** (`TestOwnerKitRules`, Owner/Creator hooks) |
| 2 | High | Advisory-lock keys and "exists for date" checks are global, not per user | C2, C3 (B3 helper **done**: `database.XactLock`/`LockKey`) |
| 3 | Medium | Anyone can lock the owner out of login, from one IP, indefinitely | S1: **done**; C6: lockouts in `auth_events` (`login_rate_limited`) **done** |
| 4 | Medium | IPv6 address rotation defeats per-IP limits and can fill the limiter (global lockout) | S1: **done** |
| 5 | Medium | TMDB API key can leak to users in lookup error messages | S3: **done** |
| 6 | Medium | Raw internal errors (SQL, constraint names) returned to clients at ~160 sites | C1–C5 (C1, C4, C5 **done**); B3 test **done** (`TestNoRawErrorsIn500s`) |
| 7 | Medium | No security headers (CSP, frame-ancestors, nosniff, Referrer-Policy); API has none at all | H1: **done** (Caddy; deploys at D1, ROLLOUT 3.5) |
| 8 | Medium | No request body size limit outside `/auth` (imports, JSON) | H1: **done** (1 MiB default, 5 MiB + 10k rows on import; Caddy 6MB) |
| 9 | Medium | Frontend local storage isn't scoped to the user; logout leaves drafts, timers, last route | S2: **done** |
| 10 | Medium | Tokens without `exp` never expire; 1-year session lifetime | B1: **done** (exp/iat/alg enforced; 365-day lifetime accepted by owner) |
| 11 | Low | JWT signature compared with `!=` (not constant-time) | B1: **done** |
| 12 | Low | `"Invalid token: " + err` echoes verification details | B1: **done** |
| 13 | Low | CSRF relies on SameSite=Lax only; JSON endpoints accept `text/plain` bodies | B1: **done** (Sec-Fetch-Site) |
| 14 | Low | Usernames are case-sensitive unique (`Jack` vs `jack`); no password policy; `max=72` counts runes, bcrypt limit is bytes | C6: **done** (lower(username) unique, charset, 10 chars min, 72 bytes max) |
| 15 | Low | Shared-catalog cover URLs are arbitrary image URLs shown to every user (tracking pixel) | C5: **done** (https + lookup image hosts only; legacy covers not served) |
| 16 | Low | `TRUSTED_PROXIES` depends on an unpinned Docker gateway IP; drift merges all clients into one IP | A2: deferred to D1 (needs a network recreate on the VPS) |
| 17 | Low | API responses carry no `Cache-Control`; personal data may sit in browser/bfcache on shared devices | H1: **done** |
| 18 | Low | Default CORS list (used when `CORS_ORIGINS` is unset) hard-codes home LAN IPs | H1: **done** |
| 19 | Low | Request logs include full query strings; login logs record attempted usernames | S1: login part **done**; C6: request logs **done** (route patterns) |
| 20 | Low | No Go vulnerability scan in CI | S4: manual run **done** (bumped x/net, x/text, Go 1.26); pipeline steps: B3 **done** (govulncheck + npm audit) |
| 21 | Info | Dev stack: backend listens on 0.0.0.0, `/benchmark` unauthenticated, dev/dev login on LAN | none |
| 22 | Info | Postgres `sslmode=disable` inside the compose network | none |
| 23 | Low | Owner DB password is in the backend container env (j-cicd pre_swap runs in the backend service) | j-cicd + deploy config (after D1) |
| 24 | Info | Compose file isn't shipped with releases (`.cicd.yaml` has no `deploy.compose_file`) | j-cicd + deploy config (after D1) |

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
- **B3 (done)**: `database.Owner`/`Creator` stamp `user_id`/`created_by` in a `BeforeCreate` hook
  (`json:"-"` on `user_id`); `TestOwnerKitRules` fails on a bind into an owner-carrying type, on
  writes of those fields elsewhere, and on new creates without `Omit(clause.Associations)`
  (existing sites allowlisted per file for the C lanes). The meal_route.go binds above are
  still there and become test failures once C2 adds `Creator`/`Owner` to those models.

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
- **B3 (done)**: `database.XactLock(ctx, tx, parts...)` / `database.LockKey`; raw `pg_advisory`
  calls are allowlisted only at the two sites above.
- **Since B2**: locks taken in a request are held until the response is sent (the request
  transaction), not just for the insert. With the date-only key, Alice's and Bob's first gym
  load of the same day would wait for each other's whole request. Still only the first load of
  a day; the per-user key removes it. The diet locks key on `client_id` (random per request), so
  users never share one.

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
- **Done (S1)**: strict counter per (source, username); the username-wide counter is a
  serialized backoff instead of a block, and a source that signed in to that username skips it
  (DECISIONS "S1 login protection"). Logging lockouts to `auth_events` stays with C6.

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
- **Done (S1)**: IPv6 keyed by /64 (IPv4-mapped treated as IPv4), for the per-IP limit and the
  spray rule. No more `capacity` rejection: a full limiter (now 100,000 entries) evicts
  lowest-count, oldest entries, never blocking ones or the current request's keys.

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
- **Done (S3)**: lookup errors' `Error()` is the bare sentinel; the detail (host, status,
  transport error with the URL's query and userinfo stripped, key/token params redacted) is only
  logged via `lookup.Detail`. Upstream bodies are no longer read. `respondLookupError` returns
  fixed messages (502 "Lookup service unavailable", 503 "Lookup is not configured", 400, 404).
  The v4 bearer token is a config choice (`TMDB_API_KEY=eyJ…` already goes in a header).

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
- **C4 (done)**: money handlers moved to `apierr` (`respondError`, the birth-year lookup); the
  raw-error allowlist entry is gone.
- **B3 (done)**: `internal/utils/apierr/rules_test.go` counts 500 responses with `.Error()` per
  file; lanes lower their entries. 4xx responses echoing database errors aren't caught by it.
- **C1 (done)**: all tracking handlers (weight, steps, water, grocery, profile, missed) are on
  `apierr` / `common.Respond`, including the 4xx paths that echoed database errors.

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
- **Done (H1)**: `deploy/Caddyfile` sets CSP (inline theme script by hash,
  `frontend/src/csp.test.ts` keeps them in sync), `X-Frame-Options`, nosniff, `Referrer-Policy:
  same-origin`, `Permissions-Policy`, `-Server`. Checked on a local prod copy: no CSP violations on any page.
  Reaches prod via the Caddy j-cicd project at D1, after the app (ROLLOUT 3.5).

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
- **Done (H1)**: `cmd/server/hardening.go`: 1 MiB default, per-route `bodyLimits` (5 MiB weight
  import), 413 on `Content-Length`; 10,000 days per import (`weight.MaxImportRows`); Caddy
  `request_body max_size 6MB` on `/api`.

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
- **Status (S2)**: done in `frontend/src/utils/userStorage.ts`. New storage uses `userKey(name)`
  (returns null until the user ID is known); `clearUserStorage()` runs on logout and the 401
  handler. Device-wide prefs stay global: `st:theme`, `st:chunk-reload-at`, dev-only `tune*`.
  A different user signing in clears the previous owner's keys (`st:owner`) before anything is
  read. A cold start whose `/auth/me` is 401 leaves the keys on disk but unreadable until the
  next sign-in (same user: kept; different user: cleared). Checked in a browser: login
  namespaces keys, in-app 401 clears them and the cookie.

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
  `AUTH_COOKIE_MAX_AGE_SEC`. Owner kept the 365-day lifetime (no sliding re-issue); it is set
  in prod through `AUTH_COOKIE_MAX_AGE_SEC` in `docker-compose.vps.yml`.

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
- **Done (C6)**: new usernames lowercased, `^[a-z0-9][a-z0-9._-]{1,31}$`, unique on
  `lower(username)`, login case-insensitive; passwords being set need 10+ characters and at most
  72 bytes; login treats >72 bytes as wrong.

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
- **Done (C5)**: `lookup.Service.CoverAllowed` accepts empty or https on image.tmdb.org,
  covers.openlibrary.org, `*.anilist.co`, `*.mzstatic.com` (plus the configured image bases, which
  is how e2e's http fake upstream works; production's are the https defaults). Checked on catalog
  create and edit (400). Stored covers that fail it are blanked in every response, so rows from
  before the rule are never handed to another browser. No free-form per-user cover (no second
  column): a user who wants an odd cover can't have one. CSP `img-src` (finding 7) still applies.

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
- **Status (A2, 2026-11-27)**: deferred to D1. Pinning `ipam` on a network that already exists
  makes compose want to recreate it (`docker compose down`/`up` on the VPS, a short outage),
  and that can't be tested from here. Do it at D1 alongside ROLLOUT step 3.2 (the compose file
  is copied by hand then anyway): pin the subnet/gateway the network has today (`docker network
  inspect simple-track_app-network`), recreate during the deploy window, and confirm
  `TRUSTED_PROXIES` still matches.

### 17. Low: API responses aren't marked uncacheable

- **Where**: only `/auth/*` sets `Cache-Control: no-store` (`auth_route.go:196-199`).
- **Failure**: personal data responses have no caching directive. Browsers won't usually cache
  them without validators, but the back/forward cache and any intermediary may keep them; on a
  shared device that's someone else's data.
- **Fix**: `Cache-Control: no-store` (or `private, no-cache`) on every `/api` response, in the
  backend or in Caddy's `/api` handler.
- **Lane**: new (H1).
- **Done (H1)**: `Cache-Control: no-store` on every backend response, set in
  `cmd/server/hardening.go`.

### 18. Low: default CORS origins include home LAN IPs

- **Where**: `backend/cmd/server/main.go:44`.
- **Failure**: production must set `CORS_ORIGINS` (`production_config.go:27`), so prod is safe.
  Any non-production deployment that forgets it allows credentials from two fixed LAN addresses.
  `ExposeHeaders: Set-Cookie` (`:53`) does nothing (browsers never expose it) and suggests the
  cookie is meant to be script-visible. `dev/env.sh` sets `CORS_ORIGINS` itself, so the default
  is unused in dev too.
- **Fix**: default to no extra origins (same-origin only) and drop `ExposeHeaders`.
- **Lane**: new (H1).
- **Done (H1)**: no default list; with `CORS_ORIGINS` unset only same-origin requests pass (other
  origins get 403). `ExposeHeaders` removed.

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
- **Login part done (S1)**: failed-login logs carry the username only if it exists, otherwise
  `unknown`.
- **Request part done (C6)**: `cmd/server/requestlog.go` logs `c.FullPath()` only, never the
  raw path or query. `auth_events` never stores attempted usernames either.

### 20. Low: no Go vulnerability scan

- **Where**: `.cicd.yaml` runs `go vet` and tests only. `npm audit` is clean today (0 findings,
  2026-11-27). `govulncheck` wasn't run during this audit (the tool fetch was blocked in this
  session), so Go module status is unchecked. Direct deps are recent (gin 1.11.0, x/crypto
  0.52.0, pgx 5.9.2, gorm 1.31.0).
- **Fix**: add a `govulncheck ./...` step (and `npm audit --omit=dev --audit-level=high`) to the
  pipeline next to `backend-test`. Run it once by hand now.
- **Lane**: B3 (already editing the CI test step).
- **Done (B3)**: `backend-vulncheck` (govulncheck v1.10.0, the version S4 audited with) and
  `frontend-audit` steps in `.cicd.yaml`. The dependency bump is S4's (below).

**S4 audit result (2026-11-27)**: govulncheck v1.10.0 (DB updated 2026-11-24), go1.27.4 locally,
node v22.23.5 / npm 10.9.9.

- Before: 8 reachable findings, all from two modules. `golang.org/x/net@v0.55.0`: GO-2026-6617,
  6612, 6611, 6610, 6603 (HTTP/2, fixed in v0.60.0; reached via `media/lookup` http.Client) and
  GO-2026-5026 (idna, fixed v0.57.0). `golang.org/x/text@v0.37.0`: GO-2026-6629 (precis, fixed
  v0.41.0) and GO-2026-5970 (norm, fixed v0.39.0; reached via pgx and gorm). Imported but not
  called: x/net GO-2026-6866, 6860 (fixed v0.62.0). Required modules only: x/crypto
  GO-2026-6355, 6354 (ssh, fixed v0.56.0), 6303 (ssh, v0.55.0), x/net GO-2026-5942 (dnsmessage,
  v0.56.0), x/crypto GO-2026-5932 (openpgp, only fixed in a pseudo-version). The Go stdlib was
  not flagged.
- Bumped (backend/go.mod): x/net v0.55.0 to v0.60.0, x/text v0.37.0 to v0.42.0 (v0.60.0
  requires it), x/crypto v0.52.0 to v0.57.0 and x/sync, x/sys (forced by x/net v0.60.0). x/net
  v0.60.0 requires Go 1.26, so the `go` directive went 1.25.0 to 1.26.0 and
  `backend/Dockerfile` went `golang:1.25-alpine` to `golang:1.26-alpine` (image build checked
  locally; pipeline steps run on the host's Go 1.27).
- After: `govulncheck ./...` reports 0 reachable vulnerabilities.
- Left: x/net GO-2026-6866 and 6860 (imported, not called; fix v0.62.0 needs x/crypto v0.58.0
  and x/text v0.43.0), x/crypto GO-2026-5932 (openpgp is unused; no tagged fix). Re-check with
  the B3 pipeline step.
- npm: `npm audit --omit=dev` and full `npm audit` both report 0 vulnerabilities in `frontend/`
  and `e2e/`. Nothing bumped.
- Pipeline step (govulncheck and npm audit in `.cicd.yaml`) is still open and stays with B3.

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

### 23. Low: the owner DB password is in the server's environment

- **Where**: `docker-compose.vps.yml` backend `DATABASE_OWNER_URL`; j-cicd agent `pre_swap`.
- **Why**: j-cicd runs `pre_swap` (`/app/migrate up`) only inside a tracked compose service, so
  migrate has to run in the `backend` service and gets the server's environment. A2 split the
  roles (the server logs in as `simpletrack_app`), but code running inside the server process
  could still read the owner DSN and bypass RLS. The split protects against query bugs, not a
  compromised server.
- **Fix**: let j-cicd run `pre_swap` in an untracked compose service that uses the new backend
  image (e.g. a `migrate` service with `image: ${BACKEND_IMAGE}` and only
  `DATABASE_OWNER_URL`), then remove `DATABASE_OWNER_URL` from `backend`. Changes in j-cicd
  (`internal/agent/preswap.go` validation), the VPS agent config and `docker-compose.vps.yml`.
- **Lane**: j-cicd + deploy config, after D1. Recorded by A2 (DECISIONS 2026-11-27).

### 24. Info: the compose file isn't shipped with releases

- **Where**: `.cicd.yaml` (no `deploy.compose_file`).
- **Why**: the VPS agent uses its own copy of `docker-compose.vps.yml`, so compose changes reach
  prod only by hand, and a rollback restores images but not the compose file that went with
  them. D1 has to work around this (ROLLOUT steps 3.2 and 4).
- **Fix**: set `deploy.compose_file: docker-compose.vps.yml` so each deploy sends the file from
  the same commit and rollbacks restore it. Check the VPS agent version supports it, then do one
  deploy with it on a quiet day. It changes how every deploy works, so it's the owner's call.
- **Lane**: j-cicd + deploy config, after D1. Would drop ROLLOUT's manual copy and rollback caveats.
