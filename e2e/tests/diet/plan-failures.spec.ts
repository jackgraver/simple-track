import { expect, test, waitForApi } from '../../fixtures';
import { createSavedMeals, dayCard, OFFSET, plannedChip } from './plan-helpers';

// /diet/edit-planned: a failed load shows Retry, a failed write rolls back (and Retry saves it).
// Queries retry twice on 5xx (3 attempts, ~1s then ~2s backoff), the plan writes once (2 attempts), so
// a fault must fail every attempt to surface; page.clock.runFor skips the backoff after a failed response.

const ATTEMPTS = 3;
const WEEK_DAYS = 7; // one /diet/logs/today query per day of the week

const dayLoad = (r: import('@playwright/test').Response) =>
  r.request().method() === 'GET' && new URL(r.url()).pathname === '/api/diet/logs/today';

test.describe('diet plan: failures', { tag: ['@diet'] }, () => {
  test('week load error: "Couldn\'t load this week" with Retry, and Retry loads it', async ({ page, api, e2e }) => {
    const { banana } = await createSavedMeals(api);
    await api.addPlannedFromSaved(banana, OFFSET.fri);
    // The days query is a gorm First on `days`. All seven day queries run, each with its own 3 attempts.
    await e2e.fault({ table: 'days', op: 'query', times: ATTEMPTS * WEEK_DAYS });
    let failures = 0;
    page.on('response', (r) => {
      if (dayLoad(r) && r.status() === 500) failures++;
    });
    await page.goto('/diet/edit-planned');
    // Each day's query backs off on its own timer: keep skipping backoff until all 21 attempts failed.
    await expect
      .poll(async () => {
        await page.clock.runFor(5_000);
        return failures;
      })
      .toBe(ATTEMPTS * WEEK_DAYS);

    const alert = page.getByRole('alert').filter({ hasText: "Couldn't load this week" });
    await expect(alert).toBeVisible();
    await expect(page.getByRole('heading', { level: 3 })).toHaveCount(0);
    await alert.getByRole('button', { name: 'Retry' }).click();
    await expect(alert).toHaveCount(0);
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toBeVisible();
    await expect(dayCard(page, 'Sat 17').getByText('Nothing planned')).toBeVisible();
  });

  test('write failure: the chip rolls back with a Retry toast, and Retry saves it', async ({ page, api, e2e }) => {
    await createSavedMeals(api);
    await page.goto('/diet/edit-planned');
    await expect(dayCard(page, 'Fri 16').getByText('Nothing planned')).toBeVisible();

    // Add-from-saved retries once by itself (client_id makes a replay safe), so two failures surface it.
    await e2e.fault({ table: 'planned_meals', op: 'create', times: 2 });
    const failedAdd = () =>
      page.waitForResponse((r) => r.request().method() === 'POST' && r.url().includes('/planned/from-saved') && r.status() === 500);
    const first = failedAdd();
    await dayCard(page, 'Fri 16').getByRole('button', { name: 'Add meal to Fri 16' }).click();
    const dialog = page.getByRole('dialog', { name: 'Add to Fri 16' });
    await dialog.getByRole('button', { name: /^Banana/ }).click();
    await dialog.getByRole('button', { name: 'Done' }).click();
    await first;
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toBeVisible(); // optimistic: held by the retry backoff
    const second = failedAdd();
    await page.clock.runFor(5_000);
    await second;

    const retry = page.getByRole('button', { name: 'Retry' });
    await expect(retry).toBeVisible();
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toHaveCount(0); // rolled back
    await expect(dayCard(page, 'Fri 16').getByText('Nothing planned')).toBeVisible();
    expect(await api.plannedMeals(OFFSET.fri)).toEqual([]);

    const saved = waitForApi(page, 'POST /diet/meals/planned/from-saved');
    await retry.click();
    await saved;
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toBeVisible();
    await page.reload();
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toBeVisible();
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Banana']); // once: no duplicate from the retries
  });
});
