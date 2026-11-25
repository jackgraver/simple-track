import type { Page } from '@playwright/test';
import type { Api } from '../../fixtures/api';

// Shared helpers for the /diet day view specs (not a spec: Playwright only runs *.spec.ts).
// Seed foods (IDs in this order): Chicken Breast 1, White Rice 2, Egg 3, Oats 4, Banana 5, Greek Yogurt 6,
// Broccoli 7, Olive Oil 8. `amount` in a saved meal is a multiple of the food's serving.
export const FOOD = { chicken: 1, rice: 2, egg: 3, oats: 4, banana: 5 } as const;

export const LOG_PLANNED = 'POST /diet/meals/meal/log-planned';
export const DELETE_LOGGED = 'DELETE /diet/meals/meal/logged';

/** Chicken 1 + Rice 1 = 295 kcal, P 33.7, C 28, F 3.9. Shown as "295 kcal · P 34 · C 28 · F 4". */
export const BOWL = {
  name: 'Chicken bowl',
  items: [
    { food_id: FOOD.chicken, amount: 1 },
    { food_id: FOOD.rice, amount: 1 },
  ],
  kcal: 295,
} as const;

/** Egg 2 + Banana 1 = 249 kcal, P 13.9, C 27.8, F 10. */
export const BREAKFAST = {
  name: 'Egg breakfast',
  items: [
    { food_id: FOOD.egg, amount: 2 },
    { food_id: FOOD.banana, amount: 1 },
  ],
  kcal: 249,
} as const;

export interface PlannedArrangement {
  savedMealId: number;
  /** The meal row the planned entry points at (what log-planned takes). */
  mealId: number;
  plannedId: number;
  dayId: number;
}

/** Create a saved meal and plan it on `offset`. */
export async function planMeal(
  api: Api,
  meal: { name: string; items: readonly { food_id: number; amount: number }[] },
  offset = 0,
): Promise<PlannedArrangement> {
  const savedMealId = await api.createSavedMeal(meal.name, [...meal.items]);
  const day = await api.addPlannedFromSaved(savedMealId, offset);
  const planned = day.day.plannedMeals.find((p) => p.meal.name === meal.name);
  if (!planned) throw new Error(`planned meal ${meal.name} not on the day`);
  return { savedMealId, mealId: planned.meal_id, plannedId: planned.ID, dayId: day.day.ID };
}

export const loggedRow = (page: Page, name: string) => page.getByRole('button', { name: `More actions for ${name}` });
export const logButton = (page: Page, name: string) => page.getByRole('button', { name: `Log ${name}`, exact: true });
export const plannedHeading = (page: Page) => page.getByRole('heading', { name: /^Planned/ });
export const loggedHeading = (page: Page) => page.getByRole('heading', { name: /^Logged/ });

/** "2,205 kcal left" lives in one paragraph: the number and the unit are separate spans. */
export const kcalLeft = (page: Page) => page.locator('p').filter({ hasText: /kcal (left|over|eaten)/ }).first();
/** A macro row of the summary card, e.g. macro(page, 'Protein') reads "Protein · 146 g left". */
export const macro = (page: Page, label: 'Protein' | 'Carbs' | 'Fat') =>
  page.getByText(new RegExp(`^${label}\\s*·`));
