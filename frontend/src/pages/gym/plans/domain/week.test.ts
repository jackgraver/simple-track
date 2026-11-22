import { describe, expect, it } from "vitest";
import type { WorkoutPlan, WorkoutProgram } from "~/types/workout";
import {
    assignedDays,
    buildWeek,
    formatWeekSummary,
    pickDefaultProgram,
    planMeta,
    weekSummary,
} from "./week";

const ex = (n: number) => Array.from({ length: n }, (_, i) => ({ ID: i + 1, name: `E${i}`, rep_rollover: 12, cues: "" }));

function plan(id: number, name: string, days: number[] | null, exercises = 4, extra: Partial<WorkoutPlan> = {}): WorkoutPlan {
    return {
        ID: id,
        name,
        day_of_week: days?.[0] ?? null,
        assigned_days: days ?? undefined,
        exercises: ex(exercises),
        ...extra,
    } as WorkoutPlan;
}

const ppl = [plan(1, "Push", [1]), plan(2, "Pull", [3], 4, { planned_cardio_type: "Bike", planned_cardio_minutes: 15 }), plan(3, "Legs", [5])];

describe("assignedDays", () => {
    it("prefers assigned_days and falls back to day_of_week", () => {
        expect(assignedDays(plan(1, "A", [2, 4]))).toEqual([2, 4]);
        expect(assignedDays({ ...plan(1, "A", null), day_of_week: 6 })).toEqual([6]);
        expect(assignedDays(plan(1, "A", null))).toEqual([]);
    });
});

describe("buildWeek", () => {
    it("lays routines on a Sunday-first week and marks today", () => {
        const week = buildWeek(ppl, 4);
        expect(week.map((s) => s.plan?.name ?? "-")).toEqual(["-", "Push", "-", "Pull", "-", "Legs", "-"]);
        expect(week.filter((s) => s.isToday).map((s) => s.short)).toEqual(["Thu"]);
    });

    it("puts a routine on every day it is assigned", () => {
        const week = buildWeek([plan(1, "Full body", [1, 3, 5])], 0);
        expect(week.filter((s) => s.plan).map((s) => s.dow)).toEqual([1, 3, 5]);
    });
});

describe("weekSummary", () => {
    it("counts sessions, exercises per session and rest days", () => {
        const s = weekSummary(buildWeek(ppl, 0));
        expect(s).toEqual({ sessions: 3, exercises: 12, restDays: 4 });
        expect(formatWeekSummary(s)).toBe("3 sessions · 12 exercises · 4 rest days");
    });

    it("says so when nothing is scheduled", () => {
        expect(formatWeekSummary(weekSummary(buildWeek([], 0)))).toBe("No training days scheduled yet");
    });

    it("uses singulars", () => {
        expect(formatWeekSummary({ sessions: 1, exercises: 1, restDays: 6 })).toBe("1 session · 1 exercise · 6 rest days");
    });
});

describe("planMeta", () => {
    it("describes exercises and cardio", () => {
        expect(planMeta(ppl[1]!)).toEqual(["4 exercises", "Bike 15 min"]);
        expect(planMeta(plan(9, "X", [], 1))).toEqual(["1 exercise"]);
    });
});

describe("pickDefaultProgram", () => {
    const prog = (ID: number, is_active: boolean, plans: WorkoutPlan[]) => ({ ID, name: `P${ID}`, is_active, plans }) as WorkoutProgram;

    it("prefers an active program that has routines", () => {
        expect(pickDefaultProgram([prog(1, true, []), prog(2, true, ppl), prog(3, false, [])])?.ID).toBe(2);
    });

    it("falls back to any active, then the first, then null", () => {
        expect(pickDefaultProgram([prog(1, false, ppl), prog(2, true, [])])?.ID).toBe(2);
        expect(pickDefaultProgram([prog(1, false, []), prog(2, false, [])])?.ID).toBe(1);
        expect(pickDefaultProgram([])).toBeNull();
    });
});
