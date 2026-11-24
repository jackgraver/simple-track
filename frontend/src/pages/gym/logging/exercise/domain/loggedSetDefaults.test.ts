import { describe, expect, it } from "vitest";
import type { LoggedSet } from "~/types/workout";
import {
    initializeWeightAndReps,
    previousRepsForSetAtWeight,
} from "./loggedSetDefaults";

function loggedSet(partial: {
    weight: number;
    reps: number;
    weight_setup?: string;
    id?: number;
}): LoggedSet {
    const id = partial.id ?? 1;
    return {
        ID: id,
        created_at: "",
        updated_at: "",
        logged_exercise_id: 1,
        weight: partial.weight,
        reps: partial.reps,
        weight_setup: partial.weight_setup ?? "",
    };
}

describe("initializeWeightAndReps", () => {
    it("returns zeros and empty strings when there are no logged sets and no previous sets", () => {
        expect(initializeWeightAndReps(undefined, undefined)).toEqual({
            weight: 0,
            reps: 0,
            weightSetup: "",
            notes: "",
        });
    });

    it("treats an empty loggedSets array like missing data and falls back to previous sets", () => {
        const previous = [
            loggedSet({ weight: 70, reps: 8, id: 1 }),
            loggedSet({ weight: 90, reps: 5, id: 2 }),
        ];
        expect(initializeWeightAndReps([], previous)).toEqual({
            weight: 90,
            reps: 5,
            weightSetup: "",
            notes: "",
        });
    });

    it("uses the last logged set's weight, reps, and weight_setup when loggedSets is non-empty", () => {
        const sets = [
            loggedSet({ weight: 60, reps: 12, weight_setup: "first", id: 1 }),
            loggedSet({ weight: 55, reps: 10, weight_setup: "last", id: 2 }),
        ];
        const previous = [loggedSet({ weight: 100, reps: 3, id: 99 })];
        expect(initializeWeightAndReps(sets, previous)).toEqual({
            weight: 55,
            reps: 10,
            weightSetup: "last",
            notes: "",
        });
    });

    it("ignores previousSets when loggedSets has entries", () => {
        const sets = [loggedSet({ weight: 40, reps: 15 })];
        expect(
            initializeWeightAndReps(sets, [
                loggedSet({ weight: 200, reps: 1, id: 2 }),
            ]),
        ).toEqual({
            weight: 40,
            reps: 15,
            weightSetup: "",
            notes: "",
        });
    });

    it("seeds from the final set's weight after a drop", () => {
        const previous = [
            loggedSet({ weight: 100, reps: 10, weight_setup: "belt", id: 1 }),
            loggedSet({ weight: 90, reps: 8, id: 2 }),
        ];
        expect(initializeWeightAndReps(undefined, previous)).toEqual({
            weight: 90,
            reps: 8,
            weightSetup: "",
            notes: "",
        });
    });

    it("seeds from the final (heaviest) set when ramping up", () => {
        const previous = [
            loggedSet({ weight: 80, reps: 6, id: 1 }),
            loggedSet({ weight: 100, reps: 5, weight_setup: "belt", id: 2 }),
        ];
        expect(initializeWeightAndReps(undefined, previous)).toEqual({
            weight: 100,
            reps: 5,
            weightSetup: "belt",
            notes: "",
        });
    });

    it("seeds from the highest reps previous set when same weight", () => {
        const previous = [
            loggedSet({ weight: 80, reps: 6, id: 1 }),
            loggedSet({ weight: 80, reps: 5, weight_setup: "belt", id: 2 }),
        ];
        expect(initializeWeightAndReps(undefined, previous)).toEqual({
            weight: 80,
            reps: 6,
            weightSetup: "",
            notes: "",
        });
    });
});

describe("previousRepsForSetAtWeight", () => {
    const sets = [
        { weight: 100, reps: 6 },
        { weight: 100, reps: 5 },
    ];

    it("returns reps from the matching set number and weight", () => {
        expect(previousRepsForSetAtWeight(sets, 0, 100)).toBe(6);
        expect(previousRepsForSetAtWeight(sets, 1, 100)).toBe(5);
    });

    it("hides previous reps when last time had no set at that weight", () => {
        expect(previousRepsForSetAtWeight(sets, 0, 105)).toBeNull();
    });

    it("falls back to the nearest set last time at that weight", () => {
        const drop = [
            { weight: 100, reps: 10 },
            { weight: 90, reps: 8 },
        ];
        expect(previousRepsForSetAtWeight(drop, 0, 90)).toBe(8);
        expect(previousRepsForSetAtWeight(drop, 1, 100)).toBe(10);
        expect(previousRepsForSetAtWeight(drop, 1, 90)).toBe(8);
    });
});
