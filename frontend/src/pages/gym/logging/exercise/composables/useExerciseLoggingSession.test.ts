import { describe, expect, it } from "vitest";
import {
    defaultsFromLoggedSets,
    pickBestLoggedSet,
} from "../domain/loggedSetDefaults";

describe("pickBestLoggedSet", () => {
    it("returns null for an empty list", () => {
        expect(pickBestLoggedSet([])).toBeNull();
    });

    it("returns the only set when there is one", () => {
        const only = { weight: 100, reps: 8 };
        expect(pickBestLoggedSet([only])).toBe(only);
    });

    it("prefers higher weight", () => {
        const sets = [
            { weight: 100, reps: 10 },
            { weight: 110, reps: 5 },
            { weight: 90, reps: 20 },
        ];
        expect(pickBestLoggedSet(sets)).toEqual({ weight: 110, reps: 5 });
    });

    it("on equal weight, prefers higher reps", () => {
        const sets = [
            { weight: 100, reps: 8 },
            { weight: 100, reps: 10 },
            { weight: 100, reps: 9 },
        ];
        expect(pickBestLoggedSet(sets)).toEqual({ weight: 100, reps: 10 });
    });

    it("on equal weight and reps, keeps the first occurrence", () => {
        const first = { weight: 100, reps: 10, id: "a" };
        const second = { weight: 100, reps: 10, id: "b" };
        expect(pickBestLoggedSet([first, second])).toBe(first);
    });
});

describe("defaultsFromLoggedSets (initialization defaults)", () => {
    it("returns zeros when there are no sets", () => {
        expect(defaultsFromLoggedSets([])).toEqual({
            weight: 0,
            reps: 0,
            weight_setup: "",
        });
    });

    it("maps the best set and normalizes weight_setup", () => {
        expect(
            defaultsFromLoggedSets([
                { weight: 60, reps: 10, weight_setup: "each side" },
                { weight: 80, reps: 5 },
            ]),
        ).toEqual({
            weight: 80,
            reps: 5,
            weight_setup: "",
        });
    });

    it("uses empty string when weight_setup is missing", () => {
        expect(
            defaultsFromLoggedSets([{ weight: 50, reps: 12 }]),
        ).toEqual({
            weight: 50,
            reps: 12,
            weight_setup: "",
        });
    });
});
