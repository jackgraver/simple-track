import { describe, expect, it, vi } from "vitest";
import { adjacentOffsets, prefetchAdjacentDays, prefetchAppStart } from "./prefetchDays";

describe("adjacentOffsets", () => {
    it("is the day before and after (offsets grow into the past)", () => {
        expect(adjacentOffsets(0)).toEqual([1, -1]);
        expect(adjacentOffsets(5)).toEqual([6, 4]);
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

describe("prefetchAppStart", () => {
    it("prefetches today's diet and gym day", () => {
        const prefetchQuery = vi.fn().mockResolvedValue(undefined);
        prefetchAppStart({ prefetchQuery });
        const keys = prefetchQuery.mock.calls.map((c) => JSON.stringify((c[0] as { queryKey: unknown }).queryKey));
        expect(keys).toHaveLength(2);
        expect(keys[0]).not.toBe(keys[1]);
    });
});
