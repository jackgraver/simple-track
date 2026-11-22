import { describe, expect, it } from "vitest";
import type { DayLog, Meal, PlannedMeal } from "~/types/diet";
import {
    averageLogged,
    dietWeek,
    formatInt,
    isMeaningfullyOver,
    macroLine,
    mealMacros,
    plannedRows,
    remaining,
} from "./dietDay";

const base = { created_at: "", updated_at: "" };

function meal(id: number, items: [number, number, number, number, number, number][]): Meal {
    return {
        ...base,
        ID: id,
        name: `Meal ${id}`,
        items: items.map(([amount, calories, protein, carbs, fat, fiber], i) => ({
            ...base,
            ID: i,
            meal_id: id,
            food_id: i,
            amount,
            food: { ...base, ID: i, name: "f", serving_type: "g", serving_amount: 1, calories, protein, carbs, fat, fiber },
        })),
    };
}

function planned(id: number, mealId: number, order?: number): PlannedMeal {
    return { ...base, ID: id, day_id: 1, meal_id: mealId, meal: meal(mealId, []), display_order: order };
}

function log(mealId: number): DayLog {
    return { ...base, ID: mealId * 10, day_id: 1, meal_id: mealId, meal: meal(mealId, []) };
}

describe("mealMacros", () => {
    it("multiplies each food by its amount", () => {
        const m = mealMacros(meal(1, [[2, 100, 10, 5, 3, 1], [0.5, 200, 20, 10, 4, 2]]));
        expect(m).toEqual({ calories: 300, protein: 30, carbs: 15, fat: 8, fiber: 3 });
    });
    it("treats missing foods as zero", () => {
        const m = meal(1, [[1, 100, 0, 0, 0, 0]]);
        m.items[0]!.food = undefined;
        expect(mealMacros(m).calories).toBe(0);
    });
});

describe("remaining", () => {
    it("reports what's left under target", () => {
        const r = remaining(774, 2500);
        expect(r).toMatchObject({ left: 1726, over: 0, noTarget: false });
        expect(r.fill).toBeCloseTo(30.96);
    });
    it("reports overage and caps the fill", () => {
        expect(remaining(2700, 2500)).toMatchObject({ left: 0, over: 200, fill: 100 });
    });
    it("flags a missing target", () => {
        expect(remaining(500, 0)).toMatchObject({ noTarget: true, fill: 100 });
        expect(remaining(0, 0)).toMatchObject({ noTarget: true, fill: 0 });
    });
});

describe("isMeaningfullyOver", () => {
    it("tolerates 5% over", () => {
        expect(isMeaningfullyOver(2600, 2500)).toBe(false);
        expect(isMeaningfullyOver(2700, 2500)).toBe(true);
        expect(isMeaningfullyOver(100, 0)).toBe(false);
    });
});

describe("plannedRows", () => {
    it("sorts by display order then id and marks logged by meal id", () => {
        const rows = plannedRows([planned(3, 30, 1), planned(1, 10, 0), planned(2, 20, 0)], [log(20)]);
        expect(rows.map((r) => [r.planned.ID, r.logged])).toEqual([
            [1, false],
            [2, true],
            [3, false],
        ]);
    });
});

describe("format", () => {
    it("formats integers and macro lines", () => {
        expect(formatInt(1726.4)).toBe("1,726");
        expect(macroLine({ calories: 774.4, protein: 68.75, carbs: 70, fat: 21.45, fiber: 1 })).toBe(
            "774 kcal · P 69 · C 70 · F 21",
        );
    });
});

describe("dietWeek", () => {
    // Friday 20 Nov 2026
    const now = new Date(2026, 10, 20, 20, 0);

    it("returns Sunday to Saturday around today", () => {
        const w = dietWeek(0, now);
        expect(w.map((d) => d.key)).toEqual([
            "2026-11-15",
            "2026-11-16",
            "2026-11-17",
            "2026-11-18",
            "2026-11-19",
            "2026-11-20",
            "2026-11-21",
        ]);
        expect(w.map((d) => d.offset)).toEqual([5, 4, 3, 2, 1, 0, -1]);
        expect(w[5]).toMatchObject({ isToday: true, isSelected: true });
        expect(w[6]).toMatchObject({ isFuture: true });
    });

    it("follows the selected day into another week", () => {
        const w = dietWeek(7, now);
        expect(w[0]!.key).toBe("2026-11-08");
        expect(w.find((d) => d.isSelected)?.key).toBe("2026-11-13");
        expect(w.some((d) => d.isToday)).toBe(false);
    });
});

describe("averageLogged", () => {
    it("ignores empty and unknown days", () => {
        expect(averageLogged([774, 0, null, 1226])).toBe(1000);
        expect(averageLogged([0, null])).toBeNull();
    });
});
