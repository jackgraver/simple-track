import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// Diet golden paths through Quick log (a single typed-in meal row): log, edit, delete, each verified
// after a reload. Quick log is optimistic (row appears before the request), so every write awaits its
// response before the reload.

const LOG = 'POST /diet/meals/quick-log';
const DELETE = 'DELETE /diet/meals/meal/logged';

async function quickLog(page: Page, name: string, calories: number, protein = 0) {
  await page.getByRole('button', { name: 'Quick log' }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByLabel('Name').fill(name);
  await dialog.getByLabel('Calories').fill(String(calories));
  await dialog.getByLabel('Protein (g)').fill(String(protein));
  await dialog.getByRole('button', { name: 'Log meal' }).click();
}

const logged = (page: Page, name: string) => page.getByRole('button', { name: `More actions for ${name}` });

test.describe('diet: quick log', { tag: ['@diet'] }, () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/diet');
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
  });

  test('log a meal: shows at once and persists across a reload', { tag: '@smoke' }, async ({ page }) => {
    const saved = waitForApi(page, LOG);
    await quickLog(page, 'Takeout burger', 800, 40);
    await expect(logged(page, 'Takeout burger')).toBeVisible();
    await saved;
    await page.reload();
    await expect(logged(page, 'Takeout burger')).toBeVisible();
    await expect(page.getByText('Nothing logged yet')).toBeHidden();
  });

  test('edit a logged meal: replaces it, no duplicate left behind', async ({ page, api }) => {
    let saved = waitForApi(page, LOG);
    await quickLog(page, 'Burger', 800);
    await saved;
    await page.reload();

    await logged(page, 'Burger').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel('Name').fill('Salad');
    await dialog.getByLabel('Calories').fill('300');
    saved = waitForApi(page, LOG);
    await dialog.getByRole('button', { name: 'Save changes' }).click();
    await expect(logged(page, 'Salad')).toBeVisible();
    await saved;

    await page.reload();
    await expect(logged(page, 'Salad')).toBeVisible();
    await expect(logged(page, 'Burger')).toHaveCount(0);
    const today = await api.get<{ day: { loggedMeals: unknown[] } }>('/diet/logs/today');
    expect(today.day.loggedMeals).toHaveLength(1);
  });

  test('delete a logged meal: stays gone, the other meal survives', async ({ page }) => {
    for (const [name, cal] of [['Burger', 800], ['Salad', 300]] as const) {
      const saved = waitForApi(page, LOG);
      await quickLog(page, name, cal);
      await saved;
    }
    await expect(logged(page, 'Salad')).toBeVisible();

    const deleted = waitForApi(page, DELETE);
    await logged(page, 'Burger').click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('dialog', { name: 'Remove quick log' }).getByRole('button', { name: 'Confirm' }).click();
    await expect(logged(page, 'Burger')).toHaveCount(0);
    await deleted;

    await page.reload();
    await expect(logged(page, 'Salad')).toBeVisible();
    await expect(logged(page, 'Burger')).toHaveCount(0);
  });
});
