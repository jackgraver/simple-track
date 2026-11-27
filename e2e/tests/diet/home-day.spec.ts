import { expect, test, waitForApi } from '../../fixtures';
import { BREAKFAST, BOWL, logButton, loggedRow, plannedHeading, planMeal } from './home-helpers';

// /diet day view: day navigation, the week strip, water, empty states and load errors.
// Epoch: Thursday 2026-01-15 (week Sun Jan 11 - Sat Jan 17). `?offset=N` is N days in the past.

const tile = (page: import('@playwright/test').Page, label: RegExp) =>
  page.getByRole('list', { name: 'Days this week' }).getByRole('link', { name: label });

test.describe('diet home: day navigation', { tag: ['@diet'] }, () => {
  test('quick log on yesterday lands on yesterday, not today', async ({ page, api }) => {
    await page.goto('/diet');
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await tile(page, /^Wednesday, January 14/).click();
    await expect(page).toHaveURL(/offset=1/);
    await expect(page.getByText('Left on Wed, Jan 14')).toBeVisible();

    await page.getByRole('button', { name: 'Quick log' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Name').fill('Late snack');
    await dialog.getByLabel('Calories').fill('300');
    const saved = waitForApi(page, 'POST /diet/meals/quick-log');
    await dialog.getByRole('button', { name: 'Log meal' }).click();
    await expect(loggedRow(page, 'Late snack')).toBeVisible();
    await saved;

    await page.reload();
    await expect(page).toHaveURL(/offset=1/);
    await expect(loggedRow(page, 'Late snack')).toBeVisible();
    await tile(page, /^Thursday, January 15/).click();
    await expect(page).not.toHaveURL(/offset=/);
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await expect(loggedRow(page, 'Late snack')).toHaveCount(0);
    expect((await api.dietDay(1)).day.loggedMeals).toHaveLength(1);
    expect((await api.dietDay(0)).day.loggedMeals).toHaveLength(0);
  });

  test('the planned Log button is offered today only', async ({ page, api }) => {
    await planMeal(api, BOWL, 0);
    await planMeal(api, BREAKFAST, 1);
    await page.goto('/diet?offset=1');
    await expect(plannedHeading(page)).toHaveText('Planned · 1 to go');
    await expect(loggedRow(page, BREAKFAST.name)).toBeVisible(); // the planned row's menu
    await expect(logButton(page, BREAKFAST.name)).toHaveCount(0);

    await page.goto('/diet');
    await expect(logButton(page, BOWL.name)).toBeVisible();
  });
});

test.describe('diet home: week strip', { tag: ['@diet'] }, () => {
  test('tiles show their kcal, the summary counts logged days, and paging works', async ({ page, api }) => {
    await api.quickLog({ name: 'Today meal', calories: 500, offset: 0 });
    await api.quickLog({ name: 'Yesterday meal', calories: 700, offset: 1 });
    await api.quickLog({ name: 'Last week meal', calories: 400, offset: 8 }); // Thu Jan 7
    await page.goto('/diet');

    await expect(page.getByRole('heading', { name: 'This week' })).toBeVisible();
    await expect(tile(page, /^Thursday, January 15, 500 of 2,500 kcal/)).toBeVisible();
    await expect(tile(page, /^Wednesday, January 14, 700 of 2,500 kcal/)).toBeVisible();
    await expect(tile(page, /^Tuesday, January 13, nothing logged/)).toBeVisible();
    await expect(tile(page, /^Thursday, January 15/)).toHaveAttribute('aria-current', 'date');
    await expect(page.getByText('avg 600 / 2,500 kcal · 2 days logged')).toBeVisible();

    await page.getByRole('button', { name: 'Previous week' }).click();
    await expect(page).toHaveURL(/offset=11/); // lands on Sunday Jan 4
    await expect(page.getByRole('heading', { name: 'Jan 4 – Jan 10' })).toBeVisible();
    await expect(tile(page, /^Wednesday, January 7, 400 of 2,500 kcal/)).toBeVisible();
    await expect(page.getByText('avg 400 / 2,500 kcal · 1 day logged')).toBeVisible();

    await page.getByRole('button', { name: 'Next week' }).click();
    await expect(page.getByRole('heading', { name: 'This week' })).toBeVisible();
    await expect(page.getByText('avg 600 / 2,500 kcal · 2 days logged')).toBeVisible();
    await expect(page).not.toHaveURL(/offset=/);
  });

  // BUG (notes/diet-home.md, "week strip creates future days"): `GET /diet/logs/today` creates the day
  // row with the plan in force right now, and the strip reads Fri and Sat too. prefetchDays.ts leaves
  // future days out for exactly this reason ("an early fetch would pin tomorrow's plan"), but the strip
  // does not. Intended: looking at the week writes nothing for days that have not happened.
  test.fail('viewing the week creates no rows for future days', async ({ page, db }) => {
    await page.goto('/diet');
    await expect(tile(page, /^Saturday, January 17/)).toBeVisible(); // the strip has loaded all 7 days
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM days WHERE date >= '2026-01-16'`);
    expect(rows[0].n).toBe(0);
  });

  test('a week with nothing logged says so', async ({ page }) => {
    await page.goto('/diet?offset=7');
    await expect(page.getByText('Nothing logged this week')).toBeVisible();
  });
});

test.describe('diet home: water', { tag: ['@diet'] }, () => {
  test('a preset logs a drink, the entries dialog lists it, deleting persists', async ({ page, api }) => {
    await api.createDrinkSize('Glass', 8);
    await page.goto('/diet');
    const saved = waitForApi(page, 'POST /tracking/water');
    await page.getByRole('button', { name: 'Log Glass, 8 oz of water' }).click();
    await saved;

    await page.reload();
    const entries = page.getByRole('button', { name: /^Water entries for/ });
    await entries.click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText('8 oz', { exact: true })).toBeVisible();
    await expect(dialog.getByRole('button', { name: 'Delete water entry 8 oz' })).toHaveCount(1);

    const deleted = waitForApi(page, 'DELETE /tracking/water/:id');
    await dialog.getByRole('button', { name: 'Delete water entry 8 oz' }).click();
    await expect(dialog.getByText('No water logged')).toBeVisible();
    await deleted;

    await page.reload();
    await entries.click();
    await expect(page.getByRole('dialog').getByText('No water logged')).toBeVisible();
  });
});

// The day view and the week strip load the same `GET /diet/logs/today?offset=N` (7 days, one shared
// query for the selected day). Queries retry twice on 5xx (3 attempts, ~1s then ~2s backoff), and each
// day read is one `days` SELECT, so a fault with `times: 7 * 3` fails every attempt of every day and is
// used up exactly; which request takes which hit does not matter.
const DAYS = 7;
const ATTEMPTS = 3;

test.describe('diet home: load error', { tag: ['@diet'] }, () => {
  test('"Couldn\'t load this day" recovers with Retry', async ({ page, api, e2e }) => {
    await api.quickLog({ name: 'Lunch', calories: 600 });
    await e2e.fault({ table: 'days', op: 'query', times: DAYS * ATTEMPTS });
    let failures = 0;
    page.on('response', (r) => {
      if (r.request().method() === 'GET' && r.url().includes('/api/diet/logs/today') && r.status() === 500) failures++;
    });
    await page.goto('/diet');
    for (let round = 1; round <= ATTEMPTS; round++) {
      await expect.poll(() => failures).toBe(round * DAYS);
      if (round < ATTEMPTS) await page.clock.runFor(5_000);
    }

    const dayError = page.getByRole('alert').filter({ hasText: "Couldn't load this day" });
    await expect(dayError).toBeVisible();
    await expect(page.getByRole('alert').filter({ hasText: "Couldn't load this week" })).toBeVisible();
    await expect(loggedRow(page, 'Lunch')).toHaveCount(0);

    await dayError.getByRole('button', { name: 'Retry' }).click();
    await expect(dayError).toHaveCount(0);
    await expect(loggedRow(page, 'Lunch')).toBeVisible();

    const weekError = page.getByRole('alert').filter({ hasText: "Couldn't load this week" });
    await weekError.getByRole('button', { name: 'Retry' }).click();
    await expect(weekError).toHaveCount(0);
    await expect(tile(page, /^Thursday, January 15, 600 of 2,500 kcal/)).toBeVisible();
  });
});

test.describe('diet home: empty states', { tag: ['@diet'] }, () => {
  test('nothing planned, nothing logged, no drink sizes', async ({ page }) => {
    await page.goto('/diet');
    await expect(page.getByText('Nothing planned')).toBeVisible();
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Add drink sizes' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Log water' })).toBeVisible();
    await page.getByRole('button', { name: 'Add drink sizes' }).click();
    await expect(page).toHaveURL(/\/diet\/water-presets/);
  });

  test('everything logged: "Nothing left to log"', async ({ page, api }) => {
    const p = await planMeal(api, BOWL);
    await api.logPlannedMeal(p.mealId, p.dayId);
    await page.goto('/diet');
    await expect(page.getByText('Nothing left to log')).toBeVisible();
    await expect(page.getByText('Nothing planned')).toHaveCount(0);
    await expect(plannedHeading(page)).toHaveText('Planned');
  });
});
