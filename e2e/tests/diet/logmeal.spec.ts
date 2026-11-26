import type { Page } from '@playwright/test';
import { expect, test, waitForApi } from '../../fixtures';
import {
  FOODS,
  MEAL_NEW,
  addFood,
  dayRow,
  itemsList,
  loggedRegion,
  mealKcal,
  mealName,
  primary,
  openLog,
  qty,
  toast,
  savedPicker,
  setQty,
} from './logmeal-helpers';

// The meal builder on /diet/log: building, logging and saving a meal from catalogue foods, grouping,
// amounts, and the sessionStorage draft. Logging navigates back to /diet, so every write is verified
// there after a reload. Amounts are in servings: 100 g of Chicken Breast is 1 serving (165 kcal).

const DRAFT_KEY = 'simpletracker:diet:logmeal-draft:v2:log';

const draftInStorage = (page: Page) => page.evaluate((key) => sessionStorage.getItem(key), DRAFT_KEY);

test.describe('diet: log meal', { tag: ['@diet'] }, () => {
  test('log a catalogue meal: right kcal on /diet after a reload', async ({ page, api }) => {
    await openLog(page);
    await mealName(page).fill('Chicken and eggs');
    await addFood(page, FOODS.chicken.name);
    await addFood(page, FOODS.egg.name);
    await setQty(page, FOODS.chicken.name, 200);
    await setQty(page, FOODS.egg.name, 3);
    // 200 g chicken (2 servings, 330 kcal) + 3 eggs (216 kcal)
    await expect(mealKcal(page)).toContainText(/546\s*kcal/);
    await expect(itemsList(page).getByRole('listitem').filter({ hasText: FOODS.chicken.name })).toContainText(/330\s*kcal/);

    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);

    await page.reload();
    await expect(dayRow(page, 'Chicken and eggs')).toBeVisible();
    await expect(loggedRegion(page).getByRole('button', { name: /^Chicken and eggs 546 kcal/ })).toBeVisible();
    await expect(page.getByText('546 / 2,500', { exact: true })).toBeVisible();
    const logged = (await api.dietDay()).day.loggedMeals;
    expect(logged).toHaveLength(1);
    expect(logged[0].meal.items.map((i: any) => [i.food_id, i.amount])).toEqual([
      [FOODS.chicken.id, 2],
      [FOODS.egg.id, 3],
    ]);
  });

  test('log for another day: ?offset= lands the meal on that day, not today', async ({ page }) => {
    await openLog(page, 'type=log&offset=2');
    await mealName(page).fill('Two days ago');
    await addFood(page, FOODS.banana.name);
    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet\?offset=2$/);

    await page.reload();
    await expect(dayRow(page, 'Two days ago')).toBeVisible();
    await page.goto('/diet');
    await expect(page.getByText('Nothing logged yet')).toBeVisible();
    await expect(dayRow(page, 'Two days ago')).toHaveCount(0);
  });

  test('"Log and save meal": logged today and offered in the Saved meals picker', async ({ page, api }) => {
    await openLog(page);
    await mealName(page).fill('Yogurt bowl');
    await addFood(page, FOODS.yogurt.name);
    await addFood(page, FOODS.banana.name);

    const saved = waitForApi(page, MEAL_NEW);
    await page.getByRole('button', { name: 'More log options' }).click();
    await page.getByRole('menuitem', { name: 'Log and save meal' }).click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);
    await expect(toast(page, 'Meal logged and saved for later!')).toBeVisible();

    await page.reload();
    await expect(loggedRegion(page).getByRole('button', { name: /^Yogurt bowl 205 kcal/ })).toBeVisible();

    await openLog(page);
    await expect(savedPicker(page).getByRole('button', { name: /^Yogurt bowl 205 kcal/ })).toBeVisible();
    const savedMeals = (await api.savedMeals()).saved_meals;
    expect(savedMeals.map((m) => m.name)).toEqual(['Yogurt bowl']);
    expect(savedMeals[0].items.map((i) => [i.food_id, i.amount])).toEqual([
      [FOODS.yogurt.id, 1],
      [FOODS.banana.id, 1],
    ]);
  });

  test('start from a saved meal, change an amount, log: the saved meal stays as it was', async ({ page, api }) => {
    await api.createSavedMeal('Chicken and rice', [
      { food_id: FOODS.chicken.id, amount: 1 },
      { food_id: FOODS.rice.id, amount: 2 },
    ]);
    await openLog(page);
    await savedPicker(page).getByRole('button', { name: /^Chicken and rice/ }).click();
    await expect(qty(page, FOODS.chicken.name)).toHaveValue('100');
    await expect(qty(page, FOODS.rice.name)).toHaveValue('200');
    await expect(mealKcal(page)).toContainText(/425\s*kcal/);

    await setQty(page, FOODS.rice.name, 300);
    await expect(mealKcal(page)).toContainText(/555\s*kcal/);
    // Picking a saved meal keeps the name already in the box (the time-of-day default), see notes.
    await mealName(page).fill('Chicken and rice, big');

    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);

    await page.reload();
    await expect(loggedRegion(page).getByRole('button', { name: /^Chicken and rice, big 555 kcal/ })).toBeVisible();
    const savedMeals = (await api.savedMeals()).saved_meals;
    expect(savedMeals).toHaveLength(1);
    expect(savedMeals[0].name).toBe('Chicken and rice');
    expect(savedMeals[0].items.map((i) => [i.food_id, i.amount])).toEqual([
      [FOODS.chicken.id, 1],
      [FOODS.rice.id, 2],
    ]);
    await openLog(page);
    await expect(savedPicker(page).getByRole('button', { name: /^Chicken and rice 425 kcal/ })).toBeVisible();
  });

  test('group two items, name the group, log: the group survives in the logged meal', async ({ page, api }) => {
    await openLog(page);
    await mealName(page).fill('Bowl');
    await addFood(page, FOODS.chicken.name);
    await addFood(page, FOODS.rice.name);
    await addFood(page, FOODS.broccoli.name);

    await page.getByRole('button', { name: 'Select', exact: true }).click();
    await page.getByRole('checkbox', { name: `Select ${FOODS.chicken.name}` }).check();
    await page.getByRole('checkbox', { name: `Select ${FOODS.rice.name}` }).check();
    await page.getByRole('button', { name: 'Group', exact: true }).click();
    await page.getByRole('textbox', { name: 'Group name' }).fill('Protein and rice');
    await expect(page.getByRole('button', { name: 'Group options' })).toBeVisible();

    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);

    await page.reload();
    const row = loggedRegion(page).getByRole('button', { name: /^Bowl 329 kcal/ });
    await row.click();
    // The expanded meal lists the group (collapsible header with its name) and its two items.
    const group = loggedRegion(page).getByRole('button', { name: /^Protein and rice/ });
    await expect(group).toBeVisible();
    await expect(group).toHaveAttribute('aria-expanded', 'true');
    await expect(loggedRegion(page).getByText('100g Chicken Breast')).toBeVisible();
    await expect(loggedRegion(page).getByText('100g White Rice (cooked)')).toBeVisible();
    await expect(loggedRegion(page).getByText('100g Broccoli')).toBeVisible();

    const items = (await api.dietDay()).day.loggedMeals[0].meal.items as {
      food_id: number;
      group_id: string;
      group_label: string;
    }[];
    const byFood = Object.fromEntries(items.map((i) => [i.food_id, i]));
    expect(byFood[FOODS.chicken.id].group_label).toBe('Protein and rice');
    expect(byFood[FOODS.chicken.id].group_id).not.toBe('');
    expect(byFood[FOODS.rice.id].group_id).toBe(byFood[FOODS.chicken.id].group_id);
    expect(byFood[FOODS.broccoli.id].group_id).toBe('');
  });

  test('quantity 0 removes the item; Halve and Double amount update the kcal', async ({ page }) => {
    await openLog(page);
    await addFood(page, FOODS.chicken.name);
    await addFood(page, FOODS.egg.name);
    const chicken = itemsList(page).getByRole('listitem').filter({ hasText: FOODS.chicken.name });
    await expect(mealKcal(page)).toContainText(/237\s*kcal/);

    await page.getByRole('button', { name: `Options for ${FOODS.chicken.name}` }).click();
    await page.getByRole('menuitem', { name: 'Double amount' }).click();
    await expect(qty(page, FOODS.chicken.name)).toHaveValue('200');
    await expect(chicken).toContainText(/330\s*kcal/);
    await expect(mealKcal(page)).toContainText(/402\s*kcal/);

    await page.getByRole('button', { name: `Options for ${FOODS.chicken.name}` }).click();
    await page.getByRole('menuitem', { name: 'Halve amount' }).click();
    await page.getByRole('button', { name: `Options for ${FOODS.chicken.name}` }).click();
    await page.getByRole('menuitem', { name: 'Halve amount' }).click();
    await expect(qty(page, FOODS.chicken.name)).toHaveValue('50');
    await expect(chicken).toContainText(/83\s*kcal/);

    await setQty(page, FOODS.egg.name, 0);
    await expect(itemsList(page).getByRole('listitem').filter({ hasText: FOODS.egg.name })).toHaveCount(0);
    await expect(page.getByRole('heading', { name: 'Items 1' })).toBeVisible();
    await expect(mealKcal(page)).toContainText(/83\s*kcal/);
  });

  test('the draft survives a reload and is cleared after a log', async ({ page }) => {
    await openLog(page);
    await mealName(page).fill('Draft lunch');
    await addFood(page, FOODS.oats.name);
    await setQty(page, FOODS.oats.name, 80);
    await expect(mealKcal(page)).toContainText(/300\s*kcal/);
    await expect.poll(() => draftInStorage(page)).not.toBeNull();

    await page.reload();
    await expect(page.getByRole('heading', { name: 'Add foods' })).toBeVisible();
    await expect(mealName(page)).toHaveValue('Draft lunch');
    await expect(qty(page, FOODS.oats.name)).toHaveValue('80');
    await expect(mealKcal(page)).toContainText(/300\s*kcal/);

    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);
    expect(await draftInStorage(page)).toBeNull();

    await openLog(page);
    await expect(page.getByText('No foods yet')).toBeVisible();
    // Opening the page again only re-saves the default name; none of the logged items come back.
    await expect.poll(async () => JSON.parse((await draftInStorage(page)) ?? '{"items":[]}').items).toEqual([]);
  });

  test('a failed log shows an error toast, keeps the draft, and the retry logs once', async ({ page, api, e2e }) => {
    await openLog(page);
    await mealName(page).fill('Flaky lunch');
    await addFood(page, FOODS.egg.name);
    await setQty(page, FOODS.egg.name, 2);

    await e2e.fault({ table: 'meals', op: 'create', times: 1 });
    await primary(page, 'Log meal').click();
    await expect(toast(page, /^Log meal failed!/)).toBeVisible();
    await expect(page).toHaveURL(/\/diet\/log/);
    await expect(qty(page, FOODS.egg.name)).toHaveValue('2');
    expect(await draftInStorage(page)).not.toBeNull();
    expect((await api.dietDay()).day.loggedMeals).toHaveLength(0);

    // The draft is what a reload brings back, so nothing the user typed is lost.
    await page.reload();
    await expect(mealName(page)).toHaveValue('Flaky lunch');
    await expect(qty(page, FOODS.egg.name)).toHaveValue('2');

    const saved = waitForApi(page, MEAL_NEW);
    await primary(page, 'Log meal').click();
    await saved;
    await expect(page).toHaveURL(/\/diet$/);
    await page.reload();
    await expect(loggedRegion(page).getByRole('button', { name: /^Flaky lunch 144 kcal/ })).toBeVisible();
    expect((await api.dietDay()).day.loggedMeals).toHaveLength(1);
  });
});
