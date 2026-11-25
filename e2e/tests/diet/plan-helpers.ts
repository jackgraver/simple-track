import type { Page } from '@playwright/test';
import type { Api } from '../../fixtures/api';

// Shared helpers for the plan lane (/diet/edit-planned). Not a spec. Epoch: Thu 2026-01-15, so the
// default week is Sun 11 .. Sat 17 (Sunday-first). Offsets are days in the past (negative = future).

export const FOOD = { chicken: 1, rice: 2, egg: 3, oats: 4, banana: 5 } as const;

/** Saved meals with easy-to-read calories: Oats bowl 150, Banana 105, Chicken and rice 460. */
export async function createSavedMeals(api: Api) {
  const oats = await api.createSavedMeal('Oats bowl', [{ food_id: FOOD.oats, amount: 1 }]);
  const banana = await api.createSavedMeal('Banana', [{ food_id: FOOD.banana, amount: 1 }]);
  const chicken = await api.createSavedMeal('Chicken and rice', [
    { food_id: FOOD.chicken, amount: 2 },
    { food_id: FOOD.rice, amount: 1 },
  ]);
  return { oats, banana, chicken };
}

export const OFFSET = { sun: 4, mon: 3, tue: 2, wed: 1, thu: 0, fri: -1, sat: -2, nextMon: -4 } as const;

/** One day's card on the plan page, by its short label ("Fri 16"). */
export const dayCard = (page: Page, label: string) => page.getByRole('region', { name: label });
/** A planned meal chip on a day's card. */
export const plannedChip = (page: Page, label: string, meal: string) =>
  dayCard(page, label).getByRole('button', { name: new RegExp(`^${meal}, [\\d,]+ kcal\\. Options$`) });
