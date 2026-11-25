import type { Page } from '@playwright/test';
import { expect, test } from '../../fixtures';

// /gym/progression: the Graph view is a canvas, so specs read the List view (one group per day, newest
// first, "- Set n: <weight>x<reps>"). Range cutoffs use the browser clock (E2E_EPOCH, Jan 15 2026).
const SQUAT = 9;

async function openExercise(page: Page, name: string) {
  await page.goto('/gym/progression');
  const search = page.getByRole('combobox', { name: 'Search exercises…' });
  await search.fill(name);
  await page.getByRole('option', { name }).click();
  await expect(page.getByText(name, { exact: true }).first()).toBeVisible();
}

const control = (page: Page, label: 'Range' | 'View') => page.getByRole('group', { name: label });
const dayGroups = (page: Page) => page.getByRole('main').getByText(/^[A-Z][a-z]{2} \d{1,2}, \d{4}/);
const sets = (page: Page) => page.getByRole('main').getByRole('listitem');

test.describe('gym progression', { tag: ['@gym'] }, () => {
  test('sets across dates show per day, and the range control filters them', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: '2026-01-08' });
    await api.logSet({ exerciseId: SQUAT, reps: 6, weight: 100, date: '2026-01-08' });
    await api.logSet({ exerciseId: SQUAT, reps: 5, weight: 105, date: '2026-01-13' });
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 95, date: '2025-11-20' }); // 3 mo, not 1 mo
    await api.logSet({ exerciseId: SQUAT, reps: 10, weight: 80, date: '2025-06-01' }); // 1 yr, not 6 mo
    await api.logSet({ exerciseId: SQUAT, reps: 0, weight: 100, date: '2026-01-14' }); // 0 reps: excluded

    await openExercise(page, 'Back Squat');
    await expect(control(page, 'Range').getByRole('button', { name: '1 mo' })).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByText(/^No (chart )?data in the/)).toHaveCount(0); // Graph view: the chart (a canvas) has data
    await control(page, 'View').getByRole('button', { name: 'List' }).click();

    await expect(dayGroups(page)).toHaveText(['Jan 13, 2026', 'Jan 8, 2026']);
    await expect(sets(page)).toHaveText(['- Set 1: 105x5', '- Set 1: 100x8', '- Set 2: 100x6']);

    await control(page, 'Range').getByRole('button', { name: '3 mo' }).click();
    await expect(dayGroups(page)).toHaveText(['Jan 13, 2026', 'Jan 8, 2026', 'Nov 20, 2025']);
    await control(page, 'Range').getByRole('button', { name: '6 mo' }).click();
    await expect(dayGroups(page)).toHaveCount(3);
    await control(page, 'Range').getByRole('button', { name: 'All' }).click();
    await expect(dayGroups(page)).toHaveText(['Jan 13, 2026', 'Jan 8, 2026', 'Nov 20, 2025', 'Jun 1, 2025']);
    await expect(sets(page).last()).toHaveText('- Set 1: 80x10');
  });

  test('an exercise without sets shows the empty state, in every range', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: '2026-01-08' }); // another exercise's data
    await openExercise(page, 'Leg Curl');
    await expect(page.getByText('No data in the last month')).toBeVisible();
    await control(page, 'Range').getByRole('button', { name: '1 yr' }).click();
    await expect(page.getByText('No data in the last year')).toBeVisible();
    await control(page, 'View').getByRole('button', { name: 'List' }).click();
    await expect(page.getByText('No data in the last year')).toBeVisible();
    await expect(sets(page)).toHaveCount(0);
  });

  // Bug (e2e/notes/gym-tracking.md): the "All" range builds "No data in the this exercise".
  test('the empty state reads correctly for the "All" range', async ({ page }) => {
    test.fail(true, 'copy bug "No data in the this exercise", see e2e/notes/gym-tracking.md');
    await openExercise(page, 'Leg Curl');
    await control(page, 'Range').getByRole('button', { name: 'All' }).click();
    await expect(page.getByText(/^No data /)).toBeVisible();
    await expect(page.getByText(/the this/)).toHaveCount(0, { timeout: 1_000 });
  });

  test('switching exercise swaps the data', async ({ page, api }) => {
    await api.logSet({ exerciseId: SQUAT, reps: 8, weight: 100, date: '2026-01-08' });
    await openExercise(page, 'Leg Curl');
    await expect(page.getByText('No data in the last month')).toBeVisible();
    const search = page.getByRole('combobox', { name: 'Search exercises…' });
    await search.fill('Back');
    await page.getByRole('option', { name: 'Back Squat' }).click();
    await expect(page.getByText('No data in the last month')).toHaveCount(0);
    await control(page, 'View').getByRole('button', { name: 'List' }).click();
    await expect(sets(page)).toHaveText(['- Set 1: 100x8']);
  });
});
