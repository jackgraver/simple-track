import { describe, expect, it } from "vitest";
import type { DietLogsTodayResponse } from "~/api/diet/api";
import { dietWeek } from "~/pages/diet/home/dietDay";
import {
    buildPlanDay,
    formatWeekSummary,
    moveId,
    planStatus,
    savedMealMacros,
    statusText,
    weekRangeLabel,
    weekSummary,
} from "./planWeek";

const food = (calories: number, protein: number) => ({ calories, protein, carbs: 0, fat: 0, fiber: 0 });

function response(meals: { id: number; order: number; kcal: number; protein: number }[], target = 2500, logged = 0) {
    return {
        day: {
            plan: { calories: target, protein: 180 },
            plannedMeals: meals.map((m) => ({
                ID: m.id,
                display_order: m.order,
                meal: { name: `M${m.id}`, items: [{ amount: 1, food: food(m.kcal, m.protein) }] },
            })),
            loggedMeals: [],
        },
        totalCalories: logged,
    } as unknown as DietLogsTodayResponse;
}

// Friday 20 Nov 2026, noon.
const NOW = new Date(2026, 10, 20, 12);
const week = dietWeek(0, NOW);

describe("planStatus", () => {
    it("bands planned kcal against the target", () => {
        expect(planStatus(0, 2500, 0)).toBe("empty");
        expect(planStatus(2000, 2500, 2)).toBe("under");
        expect(planStatus(2250, 2500, 2)).toBe("on");
        expect(planStatus(2625, 2500, 2)).toBe("on");
        expect(planStatus(2700, 2500, 2)).toBe("over");
        expect(planStatus(500, 0, 1)).toBe("none");
    });
});

describe("buildPlanDay", () => {
    it("sorts meals by display order and sums them", () => {
        const fri = week[5]!;
        const d = buildPlanDay(fri, response([
            { id: 2, order: 1, kcal: 689, protein: 72 },
            { id: 1, order: 0, kcal: 321, protein: 20 },
        ]));
        expect(d.meals.map((m) => m.id)).toEqual([1, 2]);
        expect(d.planned.calories).toBe(1010);
        expect(d.planned.protein).toBe(92);
        expect(d.status).toBe("under");
        expect(d.fill).toBeCloseTo(40.4);
        expect(d.isPast).toBe(false);
        expect(statusText(d)).toBe("1,490 kcal to plan");
    });

    it("treats missing data as an empty day", () => {
        const d = buildPlanDay(week[0]!, undefined);
        expect(d.status).toBe("empty");
        expect(d.isPast).toBe(true);
        expect(d.fill).toBe(0);
    });
});

describe("weekSummary", () => {
    it("only counts today and later", () => {
        const days = week.map((w, i) =>
            buildPlanDay(w, i === 6 ? response([{ id: 1, order: 0, kcal: 2400, protein: 150 }]) : response([])),
        );
        const s = weekSummary(days);
        expect(s).toEqual({ unplanned: 1, under: 0, onTarget: 1, upcoming: 2 });
        expect(formatWeekSummary(s)).toEqual({ text: "1 day with nothing planned", tone: "warn" });
    });

    it("is null for a past week and success when all on target", () => {
        expect(formatWeekSummary({ unplanned: 0, under: 0, onTarget: 0, upcoming: 0 })).toBeNull();
        expect(formatWeekSummary({ unplanned: 0, under: 0, onTarget: 3, upcoming: 3 })?.tone).toBe("success");
    });
});

describe("helpers", () => {
    it("moves ids within bounds", () => {
        expect(moveId([1, 2, 3], 0, 1)).toEqual([2, 1, 3]);
        expect(moveId([1, 2, 3], 0, -1)).toBeNull();
        expect(moveId([1, 2, 3], 2, 1)).toBeNull();
    });

    it("labels week ranges across months", () => {
        expect(weekRangeLabel(new Date(2026, 10, 15), new Date(2026, 10, 21))).toMatch(/15 – 21$/);
        expect(weekRangeLabel(new Date(2026, 10, 29), new Date(2026, 11, 5))).toMatch(/29 – .+ 5$/);
    });

    it("sums saved meal items", () => {
        const m = savedMealMacros({ items: [{ amount: 2, food: food(100, 10) }] } as never);
        expect(m.calories).toBe(200);
        expect(m.protein).toBe(20);
    });
});
