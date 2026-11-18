import { beforeEach, describe, expect, it, vi } from "vitest";
import { LAST_ROUTE_MAX_AGE_MS, loadLastRoute, saveLastRoute } from "./lastRoute";

const store = new Map<string, string>();
vi.stubGlobal("localStorage", {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    clear: () => store.clear(),
});

describe("lastRoute", () => {
    beforeEach(() => store.clear());

    it("round-trips a path with query", () => {
        saveLastRoute("/gym/plans?x=1", 1000);
        expect(loadLastRoute(2000)).toBe("/gym/plans?x=1");
    });
    it("expires after a week", () => {
        saveLastRoute("/diet", 0);
        expect(loadLastRoute(LAST_ROUTE_MAX_AGE_MS + 1)).toBeNull();
        expect(loadLastRoute(LAST_ROUTE_MAX_AGE_MS)).toBe("/diet");
    });
    it("rejects garbage and protocol-relative paths", () => {
        localStorage.setItem("st:last-route", "nope");
        expect(loadLastRoute()).toBeNull();
        localStorage.setItem("st:last-route", JSON.stringify({ path: "//evil.com", at: Date.now() }));
        expect(loadLastRoute()).toBeNull();
    });
});
