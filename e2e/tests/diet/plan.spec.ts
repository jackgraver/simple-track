import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import { createSavedMeals, dayCard, OFFSET, plannedChip } from './plan-helpers';

// /diet/edit-planned ("Plan meals"): put saved meals on days of a Sunday-first week. Epoch is Thu
// 2026-01-15, so the default week is Sun 11 .. Sat 17. Day edits are optimistic (the chip shows before
// the request), so each write awaits its response before the reload.

const ADD = 'POST /diet/meals/planned/from-saved';
const REMOVE = 'DELETE /diet/meals/planned';
const REORDER = 'POST /diet/meals/planned/reorder';
const COPY = 'POST /diet/meals/planned/copy-day';

const PLAN = '/diet/edit-planned';
const FRI = 'Fri 16';
const SAT = 'Sat 17';

async function openPlan(page: Page, query = '') {
  await page.goto(PLAN + query);
  await expect(page.getByRole('heading', { level: 3 }).first()).toBeVisible();
}

const dayMenu = (page: Page, label: string) => dayCard(page, label).getByRole('button', { name: `${label} actions` });

test.describe('diet plan: saved meals on days', { tag: ['@diet'] }, () => {
  test('add a saved meal to a day: persists and shows on /diet for that day', async ({ page, api }) => {
    await createSavedMeals(api);
    await openPlan(page);
    await expect(dayCard(page, FRI).getByText('Nothing planned')).toBeVisible();

    await dayCard(page, FRI).getByRole('button', { name: `Add meal to ${FRI}` }).click();
    const dialog = page.getByRole('dialog', { name: `Add to ${FRI}` });
    const saved = waitForApi(page, ADD);
    await dialog.getByRole('button', { name: /^Oats bowl/ }).click();
    await expect(dialog.getByText('150 / 2,500 kcal planned')).toBeVisible(); // the open overlay follows the add
    await saved;
    await dialog.getByRole('button', { name: 'Done' }).click();
    await expect(plannedChip(page, FRI, 'Oats bowl')).toBeVisible();

    await page.reload();
    await expect(plannedChip(page, FRI, 'Oats bowl')).toBeVisible();
    await expect(dayCard(page, FRI).getByText('150 / 2,500 kcal')).toBeVisible();
    await expect(dayCard(page, SAT).getByText('Nothing planned')).toBeVisible();

    await page.goto('/diet?offset=-1');
    await expect(page.getByRole('button', { name: 'More actions for Oats bowl' })).toBeVisible();
    await page.goto('/diet');
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await expect(page.getByRole('button', { name: 'More actions for Oats bowl' })).toHaveCount(0);
  });

  test('armed add (desktop): pick a saved meal, then tap several days', async ({ page, api, isMobile }) => {
    test.skip(isMobile, 'the saved-meal tray is desktop only (lg breakpoint)');
    await createSavedMeals(api);
    await openPlan(page);

    const tray = page.getByRole('group', { name: 'Saved meals' });
    await tray.getByRole('button', { name: /^Banana/ }).click();
    await expect(tray.getByRole('button', { name: /^Banana/ })).toHaveAttribute('aria-pressed', 'true');
    for (const day of ['Wed 14', FRI, SAT]) {
      const saved = waitForApi(page, ADD);
      await dayCard(page, day).getByRole('button', { name: `Add Banana to ${day}` }).click();
      await expect(plannedChip(page, day, 'Banana')).toBeVisible();
      await saved;
    }
    // Esc drops the pick: the buttons are plain "Add meal" again.
    await page.keyboard.press('Escape');
    await expect(dayCard(page, FRI).getByRole('button', { name: `Add meal to ${FRI}` })).toBeVisible();

    await page.reload();
    for (const day of ['Wed 14', FRI, SAT]) await expect(plannedChip(page, day, 'Banana')).toBeVisible();
    await expect(dayCard(page, 'Thu 15').getByText('Nothing planned')).toBeVisible();
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Banana']);
  });

  test('remove a planned meal, and move earlier / later: order persists', async ({ page, api }) => {
    const { oats, banana, chicken } = await createSavedMeals(api);
    for (const id of [oats, banana, chicken]) await api.addPlannedFromSaved(id, OFFSET.fri);
    await openPlan(page);
    const order = () => dayCard(page, FRI).getByRole('listitem').getByRole('button').allInnerTexts();
    await expect(plannedChip(page, FRI, 'Chicken and rice')).toBeVisible();
    expect((await order()).map((t) => t.split('\n')[0])).toEqual(['Oats bowl', 'Banana', 'Chicken and rice']);

    // Move earlier: Banana goes ahead of Oats bowl.
    await plannedChip(page, FRI, 'Banana').click();
    await expect(page.getByRole('menuitem', { name: 'Move earlier' })).toBeEnabled();
    let saved = waitForApi(page, REORDER);
    await page.getByRole('menuitem', { name: 'Move earlier' }).click();
    await saved;
    await expect.poll(async () => (await order()).map((t) => t.split('\n')[0])).toEqual(['Banana', 'Oats bowl', 'Chicken and rice']);

    // Move later: the last meal can't move later; Oats bowl goes behind Chicken and rice.
    await plannedChip(page, FRI, 'Chicken and rice').click();
    await expect(page.getByRole('menuitem', { name: 'Move later' })).toBeDisabled();
    await page.keyboard.press('Escape');
    await plannedChip(page, FRI, 'Oats bowl').click();
    saved = waitForApi(page, REORDER);
    await page.getByRole('menuitem', { name: 'Move later' }).click();
    await saved;
        await page.reload();
    await expect(plannedChip(page, FRI, 'Oats bowl')).toBeVisible();
    expect((await order()).map((t) => t.split('\n')[0])).toEqual(['Banana', 'Chicken and rice', 'Oats bowl']);
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Banana', 'Chicken and rice', 'Oats bowl']);

    // Remove the middle one.
    await plannedChip(page, FRI, 'Chicken and rice').click();
    const removed = waitForApi(page, REMOVE);
    await page.getByRole('menuitem', { name: 'Remove' }).click();
    await expect(plannedChip(page, FRI, 'Chicken and rice')).toHaveCount(0);
    await removed;
    await page.reload();
    await expect(plannedChip(page, FRI, 'Oats bowl')).toBeVisible();
    await expect(plannedChip(page, FRI, 'Chicken and rice')).toHaveCount(0);
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Banana', 'Oats bowl']);
  });

  test('clear day: removes every planned meal after a confirm, Cancel keeps them', async ({ page, api }) => {
    const { oats, banana } = await createSavedMeals(api);
    for (const id of [oats, banana]) await api.addPlannedFromSaved(id, OFFSET.fri);
    await api.addPlannedFromSaved(banana, OFFSET.sat);
    await openPlan(page);
    await expect(plannedChip(page, FRI, 'Banana')).toBeVisible();

    await dayMenu(page, FRI).click();
    await page.getByRole('menuitem', { name: 'Clear day' }).click();
    const confirm = page.getByRole('dialog', { name: `Clear ${FRI}?` });
    await expect(confirm).toContainText('Removes 2 planned meals (255 kcal)');
    await confirm.getByRole('button', { name: 'Cancel' }).click();
    await expect(plannedChip(page, FRI, 'Oats bowl')).toBeVisible();

    await dayMenu(page, FRI).click();
    await page.getByRole('menuitem', { name: 'Clear day' }).click();
    const removed = [waitForApi(page, REMOVE), waitForApi(page, REMOVE)];
    await page.getByRole('dialog', { name: `Clear ${FRI}?` }).getByRole('button', { name: 'Clear day' }).click();
    await expect(dayCard(page, FRI).getByText('Nothing planned')).toBeVisible();
    await Promise.all(removed);

    await page.reload();
    await expect(dayCard(page, FRI).getByText('Nothing planned')).toBeVisible();
    await expect(plannedChip(page, SAT, 'Banana')).toBeVisible(); // other days untouched
    expect(await api.plannedMeals(OFFSET.fri)).toEqual([]);
  });
});

