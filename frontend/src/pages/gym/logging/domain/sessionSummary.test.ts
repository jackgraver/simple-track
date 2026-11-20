import { describe, expect, it } from "vitest";
import type { ExerciseGroup } from "~/api/workout/api";
import { buildSessionSummary, topSet } from "./sessionSummary";

const sets = (...s: [number, number][]) => s.map(([weight, reps]) => ({ weight, reps }));
const group = (id: number, today?: [number, number][], prev?: [number, number][], max?: [number, number][]) =>
    ({
        planned: { ID: id, name: `Lift ${id}` },
        logged: today ? { exercise_id: id, sets: sets(...today) } : undefined,
        previous: prev ? { sets: sets(...prev) } : undefined,
        max: max ? { sets: sets(...max) } : undefined,
    }) as unknown as ExerciseGroup;

describe("topSet", () => {
    it("picks the heaviest set, then the most reps", () => {
        expect(topSet(sets([100, 10], [110, 5], [110, 6]))).toEqual({ weight: 110, reps: 6 });
        expect(topSet([])).toBeNull();
    });
});

describe("buildSessionSummary", () => {
    it("totals logged lifts and compares with last time and the best so far", () => {
        const s = buildSessionSummary([
            group(1, [[105, 10], [105, 8]], [[100, 10], [100, 8]], [[100, 10]]), // +5 lb, PR
            group(2, [[50, 12]], [[50, 10]], [[60, 8]]), // same weight, +2 reps, no PR
            group(3, undefined, [[80, 8]]), // not logged today: ignored
            group(4, [[20, 10]]), // first time: no comparison
        ]);
        expect(s.sets).toBe(4);
        expect(s.reps).toBe(40);
        expect(s.volume).toBe(105 * 18 + 50 * 12 + 200);
        expect(s.lifts.map((l) => [l.deltaWeight, l.deltaReps, l.isPR])).toEqual([
            [5, null, true],
            [0, 2, false],
            [null, null, false],
        ]);
        expect(s.prCount).toBe(1);
        // (1890 + 600) vs (1800 + 500): +8%
        expect(s.volumeDeltaPct).toBe(8);
    });

    it("is empty before anything is logged", () => {
        const s = buildSessionSummary([group(1, undefined, [[100, 10]])]);
        expect(s.lifts).toEqual([]);
        expect(s.volumeDeltaPct).toBeNull();
    });
});
