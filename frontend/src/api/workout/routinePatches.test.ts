import { describe, expect, it } from "vitest";
import type { WorkoutPlan } from "~/types/workout";
import {
    ASSIGN_FIELDS,
    assignAffected,
    assignDay,
    removeExercise,
    reorderExercises,
    restoreFields,
    unassignDay,
    type ProgramsData,
} from "./routinePatches";

const plan = (ID: number, days: number[], ex: number[] = []) =>
    ({ ID, name: `p${ID}`, day_of_week: days[0] ?? null, assigned_days: days, exercises: ex.map((e) => ({ ID: e })) }) as unknown as WorkoutPlan;
const data = (): ProgramsData => ({
    programs: [
        { ID: 1, name: "A", is_active: true, plans: [plan(1, [1]), plan(2, [3, 5])] } as never,
        { ID: 2, name: "B", is_active: false, plans: [plan(3, [3])] } as never,
    ],
});
const days = (d: ProgramsData, id: number) => d.programs.flatMap((g) => g.plans).find((p) => p.ID === id)?.assigned_days;

describe("assignDay", () => {
    it("adds the day and clears it from the program's other routines only", () => {
        const next = assignDay(data(), 1, 3);
        expect(days(next, 1)).toEqual([1, 3]);
        expect(days(next, 2)).toEqual([5]);
        expect(days(next, 3)).toEqual([3]);
    });
    it("does not mutate its input", () => {
        const d = data();
        assignDay(d, 1, 3);
        expect(days(d, 2)).toEqual([3, 5]);
    });
});

describe("unassignDay", () => {
    it("removes the day and nulls the legacy day_of_week when empty", () => {
        const next = unassignDay(data(), 1, 1);
        expect(days(next, 1)).toEqual([]);
        expect(next.programs[0]!.plans[0]!.day_of_week).toBeNull();
    });
});

describe("revert", () => {
    it("restores only the affected routines' assignment, keeping other changes", () => {
        const snap = data();
        let cur = assignDay(snap, 1, 3);
        cur = { programs: cur.programs.map((g) => ({ ...g, plans: g.plans.map((p) => (p.ID === 3 ? { ...p, name: "renamed" } : p)) })) };
        const affected = assignAffected(snap, 1, 3);
        expect(affected.sort()).toEqual([1, 2]);
        const back = restoreFields(cur, snap, affected, ASSIGN_FIELDS);
        expect(days(back, 1)).toEqual([1]);
        expect(days(back, 2)).toEqual([3, 5]);
        expect(back.programs[1]!.plans[0]!.name).toBe("renamed");
    });
});

describe("exercises", () => {
    it("reorders and removes", () => {
        const p = plan(1, [], [10, 11, 12]);
        expect(reorderExercises(p, [12, 10, 11]).exercises.map((e) => e.ID)).toEqual([12, 10, 11]);
        expect(removeExercise(p, 11).exercises.map((e) => e.ID)).toEqual([10, 12]);
    });
    it("restores exercise order from a snapshot", () => {
        const snap = data();
        snap.programs[0]!.plans[0] = plan(1, [1], [10, 11]);
        const cur = { programs: [{ ...snap.programs[0]!, plans: [reorderExercises(snap.programs[0]!.plans[0]!, [11, 10]), snap.programs[0]!.plans[1]!] }, snap.programs[1]!] };
        expect(restoreFields(cur as ProgramsData, snap, [1], ["exercises"]).programs[0]!.plans[0]!.exercises.map((e) => e.ID)).toEqual([10, 11]);
    });
});
