import type { MealItem } from "~/types/diet";

export type Macros = { calories: number; protein: number; carbs: number; fat: number; fiber: number };

type ItemLike = { amount: number; food?: Partial<Macros> | null };

/** Macros for one meal/saved-meal item (amount is in servings). */
export function itemMacros(item: ItemLike): Macros {
    const f = item.food ?? {};
    const a = Number(item.amount) || 0;
    return {
        calories: a * (f.calories ?? 0),
        protein: a * (f.protein ?? 0),
        carbs: a * (f.carbs ?? 0),
        fat: a * (f.fat ?? 0),
        fiber: a * (f.fiber ?? 0),
    };
}

/** Summed macros for a list of items. */
export function itemsMacros(items: ItemLike[]): Macros {
    const t: Macros = { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 };
    for (const it of items) {
        const m = itemMacros(it);
        t.calories += m.calories;
        t.protein += m.protein;
        t.carbs += m.carbs;
        t.fat += m.fat;
        t.fiber += m.fiber;
    }
    return t;
}

/** Kcal for one meal item (amount is in servings). */
export function itemCalories(item: MealItem): number {
    return item.amount * (item.food?.calories ?? 0);
}

/** Item's share of the meal's kcal, 0..100, rounded. 0 when the meal has no kcal. */
export function shareOfMeal(itemKcal: number, mealKcal: number): number {
    if (mealKcal <= 0 || itemKcal <= 0) return 0;
    return Math.min(100, Math.round((itemKcal / mealKcal) * 100));
}

/**
 * Stepper increment in display units (the unit the quantity field shows).
 * Gram foods step by 10 g (5 g when a serving is under 50 g); counted foods by 1, or ½ when
 * the current amount is fractional.
 */
export function amountStep(item: MealItem): number {
    const food = item.food;
    if (!food) return 1;
    if (food.serving_type === "g" || food.serving_type === "ml") {
        return (food.serving_amount || 100) < 50 ? 5 : 10;
    }
    const qty = (food.serving_amount || 1) * item.amount;
    return Number.isInteger(qty) ? 1 : 0.5;
}

export type DayFit = {
    /** Day total once this meal is counted. */
    after: number;
    target: number;
    /** Positive: room left. Negative: over target. */
    left: number;
    over: boolean;
};

/** How the day lands after this meal, for one macro. Null without a target. */
export function dayFit(after: number, target: number | undefined): DayFit | null {
    if (!target || target <= 0) return null;
    const left = target - after;
    return { after, target, left, over: left < 0 };
}

/** Default meal name for a new log, from the local hour. */
export function defaultMealName(hour: number): string {
    if (hour >= 4 && hour < 11) return "Breakfast";
    if (hour >= 11 && hour < 16) return "Lunch";
    if (hour >= 16 && hour < 22) return "Dinner";
    return "Snack";
}

export function formatKcal(n: number): string {
    return Math.round(n).toLocaleString("en-US");
}
