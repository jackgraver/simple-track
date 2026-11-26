import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';

// /diet/water-presets ("Water sizes & goal"). Drink sizes are real data (tracking API, stored in oz);
// the daily goal and the display unit are browser-local prefs (localStorage, composables/water/
// useWaterPrefs), so those are checked with a reload, not against the database.
// Notes and findings: e2e/notes/diet-targets.md.

const CREATE = 'POST /tracking/water/presets';
const UPDATE = 'PUT /tracking/water/presets/:id';
const DELETE = 'DELETE /tracking/water/presets/:id';

async function openPage(page: Page) {
  await page.goto('/diet/water-presets');
  await expect(page.getByRole('heading', { name: 'Water sizes & goal' })).toBeVisible();
}

const goal = (page: Page) => page.getByRole('textbox', { name: 'Daily goal' });
const unit = (page: Page) => page.getByRole('combobox', { name: 'Unit' });
const row = (page: Page, name: string) => page.getByRole('listitem').filter({ hasText: name });
/** The open edit form (the list item that holds the Save button). */
const editForm = (page: Page) =>
  page.getByRole('listitem').filter({ has: page.getByRole('button', { name: 'Save', exact: true }) });
/** The preset's quick-add button on /diet. */
const presetButton = (page: Page, name: string, amount: string) =>
  page.getByRole('button', { name: `Log ${name}, ${amount} of water` });

/**
 * Read a toast, then dismiss every toast so the next assertion can't match a stale message. The goal
 * field saves on `change` (blur) and the Save goal button saves again on click, so one "Save goal"
 * click toasts twice (see the notes): hence `.first()` and dismissing until none is left.
 * (After a bad goal the field is restored to the saved value, so the button's second save toasts "Goal saved".)
 */
async function expectToast(page: Page, message: string) {
  const toast = page.getByText(message, { exact: true });
  await expect(toast.first()).toBeVisible();
  const dismiss = page.getByRole('button', { name: 'Dismiss' });
  // A toast can leave on its own while we click: retry the dismissal instead of waiting on a gone button.
  await expect(async () => {
    if (await dismiss.count()) await dismiss.first().click({ timeout: 1_000 });
    await expect(toast).toHaveCount(0, { timeout: 500 });
  }).toPass({ timeout: 10_000 });
}

test.describe('diet water sizes: add, edit, delete', { tag: ['@diet'] }, () => {
  test('add a size: persists, shows as a quick-add button on /diet', async ({ page, api }) => {
    await openPage(page);
    await expect(page.getByText('No sizes yet')).toBeVisible();
    await page.getByLabel('Label').fill('Large bottle');
    await page.getByLabel('Amount').fill('24');
    const saved = waitForApi(page, CREATE);
    await page.getByRole('button', { name: 'Add size' }).click();
    await saved;
    await expectToast(page, 'Size created');
    await expect(row(page, 'Large bottle')).toContainText('24 oz');
    await expect(page.getByLabel('Label')).toHaveValue(''); // the form clears for the next size

    await page.reload();
    await expect(row(page, 'Large bottle')).toContainText('24 oz');
    expect(await api.drinkSizes()).toMatchObject([{ name: 'Large bottle', amount_oz: 24 }]);

    await page.goto('/diet');
    await expect(presetButton(page, 'Large bottle', '24 oz')).toBeVisible();
  });

  test('edit a size: new name and amount persist and reach /diet; cancel changes nothing', async ({ page, api }) => {
    await api.createDrinkSize('Large bottle', 24);
    await api.createDrinkSize('Cup', 8);
    await openPage(page);

    // Cancel first: the edit form closes and the row is untouched.
    await row(page, 'Large bottle').getByRole('button', { name: 'Edit size' }).click();
    await editForm(page).getByLabel('Label').fill('Throwaway');
    await editForm(page).getByRole('button', { name: 'Cancel' }).click();
    await expect(row(page, 'Large bottle')).toContainText('24 oz');
    await expect(page.getByText('Throwaway')).toHaveCount(0);

    await row(page, 'Large bottle').getByRole('button', { name: 'Edit size' }).click();
    const form = editForm(page);
    await expect(form.getByLabel('Label')).toHaveValue('Large bottle');
    await expect(form.getByLabel('Amount')).toHaveValue('24');
    await form.getByLabel('Label').fill('Huge bottle');
    await form.getByLabel('Amount').fill('32');
    const saved = waitForApi(page, UPDATE);
    await form.getByRole('button', { name: 'Save', exact: true }).click();
    await saved;
    await expectToast(page, 'Updated');
    await expect(row(page, 'Huge bottle')).toContainText('32 oz');

    await page.reload();
    await expect(row(page, 'Huge bottle')).toContainText('32 oz');
    await expect(page.getByText('Large bottle')).toHaveCount(0);
    await expect(row(page, 'Cup')).toContainText('8 oz'); // the other size is untouched
    const sizes = (await api.drinkSizes()).map((s) => [s.name, s.amount_oz]);
    expect(sizes).toHaveLength(2);
    expect(sizes).toContainEqual(['Huge bottle', 32]);
    expect(sizes).toContainEqual(['Cup', 8]);

    await page.goto('/diet');
    await expect(presetButton(page, 'Huge bottle', '32 oz')).toBeVisible();
    await expect(presetButton(page, 'Large bottle', '24 oz')).toHaveCount(0);
  });

  test('delete a size: confirm removes it for good, the other size survives', async ({ page, api }) => {
    await api.createDrinkSize('Large bottle', 24);
    await api.createDrinkSize('Cup', 8);
    await openPage(page);

    // Cancel keeps it.
    await row(page, 'Large bottle').getByRole('button', { name: 'Delete size' }).click();
    const dialog = page.getByRole('dialog', { name: 'Delete drink size?' });
    await dialog.getByRole('button', { name: 'Cancel' }).click();
    await expect(dialog).toBeHidden();
    await expect(row(page, 'Large bottle')).toBeVisible();

    await row(page, 'Large bottle').getByRole('button', { name: 'Delete size' }).click();
    const deleted = waitForApi(page, DELETE);
    await dialog.getByRole('button', { name: 'Delete' }).click();
    await deleted;
    await expectToast(page, 'Deleted');
    await expect(row(page, 'Large bottle')).toHaveCount(0);

    await page.reload();
    await expect(row(page, 'Large bottle')).toHaveCount(0);
    await expect(row(page, 'Cup')).toContainText('8 oz');
    expect((await api.drinkSizes()).map((s) => s.name)).toEqual(['Cup']);

    await page.goto('/diet');
    await expect(presetButton(page, 'Cup', '8 oz')).toBeVisible();
    await expect(presetButton(page, 'Large bottle', '24 oz')).toHaveCount(0);
  });
});

