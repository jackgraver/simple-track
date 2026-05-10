import type { Meal } from "~/types/diet";

export function mealHasQuickEntryFood(meal: Meal): boolean {
    return meal.items.some((i) => i.food?.quick_entry === true);
}
