import { describe, expect, it } from "vitest";
import type { Food, MealItem } from "~/types/diet";
import { amountStep, dayFit, defaultMealName, formatKcal, itemCalories, itemsMacros, shareOfMeal } from "./mealCard";

function item(food: Partial<Food>, amount: number): MealItem {
    return {
        ID: 0,
        meal_id: 0,
        food_id: 1,
        amount,
        food: {
            ID: 1,
            name: "x",
            serving_type: "g",
            serving_amount: 100,
            calories: 0,
            protein: 0,
            fiber: 0,
            carbs: 0,
            fat: 0,
            ...food,
        } as Food,
    } as MealItem;
}

describe("itemCalories / shareOfMeal", () => {
    it("scales by servings", () => {
        expect(itemCalories(item({ calories: 165 }, 2))).toBe(330);
    });
    it("is 0 for no food", () => {
        expect(itemCalories({ amount: 3 } as MealItem)).toBe(0);
    });
    it("rounds the share and guards empty meals", () => {
        expect(shareOfMeal(330, 689)).toBe(48);
        expect(shareOfMeal(34, 689)).toBe(5);
        expect(shareOfMeal(10, 0)).toBe(0);
    });
});

describe("amountStep", () => {
    it("steps grams by 10, small servings by 5", () => {
        expect(amountStep(item({ serving_type: "g", serving_amount: 100 }, 2))).toBe(10);
        expect(amountStep(item({ serving_type: "g", serving_amount: 30 }, 1))).toBe(5);
    });
    it("steps counted foods by 1, or ½ when fractional", () => {
        expect(amountStep(item({ serving_type: "each", serving_amount: 1 }, 3))).toBe(1);
        expect(amountStep(item({ serving_type: "each", serving_amount: 1 }, 1.5))).toBe(0.5);
    });
});

describe("dayFit", () => {
    it("reports room left", () => {
        expect(dayFit(1929, 2500)).toEqual({ after: 1929, target: 2500, left: 571, over: false });
    });
    it("reports over target", () => {
        expect(dayFit(2600, 2500)?.over).toBe(true);
        expect(dayFit(2600, 2500)?.left).toBe(-100);
    });
    it("is null without a target", () => {
        expect(dayFit(100, 0)).toBeNull();
        expect(dayFit(100, undefined)).toBeNull();
    });
});

describe("defaultMealName", () => {
    it("names by hour", () => {
        expect(defaultMealName(8)).toBe("Breakfast");
        expect(defaultMealName(12)).toBe("Lunch");
        expect(defaultMealName(19)).toBe("Dinner");
        expect(defaultMealName(23)).toBe("Snack");
        expect(defaultMealName(2)).toBe("Snack");
    });
});

describe("formatKcal", () => {
    it("rounds and groups", () => {
        expect(formatKcal(1929.4)).toBe("1,929");
    });
});

describe("itemsMacros", () => {
    it("sums items scaled by amount", () => {
        const t = itemsMacros([
            item({ calories: 165, protein: 31, fat: 3.6 }, 2),
            item({ calories: 34, protein: 2.8, carbs: 7, fiber: 2.6 }, 1),
        ]);
        expect(t.calories).toBe(364);
        expect(t.protein).toBeCloseTo(64.8);
        expect(t.carbs).toBe(7);
        expect(t.fat).toBeCloseTo(7.2);
        expect(t.fiber).toBeCloseTo(2.6);
    });
    it("is all zeros for no items", () => {
        expect(itemsMacros([])).toEqual({ calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 });
    });
});
