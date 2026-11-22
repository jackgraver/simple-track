import { describe, expect, it } from "vitest";
import type { Exercise, WorkoutPlan, WorkoutProgram } from "~/types/workout";
import {
    addListItem,
    cardioChanged,
    dayToggles,
    findRoutine,
    formatDays,
    moveItem,
    normaliseCardio,
    otherDaysLine,
    parseId,
    routineMeta,
} from "./routine";

const base = { created_at: "", updated_at: "" };
const ex = (ID: number): Exercise => ({ ...base, ID, name: `E${ID}`, rep_rollover: 12, cues: "" });
const plan = (ID: number, name: string, days: number[], exercises: Exercise[] = []): WorkoutPlan => ({
    ...base,
    ID,
    name,
    day_of_week: days[0] ?? null,
    assigned_days: days,
    exercises,
});
const program = (ID: number, name: string, plans: WorkoutPlan[]): WorkoutProgram => ({ ...base, ID, name, is_active: true, plans });

const push = plan(1, "Push", [1, 2], [ex(1), ex(2)]);
const pull = plan(2, "Pull", [3]);
const legs = plan(3, "Legs", [5]);
const ppl = program(2, "Push Pull Legs", [push, pull, legs]);
const other = program(5, "Upper Lower", [plan(9, "Upper", [1])]);

describe("findRoutine", () => {
    it("finds the routine in its program", () => {
        expect(findRoutine([other, ppl], 2, 1)).toEqual({ program: ppl, plan: push });
    });
    it("falls back to any program when the program id is missing or wrong", () => {
        expect(findRoutine([other, ppl], null, 2)?.program).toBe(ppl);
        expect(findRoutine([other, ppl], 5, 2)?.program).toBe(ppl);
    });
    it("returns null for unknown or missing ids", () => {
        expect(findRoutine([ppl], 2, 99)).toBeNull();
        expect(findRoutine([ppl], 2, null)).toBeNull();
    });
});

describe("parseId", () => {
    it("accepts positive integers only", () => {
        expect(parseId("12")).toBe(12);
        expect(parseId("0")).toBeNull();
        expect(parseId("1e3")).toBeNull();
        expect(parseId(["1"])).toBeNull();
        expect(parseId(undefined)).toBeNull();
    });
});

describe("dayToggles", () => {
    const t = dayToggles(push, ppl.plans);
    it("starts on Sunday and marks this routine's days", () => {
        expect(t.map((d) => d.short)).toEqual(["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]);
        expect(t.filter((d) => d.on).map((d) => d.dow)).toEqual([1, 2]);
    });
    it("names the routine that holds another day", () => {
        expect(t.find((d) => d.dow === 3)?.takenBy?.name).toBe("Pull");
        expect(t.find((d) => d.dow === 1)?.takenBy).toBeNull();
    });
    it("formats the summary lines", () => {
        expect(formatDays(t)).toBe("Mon, Tue");
        expect(formatDays(dayToggles(plan(7, "X", []), []))).toBe("No days yet");
        expect(otherDaysLine(t)).toBe("Wed Pull · Fri Legs");
    });
});

describe("routineMeta", () => {
    it("joins program and exercise count", () => {
        expect(routineMeta(ppl, push)).toBe("Push Pull Legs · 2 exercises");
        expect(routineMeta(ppl, pull)).toBe("Push Pull Legs · 0 exercises");
    });
});

describe("moveItem", () => {
    it("moves up and down", () => {
        expect(moveItem([1, 2, 3, 4], 0, 2)).toEqual([2, 3, 1, 4]);
        expect(moveItem([1, 2, 3, 4], 3, 1)).toEqual([1, 4, 2, 3]);
    });
    it("ignores no-op and out-of-range moves", () => {
        const list = [1, 2];
        expect(moveItem(list, 0, 0)).toBe(list);
        expect(moveItem(list, 1, 2)).toBe(list);
        expect(moveItem(list, -1, 0)).toBe(list);
    });
});

describe("addListItem", () => {
    it("trims, collapses spaces and skips duplicates", () => {
        expect(addListItem(["Cat-cow"], "  Arm   circles ")).toEqual(["Cat-cow", "Arm circles"]);
        const list = ["Cat-cow"];
        expect(addListItem(list, "cat-COW")).toBe(list);
        expect(addListItem(list, "   ")).toBe(list);
    });
});

describe("cardio", () => {
    it("normalises and detects changes", () => {
        expect(normaliseCardio(" Bike ", 14.6)).toEqual({ type: "Bike", minutes: 15 });
        expect(normaliseCardio("", null)).toEqual({ type: "", minutes: 0 });
        expect(normaliseCardio("Run", -5)).toEqual({ type: "Run", minutes: 0 });
        const p = { ...pull, planned_cardio_type: "Bike", planned_cardio_minutes: 15 };
        expect(cardioChanged(p, { type: "Bike", minutes: 15 })).toBe(false);
        expect(cardioChanged(p, { type: "Bike", minutes: 20 })).toBe(true);
        expect(cardioChanged(push, { type: "", minutes: 0 })).toBe(false);
    });
});
