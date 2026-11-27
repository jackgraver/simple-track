import { beforeEach, describe, expect, it, vi } from "vitest";
import { isExpiredLoggingResume, LAST_ROUTE_MAX_AGE_MS, LOGGING_RESUME_WINDOW_MS, loadLastRoute, saveLastRoute } from "./lastRoute";
import { clearUserStorage, setUserStorageOwner, userKey } from "./userStorage";

const store = new Map<string, string>();
const fakeStorage = {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
    key: (i: number) => [...store.keys()][i] ?? null,
    get length() {
        return store.size;
    },
    clear: () => store.clear(),
};
vi.stubGlobal("localStorage", fakeStorage);
vi.stubGlobal("window", { localStorage: fakeStorage, sessionStorage: { ...fakeStorage, getItem: () => null, key: () => null, length: 0 } });

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
        clearUserStorage();
        setUserStorageOwner(1);
    });

    it("ignores another user's route, in storage and in the cookie", () => {
        saveLastRoute("/finances", 1000);
        setUserStorageOwner(2);
        expect(loadLastRoute(2000)).toBeNull();
    });

    it("returns nothing and saves nothing before the user is known", () => {
        saveLastRoute("/finances", 1000);
        clearUserStorage();
        expect(loadLastRoute(2000)).toBeNull();
        cookie = "";
        saveLastRoute("/diet", 1000);
        expect(store.size).toBe(0);
        expect(cookie).toBe("");
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
        const key = userKey("last-route") ?? "";
        localStorage.setItem(key, "nope");
        expect(loadLastRoute()).toBeNull();
        localStorage.setItem(key, JSON.stringify({ path: "//evil.com", at: Date.now(), u: 1 }));
        expect(loadLastRoute()).toBeNull();
    });

    describe("isExpiredLoggingResume", () => {
        const logging = "/gym/logging/12?offset=0";

        it("keeps a logging page resumable inside the window", () => {
            expect(isExpiredLoggingResume({ path: logging, at: 0 }, LOGGING_RESUME_WINDOW_MS)).toBe(false);
        });

        it("expires a logging page just past the window", () => {
            expect(isExpiredLoggingResume({ path: logging, at: 0 }, LOGGING_RESUME_WINDOW_MS + 1)).toBe(true);
        });

        it("never expires other pages", () => {
            expect(isExpiredLoggingResume({ path: "/gym/plans", at: 0 }, LOGGING_RESUME_WINDOW_MS + 1)).toBe(false);
        });
    });
});
