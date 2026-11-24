import { describe, expect, it } from "vitest";
import {
    addPlate,
    baseWeightFor,
    cleanPlates,
    formatPlates,
    loadFromPlates,
    normalizeLoadType,
    parseLegacyPlates,
    plateConfigFor,
    plateMismatch,
    platesForWeight,
    plateSetupText,
    platesOfSet,
    plateStack,
    weightStepFor,
} from "./plates";

const bar = { base: 45, mode: "per_side" } as const;
const sled = { base: 0, mode: "per_side" } as const;
const peg = { base: 0, mode: "total" } as const;

describe("exercise settings", () => {
    it("maps legacy load types to per side", () => {
        expect(normalizeLoadType("plate_loaded_with_bar")).toBe("plate_loaded_per_side");
        expect(normalizeLoadType("plate_loaded_without_bar")).toBe("plate_loaded_per_side");
        expect(normalizeLoadType(undefined)).toBe("plate_loaded_per_side");
        expect(normalizeLoadType("weight_stack")).toBe("weight_stack");
    });

    it("defaults the base weight by load type, and keeps an explicit 0", () => {
        expect(baseWeightFor({ load_type: "plate_loaded_per_side" })).toBe(45);
        expect(baseWeightFor({ load_type: "plate_loaded_without_bar" })).toBe(0);
        expect(baseWeightFor({ load_type: "plate_loaded_per_side", base_weight: 0 })).toBe(0);
        expect(baseWeightFor({ load_type: "plate_loaded_per_side", base_weight: 25 })).toBe(25);
        expect(baseWeightFor({ load_type: "plate_loaded_total" })).toBe(0);
    });

    it("has no plate config for stacks and dumbbells", () => {
        expect(plateConfigFor({ load_type: "weight_stack" })).toBeNull();
        expect(plateConfigFor({ load_type: "free_weights" })).toBeNull();
        expect(plateConfigFor({ load_type: "plate_loaded_total", base_weight: 10 })).toEqual({ base: 10, mode: "total" });
    });

    it("uses the exercise's step, else 5 lb", () => {
        expect(weightStepFor({ load_type: "plate_loaded_per_side" })).toBe(5);
        expect(weightStepFor({ load_type: "weight_stack" })).toBe(5);
        expect(weightStepFor({ load_type: "plate_loaded_per_side", weight_step: 10 })).toBe(10);
        expect(weightStepFor({ load_type: "weight_stack", weight_step: 15 })).toBe(15);
        expect(weightStepFor({ load_type: "weight_stack", weight_step: 0 })).toBe(5);
        expect(weightStepFor(null)).toBe(5);
    });
});

describe("loadFromPlates", () => {
    it("counts per-side plates twice, plus the base", () => {
        expect(loadFromPlates(bar, { 45: 2, 10: 1 })).toBe(245);
        expect(loadFromPlates(sled, { 45: 2 })).toBe(180);
        expect(loadFromPlates(bar, {})).toBe(45);
    });

    it("counts total-mode plates once", () => {
        expect(loadFromPlates(peg, { 45: 1, 25: 1 })).toBe(70);
    });
});

