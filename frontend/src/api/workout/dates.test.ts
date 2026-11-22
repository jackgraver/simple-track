import { describe, expect, it } from "vitest";
import { dateToOffset, offsetToDate } from "./dates";

describe("workout day dates", () => {
    const now = new Date(2026, 10, 19, 23, 59);

    it("maps offsets to local dates (positive = past)", () => {
        expect(offsetToDate(0, now)).toBe("2026-11-19");
        expect(offsetToDate(1, now)).toBe("2026-11-18");
        expect(offsetToDate(-1, now)).toBe("2026-11-20");
        expect(offsetToDate(19, now)).toBe("2026-10-31");
    });

    it("round-trips and tracks the clock, not the original offset", () => {
        for (const o of [-3, 0, 1, 40]) expect(dateToOffset(offsetToDate(o, now), now)).toBe(o);
        const tomorrow = new Date(2026, 10, 20, 0, 1);
        expect(dateToOffset("2026-11-19", tomorrow)).toBe(1);
    });

    it("crosses month and year boundaries", () => {
        expect(offsetToDate(1, new Date(2027, 0, 1))).toBe("2026-12-31");
        expect(dateToOffset("2026-12-31", new Date(2027, 0, 1))).toBe(1);
    });
});
