import { describe, expect, it } from "vitest";

import type { DayLog, Food, Meal, PlannedMeal, SavedMeal } from "~/types/diet";
import type { DietLogsTodayResponse } from "./api";
import {
    applyAddPlanned,
    applyDeletePlanned,
    applyLogPlanned,
    applyQuickLog,
    applyReorder,
    applyUnlog,
    isTempId,
    revertAddPlanned,
    revertDeletePlanned,
    revertLogPlanned,
    revertQuickLog,
    revertReorder,
    revertUnlog,
    resolveLoggedMealId,
} from "./dayPatches";

const base = { created_at: "", updated_at: "" };
const food = (ID: number, calories: number, protein = 0): Food => ({
    ...base, ID, name: "f" + ID, serving_type: "g", serving_amount: 1, calories, protein, fiber: 0, carbs: 0, fat: 0,
});
const meal = (ID: number, f: Food, amount = 1): Meal => ({
    ...base, ID, name: "m" + ID, items: [{ ...base, ID: ID * 10, meal_id: ID, food_id: f.ID, food: f, amount }],
});
const planned = (ID: number, m: Meal, order: number): PlannedMeal => ({
    ...base, ID, day_id: 1, meal_id: m.ID, meal: m, display_order: order,
});
const logged = (ID: number, m: Meal): DayLog => ({ ...base, ID, day_id: 1, meal_id: m.ID, meal: m });

function day(): DietLogsTodayResponse {
    const m1 = meal(11, food(1, 200, 10));
    const m2 = meal(12, food(2, 300, 20), 2); // 600 kcal, 40 g
    const l = logged(5, m1);
    return {
        day: {
            ...base, ID: 1, date: "2026-11-22",
            plan: { ...base, ID: 1, name: "p", calories: 2000, protein: 150, fiber: 30, carbs: 200, fat: 60 },
            plannedMeals: [planned(101, m1, 1), planned(102, m2, 2)],
            loggedMeals: [l],
        },
        totalCalories: 200, totalProtein: 10, totalFiber: 0, totalCarbs: 0, totalFat: 0,
    };
}

describe("log planned", () => {
    it("adds a logged row and the meal's macros, and is idempotent", () => {
        const next = applyLogPlanned(day(), 12);
        expect(next.day.loggedMeals.map((l) => l.meal_id)).toEqual([11, 12]);
        expect(isTempId(next.day.loggedMeals[1].ID)).toBe(true);
        expect(next.totalCalories).toBe(800);
        expect(next.totalProtein).toBe(50);
        expect(applyLogPlanned(next, 12)).toBe(next);
    });
    it("ignores a meal that is not planned", () => {
        const d = day();
        expect(applyLogPlanned(d, 999)).toBe(d);
    });
    it("moves the meal out of the planned list at once", () => {
        const next = applyLogPlanned(day(), 12);
        expect(next.day.plannedMeals.map((p) => p.meal_id)).toEqual([11]);
    });
    it("revert removes only the temp row and its macros, and re-plans the meal", () => {
        const next = applyLogPlanned(day(), 12);
        expect(revertLogPlanned(next, 12, day())).toEqual(day());
        // Without a snapshot (or already re-planned) the planned list is left alone.
        expect(revertLogPlanned(next, 12).day.plannedMeals.map((p) => p.meal_id)).toEqual([11]);
        const back = revertLogPlanned(next, 12, day());
        expect(revertLogPlanned(back, 12, day())).toEqual(back);
    });
});

describe("unlog", () => {
    it("removes the row and subtracts, never below zero", () => {
        const d = day();
        const next = applyUnlog(d, 11);
        expect(next.day.loggedMeals).toEqual([]);
        expect(next.totalCalories).toBe(0);
        expect(applyUnlog({ ...d, totalCalories: 50 }, 11).totalCalories).toBe(0);
        expect(applyUnlog(next, 11)).toBe(next);
    });
    it("revert reinserts at the original index on top of another patch", () => {
        const d = applyLogPlanned(day(), 12); // sibling: logged 12 after 11
        const afterUnlog = applyUnlog(d, 11);
        const back = revertUnlog(afterUnlog, 11, d);
        expect(back.day.loggedMeals.map((l) => l.meal_id).sort()).toEqual([11, 12]);
        expect(back.totalCalories).toBe(800);
        expect(revertUnlog(back, 11, d)).toBe(back);
    });
});