test.describe('diet plan: copy a day', { tag: ['@diet'] }, () => {
  test('Copy to other days: picked days get the meals appended, Select empty days skips planned ones', async ({ page, api }) => {
    const { oats, banana } = await createSavedMeals(api);
    for (const id of [oats, banana]) await api.addPlannedFromSaved(id, OFFSET.thu);
    await api.addPlannedFromSaved(banana, OFFSET.sat); // Sat already has a meal
    await openPlan(page);
    await expect(plannedChip(page, 'Thu 15', 'Banana')).toBeVisible();

    await dayMenu(page, 'Thu 15').click();
    await page.getByRole('menuitem', { name: 'Copy to other days…' }).click();
    const dialog = page.getByRole('dialog', { name: 'Copy Thu 15 to…' });
    await expect(dialog).toContainText('Adds its 2 meals (255 kcal) after anything already planned.');
    await expect(dialog.getByRole('checkbox', { name: /^Sat 17 Jan, 1 planned/ })).toBeVisible();
    await dialog.getByRole('button', { name: 'Select empty days' }).click();
    await expect(dialog.getByRole('checkbox', { name: /^Fri 16 Jan, nothing planned/ })).toBeChecked();
    await expect(dialog.getByRole('checkbox', { name: /^Sat 17 Jan, 1 planned/ })).not.toBeChecked();
    // 14 candidate days after Thu 15, Sat 17 is taken.
    await expect(dialog.getByRole('button', { name: 'Copy to 13 days' })).toBeVisible();
    await dialog.getByRole('button', { name: 'Clear', exact: true }).click();
    await expect(dialog.getByRole('button', { name: 'Copy', exact: true })).toBeDisabled();

    await dialog.getByRole('checkbox', { name: /^Fri 16 Jan/ }).check();
    await dialog.getByRole('checkbox', { name: /^Sat 17 Jan/ }).check();
    const copied = waitForApi(page, COPY);
    await dialog.getByRole('button', { name: 'Copy to 2 days' }).click();
    await copied;
    await expect(page.getByText('Copied Thu 15 to 2 days')).toBeVisible();
    await expect(dialog).toBeHidden();

    await page.reload();
    await expect(plannedChip(page, 'Fri 16', 'Oats bowl')).toBeVisible();
    await expect(plannedChip(page, 'Fri 16', 'Banana')).toBeVisible();
    // Sat keeps its own meal first and gets the copies after it.
    await expect(dayCard(page, SAT).getByRole('listitem')).toHaveCount(3);
    expect(await api.plannedMeals(OFFSET.sat)).toEqual(['Banana', 'Oats bowl', 'Banana']);
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Oats bowl', 'Banana']);
    expect(await api.plannedMeals(OFFSET.thu)).toEqual(['Oats bowl', 'Banana']); // the source is unchanged
    expect(await api.plannedMeals(-3)).toEqual([]); // Sun 18, not selected
  });

  test('Copy from the previous day: shown on an empty day after a planned one', async ({ page, api }) => {
    const { oats, chicken } = await createSavedMeals(api);
    for (const id of [oats, chicken]) await api.addPlannedFromSaved(id, OFFSET.fri);
    await openPlan(page);
    await expect(plannedChip(page, FRI, 'Chicken and rice')).toBeVisible();

    // A day whose neighbour is empty has no such entry; a planned day never offers it.
    await dayMenu(page, 'Thu 15').click();
    await expect(page.getByRole('menuitem', { name: 'Copy from Wed' })).toHaveCount(0);
    await page.keyboard.press('Escape');
    await dayMenu(page, FRI).click();
    await expect(page.getByRole('menuitem', { name: /^Copy from/ })).toHaveCount(0);
    await page.keyboard.press('Escape');

    await dayMenu(page, SAT).click();
    const copied = waitForApi(page, COPY);
    await page.getByRole('menuitem', { name: 'Copy from Fri' }).click();
    await copied;
    await expect(page.getByText('Copied Fri 16 to 1 day')).toBeVisible();
    await expect(plannedChip(page, SAT, 'Chicken and rice')).toBeVisible();

    await page.reload();
    await expect(plannedChip(page, SAT, 'Oats bowl')).toBeVisible();
    expect(await api.plannedMeals(OFFSET.sat)).toEqual(['Oats bowl', 'Chicken and rice']);
    expect(await api.plannedMeals(OFFSET.fri)).toEqual(['Oats bowl', 'Chicken and rice']);
  });
});

