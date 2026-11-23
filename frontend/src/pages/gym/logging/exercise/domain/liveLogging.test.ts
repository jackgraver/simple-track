import { describe, expect, it } from "vitest";
import {
    formatWeightDelta,
    overloadReason,
    overloadStatus,
    plannedSetCount,
    progressVerdict,
    restBarPercent,
    restStage,
    setsTally,
    visibleMachineSettings,
} from "./liveLogging";

describe("restStage", () => {
    it("is idle without a running clock", () => {
        expect(restStage(90_000, false)).toBe("idle");
    });
    it("goes wait → near → ready at 1:00 and 2:00", () => {
        expect(restStage(0, true)).toBe("wait");
        expect(restStage(59_999, true)).toBe("wait");
        expect(restStage(60_000, true)).toBe("near");
        expect(restStage(119_999, true)).toBe("near");
        expect(restStage(120_000, true)).toBe("ready");
        expect(restStage(400_000, true)).toBe("ready");
    });
});

describe("restBarPercent", () => {
    it("fills to 3:00 and clamps", () => {
        expect(restBarPercent(0)).toBe(0);
        expect(restBarPercent(90_000)).toBe(50);
        expect(restBarPercent(500_000)).toBe(100);
        expect(restBarPercent(Number.NaN)).toBe(0);
    });
});

describe("plannedSetCount", () => {
    it("uses last session's count, else 2", () => {
        expect(plannedSetCount([{ weight: 1, reps: 1 }, { weight: 1, reps: 1 }, { weight: 1, reps: 1 }])).toBe(3);
        expect(plannedSetCount([])).toBe(2);
        expect(plannedSetCount(undefined)).toBe(2);
    });
});

describe("setsTally", () => {
    it("counts against the plan, then just counts past it", () => {
        expect(setsTally(2, 4)).toBe("2/4");
        expect(setsTally(4, 4)).toBe("4/4");
        expect(setsTally(5, 4)).toBe("5");
    });
});

describe("overloadStatus", () => {
    const prev = [
        { weight: 100, reps: 10 },
        { weight: 100, reps: 10 },
        { weight: 100, reps: 9 },
    ];
    it("is none without history or weight", () => {
        expect(overloadStatus([], 10, 100)).toEqual({ kind: "none" });
        expect(overloadStatus(prev, 10, 0)).toEqual({ kind: "none" });
    });
    it("reports an earned increase", () => {
        const all = prev.map((s) => ({ ...s, reps: 10 }));
        expect(overloadStatus(all, 10, 105)).toMatchObject({ kind: "up", delta: 5, hit: 3, total: 3, earned: true });
    });
    it("holds when not every set hit the rollover", () => {
        expect(overloadStatus(prev, 10, 100)).toMatchObject({ kind: "same", delta: 0, hit: 2, total: 3, earned: false });
    });
    it("handles a manual drop and no rollover", () => {
        expect(overloadStatus(prev, 0, 90)).toMatchObject({ kind: "down", delta: -10, hit: null, rollover: null });
    });
    it("compares against the top weight last time", () => {
        const mixed = [
            { weight: 80, reps: 12 },
            { weight: 100, reps: 10 },
        ];
        expect(overloadStatus(mixed, 10, 100)).toMatchObject({ previousWeight: 100, hit: 1, total: 1, earned: true });
    });
});

describe("formatWeightDelta", () => {
    it("signs the delta", () => {
        expect(formatWeightDelta(5)).toBe("+5");
        expect(formatWeightDelta(-2.5)).toBe("−2.5");
        expect(formatWeightDelta(0)).toBe("");
    });
});

describe("progressVerdict", () => {
    it("counts today's sets at the current weight", () => {
        const sets = [
            { weight: 105, reps: 10 },
            { weight: 105, reps: 9 },
            { weight: 95, reps: 12 },
        ];
        expect(progressVerdict(sets, 10, 105)).toEqual({ hit: 1, total: 2, rollover: 10, onTrack: false });
        expect(progressVerdict(sets.slice(0, 1), 10, 105)?.onTrack).toBe(true);
    });
    it("is null without a rollover or sets", () => {
        expect(progressVerdict([{ weight: 1, reps: 1 }], 0, 1)).toBeNull();
        expect(progressVerdict([], 10, 100)).toBeNull();
    });
});

describe("visibleMachineSettings", () => {
    it("drops unlabeled entries and handles null", () => {
        expect(visibleMachineSettings(null)).toEqual([]);
        expect(
            visibleMachineSettings([
                { label: "Seat", value: "4" },
                { label: " ", value: "x" },
            ]),
        ).toEqual([{ label: "Seat", value: "4" }]);
    });
});

describe("overloadReason", () => {
    it("explains an earned increase and a hold", () => {
        const all = [
            { weight: 100, reps: 10 },
            { weight: 100, reps: 10 },
        ];
        expect(overloadReason(overloadStatus(all, 10, 105))).toBe("Last time 2/2 sets hit 10 at 100 lb, so +5 today.");
        expect(overloadReason(overloadStatus([{ weight: 100, reps: 8 }, ...all], 10, 100))).toBe(
            "Last time 2/3 sets hit 10. Hit 10 on every set to go up.",
        );
        expect(overloadReason({ kind: "none" })).toMatch(/No history/);
    });
});
