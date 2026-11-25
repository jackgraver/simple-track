# e2e backlog (temporary)

Scratch list of bugs and follow-ups the B-scenarios surfaced, to address once the e2e suite is better
established. Delete items as they are fixed or turned into TASKS.md entries; delete this file when empty.
Details and reasoning live in the Findings section of [TASKS.md](TASKS.md).

## App bugs (found by specs, not fixed)

- [ ] **Finances: deposit + balance is not atomic.** Two client requests (`POST .../transactions`, then
      `PUT .../balances`). If the second fails the transaction is saved, the balance is stale and a retry
      double-counts. Fix: one server endpoint in a DB transaction, then drop the `test.fail` in
      `tests/finances/transactions.spec.ts`.
- [ ] **Gym reps slider races its smooth scroll.** A tap on 9 sometimes saved 11 (`NumberSlider`
      `syncingScroll` clears after one frame). The spec waits for scrolling to stop as a workaround.
- [ ] **`POST /diet/foods` with a duplicate name returns 500** (`idx_foods_name`; `PostFood` maps every
      error to internal). Decide the status (409?), then pin it in a spec.
- [ ] **Diet meal edit is create-then-delete** (`replace_meal_id`). A failure between the two could lose or
      duplicate the meal. Needs a fault test, then probably a single server call.
- [ ] **Gym: switching a day's plan returns 400 with the raw DB error.** `controller.SwitchPlan` maps every
      error to 400 and puts `err.Error()` in the body, so the toast shows a Postgres message instead of
      "Couldn’t switch the plan", and 5xx monitoring misses it. Fix: 404 for a missing plan, 500 otherwise;
      then drop the `test.fail` and `allowedErrors` in `tests/gym/home.spec.ts`.

## Product decisions needed

- [ ] **Logout button.** None exists (`useAuth().logout` is never called from a component); the auth spec
      posts `/api/auth/logout`. Add one, then drive it through the UI.
- [ ] **Grocery delete control.** Only the API can delete; the spec arranges via API. Intended?
- [ ] **No edit-transaction** in finances. Intended?

## Test-suite follow-ups

- [ ] Extend `aria-busy` on pending rows (grocery has it via `ListRow`) to gym sets, diet and finances, so
      those specs can stop using `waitForApi` before reloading.
- [ ] `Checkbox.vue`: the slot replaces the `label` prop, so the accessible name is "Milk", not
      "Complete Milk". Fix the component or accept it.
- [ ] Media fakes for TMDB, OpenLibrary and iTunes (only AniList is faked); cover images are not faked
      (hermetic routing blocks them).
- [ ] Uncovered scenarios: settings profile form, register, rate-limit lockout; grocery paste dialog,
      suggestions and offline complete; diet catalogue-food log, planned-meal log, day offset, water,
      fault/retry, targets, saved meals; gym scenario todos listed under B2 in TASKS.md.
- [ ] Rollback-with-Retry is only covered by the grocery-add example; add it for gym, diet and finances.
- [ ] Cleanup bug seen once: "missing template DB" errors from an `e2e-baseline-f8` run; a rerun passed.
