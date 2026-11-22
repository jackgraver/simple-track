import { describe, expect, it } from "vitest";
import type { BodyWeightLog, WaterLog } from "~/api/tracking/types";
import {
    addWaterPatch,
    deleteWaterPatch,
    reconcileByDay,
    reconcileWater,
    revertAddWater,
    revertDeleteWater,
    revertUpsertByDay,
    saveStepsPatch,
    saveWeightPatch,
    tempWaterId,
    type AddWaterVars,
} from "./patches";

const w = (ID: number, oz: number, client_id?: string): WaterLog => ({ ID, date: "2026-11-22", amount_oz: oz, preset_id: null, client_id });
const add = (id: string, oz = 16): AddWaterVars => ({ client_id: id, date: "2026-11-22", amountOz: oz });
const A = "aaaaaaaa-0000-4000-8000-000000000001";
const B = "bbbbbbbb-0000-4000-8000-000000000002";

describe("water add", () => {
    it("appends a row with a negative temp id, and is idempotent per client_id", () => {
        const once = addWaterPatch([w(1, 8)], add(A));
        expect(once).toHaveLength(2);
        expect(once[1]).toMatchObject({ client_id: A, amount_oz: 16, ID: tempWaterId(A) });
        expect(once[1].ID).toBeLessThan(0);
        expect(addWaterPatch(once, add(A))).toBe(once);
        expect(tempWaterId(A)).not.toBe(tempWaterId(B));
    });

    it("revert removes only its own row, leaving a sibling's", () => {
        const both = addWaterPatch(addWaterPatch([], add(A)), add(B));
        expect(revertAddWater(both, add(A)).map((r) => r.client_id)).toEqual([B]);
    });

    it("reconcile swaps the temp row for the server row and keeps siblings", () => {
        const both = addWaterPatch(addWaterPatch([w(1, 8)], add(A)), add(B));
        const next = reconcileWater(both, add(A), w(42, 16));
        expect(next.map((r) => r.ID)).toEqual([1, 42, tempWaterId(B)]);
        expect(next[1].client_id).toBe(A);
    });

    it("reconcile appends if the temp row vanished (refetch) and the server row is not there yet", () => {
        expect(reconcileWater([w(1, 8)], add(A), w(42, 16)).map((r) => r.ID)).toEqual([1, 42]);
        // Refetch already brought the server row: no duplicate.
        expect(reconcileWater([w(1, 8), w(42, 16, A)], add(A), w(42, 16)).map((r) => r.ID)).toEqual([1, 42]);
    });
});

describe("water delete", () => {
    it("removes by id and revert restores from the snapshot without wiping other changes", () => {
        const snap = [w(1, 8), w(2, 12)];
        const patched = deleteWaterPatch(snap, { id: 2, date: "d" });
        expect(patched.map((r) => r.ID)).toEqual([1]);
        const withSibling = addWaterPatch(patched, add(A));
        expect(revertDeleteWater(withSibling, { id: 2, date: "d" }, snap).map((r) => r.ID)).toEqual([1, tempWaterId(A), 2]);
        // Already back (refetch): unchanged.
        expect(revertDeleteWater(snap, { id: 2, date: "d" }, snap)).toBe(snap);
    });
});

const wt = (date: string, lbs: number, ID = 1): BodyWeightLog => ({ ID, date, weight_lbs: lbs });

describe("weight / steps upsert by day", () => {
    it("overwrites the day's row, keeping other fields (Renpho metrics)", () => {
        const old = [{ ...wt("2026-11-22T00:00:00Z", 180, 5), bmi: 24 }, wt("2026-11-21T00:00:00Z", 181, 4)];
        const next = saveWeightPatch(old, { date: "2026-11-22", weightLbs: 179 });
        expect(next).toHaveLength(2);
        expect(next[0]).toMatchObject({ ID: 5, weight_lbs: 179, bmi: 24 });
    });

    it("inserts a new day newest-first", () => {
        const old = [wt("2026-11-22T00:00:00Z", 180), wt("2026-11-20T00:00:00Z", 182)];
        const next = saveWeightPatch(old, { date: "2026-11-21", weightLbs: 181 });
        expect(next.map((r) => r.date.slice(0, 10))).toEqual(["2026-11-22", "2026-11-21", "2026-11-20"]);
        expect(saveStepsPatch([], { date: "2026-11-21", steps: 9000 })).toEqual([{ ID: 0, date: "2026-11-21", steps: 9000 }]);
    });

    it("revert restores the previous row, or removes an inserted one, leaving other days alone", () => {
        const snap = [wt("2026-11-22T00:00:00Z", 180, 5)];
        const overwritten = saveWeightPatch(snap, { date: "2026-11-22", weightLbs: 170 });
        expect(revertUpsertByDay(overwritten, "2026-11-22", snap)).toEqual(snap);

        const inserted = saveWeightPatch(snap, { date: "2026-11-21", weightLbs: 181 });
        const other = saveWeightPatch(inserted, { date: "2026-11-20", weightLbs: 190 });
        expect(revertUpsertByDay(other, "2026-11-21", snap).map((r) => r.date.slice(0, 10))).toEqual(["2026-11-22", "2026-11-20"]);
    });

    it("reconcile writes the server row (real ID) over the optimistic one", () => {
        const optimistic = saveWeightPatch([], { date: "2026-11-22", weightLbs: 179 });
        const next = reconcileByDay(optimistic, wt("2026-11-22T00:00:00Z", 179, 77));
        expect(next).toEqual([wt("2026-11-22T00:00:00Z", 179, 77)]);
    });
});
