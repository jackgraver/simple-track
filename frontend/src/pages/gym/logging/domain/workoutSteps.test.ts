import { describe, expect, it } from "vitest";
import type { ExerciseGroup } from "~/api/workout/api";
import { buildWorkoutSteps, formatSets, nextStepIndex } from "./workoutSteps";

const set = (weight: number, reps: number) => ({ weight, reps });

describe("formatSets", () => {
    it("groups consecutive sets at the same weight", () => {
        expect(formatSets([set(135, 8), set(135, 8), set(140, 6)])).toBe("135 lb × 8, 8 · 140 lb × 6");
    });
    it("shows bodyweight for weight 0 and nothing for no sets", () => {
        expect(formatSets([set(0, 12)])).toBe("BW × 12");
        expect(formatSets([])).toBe("");
        expect(formatSets(undefined)).toBe("");
    });
    it("reads unilateral sets per pair, and spells out uneven weights", () => {
        const side = (weight: number, reps: number, s: "left" | "right") => ({ weight, reps, side: s });
        expect(formatSets([side(30, 10, "left"), side(30, 9, "right"), side(30, 10, "left"), side(30, 10, "right")])).toBe(
            "30 lb × L10 R9, L10 R10",
        );
        expect(formatSets([side(30, 10, "left"), side(30, 9, "right"), side(30, 8, "left"), side(25, 10, "right")])).toBe(
            "30 lb × L10 R9 · L 30 lb × 8 R 25 lb × 10",
        );
    });
});

describe("buildWorkoutSteps", () => {
    const lift = (id: number, name: string, logged?: number[][], previous?: number[][]): ExerciseGroup =>
        ({
            planned: { ID: id, name },
            logged: logged ? { exercise_id: id, sets: logged.map(([w, r]) => set(w!, r!)) } : undefined,
            previous: previous ? { exercise_id: id, sets: previous.map(([w, r]) => set(w!, r!)) } : undefined,
        }) as unknown as ExerciseGroup;

    it("orders warmup, lifts, cardio, cooldown and finds the next step", () => {
        const steps = buildWorkoutSteps({
            log: [lift(1, "Row", [[135, 8]]), lift(2, "Curl", undefined, [[30, 10], [30, 9]])],
            plannedCardio: { type: "Bike", minutes: 15 },
            loggedPreMobility: { title: "Warmup", items: ["a", "b"], checked: ["a", "b"] },
            plannedPostMobility: { title: "Stretch", items: ["c"] },
        });
        expect(steps.map((s) => s.kind)).toEqual(["warmup", "lift", "lift", "cardio", "cooldown"]);
        expect(steps.map((s) => s.done)).toEqual([true, true, false, false, false]);
        expect(steps[1]!.detail).toBe("135 lb × 8");
        expect(steps[2]!.detail).toBe("Last 30 lb × 10, 9");
        expect(steps[3]!.detail).toBe("15 min planned");
        expect(nextStepIndex(steps)).toBe(2);
    });

    it("reports partial mobility progress and skips empty routines", () => {
        const steps = buildWorkoutSteps({
            log: [],
            loggedPreMobility: { title: "", items: ["a", "b", "c"], checked: ["a"] },
            loggedPostMobility: { title: "", items: [], checked: [] },
        });
        expect(steps).toHaveLength(1);
        expect(steps[0]!.name).toBe("Dynamic warmup");
        expect(steps[0]!.detail).toBe("1 of 3 done");
        expect(nextStepIndex([{ ...steps[0]!, done: true }])).toBe(-1);
    });
});
