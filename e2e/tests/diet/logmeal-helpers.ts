import type { Locator, Page } from '@playwright/test';
import { expect } from '../../fixtures';

// Shared helpers for the meal builder (/diet/log). Not a spec: Playwright only runs *.spec.ts.
// Seed foods (IDs follow this order). `amount` on a meal item is in servings: 1 serving of Chicken
// Breast is 100 g (165 kcal), of Oats 40 g, of Egg one piece.

export const FOODS = {
  chicken: { id: 1, name: 'Chicken Breast', kcal: 165 },
  rice: { id: 2, name: 'White Rice (cooked)', kcal: 130 },
  egg: { id: 3, name: 'Egg', kcal: 72 },
  oats: { id: 4, name: 'Oats', kcal: 150 },
  banana: { id: 5, name: 'Banana', kcal: 105 },
  yogurt: { id: 6, name: 'Greek Yogurt', kcal: 100 },
  broccoli: { id: 7, name: 'Broccoli', kcal: 34 },
  oil: { id: 8, name: 'Olive Oil', kcal: 119 },
} as const;

export const MEAL_NEW = 'POST /diet/meals/meal/new';
export const MEAL_EDIT_LOGGED = 'POST /diet/meals/meal/editlogged';
export const MEAL_LOG_EDITED = 'POST /diet/meals/meal/logedited';
export const SAVED_MEAL_NEW = 'POST /diet/meals/saved-meal/new';
export const FOOD_NEW = 'POST /diet/foods';

/** The meal builder, opened in `?type=log` mode (or whatever query the test passes). */
export async function openLog(page: Page, query = 'type=log') {
  await page.goto(`/diet/log?${query}`);
  await expect(page.getByRole('heading', { name: 'Add foods' })).toBeVisible();
}

/** The "Add foods" and "Saved meals" side panels (two <aside>s, told apart by their heading). */
export const foodPicker = (page: Page) =>
  page.getByRole('complementary').filter({ has: page.getByRole('heading', { name: 'Add foods' }) });
export const savedPicker = (page: Page) =>
  page.getByRole('complementary').filter({ has: page.getByRole('heading', { name: 'Saved meals' }) });

/** The builder's primary action ("Log meal", "Save meal", "Update", "Log"). On mobile the page header
 *  has its own "Log meal" menu button, so scope to the meal card. */
export const primary = (page: Page, label: string) =>
  page.getByRole('article').getByRole('button', { name: label, exact: true });

export const mealName =(page: Page) => page.getByRole('textbox', { name: 'Meal name' });

/** Search the food picker and tap the matching row (adds one serving, or one more if already in). */
export async function addFood(page: Page, name: string) {
  await foodPicker(page).getByRole('textbox', { name: 'Search' }).fill(name);
  await foodPicker(page).getByRole('button', { name: new RegExp(`^${escapeRe(name)}`) }).click();
  await expect(foodPicker(page).getByRole('textbox', { name: 'Search' })).toHaveValue('');
}

/** The item list of the meal being built. */
export const itemsList = (page: Page) => page.getByRole('region', { name: /^Items/ });

/** Quantity field of one meal item (in grams for gram foods, pieces otherwise). */
// The spinbutton itself has no accessible name; the surrounding group carries it (see notes: follow-ups).
export const qty = (page: Page, food: string) =>
  page.getByRole('group', { name: `${food} quantity` }).getByRole('spinbutton');

/** Type a quantity and commit it (the field commits on blur / Enter). */
export async function setQty(page: Page, food: string, value: number) {
  const field = qty(page, food);
  await field.fill(String(value));
  await field.press('Enter');
}

/** The "This meal" summary in the builder's header (kcal and macros; assert with toContainText). */
export const mealKcal = (page: Page): Locator => page.getByRole('region', { name: 'This meal' });

/** The "Logged · n" list on the /diet day view. */
export const loggedRegion = (page: Page): Locator => page.getByRole('region', { name: /^Logged/ });

/**
 * A toast by its message. Reka also mirrors each toast into a hidden live region whose text starts with
 * "Notifications", so a string matches exactly and a regex must be anchored (`/^Update Failed!/`).
 * Not scoped to the Notifications region: an open modal dialog hides that region from the a11y tree.
 */
export const toast = (page: Page, text: string | RegExp): Locator =>
  page.getByText(text, { exact: typeof text === 'string' });

/** A logged or planned meal row's "More actions" button on the /diet day view. */
export const dayRow = (page: Page, name: string) => page.getByRole('button', { name: `More actions for ${name}` });

function escapeRe(s: string) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