test.describe('diet water sizes: validation', { tag: ['@diet'] }, () => {
  test('add: an empty name or a bad amount toasts and creates nothing', async ({ page, api }) => {
    await openPage(page);
    const add = page.getByRole('button', { name: 'Add size' });

    await page.getByLabel('Amount').fill('24');
    await add.click();
    await expectToast(page, 'Name is required');

    await page.getByLabel('Label').fill('Large bottle');
    for (const bad of ['', '0', '-5', 'abc']) {
      await page.getByLabel('Amount').fill(bad);
      await add.click();
      await expectToast(page, 'Enter a valid amount');
    }
    // The typed name is kept for the retry.
    await expect(page.getByLabel('Label')).toHaveValue('Large bottle');
    await expect(page.getByText('No sizes yet')).toBeVisible();
    expect(await api.drinkSizes()).toHaveLength(0);
  });

  test('edit: an empty name or a bad amount toasts, keeps the form open and changes nothing', async ({ page, api }) => {
    await api.createDrinkSize('Large bottle', 24);
    await openPage(page);
    await row(page, 'Large bottle').getByRole('button', { name: 'Edit size' }).click();
    const form = editForm(page);
    const save = form.getByRole('button', { name: 'Save', exact: true });

    await form.getByLabel('Label').fill('   ');
    await save.click();
    await expectToast(page, 'Name is required');

    await form.getByLabel('Label').fill('Huge bottle');
    await form.getByLabel('Amount').fill('0');
    await save.click();
    await expectToast(page, 'Enter a valid amount');
    await expect(form).toBeVisible();

    expect(await api.drinkSizes()).toMatchObject([{ name: 'Large bottle', amount_oz: 24 }]);
  });
});

test.describe('diet water goal and unit (local prefs)', { tag: ['@diet'] }, () => {
  test('daily goal in oz: saved, kept after reload, shown on /diet', async ({ page }) => {
    await openPage(page);
    await expect(goal(page)).toHaveValue('64'); // default
    await goal(page).fill('100');
    await page.getByRole('button', { name: 'Save goal' }).click();
    await expectToast(page, 'Goal saved');

    await page.reload();
    await expect(goal(page)).toHaveValue('100');
    await page.goto('/diet');
    await expect(page.getByText('0 / 100 oz')).toBeVisible();
  });

  test('unit ml: goal and sizes are shown in ml after reload and on /diet (amounts stay oz underneath)', async ({ page, api }) => {
    await api.createDrinkSize('Cup', 8);
    await openPage(page);
    await expect(row(page, 'Cup')).toContainText('8 oz');

    await unit(page).selectOption('ml');
    await expect(row(page, 'Cup')).toContainText('236.6 ml');
    await expect(goal(page)).toHaveValue('1892.7'); // the 64 oz default
    await goal(page).fill('2000');
    await page.getByRole('button', { name: 'Save goal' }).click();
    await expectToast(page, 'Goal saved');

    await page.reload();
    await expect(unit(page)).toHaveValue('ml');
    await expect(goal(page)).toHaveValue('2000');
    await expect(row(page, 'Cup')).toContainText('236.6 ml');
    expect((await api.drinkSizes())[0].amount_oz).toBe(8); // the stored amount did not change

    await page.goto('/diet');
    await expect(page.getByText('0 / 2000 ml')).toBeVisible();
    await expect(presetButton(page, 'Cup', '236.6 ml')).toBeVisible();
  });

  test('a bad goal toasts, keeps the saved goal and restores the field', async ({ page }) => {
    await openPage(page);
    for (const bad of ['0', '-3', 'abc', '']) {
      await goal(page).fill(bad);
      await page.getByRole('button', { name: 'Save goal' }).click();
      await expectToast(page, 'Enter a positive goal');
      await expect(goal(page)).toHaveValue('64');
    }
    await page.goto('/diet');
    await expect(page.getByText('0 / 64 oz')).toBeVisible();
  });
});
