import { describe, expect, it } from "vitest";
import {
    computeBarbellTotalLbs,
    formatWeightSetup,
    parseWeightSetup,
    perSideLoadLbs,
    weightSetupMismatchLbs,
} from "./weightSetup";

describe("computeBarbellTotalLbs", () => {
    it("adds bar plus twice per-side load", () => {
        expect(computeBarbellTotalLbs(45, { 45: 2, 10: 1 })).toBe(185);
    });

    it("handles empty plates as bar only", () => {
        expect(computeBarbellTotalLbs(45, {})).toBe(45);
    });

    it("omits bar weight when bar is zero", () => {
        expect(computeBarbellTotalLbs(0, { 45: 2, 10: 1 })).toBe(180);
    });
});

describe("formatWeightSetup", () => {
    it("formats counts and omits default bar", () => {
        expect(formatWeightSetup(45, { 45: 2, 10: 1 })).toBe("2×45 + 10");
    });

    it("includes non-default bar", () => {
        expect(formatWeightSetup(35, { 45: 1 })).toBe("45, bar 35");
    });

    it("marks plates-only setups with bar 0", () => {
        expect(formatWeightSetup(0, { 45: 2 })).toBe("2×45, bar 0");
    });

    it("returns empty for plates-only with no plates", () => {
        expect(formatWeightSetup(0, {})).toBe("");
    });
});

describe("parseWeightSetup", () => {
    it("parses multiplied and single plates", () => {
        expect(parseWeightSetup("2x45 + 10")).toEqual({
            barLbs: 45,
            platesPerSide: { 45: 2, 10: 1 },
            parsed: true,
        });
    });

    it("parses bar suffix", () => {
        expect(parseWeightSetup("45, bar 35")).toEqual({
            barLbs: 35,
            platesPerSide: { 45: 1 },
            parsed: true,
        });
    });

    it("parses plates-only bar zero suffix", () => {
        expect(parseWeightSetup("2×45 + 10, bar 0")).toEqual({
            barLbs: 0,
            platesPerSide: { 45: 2, 10: 1 },
            parsed: true,
        });
    });

    it("returns unparsed for invalid tokens", () => {
        expect(parseWeightSetup("each side")).toEqual({
            barLbs: 45,
            platesPerSide: {},
            parsed: false,
        });
    });
});

describe("weightSetupMismatchLbs", () => {
    it("returns null when within half pound", () => {
        expect(weightSetupMismatchLbs(185, 185)).toBeNull();
    });

    it("returns delta when mismatched", () => {
        expect(weightSetupMismatchLbs(185, 180)).toBe(5);
    });
});

describe("perSideLoadLbs", () => {
    it("sums plate weight times count", () => {
        expect(perSideLoadLbs({ 45: 2, 10: 1 })).toBe(100);
    });
});
