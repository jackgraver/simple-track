import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// The /gym home page: the Today card (state, plan switching) and the week strip (status, streak,
// day and week navigation). Seed: the epoch is Thu 2026-01-15; the active program "Push Pull Legs"
// schedules Push on Monday, Pull on Wednesday and Legs (Back Squat 9, Romanian Deadlift 10,
// Leg Press 11, Leg Curl 12) on Thursday. Weeks start on Sunday. Insights (mock data) and the
// heatmap are out of scope.
const TODAY = '2026-01-15';
const MONDAY = '2026-01-12';
const LEGS = [9, 10, 11, 12];
const LEGS_NAMES = ['Back Squat', 'Romanian Deadlift', 'Leg Press', 'Leg Curl'];
const SWITCH = 'PATCH /workout/logs/switch-plan';

const planHeading = (page: Page, name: string) => page.getByRole('heading', { level: 2, name, exact: true });
const tile = (page: Page, weekday: string) =>
  page.getByRole('list', { name: 'Days this week' }).getByRole('link', { name: new RegExp(`^${weekday},`) });
const changeButton = (page: Page) => page.getByRole('button', { name: 'Change workout' });
const stepsList = (page: Page) => page.getByRole('list').filter({ has: page.getByText(/Back Squat|Bench Press/) }).last();

async function openHome(page: Page, query = '') {
  await page.goto(`/gym${query}`);
  await expect(page.getByRole('list', { name: 'Days this week' })).toBeVisible();
}

async function showExercises(page: Page) {
  await page.getByRole('button', { name: 'Show exercises' }).click();
  await expect(page.getByRole('button', { name: 'Hide exercises' })).toHaveAttribute('aria-expanded', 'true');
}

/** Pick an entry from the Today card's plan menu ("Change" or "Train anyway") and wait for the write. */
async function switchTo(page: Page, trigger: 'Change workout' | 'Train anyway', item: string | RegExp) {
  await page.getByRole('button', { name: trigger }).click();
  const saved = waitForApi(page, SWITCH);
  await page.getByRole('menuitem', { name: item }).click();
  await saved;
}

