import { describe, expect, it } from "vitest";

import type { WorkoutLogsPreviousResponse } from "~/api/workout/api";
import type { Exercise, LoggedExercise } from "~/types/workout";
import {
    addExerciseGroup,
    applyAddedExercise,
    removeExerciseGroup,
    restoreCardio,
    restoreExerciseGroup,
    restoreMobility,
    setCardio,
    setMobilityChecked,
} from "./dayCache";

const ex = (ID: number): Exercise => ({ ID, name: `E${ID}`, rep_rollover: 10, cues: "", created_at: "", updated_at: "" });
const lg = (exercise_id: number): LoggedExercise => ({
    ID: 90 + exercise_id,
    workout_log_id: 3,
    exercise_id,
    exercise: ex(exercise_id),
    sets: [],
    notes: "",
    created_at: "",
    updated_at: "",
});
const base = (over: Partial<WorkoutLogsPreviousResponse> = {}): WorkoutLogsPreviousResponse => ({
    day: { ID: 3 } as WorkoutLogsPreviousResponse["day"],
    planned_exercises: [],
    planned_cardio: null,
    logged_cardio: null,
    planned_pre_mobility: null,
    logged_pre_mobility: null,
    planned_post_mobility: null,
    logged_post_mobility: null,
    ...over,
});

describe("mobility", () => {
    it("ticks the logged row, or starts from the planned routine", () => {
        const withLogged = base({ logged_pre_mobility: { title: "Pre", items: ["a", "b"], checked: ["a"] } });
        expect(setMobilityChecked(withLogged, "pre", ["a", "b"]).logged_pre_mobility?.checked).toEqual(["a", "b"]);
        const planned = base({ planned_post_mobility: { title: "Post", items: ["x"] } });
        const next = setMobilityChecked(planned, "post", ["x"]);
        expect(next.logged_post_mobility).toEqual({ title: "Post", items: ["x"], checked: ["x"] });
        expect(next.logged_pre_mobility).toBeNull();
    });
    it("does nothing when the slot has no routine", () => {
        const d = base();
        expect(setMobilityChecked(d, "pre", ["a"])).toBe(d);
    });
    it("restores only its slot from the snapshot", () => {
        const snap = base({ logged_pre_mobility: { title: "P", items: ["a"], checked: [] } });
        const cur = base({ logged_pre_mobility: { title: "P", items: ["a"], checked: ["a"] }, logged_cardio: { ID: 1 } as never });
        const out = restoreMobility(cur, snap, "pre");
        expect(out.logged_pre_mobility?.checked).toEqual([]);
        expect(out.logged_cardio).toBe(cur.logged_cardio);
    });
});

describe("cardio", () => {
    it("creates a logged cardio with the planned type", () => {
        const d = base({ planned_cardio: { type: "Run", minutes: 20 } });
        const c = setCardio(d, { date: "d", minutes: 25, notes: "n" }).logged_cardio!;
        expect(c).toMatchObject({ minutes: 25, type: "Run", notes: "n", workout_log_id: 3 });
    });
    it("keeps existing type/notes when omitted and restores from the snapshot", () => {
        const prev = { ID: 5, workout_log_id: 3, minutes: 10, type: "Bike", notes: "old", created_at: "", updated_at: "" };
        const d = base({ logged_cardio: prev });
        const next = setCardio(d, { date: "d", minutes: 30 });
        expect(next.logged_cardio).toMatchObject({ ID: 5, minutes: 30, type: "Bike", notes: "old" });
        expect(restoreCardio(next, d).logged_cardio).toBe(prev);
    });
});

describe("add / remove exercise", () => {
    it("appends an ad-hoc group, attaches to a planned one, ignores a logged one", () => {
        expect(addExerciseGroup(base(), { date: "d", exerciseId: 1 })).toEqual(base());
        const added = addExerciseGroup(base(), { date: "d", exerciseId: 1, exercise: ex(1) });
        expect(added.planned_exercises[0].logged).toMatchObject({ exercise_id: 1, ID: 0, sets: [] });
        const planned = base({ planned_exercises: [{ planned: ex(2) }] });
        const p2 = addExerciseGroup(planned, { date: "d", exerciseId: 2, exercise: ex(2) });
        expect(p2.planned_exercises).toHaveLength(1);
        expect(p2.planned_exercises[0].logged?.exercise_id).toBe(2);
        const logged = base({ planned_exercises: [{ logged: lg(3) }] });
        expect(addExerciseGroup(logged, { date: "d", exerciseId: 3, exercise: ex(3) })).toBe(logged);
    });
    it("swaps the temporary row for the saved one", () => {
        const d = addExerciseGroup(base(), { date: "d", exerciseId: 1, exercise: ex(1) });
        expect(applyAddedExercise(d, 1, lg(1)).planned_exercises[0].logged?.ID).toBe(91);
    });
    it("removes ad-hoc groups, keeps planned ones, ignores unlogged ones", () => {
        const d = base({ planned_exercises: [{ logged: lg(1) }, { planned: ex(2), logged: lg(2) }, { planned: ex(3) }] });
        const out = removeExerciseGroup(removeExerciseGroup(d, 1), 2);
        expect(out.planned_exercises).toEqual([{ planned: ex(2) }, { planned: ex(3) }]);
        expect(removeExerciseGroup(out, 3)).toBe(out);
    });
    it("restores the group in place or just its logged row", () => {
        const d = base({ planned_exercises: [{ logged: lg(1) }, { planned: ex(2), logged: lg(2) }, { logged: lg(4) }] });
        const removed = removeExerciseGroup(removeExerciseGroup(d, 1), 2);
        const back = restoreExerciseGroup(restoreExerciseGroup(removed, d, 1), d, 2);
        expect(back.planned_exercises).toEqual(d.planned_exercises);
    });
});
