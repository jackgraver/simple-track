import { beforeEach, describe, expect, it, vi } from "vitest";
import { LAST_ROUTE_MAX_AGE_MS, loadLastRoute, saveLastRoute } from "./lastRoute";

const store = new Map<string, string>();
vi.stubGlobal("localStorage", {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    clear: () => store.clear(),
});

let cookie = "";
vi.stubGlobal("document", {
    get cookie() {
        return cookie;
    },
    set cookie(v: string) {
        cookie = v.split(";")[0] ?? "";
    },
});
vi.stubGlobal("location", { protocol: "https:" });

describe("lastRoute", () => {
    beforeEach(() => {
        store.clear();
        cookie = "";
    });

    it("falls back to the cookie when localStorage is empty", () => {
        saveLastRoute("/finances", 1000);
        store.clear();
        expect(loadLastRoute(2000)?.path).toBe("/finances");
    });

    it("round-trips a path with query", () => {
        saveLastRoute("/gym/plans?x=1", 1000);
        expect(loadLastRoute(2000)?.path).toBe("/gym/plans?x=1");
    });
    it("expires after a week", () => {
        saveLastRoute("/diet", 0);
        expect(loadLastRoute(LAST_ROUTE_MAX_AGE_MS + 1)).toBeNull();
        expect(loadLastRoute(LAST_ROUTE_MAX_AGE_MS)?.path).toBe("/diet");
    });
    it("rejects garbage and protocol-relative paths", () => {
        localStorage.setItem("st:last-route", "nope");
        expect(loadLastRoute()).toBeNull();
        localStorage.setItem("st:last-route", JSON.stringify({ path: "//evil.com", at: Date.now() }));
        expect(loadLastRoute()).toBeNull();
    });
});