describe("quick log", () => {
    const entry = { client_id: "c1", name: "Burger", calories: 700, protein: 30, carbs: 50, fat: 40, fiber: 2 };
    it("appends a quick-entry meal with exact totals, idempotent by client_id", () => {
        const next = applyQuickLog(day(), entry);
        const row = next.day.loggedMeals[1];
        expect(row.meal.client_id).toBe("c1");
        expect(row.meal.items[0].food?.quick_entry).toBe(true);
        expect(next.totalCalories).toBe(900);
        expect(next.totalFiber).toBe(2);
        expect(applyQuickLog(next, entry)).toBe(next);
    });
    it("revert removes by client_id", () => {
        expect(revertQuickLog(applyQuickLog(day(), entry), entry, day())).toEqual(day());
    });
    it("replace swaps the edited row in place and revert restores it", () => {
        const e = { ...entry, replace_meal_id: 11 };
        const d = day();
        const next = applyQuickLog(d, e);
        expect(next.day.loggedMeals).toHaveLength(1);
        expect(next.day.loggedMeals[0].meal.client_id).toBe("c1");
        expect(next.totalCalories).toBe(700);
        expect(revertQuickLog(next, e, d)).toEqual(d);
    });
});

describe("planned meals", () => {
    const saved: SavedMeal = {
        ...base, ID: 7, name: "Oats",
        items: [{ ...base, ID: 1, saved_meal_id: 7, food_id: 3, food: food(3, 100), amount: 2 }],
    };
    it("add appends with the next display order and client_id; revert removes it", () => {
        const next = applyAddPlanned(day(), saved, "c9");
        const row = next.day.plannedMeals[2];
        expect(row.client_id).toBe("c9");
        expect(row.display_order).toBe(3);
        expect(row.meal.items[0].amount).toBe(2);
        expect(next.totalCalories).toBe(200); // planned does not touch logged totals
        expect(applyAddPlanned(next, saved, "c9")).toBe(next);
        expect(revertAddPlanned(next, "c9")).toEqual(day());
    });
    it("delete removes by id; revert puts the snapshot row back", () => {
        const d = day();
        const next = applyDeletePlanned(d, 101);
        expect(next.day.plannedMeals.map((p) => p.ID)).toEqual([102]);
        const back = revertDeletePlanned(next, 101, d);
        expect(back.day.plannedMeals.map((p) => p.ID).sort()).toEqual([101, 102]);
        expect(revertDeletePlanned(back, 101, d)).toBe(back);
    });
    it("reorder rewrites display_order and revert restores it", () => {
        const d = day();
        const next = applyReorder(d, [102, 101]);
        const order = (r: DietLogsTodayResponse) =>
            [...r.day.plannedMeals].sort((a, b) => a.display_order! - b.display_order!).map((p) => p.ID);
        expect(order(next)).toEqual([102, 101]);
        expect(order(revertReorder(next, d))).toEqual([101, 102]);
    });
    it("reorder keeps rows it does not name after the named ones", () => {
        const next = applyReorder(day(), [102]);
        const sorted = [...next.day.plannedMeals].sort((a, b) => a.display_order! - b.display_order!);
        expect(sorted.map((p) => p.ID)).toEqual([102, 101]);
    });
});

describe("resolveLoggedMealId", () => {
    it("keeps a real id, maps a temp quick log to the saved meal by client_id, null when never saved", () => {
        expect(resolveLoggedMealId(11, undefined, [])).toBe(11);
        const saved = day();
        const quick = { ...meal(77, food(3, 100)), client_id: "q1" };
        saved.day.loggedMeals = [...(saved.day.loggedMeals ?? []), logged(9, quick)];
        expect(resolveLoggedMealId(-3, "q1", [undefined, saved])).toBe(77);
        expect(resolveLoggedMealId(-3, "q2", [saved])).toBeNull();
        expect(resolveLoggedMealId(-3, null, [saved])).toBeNull();
    });
});
