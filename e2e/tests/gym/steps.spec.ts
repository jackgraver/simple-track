import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// /gym/steps has no history list (a 14-day canvas bar chart) and the input is not prefilled. The text
// the page exposes for a saved day is the overwrite prompt ("You already logged 8,420 steps for …"),
// so persistence is asserted through it, plus the DB for "exactly one row".
const SAVE = 'POST /tracking/steps';
const TODAY_LABEL = 'Thursday January 15th, 2026';

const stepsInput = (page: Page) => page.getByLabel(/^Steps \(/);
const saveButton = (page: Page) => page.getByRole('button', { name: 'Save', exact: true });

test.describe('gym steps', { tag: ['@gym'] }, () => {
  test('save, then save again for the same day: overwrite prompt, exactly one row', async ({ page, db }) => {
    await page.goto('/gym/steps');
    await expect(page.getByText(`Steps (${TODAY_LABEL})`)).toBeVisible();
    await stepsInput(page).fill('8420');
    const first = waitForApi(page, SAVE);
    await saveButton(page).click();
    await first;
    await expect(stepsInput(page)).toHaveValue(''); // the form clears after a save

    await page.reload();
    await stepsInput(page).fill('9 000'); // spaces are stripped
    await saveButton(page).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText(`You already logged 8,420 steps for ${TODAY_LABEL}`);
    const second = waitForApi(page, SAVE);
    await dialog.getByRole('button', { name: 'Overwrite' }).click();
    await second;

    await page.reload();
    await stepsInput(page).fill('1');
    await saveButton(page).click();
    await expect(dialog).toContainText(`You already logged 9,000 steps for ${TODAY_LABEL}`);
    await dialog.getByRole('button', { name: 'Cancel' }).click();
    await expect(dialog).toHaveCount(0);

    const { rows } = await db.query(
      `SELECT to_char(date AT TIME ZONE 'America/Toronto', 'YYYY-MM-DD') AS day, steps::int AS steps FROM step_logs`,
    );
    expect(rows).toEqual([{ day: '2026-01-15', steps: 9000 }]);
  });

  test('cancelling the overwrite keeps the saved value', async ({ page, db }) => {
    await page.goto('/gym/steps');
    await stepsInput(page).fill('5000');
    const saved = waitForApi(page, SAVE);
    await saveButton(page).click();
    await saved;
    await stepsInput(page).fill('7000');
    await saveButton(page).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('You already logged 5,000 steps');
    await dialog.getByRole('button', { name: 'Cancel' }).click();

    await page.reload();
    await stepsInput(page).fill('1');
    await saveButton(page).click();
    await expect(dialog).toContainText('You already logged 5,000 steps');
    await dialog.getByRole('button', { name: 'Cancel' }).click();
    const { rows } = await db.query(`SELECT steps::int AS steps FROM step_logs`);
    expect(rows).toEqual([{ steps: 5000 }]);
  });

  test('an invalid count is refused without a request', async ({ page, db }) => {
    await page.goto('/gym/steps');
    await stepsInput(page).fill('lots');
    await saveButton(page).click();
    await expect(page.getByText('Enter a valid step count', { exact: true })).toBeVisible();
    const { rows } = await db.query(`SELECT count(*)::int AS n FROM step_logs`);
    expect(rows[0].n).toBe(0);
  });
});
