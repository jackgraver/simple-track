import { describe, expect, it } from "vitest";
import type { ExerciseGroup } from "../../store/useWorkoutStore";
import { buildWorkoutSteps } from "../../domain/workoutSteps";
import { buildUpNext, openingWeight } from "./upNext";

const set = (weight: number, reps: number) => ({ weight, reps }) as never;
const group = (id: number, name: string, prev: [number, number][], logged: [number, number][] = [], rollover = 10): ExerciseGroup =>
    ({
        planned: { ID: id, name, rep_rollover: rollover, cues: "", machine_settings: [{ label: "Seat", value: "4" }] },
        previous: prev.length ? { sets: prev.map(([w, r]) => set(w, r)), exercise: { rep_rollover: rollover } } : undefined,
        logged: logged.length ? { exercise_id: id, sets: logged.map(([w, r]) => set(w, r)) } : undefined,
    }) as never;

describe("openingWeight", () => {
    it("adds the increase when every set hit the rollover", () => {
        expect(openingWeight(group(1, "A", [[100, 10], [100, 10]]))).toBe(105);
        expect(openingWeight(group(1, "A", [[100, 10], [100, 9]]))).toBe(100);
    });
    it("uses today's last set once logging started", () => {
        expect(openingWeight(group(1, "A", [[100, 10]], [[110, 6]]))).toBe(110);
    });
});

describe("buildUpNext", () => {
    const log = [
        group(1, "Incline Press", [[100, 10]], [[105, 10]]),
        group(2, "Seated Row", [[120, 10], [120, 9]]),
        group(3, "Chest Fly", []),
        group(4, "Lat Pulldown", [[130, 10]]),
    ];
    const steps = buildWorkoutSteps({ log, plannedCardio: { type: "Bike", minutes: 10 } });

    it("skips done and current exercises, keeps plan order, ends with cardio", () => {
        const items = buildUpNext(steps, log, 2);
        expect(items.map((i) => i.name)).toEqual(["Chest Fly", "Lat Pulldown", "Bike"]);
        expect(items[1]).toMatchObject({ weight: 135, exerciseId: 4, detail: "Last 130 lb × 10" });
        expect(items[1].machineSettings).toEqual([{ label: "Seat", value: "4" }]);
        expect(items[0]).toMatchObject({ weight: null, detail: "No history" });
    });
});
