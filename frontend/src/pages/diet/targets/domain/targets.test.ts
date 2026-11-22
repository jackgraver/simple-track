import { describe, expect, it } from "vitest";
import {
    calorieGapLabel,
    fmtTrend,
    gramsPerLb,
    intakeSummary,
    planHistory,
    observedMaintenance,
    sortPlansByEffective,
    splitSegments,
    targetsEqual,
    weightTrendLbPerWeek,
} from "./targets";

const now = new Date("2026-11-20T12:00:00Z");
const day = (n: number) => new Date(now.getTime() - n * 86400000).toISOString();

describe("weightTrendLbPerWeek", () => {
    it("fits a slope in lb/week", () => {
        // −0.1 lb a day = −0.7 lb/wk
        const logs = Array.from({ length: 15 }, (_, i) => ({ date: day(i), weight_lbs: 180 + i * 0.1 }));
        expect(weightTrendLbPerWeek(logs, now)).toBeCloseTo(-0.7, 5);
    });
    it("needs 3 logs spanning a week", () => {
        expect(weightTrendLbPerWeek(undefined, now)).toBeNull();
        expect(weightTrendLbPerWeek([{ date: day(0), weight_lbs: 180 }, { date: day(10), weight_lbs: 181 }], now)).toBeNull();
        const short = [0, 1, 2, 3].map((i) => ({ date: day(i), weight_lbs: 180 }));
        expect(weightTrendLbPerWeek(short, now)).toBeNull();
    });
    it("ignores logs outside the window and zero weights", () => {
        const logs = [
            { date: day(0), weight_lbs: 179 },
            { date: day(7), weight_lbs: 180 },
            { date: day(14), weight_lbs: 181 },
            { date: day(14), weight_lbs: 0 },
            { date: day(60), weight_lbs: 150 },
        ];
        expect(weightTrendLbPerWeek(logs, now)).toBeCloseTo(-1, 5);
    });
});

describe("intakeSummary / observedMaintenance", () => {
    const days = [2400, 2500, 0, 2300, 2400, 0, 2600].map((c) => ({ totalCalories: c, totalProtein: c / 15 }));
    it("averages only logged days", () => {
        const s = intakeSummary(days);
        expect(s).toMatchObject({ loggedDays: 5, totalDays: 7, avgCalories: 2440 });
    });
    it("adds back the deficit implied by weight loss", () => {
        // 2440 eaten, losing 0.5 lb/wk = 250 kcal/day deficit
        expect(observedMaintenance(intakeSummary(days), -0.5)).toBe(2690);
    });
    it("needs enough logged days and a trend", () => {
        expect(observedMaintenance(intakeSummary(days.slice(0, 3)), -0.5)).toBeNull();
        expect(observedMaintenance(intakeSummary(days), null)).toBeNull();
    });
});

describe("splitSegments", () => {
    it("gives kcal shares", () => {
        const s = splitSegments({ protein: 180, carbs: 280, fat: 70 });
        expect(s.map((x) => x.kcal)).toEqual([720, 1120, 630]);
        expect(s.reduce((a, x) => a + x.pct, 0)).toBeCloseTo(100);
        expect(Math.round(s[0]!.pct)).toBe(29);
    });
    it("is all zero without macros", () => {
        expect(splitSegments({ protein: 0, carbs: 0, fat: 0 }).every((x) => x.pct === 0)).toBe(true);
    });
});

describe("formatters and helpers", () => {
    it("gramsPerLb", () => {
        expect(gramsPerLb(180, 179)).toBe(1.01);
        expect(gramsPerLb(180, 0)).toBeNull();
    });
    it("calorieGapLabel", () => {
        expect(calorieGapLabel(0)).toBe("matches target");
        expect(calorieGapLabel(-30)).toBe("30 under");
        expect(calorieGapLabel(1200)).toBe("1,200 over");
    });
    it("fmtTrend", () => {
        expect(fmtTrend(-0.754)).toBe("↓0.75 lb/wk");
        expect(fmtTrend(0.2)).toBe("↑0.2 lb/wk");
        expect(fmtTrend(0.01)).toBe("steady");
    });
    it("targetsEqual rounds", () => {
        const a = { calories: 2500, protein: 180, carbs: 280, fat: 70, fiber: 35 };
        expect(targetsEqual(a, { ...a, protein: 180.2 })).toBe(true);
        expect(targetsEqual(a, { ...a, fat: 71 })).toBe(false);
    });
    it("sortPlansByEffective", () => {
        const plans = [
            { ID: 1, effective_from: "2026-10-20" },
            { ID: 2 },
            { ID: 3, effective_from: "2026-11-01" },
        ];
        expect(sortPlansByEffective(plans).map((p) => p.ID)).toEqual([3, 1, 2]);
    });
    it("planHistory keeps the last save of each day", () => {
        const plans = [
            { ID: 1, effective_from: "2026-10-20T00:00:00" },
            { ID: 2, effective_from: "2026-11-20T00:00:00" },
            { ID: 3, effective_from: "2026-11-20T00:00:00" },
        ];
        expect(planHistory(plans).map((p) => p.ID)).toEqual([3, 1]);
    });
});