test.describe('gym home: today card', { tag: ['@gym'] }, () => {
  test('fresh day: today\'s plan with its lifts in order, first one up next', { tag: ['@smoke'] }, async ({ page }) => {
    await openHome(page);
    await expect(page.getByText('Today · Thursday, Jan 15', { exact: true })).toBeVisible();
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await expect(page.getByText('4 lifts', { exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Start workout' })).toBeVisible();
    await showExercises(page);
    await expect(stepsList(page).getByRole('listitem')).toHaveText([
      /^1\s*Back Squat\s*\(up next\)/,
      /^2\s*Romanian Deadlift\s*$/,
      /^3\s*Leg Press\s*$/,
      /^4\s*Leg Curl\s*$/,
    ]);
  });

  test('in progress: one lift logged -> Continue workout, 1 of 4, next lift up next', async ({ page, api }) => {
    await api.logSet({ exerciseId: 9, reps: 8, weight: 100, date: TODAY });
    await openHome(page);
    await expect(page.getByRole('link', { name: 'Continue workout' })).toBeVisible();
    await expect(page.getByText('In progress', { exact: true })).toBeVisible();
    await expect(page.getByText('1 of 4', { exact: true })).toBeVisible();
    await showExercises(page);
    const items = stepsList(page).getByRole('listitem');
    await expect(items.nth(0)).toHaveText(/^Back Squat\s*\(done\)\s*100 lb × 8$/);
    await expect(items.nth(1)).toHaveText(/^2\s*Romanian Deadlift\s*\(up next\)/);
  });

  test('complete: every lift logged -> Review workout, All done, 4 of 4', async ({ page, api }) => {
    for (const id of LEGS) await api.logSet({ exerciseId: id, reps: 8, weight: 100, date: TODAY });
    await openHome(page);
    await expect(page.getByRole('link', { name: 'Review workout' })).toBeVisible();
    await expect(page.getByText('All done', { exact: true })).toBeVisible();
    await expect(page.getByText('4 of 4', { exact: true })).toBeVisible();
  });

  test('Start workout opens the logging page with today\'s plan', async ({ page }) => {
    await openHome(page);
    await page.getByRole('link', { name: 'Start workout' }).click();
    await expect(page).toHaveURL(/\/gym\/logging$/);
    await expect(page.getByRole('heading', { level: 1, name: 'Legs' })).toBeVisible();
    for (const name of LEGS_NAMES) await expect(page.getByText(name, { exact: true }).first()).toBeVisible();
  });
});

test.describe('gym home: plan switching', { tag: ['@gym'] }, () => {
  test('Change -> Push: swapped for today, persists, the week still schedules Legs', async ({ page }) => {
    await openHome(page);
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await switchTo(page, 'Change workout', 'Push');
    await expect(page.getByText('Workout plan updated')).toBeVisible();
    await expect(planHeading(page, 'Push')).toBeVisible();
    await expect(page.getByText('4 lifts · swapped from Legs', { exact: true })).toBeVisible();
    await page.reload();
    await expect(planHeading(page, 'Push')).toBeVisible();
    await expect(page.getByText('4 lifts · swapped from Legs', { exact: true })).toBeVisible();
    await expect(tile(page, 'Thursday')).toHaveAccessibleName(/, Legs, planned$/);
  });

  test('Make it a rest day: persists; Train anyway -> Legs restores it', async ({ page }) => {
    await openHome(page);
    await switchTo(page, 'Change workout', 'Make it a rest day');
    await expect(page.getByText('Marked as a rest day')).toBeVisible();
    await expect(planHeading(page, 'Rest day')).toBeVisible();
    await expect(page.getByText('Next up: Push, Monday', { exact: true })).toBeVisible();
    await page.reload();
    await expect(planHeading(page, 'Rest day')).toBeVisible();
    await expect(page.getByText('Next up: Push, Monday', { exact: true })).toBeVisible();

    await switchTo(page, 'Train anyway', 'Legs');
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await page.reload();
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await expect(page.getByText('4 lifts', { exact: true })).toBeVisible();
  });

  test('data loss: swapping away and back keeps the logged set', async ({ page, api, db }) => {
    await api.logSet({ exerciseId: 9, reps: 8, weight: 100, date: TODAY });
    await openHome(page);
    await expect(page.getByText('1 of 4', { exact: true })).toBeVisible();
    await switchTo(page, 'Change workout', 'Push');
    await expect(planHeading(page, 'Push')).toBeVisible();
    await switchTo(page, 'Change workout', /^Legs/);
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await page.reload();
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await expect(page.getByText('1 of 4', { exact: true })).toBeVisible();
    await showExercises(page);
    await expect(stepsList(page).getByRole('listitem').first()).toHaveText(/^Back Squat\s*\(done\)\s*100 lb × 8$/);
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM logged_sets WHERE deleted_at IS NULL`);
    expect(rows[0].n).toBe(1);
  });

  test.describe(() => {
    // SwitchPlan maps every error to 400, so the browser logs a 4xx (see notes: bug).
    test.use({ allowedErrors: [/status of 400/] });
    test('switch failure: error toast, the plan stays Legs', async ({ page, e2e }) => {
      await openHome(page);
      await expect(planHeading(page, 'Legs')).toBeVisible();
      await e2e.fault({ table: 'workout_logs', op: 'update', times: 1 });
      await page.getByRole('button', { name: 'Change workout' }).click();
      await page.getByRole('menuitem', { name: 'Push' }).click();
      // The toast shows the server's raw error text (see notes: SwitchPlan returns DB errors as 400).
      await expect(page.getByText('e2e fault: update workout_logs', { exact: true })).toBeVisible();
      await expect(page.getByText('Workout plan updated')).toHaveCount(0);
      await expect(planHeading(page, 'Legs')).toBeVisible();
      await page.reload();
      await expect(planHeading(page, 'Legs')).toBeVisible();
      await expect(page.getByText('4 lifts', { exact: true })).toBeVisible();
    });
  });
});

test.describe('gym home: week strip', { tag: ['@gym'] }, () => {
  test('week status comes from logged sets and the program schedule', async ({ page, api }) => {
    await api.logSet({ exerciseId: 1, reps: 8, weight: 100, date: MONDAY });
    await openHome(page);
    await expect(page.getByText('1 of 3 sessions · Push Pull Legs', { exact: true })).toBeVisible();
    const expected: [string, RegExp][] = [
      ['Sunday', /^Sunday, January 11, rest day$/],
      ['Monday', /^Monday, January 12, Push, trained$/],
      ['Tuesday', /^Tuesday, January 13, rest day$/],
      ['Wednesday', /^Wednesday, January 14, Pull, missed$/],
      ['Thursday', /^Thursday, January 15, Legs, planned$/],
      ['Friday', /^Friday, January 16, rest day$/],
      ['Saturday', /^Saturday, January 17, rest day$/],
    ];
    for (const [day, label] of expected) await expect(tile(page, day)).toHaveAccessibleName(label);
    await expect(tile(page, 'Thursday')).toHaveAttribute('aria-current', 'date');
  });

  const streakBadge = (page: Page) => page.getByTitle(/sessions in a row/);

  test('streak: consecutive scheduled sessions count, rest days in between don\'t break it', async ({ page, api }) => {
    // Back from today (not trained yet, doesn't count): Wed 14, Mon 12, Thu 8, Wed 7 trained; Tue/Fri-Sun
    // are rest days; Mon 5 was scheduled and missed, so the streak stops at 4.
    for (const date of ['2026-01-14', '2026-01-12', '2026-01-08', '2026-01-07']) {
      await api.logSet({ exerciseId: 1, reps: 8, weight: 100, date });
    }
    await openHome(page);
    await expect(streakBadge(page)).toHaveText(/^4\s*in a row$/);
  });

  test('streak: a missed scheduled day breaks it', async ({ page, api }) => {
    // Same history without Monday 12: Wed 14 counts, the missed Monday ends the streak at 1 (no badge).
    for (const date of ['2026-01-14', '2026-01-08', '2026-01-07']) {
      await api.logSet({ exerciseId: 1, reps: 8, weight: 100, date });
    }
    await openHome(page);
    await expect(tile(page, 'Monday')).toHaveAccessibleName(/, Push, missed$/);
    await expect(tile(page, 'Wednesday')).toHaveAccessibleName(/, Pull, trained$/);
    await expect(streakBadge(page)).toHaveCount(0);
  });

  test('day navigation: a past training day, a future rest day, a future training day', async ({ page }) => {
    await openHome(page);
    await tile(page, 'Monday').click();
    await expect(page).toHaveURL(/\/gym\?offset=3$/);
    await expect(planHeading(page, 'Push')).toBeVisible();
    await expect(page.getByText('Monday · Jan 12', { exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Log this workout' })).toBeVisible();
    await page.reload();
    await expect(planHeading(page, 'Push')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Log this workout' })).toBeVisible();

    await tile(page, 'Friday').click();
    await expect(page).toHaveURL(/\/gym\?offset=-1$/);
    await expect(planHeading(page, 'Rest day')).toBeVisible();
    await expect(page.getByText('Tomorrow · Friday, Jan 16', { exact: true })).toBeVisible();
    await expect(page.getByText('Next up: Push, Monday', { exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Next week' }).click();
    await tile(page, 'Monday').click();
    await expect(page).toHaveURL(/\/gym\?offset=-4$/);
    await expect(planHeading(page, 'Push')).toBeVisible();
    await expect(page.getByRole('link', { name: 'View plan' })).toBeVisible();
  });

  test('week paging: previous week shows its range, next week returns to This week', async ({ page }) => {
    await openHome(page);
    await expect(page.getByRole('heading', { name: 'This week' })).toBeVisible();
    await page.getByRole('button', { name: 'Previous week' }).click();
    await expect(page).toHaveURL(/\/gym\?offset=7$/);
    await expect(page.getByRole('heading', { name: 'Jan 4 – Jan 10' })).toBeVisible();
    await expect(tile(page, 'Thursday')).toHaveAccessibleName(/^Thursday, January 8, Legs, missed$/);
    await page.getByRole('button', { name: 'Next week' }).click();
    await expect(page).toHaveURL(/\/gym$/);
    await expect(page.getByRole('heading', { name: 'This week' })).toBeVisible();
  });
});

test.describe('gym home: load failures', { tag: ['@gym'] }, () => {
  // Queries retry twice (shouldRetryQuery: 3 attempts, 1s + 2s backoff in browser time), so a fault
  // that fires 3 times exhausts them and the error state shows; Retry's 4th attempt succeeds.
  test('week fails to load, Retry recovers', async ({ page, e2e, api }) => {
    await api.logSet({ exerciseId: 1, reps: 8, weight: 100, date: MONDAY });
    // The activity query is the only raw statement on logged_sets.
    await e2e.fault({ table: 'logged_sets', op: 'row', times: 3 });
    await page.goto('/gym');
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load your week" });
    await expect(alert).toBeVisible({ timeout: 10_000 });
    await expect(planHeading(page, 'Legs')).toBeVisible(); // the Today card is unaffected
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(tile(page, 'Monday')).toHaveAccessibleName(/, Push, trained$/);
    await expect(alert).toHaveCount(0);
  });

  test('day fails to load, Retry recovers', async ({ page, e2e }) => {
    await e2e.fault({ table: 'workout_logs', op: 'query', times: 3 });
    await page.goto('/gym');
    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load this day's workout" });
    await expect(alert).toBeVisible({ timeout: 10_000 });
    await expect(tile(page, 'Thursday')).toBeVisible(); // the week strip is unaffected
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(planHeading(page, 'Legs')).toBeVisible();
    await expect(alert).toHaveCount(0);
  });
});