describe("platesForWeight", () => {
    it("loads heaviest first from nothing", () => {
        expect(platesForWeight(225, bar)).toEqual({ plates: { 45: 2 }, load: 225, exact: true });
        expect(platesForWeight(185, bar).plates).toEqual({ 45: 1, 25: 1 });
        expect(platesForWeight(155, bar).plates).toEqual({ 45: 1, 10: 1 });
    });

    it("only uses 2.5s when nothing heavier lands", () => {
        expect(platesForWeight(140, bar).plates).toEqual({ 45: 1, 2.5: 1 });
        expect(platesForWeight(145, bar).plates).toEqual({ 45: 1, 5: 1 });
    });

    it("adds on top of the current plates going up", () => {
        // 45+25 a side → +10 total adds a 5, rather than swapping to 45+25+5 from scratch.
        expect(platesForWeight(195, bar, { 45: 1, 25: 1 }).plates).toEqual({ 45: 1, 25: 1, 5: 1 });
        // +20 adds a 10, not a swap of the 25 for a 35.
        expect(platesForWeight(205, bar, { 45: 1, 25: 1 }).plates).toEqual({ 45: 1, 25: 1, 10: 1 });
    });

    it("takes plates off going down", () => {
        expect(platesForWeight(185, bar, { 45: 1, 25: 1, 10: 1 }).plates).toEqual({ 45: 1, 25: 1 });
        expect(platesForWeight(155, bar, { 45: 1, 10: 1, 5: 1 }).plates).toEqual({ 45: 1, 10: 1 });
    });

    it("reloads when plates can't come off exactly", () => {
        // 45+25 a side → 175 needs 65 a side: no plates on the bar sum to 5, so it reloads.
        expect(platesForWeight(175, bar, { 45: 1, 25: 1 })).toEqual({
            plates: { 45: 1, 10: 2 },
            load: 175,
            exact: true,
        });
    });

    it("keeps the current plates when they already match", () => {
        const plates = { 35: 1, 10: 1 };
        expect(platesForWeight(135, bar, plates).plates).toEqual(plates);
    });

    it("returns the nearest loadable weight when the target can't be loaded", () => {
        // 137.5 needs 46.25 a side; 135 and 140 are equally near, so the lower wins.
        expect(platesForWeight(137.5, bar)).toEqual({ plates: { 45: 1 }, load: 135, exact: false });
        expect(platesForWeight(139, bar)).toMatchObject({ load: 140, exact: false });
        expect(platesForWeight(72, peg)).toMatchObject({ load: 72.5, exact: false });
    });

    it("gives an empty bar at or below the base", () => {
        expect(platesForWeight(45, bar)).toEqual({ plates: {}, load: 45, exact: true });
        expect(platesForWeight(30, bar)).toEqual({ plates: {}, load: 45, exact: false });
    });

    it("works in total mode", () => {
        expect(platesForWeight(70, peg).plates).toEqual({ 45: 1, 25: 1 });
        expect(platesForWeight(75, peg, { 45: 1, 25: 1 }).plates).toEqual({ 45: 1, 25: 1, 5: 1 });
    });
});

describe("plate helpers", () => {
    it("adds and removes single plates", () => {
        expect(addPlate({ 45: 1 }, 10, 1)).toEqual({ 45: 1, 10: 1 });
        expect(addPlate({ 45: 1, 10: 1 }, 10, -1)).toEqual({ 45: 1 });
        expect(addPlate({}, 10, -1)).toEqual({});
    });

    it("cleans counts and keys", () => {
        expect(cleanPlates({ "45.0": 1, 10: 0, x: 2, "2.5": -1 })).toEqual({ 45: 1 });
    });

    it("stacks heaviest first", () => {
        expect(plateStack({ 10: 1, 45: 2 })).toEqual([45, 45, 10]);
    });

    it("flags a weight that doesn't match its plates", () => {
        expect(plateMismatch(140, 135)).toBe(5);
        expect(plateMismatch(135, 135)).toBeNull();
        expect(plateMismatch(135.2, 135)).toBeNull();
    });
});

describe("text", () => {
    it("formats heaviest first with counts", () => {
        expect(formatPlates({ 10: 2, 45: 1, 2.5: 1 })).toBe("45 + 2×10 + 2.5");
        expect(formatPlates({})).toBe("");
    });

    it("adds the bar to the legacy text only when it isn't 45", () => {
        expect(plateSetupText(bar, { 45: 1 })).toBe("45");
        expect(plateSetupText(sled, { 45: 2 })).toBe("2×45, bar 0");
        expect(plateSetupText({ base: 25, mode: "per_side" }, {})).toBe("bar 25");
        expect(plateSetupText(peg, { 45: 1 })).toBe("45");
    });

    it("parses old text and round-trips the new", () => {
        expect(parseLegacyPlates("2×45 + 10")).toEqual({ 45: 2, 10: 1 });
        expect(parseLegacyPlates("2x45+10, bar 35")).toEqual({ 45: 2, 10: 1 });
        expect(parseLegacyPlates("45 + 45")).toEqual({ 45: 2 });
        expect(parseLegacyPlates("bar 0")).toEqual({});
        expect(parseLegacyPlates("")).toEqual({});
        expect(parseLegacyPlates("belt, slow")).toBeNull();
        const plates = { 45: 1, 10: 2, 2.5: 1 };
        expect(parseLegacyPlates(plateSetupText(sled, plates))).toEqual(plates);
    });

    it("reads a set's plates from its setup, else its text", () => {
        expect(platesOfSet({ plate_setup: { base: 45, mode: "per_side", plates: { 25: 1 } }, weight_setup: "45" })).toEqual({ 25: 1 });
        expect(platesOfSet({ plate_setup: null, weight_setup: "45 + 10" })).toEqual({ 45: 1, 10: 1 });
        expect(platesOfSet({ weight_setup: "" })).toBeNull();
        expect(platesOfSet({ weight_setup: "belt" })).toBeNull();
    });
});