test.describe('diet plan: weeks', { tag: ['@diet'] }, () => {
  test('week paging: Previous / Next week, "This week", past days read "No plan"', async ({ page, api }) => {
    const { banana } = await createSavedMeals(api);
    await api.addPlannedFromSaved(banana, OFFSET.fri);
    await openPlan(page);
    await expect(page.getByText('This week', { exact: true })).toBeVisible(); // the header says so; no jump-back button
    await expect(page.getByRole('button', { name: 'This week' })).toHaveCount(0);
    await expect(page.getByText('Jan 11 – 17').first()).toBeVisible();
    await expect(dayCard(page, 'Thu 15').getByText('Today')).toBeVisible();
    // Days before today are read-only history: nothing planned reads "No plan", not a warning badge.
    for (const d of ['Sun 11', 'Mon 12', 'Tue 13', 'Wed 14']) await expect(dayCard(page, d).getByText('No plan')).toBeVisible();
    await expect(dayCard(page, 'Thu 15').getByText('Nothing planned')).toBeVisible();

    await page.getByRole('button', { name: 'Previous week' }).click();
    await expect(page).toHaveURL(/offset=7/);
    await expect(page.getByRole('list', { name: 'Week of Jan 4 – 10' })).toBeVisible();
    await expect(dayCard(page, 'Sun 4').getByText('No plan')).toBeVisible();
    await expect(dayCard(page, 'Sat 10').getByText('No plan')).toBeVisible();
    await expect(page.getByText('This week', { exact: true })).toHaveCount(1); // only the jump-back button

    // The week is in the URL: it survives a reload.
    await page.reload();
    await expect(page.getByRole('list', { name: 'Week of Jan 4 – 10' })).toBeVisible();

    await page.getByRole('button', { name: 'This week' }).click();
    await expect(page.getByRole('list', { name: 'Week of Jan 11 – 17' })).toBeVisible();
    await expect(plannedChip(page, FRI, 'Banana')).toBeVisible();

    await page.getByRole('button', { name: 'Next week' }).click();
    await expect(page.getByRole('list', { name: 'Week of Jan 18 – 24' })).toBeVisible();
    await expect(dayCard(page, 'Sun 18').getByText('Nothing planned')).toBeVisible();
    await expect(page.getByText('No plan')).toHaveCount(0); // all future
  });

  test('a past day can still be edited from the plan page (no read-only guard)', { tag: '@diet' }, async ({ page, api }) => {
    // Pins today's behaviour: only the empty label differs for past days, Add is still offered.
    await createSavedMeals(api);
    await openPlan(page);
    await dayCard(page, 'Tue 13').getByRole('button', { name: 'Add meal to Tue 13' }).click();
    const saved = waitForApi(page, ADD);
    await page.getByRole('dialog', { name: 'Add to Tue 13' }).getByRole('button', { name: /^Banana/ }).click();
    await saved;
    await page.reload();
    await expect(plannedChip(page, 'Tue 13', 'Banana')).toBeVisible();
    expect(await api.plannedMeals(OFFSET.tue)).toEqual(['Banana']);
  });
});

test.describe('diet plan: empty and failing', { tag: ['@diet'] }, () => {
  test('no saved meals: empty state with a New saved meal link', async ({ page, isMobile }) => {
    await openPlan(page);
    if (!isMobile) {
      // Desktop also shows the tray.
      await expect(page.getByText('No saved meals yet.')).toBeVisible();
    }
    await dayCard(page, FRI).getByRole('button', { name: `Add meal to ${FRI}` }).click();
    const dialog = page.getByRole('dialog', { name: `Add to ${FRI}` });
    await expect(dialog.getByText('No saved meals yet')).toBeVisible();
    await dialog.getByRole('link', { name: 'New saved meal' }).click();
    // Lands on the log page in create-saved-meal mode, for that day.
    await expect(page).toHaveURL(/\/diet\/log\?.*type=/);
    await expect(page).toHaveURL(/offset=-1/);
  });
});
