import { describe, expect, it } from "vitest";

import type { MediaItem } from "./api";
import { patchItem, patchList, replaceInList, revertItem, revertList } from "./cache";

const item = (ID: number, over: Partial<MediaItem> = {}): MediaItem => ({
    ID,
    CreatedAt: "",
    UpdatedAt: "",
    type: "movie",
    title: `t${ID}`,
    creator: "",
    release_year: null,
    genres: "",
    cover_url: "",
    synopsis: "",
    status: "planned",
    rating: null,
    review: "",
    favorite: false,
    progress: null,
    length: null,
    started_at: null,
    finished_at: null,
    times_consumed: 0,
    source: "",
    external_id: "",
    ...over,
});

describe("media cache patches", () => {
    it("patches only the given fields, ignoring undefined", () => {
        const out = patchItem(item(1, { rating: 3 }), { status: "completed", favorite: undefined });
        expect(out.status).toBe("completed");
        expect(out.rating).toBe(3);
        expect(out.favorite).toBe(false);
    });

    it("can clear a rating with null", () => {
        expect(patchItem(item(1, { rating: 4 }), { rating: null }).rating).toBeNull();
    });

    it("patchList changes one row and keeps others by identity", () => {
        const a = item(1);
        const b = item(2);
        const out = patchList([a, b], 2, { favorite: true });
        expect(out[0]).toBe(a);
        expect(out[1].favorite).toBe(true);
        expect(patchList([a], 9, { favorite: true })[0]).toBe(a);
    });

    it("revertItem restores only our fields, keeping a sibling's change", () => {
        const snapshot = item(1, { status: "planned", favorite: false });
        // we set status; a sibling later set favorite
        const current = item(1, { status: "completed", favorite: true });
        const out = revertItem(current, { status: "completed" }, snapshot);
        expect(out.status).toBe("planned");
        expect(out.favorite).toBe(true);
    });

    it("revertList reverts the row from the snapshot list", () => {
        const snapshot = [item(1, { rating: 2 }), item(2)];
        const current = patchList(snapshot, 1, { rating: 5 });
        expect(revertList(current, 1, { rating: 5 }, snapshot)[0].rating).toBe(2);
        expect(revertList(current, 7, { rating: 5 }, snapshot)).toBe(current);
    });

    it("replaceInList swaps in the server item", () => {
        const server = item(1, { status: "completed", finished_at: "2026-11-22" });
        expect(replaceInList([item(1), item(2)], server)[0]).toBe(server);
    });
});
