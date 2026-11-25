# e2e notes: diet targets lane (`/diet/targets`, `/diet/water-presets`)

Specs: `tests/diet/targets.spec.ts`, `tests/diet/water-presets.spec.ts`. API helpers (end of
`fixtures/api.ts`, `// diet targets`): `saveProfile`, `profile`, `dietDay`, `dietPlans`,
`createDrinkSize`, `drinkSizes`. All tests are tagged `@diet`, desktop and mobile.

## Scenarios

| # | Scenario | Status | Spec |
|---|---|---|---|
| 1a | Change calories + split (calories field, protein, fat, fiber; carbs auto-fills), Save, reload: values persist; History shows the new snapshot as "Current", effective today, old plan still listed (2 plan rows in the API) | done | targets: "change calories and split, save..." |
| 1b | Goal segment (-500) sets calories from the formula TDEE (needs profile + a weigh-in), Save, reload: value persists and the -500 chip is pressed again | done | targets: "goal segment sets calories..." |
| 2 | Today's `/diet` summary uses the new targets (kcal, protein, fat); past days keep the old plan, both an already-existing day (offset 3) and one created lazily after the save (offset 5) | done | targets: "today's /diet summary..." |
| 3 | "Unsaved changes" badge, Save disabled until dirty, Reset restores every field | done | targets: "unsaved changes..." |
| 4 | Save failure (`plans` create fault): "Could not save targets", edits stay dirty, nothing saved (still 1 plan row, today's plan unchanged), a retry succeeds | done | targets: "a failed save..." |
| 5 | Load error + Retry (fault on `days` query) | done | targets: "load error" |
| 6 | Water sizes: add, edit (incl. Cancel), delete (incl. Cancel); persists after reload and appears/updates/disappears as a quick-add button on `/diet` | done | water-presets: "add / edit / delete a size" |
| 7 | Validation: add (empty name, empty/0/negative/non-numeric amount) and edit (blank name, amount 0) toast and change nothing; bad daily goal toasts and restores the field | done | water-presets: "validation", "a bad goal..." |
| 8 | Daily goal and unit: oz goal and ml unit saved, kept after reload, shown in the `/diet` water line and preset buttons (local prefs, reload check only); stored amounts stay in oz | done | water-presets: "daily goal in oz", "unit ml..." |
| 9 | Activity level (Select on the targets page) saves the profile, persists after reload | done | targets: "activity level" |

Adaptations to the real UI:

- Saving does not edit the plan: `UpdatePlanMacros` creates a new plan row (same name) effective
  today and re-points every `days` row with `date >= today` at it. The tests pin that snapshot
  semantics (2 plan rows, past days keep the old one).
- Every calorie / protein / fat edit refills carbs (`(kcal - 4P - 9F) / 4`) and clamps fiber to carbs.
  The tests type calories first and pick numbers that leave integer carbs (2200 kcal, P200, F60 -> C215).
- Load-error fault: the page also loads the last 7 days (same endpoint, offsets 1-7) at the same time,
  so a `times: 3` fault gets split across those requests. The test fails all 8 requests x 3 attempts
  (`times: 24`), asserts the fault drained (hits === 24), then Retry succeeds.
- The goal chips are a `group` of toggle buttons (`aria-pressed`), not radios.

## Bugs found

None that break data. Minor findings (not pinned with `test.fail`, they are copy/UX nits):

1. **"Save goal" toasts twice** (`pages/diet/water-presets/index.vue`). The goal input runs
   `applyGoalFromInput` on `@change` (fires on blur) and the Save goal button runs it again on click, so
   typing a goal and clicking the button shows two "Goal saved" toasts (or two "Enter a positive goal"
   errors). Fix: drop the button's duplicate path (save on change only, or button only), or dedupe.
   Observed in `daily goal in oz` and `a bad goal toasts`, which tolerate and dismiss both.
2. **Calories can't be undone by typing**: changing calories refills carbs, so typing the original
   calorie number back doesn't make the form clean again (seed carbs 280 vs suggested 287). Only Reset
   does. Probably fine; worth knowing when writing "dirty" tests.
3. Amount fields accept `parseFloat` prefixes (`12abc` saves 12). Not tested, harmless.

## Follow-ups

- Not covered: the Protein / Fat g-per-lb preset buttons, lock toggles, the "Fill to N g" carbs button,
  the observed-maintenance line (needs 4+ logged days and a weight trend), the history "since" label for
  plans without `effective_from`, "No diet plan yet" empty state (not reachable: the seed always has a
  plan), `Last 7 days` intake line and its Retry (7 per-day queries; easy to add with a day fault).
- Not covered: deleting a drink size that has water logs (the delete invalidates `waterLogsAll`; the
  backend FK behaviour for `preset_id` is untested), logging water through a preset button (belongs to
  the `home` lane).
- The save writes a new plan row on every save, even several times a day (plans pile up in History, no
  de-dup for same-day saves). Product question: should a second save on the same day replace today's
  snapshot instead of adding another?
- The water goal / unit live only in localStorage: a different browser or device starts at 64 oz. Product
  question, mentioned in the page hint only as "amounts are stored in oz".
- The mobile Save button has no accessible state beyond its label ("Targets saved" when clean); tests
  use a regex name for it.
