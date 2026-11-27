import { beforeEach, describe, expect, it, vi } from "vitest";
import {
    clearUserStorage,
    currentUserId,
    LAST_ROUTE_COOKIE,
    setUserStorageOwner,
    userKey,
} from "./userStorage";

function fake() {
    const m = new Map<string, string>();
    return {
        m,
        api: {
            getItem: (k: string) => m.get(k) ?? null,
            setItem: (k: string, v: string) => void m.set(k, v),
            removeItem: (k: string) => void m.delete(k),
            key: (i: number) => [...m.keys()][i] ?? null,
            get length() {
                return m.size;
            },
        },
    };
}

const local = fake();
const session = fake();
let cookie = "";
vi.stubGlobal("window", { localStorage: local.api, sessionStorage: session.api });
vi.stubGlobal("document", {
    get cookie() {
        return cookie;
    },
    set cookie(v: string) {
        if (v.includes("Max-Age=0")) cookie = "";
        else cookie = v.split(";")[0] ?? "";
    },
});

describe("userStorage", () => {
    beforeEach(() => {
        local.m.clear();
        session.m.clear();
        cookie = "";
        clearUserStorage();
    });

    it("namespaces keys by user ID and returns null until the ID is known", () => {
        expect(userKey("x")).toBeNull();
        setUserStorageOwner(7);
        expect(userKey("x")).toBe("st:u:7:x");
        expect(currentUserId.value).toBe(7);
    });

    it("clear removes scoped keys, legacy keys, the owner marker and the cookie", () => {
        setUserStorageOwner(1);
        local.m.set("st:u:1:last-route", "a");
        local.m.set("st:last-route", "a");
        local.m.set("gym-rest-timer:global", "a");
        local.m.set("simpletracker:workout:hidden-exercises:5", "a");
        session.m.set("st:u:1:diet:logmeal-draft:v2:log", "a");
        session.m.set("simpletracker:diet:logmeal-draft:v2:create", "a");
        session.m.set("gym-draft:v2:date:2026-01-01:exercise:3", "a");
        cookie = `${LAST_ROUTE_COOKIE}=x`;
        clearUserStorage();
        expect([...local.m.keys(), ...session.m.keys()]).toEqual([]);
        expect(cookie).toBe("");
        expect(currentUserId.value).toBeNull();
    });

    it("keeps device-wide preferences", () => {
        setUserStorageOwner(1);
        local.m.set("st:theme", "light");
        local.m.set("tune:values", "{}");
        session.m.set("st:chunk-reload-at", "1");
        session.m.set("tune", "1");
        clearUserStorage();
        expect(local.m.get("st:theme")).toBe("light");
        expect(local.m.get("tune:values")).toBe("{}");
        expect(session.m.get("st:chunk-reload-at")).toBe("1");
        expect(session.m.get("tune")).toBe("1");
    });

    it("keeps the same user's data across reloads", () => {
        setUserStorageOwner(1);
        local.m.set("st:u:1:x", "mine");
        setUserStorageOwner(1);
        expect(local.m.get("st:u:1:x")).toBe("mine");
    });

    it("clears the previous user's data when a different user signs in", () => {
        setUserStorageOwner(1);
        local.m.set("st:u:1:x", "alice");
        session.m.set("st:u:1:y", "alice");
        cookie = `${LAST_ROUTE_COOKIE}=x`;
        setUserStorageOwner(2);
        expect(local.m.has("st:u:1:x")).toBe(false);
        expect(session.m.size).toBe(0);
        expect(cookie).toBe("");
        expect(userKey("x")).toBe("st:u:2:x");
    });

    it("treats legacy data with no recorded owner as unowned and clears it", () => {
        local.m.set("gym-rest-timer:global", "alice");
        session.m.set("gym-draft:v2:date:d:exercise:1", "alice");
        setUserStorageOwner(2);
        expect(local.m.has("gym-rest-timer:global")).toBe(false);
        expect(session.m.size).toBe(0);
    });
});
