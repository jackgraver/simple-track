import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import {
  FOODS,
  MEAL_EDIT_LOGGED,
  MEAL_LOG_EDITED,
  dayRow,
  loggedRegion,
  primary,
  qty,
  setQty,
  toast,
} from './logmeal-helpers';

// Editing on /diet/log: a logged catalogue meal ("Update", `?type=edit`) and a planned meal
// ("Edit and log", `?type=editlogged`). The server repoints the day's log at a fresh copy of the meal in
// one transaction (EditLoggedMeal), so a failed edit must leave the old meal exactly as it was.

const DINNER = [
  { food_id: FOODS.chicken.id, amount: 1 },
  { food_id: FOODS.rice.id, amount: 1 },
]; // 165 + 130 = 295 kcal

async function openEdit(page: Page, name: string) {
  await page.goto('/diet');
  await dayRow(page, name).click();
  await page.getByRole('menuitem', { name: 'Edit', exact: true }).click();
  await expect(page.getByRole('article').getByRole('heading', { name: 'Edit Logged Meal' })).toBeVisible();
  await expect(qty(page, FOODS.chicken.name)).toHaveValue('100');
}

const amounts = (day: any) =>
  (day.day.loggedMeals as any[]).map((l) => l.meal.items.map((i: any) => [i.food_id, i.amount]));

test.describe('diet: edit a logged meal', { tag: ['@diet'] }, () => {
  test('Update replaces the meal: after a reload exactly one logged meal with the new amounts', async ({ page, api, db }) => {
    await api.logCatalogueMeal('Dinner', DINNER);
    await openEdit(page, 'Dinner');
    await expect(page.getByRole('textbox', { name: 'Meal name' })).toHaveValue('Dinner');

    await setQty(page, FOODS.chicken.name, 200);
    await setQty(page, FOODS.rice.name, 150);
    const saved = waitForApi(page, MEAL_EDIT_LOGGED);
    await primary(page, 'Update').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);

    await page.reload();
    // 200 g chicken (330) + 150 g rice (195)
    await expect(loggedRegion(page).getByRole('button', { name: /^Dinner 525 kcal/ })).toBeVisible();
    await expect(loggedRegion(page).getByRole('button', { name: /^Dinner/ })).toHaveCount(1);
    await expect(dayRow(page, 'Dinner')).toHaveCount(1);
    expect(amounts(await api.dietDay())).toEqual([
      [
        [FOODS.chicken.id, 2],
        [FOODS.rice.id, 1.5],
      ],
    ]);
    const logs = await db.query('select count(*)::int as n from day_logs where deleted_at is null');
    expect(logs.rows[0].n).toBe(1);
  });

  // The old "create, then delete" edit is now one transaction. Whichever statement fails, the user
  // keeps the original meal (no loss, no duplicate) and a retry then goes through.
  for (const fault of [
    { table: 'meals', op: 'create', what: 'creating the new copy of the meal' },
    { table: 'day_logs', op: 'update', what: 'pointing the day at the new copy' },
  ] as const) {
    test(`a failed Update (${fault.what}) loses nothing and duplicates nothing; the retry works`, async ({ page, api, db, e2e }) => {
      await api.logCatalogueMeal('Dinner', DINNER);
      await openEdit(page, 'Dinner');
      await setQty(page, FOODS.chicken.name, 200);
      const mealsBefore = (await db.query('select count(*)::int as n from meals')).rows[0].n;

      await e2e.fault({ table: fault.table, op: fault.op, times: 1 });
      await primary(page, 'Update').click();
      await expect(toast(page, /^Update Failed!/)).toBeVisible();
      await expect(page).toHaveURL(/\/diet\/log\?type=edit/);
      // The edit is still on screen so the user can retry.
      await expect(qty(page, FOODS.chicken.name)).toHaveValue('200');

      // Nothing changed server-side: one logged meal with its old amounts, no half-written copy.
      expect(amounts(await api.dietDay())).toEqual([
        [
          [FOODS.chicken.id, 1],
          [FOODS.rice.id, 1],
        ],
      ]);
      expect((await db.query('select count(*)::int as n from meals')).rows[0].n).toBe(mealsBefore);
      await page.goto('/diet');
      await expect(loggedRegion(page).getByRole('button', { name: /^Dinner 295 kcal/ })).toBeVisible();

      // Retry from a fresh edit page (the fault was one-shot).
      await openEdit(page, 'Dinner');
      await setQty(page, FOODS.chicken.name, 200);
      const saved = waitForApi(page, MEAL_EDIT_LOGGED);
      await primary(page, 'Update').click();
      await saved;
      await page.reload();
      await expect(loggedRegion(page).getByRole('button', { name: /^Dinner 460 kcal/ })).toBeVisible();
      await expect(dayRow(page, 'Dinner')).toHaveCount(1);
      expect(amounts(await api.dietDay())).toHaveLength(1);
    });
  }
});

test.describe('diet: edit and log a planned meal', { tag: ['@diet'] }, () => {
  test('logs the edited copy; the planned template and the saved meal are untouched', async ({ page, api, db }) => {
    const savedId = await api.createSavedMeal('Plan A', DINNER);
    await api.addPlannedFromSaved(savedId, 0);
    await page.goto('/diet');
    await expect(page.getByRole('region', { name: 'Planned' })).toContainText('Plan A');
    await dayRow(page, 'Plan A').click();
    await page.getByRole('menuitem', { name: 'Edit and log' }).click();
    await expect(page.getByRole('article').getByRole('heading', { name: 'Log Meal' })).toBeVisible();
    await expect(qty(page, FOODS.chicken.name)).toHaveValue('100');

    await setQty(page, FOODS.chicken.name, 200);
    const saved = waitForApi(page, MEAL_LOG_EDITED);
    await primary(page, 'Log').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);

    await page.reload();
    // Logged with the edited amounts; the planned list has nothing left to log.
    await expect(loggedRegion(page).getByRole('button', { name: /^Plan A 460 kcal/ })).toBeVisible();
    await expect(page.getByRole('region', { name: 'Planned' })).toContainText('Nothing left to log');
    expect(amounts(await api.dietDay())).toEqual([
      [
        [FOODS.chicken.id, 2],
        [FOODS.rice.id, 1],
      ],
    ]);

    // The planned meal is marked logged, and its own meal still has the planned amounts.
    const planned = await db.query(
      `select pm.logged, mi.food_id, mi.amount::float8 as amount
         from planned_meals pm join meal_items mi on mi.meal_id = pm.meal_id
        where pm.deleted_at is null order by mi.id`,
    );
    expect(planned.rows.map((r) => [r.logged, r.food_id, r.amount])).toEqual([
      [true, String(FOODS.chicken.id), 1],
      [true, String(FOODS.rice.id), 1],
    ]);
    const savedMeals = (await api.savedMeals()).saved_meals;
    expect(savedMeals).toHaveLength(1);
    expect(savedMeals[0].items.map((i) => [i.food_id, i.amount])).toEqual([
      [FOODS.chicken.id, 1],
      [FOODS.rice.id, 1],
    ]);
  });
});
