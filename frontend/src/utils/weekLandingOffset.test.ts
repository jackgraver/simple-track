import { describe, expect, it } from "vitest";
import { weekLandingOffset } from "./dateUtil";

// Offsets: 0 = today, positive = past. Say today is a Thursday, so this week's Sunday is 4.
describe("weekLandingOffset", () => {
    it("lands on next week's Sunday", () => expect(weekLandingOffset(4, 1)).toBe(-3));
    it("lands on last week's Sunday", () => expect(weekLandingOffset(4, -1)).toBe(11));
    it("lands on today when coming back to the current week", () => {
        expect(weekLandingOffset(-3, -1)).toBe(0);
        expect(weekLandingOffset(11, 1)).toBe(0);
    });
    it("lands on today when today is Sunday or Saturday of the new week", () => {
        expect(weekLandingOffset(7, 1)).toBe(0); // today is Sunday
        expect(weekLandingOffset(13, 1)).toBe(0); // today is Saturday
    });
});
