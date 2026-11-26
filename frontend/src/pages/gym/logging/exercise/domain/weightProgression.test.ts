import { describe, expect, it } from "vitest";
import { getWeightProgression, sessionOutcome } from "./weightProgression";

const sets = (...pairs: [number, number][]) => pairs.map(([weight, reps]) => ({ weight, reps }));

describe("sessionOutcome", () => {
    it("is null without sets or a weight", () => {
        expect(sessionOutcome([], 10)).toBeNull();
        expect(sessionOutcome(undefined, 10)).toBeNull();
        expect(sessionOutcome(sets([0, 10]), 10)).toBeNull();
    });

    it("earns the increase when every set is at one weight and hits the rollover", () => {
        expect(sessionOutcome(sets([100, 10], [100, 12]), 10)).toMatchObject({
            startWeight: 100,
            sameWeight: true,
            dropped: false,
            hit: 2,
            total: 2,
            earned: true,
        });
    });

    it("holds when a set misses the rollover", () => {
        expect(sessionOutcome(sets([100, 10], [100, 8]), 10)).toMatchObject({ startWeight: 100, hit: 1, earned: false });
    });

    it("starts the next session at a dropped weight, without an increase", () => {
        expect(sessionOutcome(sets([100, 10], [90, 10]), 10)).toMatchObject({
            startWeight: 90,
            sameWeight: false,
            dropped: true,
            hit: 2,
            earned: false,
        });
    });

    it("an off-day bounce back up starts at the final weight, without an increase", () => {
        expect(sessionOutcome(sets([100, 10], [90, 8], [100, 10]), 10)).toMatchObject({
            startWeight: 100,
            sameWeight: false,
            dropped: false,
            earned: false,
        });
    });

    it("logged ramp-up sets block the increase", () => {
        expect(sessionOutcome(sets([80, 10], [90, 10], [100, 10]), 10)).toMatchObject({ startWeight: 100, earned: false });
    });

    it("never earns without a rollover", () => {
        expect(sessionOutcome(sets([100, 12]), undefined)).toMatchObject({ hit: null, rollover: null, earned: false });
    });
});

describe("getWeightProgression", () => {
    it("adds five pounds when the session earned it", () => {
        expect(getWeightProgression(sets([100, 10], [100, 12]), 10)).toEqual({
            previousWeight: 100,
            nextWeight: 105,
            increase: 5,
        });
    });

    it("a proving session at the dropped weight earns the increase from there", () => {
        expect(getWeightProgression(sets([90, 10], [90, 10]), 10)?.nextWeight).toBe(95);
    });

    it("does not progress after a drop, a miss, or without a rollover", () => {
        expect(getWeightProgression(sets([100, 10], [90, 10]), 10)).toBeNull();
        expect(getWeightProgression(sets([100, 10], [100, 9]), 10)).toBeNull();
        expect(getWeightProgression(sets([100, 12]), undefined)).toBeNull();
    });
});
