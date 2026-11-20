import { describe, expect, it } from "vitest";
import { dateKey, sessionStreak } from "./useGymWeek";

// Thursday 2026-11-19. Schedule: Mon (1), Wed (3), Fri (5).
const today = new Date(2026, 10, 19);
const mwf = new Set([1, 3, 5]);
const days = (...d: number[]) => new Set(d.map((n) => dateKey(new Date(2026, 10, n))));

describe("sessionStreak", () => {
    it("skips rest days and today", () => {
        // Fri 13, Mon 16, Wed 18 trained; Sat/Sun/Tue are rest days; today (Thu) not trained yet.
        expect(sessionStreak(days(13, 16, 18), mwf, today)).toBe(3);
    });
    it("breaks on a missed scheduled day", () => {
        // Mon 16 missed.
        expect(sessionStreak(days(13, 18), mwf, today)).toBe(1);
    });
    it("counts a bonus session on a rest day", () => {
        expect(sessionStreak(days(13, 15, 16, 18), mwf, today)).toBe(4);
    });
    it("without a schedule, a week with no session breaks it", () => {
        expect(sessionStreak(days(5, 18), new Set(), today)).toBe(1);
        expect(sessionStreak(days(12, 18), new Set(), today)).toBe(2);
    });
});
