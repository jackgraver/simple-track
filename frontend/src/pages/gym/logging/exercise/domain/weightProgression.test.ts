import { describe, expect, it } from "vitest";
import { getWeightProgression } from "./weightProgression";

describe("getWeightProgression", () => {
    it("adds five pounds when every set at the current weight met rollover", () => {
        expect(
            getWeightProgression(
                [
                    { weight: 95, reps: 8 },
                    { weight: 100, reps: 10 },
                    { weight: 100, reps: 12 },
                ],
                10,
                100,
            ),
        ).toEqual({
            previousWeight: 100,
            nextWeight: 105,
            increase: 5,
        });
    });

    it("does not progress when any set at the current weight missed rollover", () => {
        expect(
            getWeightProgression(
                [
                    { weight: 100, reps: 10 },
                    { weight: 100, reps: 9 },
                ],
                10,
                100,
            ),
        ).toBeNull();
    });

    it("does not progress without matching previous sets or a valid rollover", () => {
        expect(
            getWeightProgression([{ weight: 95, reps: 12 }], 10, 100),
        ).toBeNull();
        expect(
            getWeightProgression([{ weight: 100, reps: 12 }], undefined, 100),
        ).toBeNull();
    });
});
