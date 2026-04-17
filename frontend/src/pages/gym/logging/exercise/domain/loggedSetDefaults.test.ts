import { describe, expect, it } from "vitest";
import type { LoggedSet } from "~/types/workout";
import { initializeWeightAndReps } from "./loggedSetDefaults";

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

    it("seeds from the heighest previous set when different weights", () => {
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
