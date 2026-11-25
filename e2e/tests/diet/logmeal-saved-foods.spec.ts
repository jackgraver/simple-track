import { expect, test, waitForApi } from '../../fixtures';
import {
  FOODS,
  FOOD_NEW,
  SAVED_MEAL_NEW,
  addFood,
  foodPicker,
  mealName,
  openLog,
  primary,
  qty,
  savedPicker,
  setQty,
  toast,
} from './logmeal-helpers';

// New saved meals (`/diet/log?type=create`) and new foods created from the picker's search.

test.describe('diet: new saved meal', { tag: ['@diet'] }, () => {
  test('name + items, "Save meal": offered in the Saved meals picker after a reload', async ({ page, api }) => {
    await openLog(page, 'type=create');
    await expect(page.getByRole('article').getByRole('heading', { name: 'Create New Meal' })).toBeVisible();
    // A saved meal isn't tied to a day: no "Today after logging", no "Log and save" menu.
    await expect(page.getByRole('region', { name: 'Today after logging' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'More log options' })).toHaveCount(0);

    await mealName(page).fill('Overnight oats');
    await addFood(page, FOODS.oats.name);
    await addFood(page, FOODS.banana.name);
    await setQty(page, FOODS.oats.name, 80);

    const saved = waitForApi(page, SAVED_MEAL_NEW);
    await primary(page, 'Save meal').click();
    await saved;
    await expect(toast(page, 'Saved meal created!')).toBeVisible();
    // The form empties, ready for the next one.
    await expect(page.getByText('No foods yet')).toBeVisible();
    await expect(mealName(page)).toHaveValue('');

    // 2 servings of oats (300) + a banana (105)
    await openLog(page);
    await expect(savedPicker(page).getByRole('button', { name: /^Overnight oats 405 kcal.*2 items/ })).toBeVisible();
    const savedMeals = (await api.savedMeals()).saved_meals;
    expect(savedMeals.map((m) => m.name)).toEqual(['Overnight oats']);
    expect(savedMeals[0].items.map((i) => [i.food_id, i.amount])).toEqual([
      [FOODS.oats.id, 2],
      [FOODS.banana.id, 1],
    ]);
  });

  test('needs a name and at least one food: a toast, and nothing is saved', async ({ page, api }) => {
    await openLog(page, 'type=create');
    const needs = toast(page, 'Add a name and at least one food.');

    // Nothing at all.
    await primary(page, 'Save meal').click();
    await expect(needs).toBeVisible();
    await page.getByRole('button', { name: 'Dismiss' }).click();
    await expect(needs).toHaveCount(0);

    // Foods but no name.
    await addFood(page, FOODS.egg.name);
    await primary(page, 'Save meal').click();
    await expect(needs).toBeVisible();
    await page.getByRole('button', { name: 'Dismiss' }).click();
    await expect(needs).toHaveCount(0);

    // A name but no foods.
    await qty(page, FOODS.egg.name).fill('0');
    await qty(page, FOODS.egg.name).press('Enter');
    await expect(page.getByText('No foods yet')).toBeVisible();
    await mealName(page).fill('Empty meal');
    await primary(page, 'Save meal').click();
    await expect(needs).toBeVisible();

    expect((await api.savedMeals()).saved_meals).toHaveLength(0);
  });
});

test.describe('diet: create a food from the picker', { tag: ['@diet'] }, () => {
  test('search, "Create", fill the dialog: added to the meal and in the catalogue after a reload', async ({ page, api }) => {
    await openLog(page);
    await foodPicker(page).getByRole('textbox', { name: 'Search' }).fill('Protein Bar');
    await foodPicker(page).getByRole('button', { name: 'Create "Protein Bar"' }).click();

    const dialog = page.getByRole('dialog', { name: 'Create Protein Bar' });
    await expect(dialog.getByLabel('Food name')).toHaveValue('Protein Bar');
    await dialog.getByLabel('Calories').fill('200');
    await dialog.getByLabel('Protein').fill('20');
    await dialog.getByLabel('Carbs').fill('15');
    await dialog.getByLabel('Fat').fill('7');
    await dialog.getByLabel('Fiber').fill('3');
    await dialog.getByLabel('Serving type').selectOption('Grams');
    await dialog.getByLabel('Serving amount').fill('50');

    const created = waitForApi(page, FOOD_NEW);
    await dialog.getByRole('button', { name: 'Create', exact: true }).click();
    await created;
    await expect(toast(page, 'Food Created Successfully and Added!')).toBeVisible();
    await expect(dialog).toBeHidden();

    // One serving (50 g, 200 kcal) is in the meal.
    await expect(qty(page, 'Protein Bar')).toHaveValue('50');
    await expect(page.getByRole('region', { name: 'This meal' })).toContainText(/200\s*kcal/);

    await page.reload();
    await foodPicker(page).getByRole('textbox', { name: 'Search' }).fill('Protein');
    await expect(foodPicker(page).getByRole('button', { name: /^Protein Bar 200C 20P 15C 7F 3Fi/ })).toBeVisible();
    const bars = (await api.allFoods()).foods.filter((f) => f.name === 'Protein Bar');
    expect(bars).toHaveLength(1);
    expect(bars[0]).toMatchObject({ calories: 200 });
  });

  // PostFood maps the unique-index violation to 409 (the BACKLOG note that this was a 500 is out of
  // date). The browser logs every 4xx as a console error, hence allowedErrors.
  test.describe('duplicate name', () => {
    test.use({ allowedErrors: [/409/] });

    test('is refused with a readable message, the dialog stays open, no second food', async ({ page, api }) => {
      await openLog(page);
      await foodPicker(page).getByRole('textbox', { name: 'Search' }).fill(FOODS.egg.name);
      await foodPicker(page).getByRole('button', { name: `Create "${FOODS.egg.name}"` }).click();

      const dialog = page.getByRole('dialog', { name: `Create ${FOODS.egg.name}` });
      await dialog.getByLabel('Calories').fill('80');
      await dialog.getByLabel('Serving type').selectOption('Unit');
      await dialog.getByLabel('Serving amount').fill('1');
      await dialog.getByRole('button', { name: 'Create', exact: true }).click();

      await expect(toast(page, 'a food with this name already exists')).toBeVisible();
      await expect(dialog).toBeVisible();
      await expect(page.getByText('No foods yet')).toBeVisible();

      expect((await api.allFoods()).foods.filter((f) => f.name === FOODS.egg.name)).toHaveLength(1);
      // The dialog's own Create button is usable again (not stuck loading) so the user can rename.
      await dialog.getByLabel('Food name').fill('Duck Egg');
      const created = waitForApi(page, FOOD_NEW);
      await dialog.getByRole('button', { name: 'Create', exact: true }).click();
      await created;
      await expect(qty(page, 'Duck Egg')).toHaveValue('1');
    });
  });
});
