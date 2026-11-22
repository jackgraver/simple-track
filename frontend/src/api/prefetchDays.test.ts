import { describe, expect, it, vi } from "vitest";
import { adjacentOffsets, prefetchAdjacentDays } from "./prefetchDays";

describe("adjacentOffsets", () => {
    it("is the day before and after (offsets grow into the past)", () => {
        expect(adjacentOffsets(5)).toEqual([6, 4]);
    });

    it("never includes a future day: the day GETs create the row and pin its plan (nfix #1)", () => {
        expect(adjacentOffsets(0)).toEqual([1]);
        expect(adjacentOffsets(-1)).toEqual([0]);
        expect(adjacentOffsets(-3)).toEqual([]);
    });
});

describe("prefetchAdjacentDays", () => {
    it("prefetches both neighbours through the given options factory", () => {
        const prefetchQuery = vi.fn().mockResolvedValue(undefined);
        const day = vi.fn((o: number) => ({ queryKey: ["d", o] }));
        prefetchAdjacentDays({ prefetchQuery }, 2, day);
        expect(prefetchQuery.mock.calls.map((c) => (c[0] as { queryKey: unknown }).queryKey)).toEqual([
            ["d", 3],
            ["d", 1],
        ]);
    });
});
