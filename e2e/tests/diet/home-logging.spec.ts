import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import {
  BOWL,
  BREAKFAST,
  DELETE_LOGGED,
  LOG_PLANNED,
  kcalLeft,
  logButton,
  loggedRow,
  macro,
  plannedHeading,
  planMeal,
} from './home-helpers';

// /diet day view: logging a planned meal and deleting a logged one. Both are optimistic (the row moves
// before the request), so every persistence check waits for the response (or holds it) before reloading.
// Seed: Maintenance plan 2500 kcal / P180 / C280 / F70; no meals until a test plans them.

const failedLog = (page: Page) =>
  page.waitForResponse((r) => r.request().method() === 'POST' && r.url().includes('/meal/log-planned') && r.status() === 500);

async function logCount(db: import('pg').Client, name?: string) {
  const { rows } = await db.query(
    `SELECT count(*)::int AS n FROM day_logs dl JOIN meals m ON m.id = dl.meal_id
     WHERE dl.deleted_at IS NULL AND ($1::text IS NULL OR m.name = $1)`,
    [name ?? null],
  );
  return rows[0].n as number;
}

test.describe('diet home: log a planned meal', { tag: ['@diet'] }, () => {
  test('Log moves the meal to Logged, updates the totals, and persists once', async ({ page, api, db }) => {
    await planMeal(api, BOWL);
    await page.goto('/diet');
    await expect(plannedHeading(page)).toHaveText('Planned · 1 to go');
    await expect(loggedRow(page, BOWL.name)).toBeVisible(); // planned row (its menu) before logging
    await expect(page.getByText('Nothing logged yet')).toBeVisible();

    const saved = waitForApi(page, LOG_PLANNED);
    await logButton(page, BOWL.name).click();
    await expect(page.getByText('Nothing left to log')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    await expect(logButton(page, BOWL.name)).toHaveCount(0);
    await saved;

    await page.reload();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    await expect(plannedHeading(page)).toHaveText('Planned'); // no "N to go"
    await expect(page.getByText('Nothing left to log')).toBeVisible();
    await expect(logButton(page, BOWL.name)).toHaveCount(0);
    expect(await logCount(db, BOWL.name)).toBe(1);
  });

  test('double tap with the request held: one logged row', async ({ page, api, e2e, db }) => {
    await planMeal(api, BOWL);
    await page.goto('/diet');
    const requests: string[] = [];
    page.on('request', (r) => r.url().includes('/meal/log-planned') && requests.push(r.method()));

    const gate = await e2e.hold(LOG_PLANNED);
    await logButton(page, BOWL.name).dblclick({ noWaitAfter: true }).catch(() => undefined); // the row leaves after tap 1
    await gate.arrived();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    const saved = waitForApi(page, LOG_PLANNED);
    await gate.release();
    await saved;

    await page.reload();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    expect(requests).toHaveLength(1);
    expect(await logCount(db)).toBe(1);
  });

  // The Log request carries no client_id (the frontend generates one but api.logPlannedMeal never sends
  // it); a replay is de-duplicated by (day, meal) on the server instead, which is what this pins.
  test('dropped response: the retry does not duplicate the log', async ({ page, api, e2e, db }) => {
    await planMeal(api, BOWL);
    await page.goto('/diet');
    const gate = await e2e.hold(LOG_PLANNED, { phase: 'after' });
    await logButton(page, BOWL.name).click();
    await gate.arrived(); // committed; the response is held
    const saved = waitForApi(page, LOG_PLANNED);
    await gate.drop();
    await page.clock.fastForward(5_000);
    await saved;

    await page.reload();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    expect(await logCount(db)).toBe(1);
  });

  test('DB fault: the row rolls back with a Retry toast, and Retry persists it', async ({ page, api, e2e, db }) => {
    await planMeal(api, BOWL);
    await page.goto('/diet');
    await e2e.fault({ table: 'day_logs', op: 'create', times: 2 }); // the first try and its one automatic retry
    const first = failedLog(page);
    await logButton(page, BOWL.name).click();
    await expect(page.getByText('Nothing left to log')).toBeVisible(); // optimistic
    await first;
    const second = failedLog(page);
    await page.clock.runFor(5_000); // skip the retry backoff
    await second;

    const retry = page.getByRole('button', { name: 'Retry' });
    await expect(retry).toBeVisible();
    await expect(page.getByText('Couldn’t log that meal', { exact: true })).toBeVisible();
    await expect(page.getByText('Nothing logged yet')).toBeVisible(); // rolled back
    await expect(logButton(page, BOWL.name)).toBeVisible();
    expect(await logCount(db)).toBe(0);

    const saved = waitForApi(page, LOG_PLANNED);
    await retry.click();
    await saved;
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();
    expect(await logCount(db, BOWL.name)).toBe(1);
  });
});

test.describe('diet home: delete a logged meal', { tag: ['@diet'] }, () => {
  async function remove(page: Page, name: string) {
    await loggedRow(page, name).click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('dialog', { name: 'Delete Logged Meal' }).getByRole('button', { name: 'Confirm' }).click();
  }

  test('a catalogue meal: gone after a reload; its foods and saved meal survive', async ({ page, api, db }) => {
    const p = await planMeal(api, BOWL);
    await api.logPlannedMeal(p.mealId, p.dayId);
    await page.goto('/diet');
    await expect(page.getByRole('heading', { name: 'Logged · 1' })).toBeVisible();

    const deleted = waitForApi(page, DELETE_LOGGED);
    await remove(page, BOWL.name);
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await deleted;

    await page.reload();
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    expect(await logCount(db)).toBe(0);
    const foods = await db.query(`SELECT count(*)::int AS n FROM foods WHERE deleted_at IS NULL`);
    expect(foods.rows[0].n).toBe(8);
    const saved = await db.query(`SELECT count(*)::int AS n FROM saved_meals WHERE id = $1 AND deleted_at IS NULL`, [p.savedMealId]);
    expect(saved.rows[0].n).toBe(1);
    const items = await db.query(`SELECT count(*)::int AS n FROM saved_meal_items WHERE saved_meal_id = $1`, [p.savedMealId]);
    expect(items.rows[0].n).toBe(2);
  });

  test('a row still saving: deleted behind its log, nothing left after release and reload', async ({ page, api, e2e, db }) => {
    await planMeal(api, BOWL);
    await page.goto('/diet');
    const gate = await e2e.hold(LOG_PLANNED);
    await logButton(page, BOWL.name).click();
    await gate.arrived();
    await expect(page.getByRole('listitem').filter({ hasText: BOWL.name })).toHaveAttribute('aria-busy', 'true');

    await remove(page, BOWL.name); // queued behind the held log
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    const deleted = waitForApi(page, DELETE_LOGGED);
    await gate.release();
    await deleted;

    await page.reload();
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    expect(await logCount(db)).toBe(0);
  });

  test('totals: logging moves the kcal and macro lines, deleting restores them', async ({ page, api }) => {
    await planMeal(api, BOWL);
    await planMeal(api, BREAKFAST);
    await page.goto('/diet');
    await expect(kcalLeft(page)).toHaveText(/^2,500\s*kcal left$/);
    await expect(macro(page, 'Protein')).toContainText('180 g left');

    let saved = waitForApi(page, LOG_PLANNED);
    await logButton(page, BOWL.name).click();
    await expect(kcalLeft(page)).toHaveText(/^2,205\s*kcal left$/); // 2500 - 295
    await expect(macro(page, 'Protein')).toContainText('146 g left'); // 180 - 33.7
    await expect(macro(page, 'Carbs')).toContainText('252 g left'); // 280 - 28
    await expect(macro(page, 'Fat')).toContainText('66 g left'); // 70 - 3.9
    await saved;

    saved = waitForApi(page, LOG_PLANNED);
    await logButton(page, BREAKFAST.name).click();
    await expect(kcalLeft(page)).toHaveText(/^1,956\s*kcal left$/); // minus 249
    await saved;
    await page.reload();
    await expect(kcalLeft(page)).toHaveText(/^1,956\s*kcal left$/);
    await expect(macro(page, 'Protein')).toContainText('132 g left'); // 180 - 47.6

    const deleted = waitForApi(page, DELETE_LOGGED);
    await remove(page, BREAKFAST.name);
    await expect(kcalLeft(page)).toHaveText(/^2,205\s*kcal left$/);
    await deleted;
    await page.reload();
    await expect(kcalLeft(page)).toHaveText(/^2,205\s*kcal left$/);
    await expect(macro(page, 'Protein')).toContainText('146 g left');
    await expect(macro(page, 'Carbs')).toContainText('252 g left');
    await expect(macro(page, 'Fat')).toContainText('66 g left');
  });
});
