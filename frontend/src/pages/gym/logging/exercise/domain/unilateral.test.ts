import { describe, expect, it } from "vitest";
import {
    carriedSideWeight,
    closesPair,
    collapsePair,
    completePairCount,
    defaultFirstSide,
    formatPair,
    lastOnSide,
    nextPairNumber,
    nextSide,
    pairIndexOfRow,
    pairSets,
    previousSideSet,
    setsForProgression,
    sideSet,
    type SidedSet,
} from "./unilateral";

const L = (reps: number, weight = 30): SidedSet => ({ weight, reps, side: "left" });
const R = (reps: number, weight = 30): SidedSet => ({ weight, reps, side: "right" });
const B = (reps: number, weight = 30): SidedSet => ({ weight, reps });

/** A side with its own id (`key`) and the pair id it was logged with. */
const side = (s: "left" | "right", key: string, pair: string, reps = 10): SidedSet => ({
    weight: 30,
    reps,
    side: s,
    key,
    pair_id: pair,
});

describe("pairSets with stored pair ids", () => {
    const today = [side("left", "a", "a"), side("right", "b", "a"), side("left", "c", "c"), side("right", "d", "c")];

    it("pairs by pair_id", () => {
        const pairs = pairSets(today);
        expect(pairs.map((p) => [p.left?.set.key, p.right?.set.key])).toEqual([
            ["a", "b"],
            ["c", "d"],
        ]);
    });

    it("deleting one side of an earlier set leaves every later pair alone", () => {
        const pairs = pairSets(today.filter((s) => s.key !== "a"));
        expect(pairs.map((p) => [p.left?.set.key, p.right?.set.key])).toEqual([
            [undefined, "b"],
            ["c", "d"],
        ]);
        expect(completePairCount(today.filter((s) => s.key !== "a"))).toBe(1);
        // The next log starts set 3, on today's first side.
        expect(nextPairNumber(today.filter((s) => s.key !== "a"))).toBe(3);
    });

    it("moving one side of a pair to the other side doesn't re-pair later sets", () => {
        const moved = today.map((s) => (s.key === "a" ? { ...s, side: "right" as const } : s));
        const pairs = pairSets(moved);
        expect(pairs).toHaveLength(3);
        expect(pairs.slice(1).map((p) => [p.left?.set.key, p.right?.set.key])).toEqual([
            [undefined, "b"],
            ["c", "d"],
        ]);
    });

    it("joins a second side to a first side saved without a pair_id (by its own id)", () => {
        const pairs = pairSets([{ weight: 30, reps: 10, side: "left", key: "a" }, side("right", "b", "a")]);
        expect(pairs).toHaveLength(1);
        expect(pairs[0].right?.set.key).toBe("b");
    });
});

describe("pairSets", () => {
    it("pairs alternating sides in log order, whichever goes first", () => {
        const pairs = pairSets([R(10), L(9), R(8), L(8)]);
        expect(pairs).toHaveLength(2);
        expect(pairs[0].first).toBe("right");
        expect(pairs[0].right?.index).toBe(0);
        expect(pairs[0].left?.index).toBe(1);
        expect(pairs[1].right?.index).toBe(2);
    });

    it("starts a new pair when the same side is logged twice", () => {
        const pairs = pairSets([L(10), L(9), R(9)]);
        expect(pairs).toHaveLength(2);
        expect(pairs[0].right).toBeUndefined();
        expect(pairs[1].left?.index).toBe(1);
        expect(pairs[1].right?.index).toBe(2);
    });

    it("keeps a bilateral set as its own pair that counts for both sides", () => {
        const pairs = pairSets([B(10), B(9)]);
        expect(pairs).toHaveLength(2);
        expect(sideSet(pairs[0], "left")?.reps).toBe(10);
        expect(sideSet(pairs[0], "right")?.reps).toBe(10);
    });

    it("maps a row index back to its pair", () => {
        expect(pairIndexOfRow([L(10), R(10), L(9)], 1)).toBe(0);
        expect(pairIndexOfRow([L(10), R(10), L(9)], 2)).toBe(1);
    });
});

describe("next side and set number", () => {
    it("before anything is logged: the chosen first side, set 1", () => {
        expect(nextSide([], "right")).toBe("right");
        expect(nextPairNumber([])).toBe(1);
    });

    it("after the first side: the other side, same set number", () => {
        expect(nextSide([R(10)], "left")).toBe("left");
        expect(nextPairNumber([R(10)])).toBe(1);
        expect(closesPair([R(10)], "left")).toBe(true);
        expect(closesPair([R(10)], "right")).toBe(false);
    });

    it("after a full pair: today's first side again, next set number", () => {
        expect(nextSide([R(10), L(10)], "left")).toBe("right");
        expect(nextPairNumber([R(10), L(10)])).toBe(2);
        expect(completePairCount([R(10), L(10), R(9)])).toBe(1);
    });

    it("defaults the first side to last session's", () => {
        expect(defaultFirstSide([R(10), L(10)])).toBe("right");
        expect(defaultFirstSide([B(10)])).toBe("left");
        expect(defaultFirstSide(undefined)).toBe("left");
    });

    it("finds the last set on a side", () => {
        expect(lastOnSide([L(10, 30), R(10, 25), L(9, 30)], "right")?.weight).toBe(25);
        expect(lastOnSide([B(10)], "left")).toBeUndefined();
    });
});

describe("progression view", () => {
    it("collapses a pair to its weaker side", () => {
        expect(collapsePair(pairSets([L(10), R(8)])[0])).toMatchObject({ reps: 8, weight: 30 });
        expect(collapsePair(pairSets([L(8, 30), R(12, 25)])[0])).toMatchObject({ reps: 12, weight: 25 });
        expect(collapsePair(pairSets([L(10)])[0])).toMatchObject({ reps: 10 });
    });

    it("leaves bilateral history alone and collapses sided history", () => {
        expect(setsForProgression([B(10), B(9)])).toHaveLength(2);
        expect(setsForProgression([L(10), R(9), L(10), R(10)]).map((s) => s.reps)).toEqual([9, 10]);
    });
});

describe("formatPair", () => {
    it("shares the weight when both sides match", () => {
        expect(formatPair(pairSets([L(10), R(9)])[0])).toEqual({ weight: 30, text: "L10 R9" });
    });
    it("spells out each side when the weights differ", () => {
        expect(formatPair(pairSets([L(10, 30), R(9, 25)])[0])).toEqual({
            weight: null,
            text: "L 30 lb × 10 R 25 lb × 9",
        });
    });
});

describe("carriedSideWeight", () => {
    it("keeps each side's weight after an uneven pair", () => {
        const today = [L(10, 30), R(9, 25)];
        expect(carriedSideWeight(today, "left")?.weight).toBe(30);
        expect(carriedSideWeight([...today, L(10, 30)], "right")?.weight).toBe(25);
    });
    it("carries nothing after an even pair, or before one is complete", () => {
        expect(carriedSideWeight([L(10, 30), R(10, 30)], "left")).toBeNull();
        expect(carriedSideWeight([L(10, 35)], "right")).toBeNull();
    });
});

describe("previousSideSet", () => {
    it("reads last session's same side, or the bilateral set for both", () => {
        expect(previousSideSet([R(10), L(8)], 0, "left")?.reps).toBe(8);
        expect(previousSideSet([B(10), B(9)], 1, "right")?.reps).toBe(9);
        expect(previousSideSet([B(10)], 3, "right")).toBeUndefined();
    });
});
