import { beforeEach, describe, expect, it, vi } from "vitest";

const m = new Map<string, string>();
const api = {
    getItem: (k: string) => m.get(k) ?? null,
    setItem: (k: string, v: string) => void m.set(k, v),
    removeItem: (k: string) => void m.delete(k),
    key: (i: number) => [...m.keys()][i] ?? null,
    get length() {
        return m.size;
    },
};
vi.stubGlobal("window", { localStorage: api, sessionStorage: { ...api, key: () => null, length: 0 } });
vi.stubGlobal("document", { cookie: "" });

import { currentUserId, clearUserStorage } from "~/utils/userStorage";
import { markAuthenticated, markUnauthorized } from "./session";

describe("session wiring", () => {
    beforeEach(() => {
        m.clear();
        clearUserStorage();
    });

    it("publishes the user ID on sign-in and forgets it when unauthorized", () => {
        markAuthenticated("alice", 1);
        expect(currentUserId.value).toBe(1);
        markUnauthorized();
        expect(currentUserId.value).toBeNull();
    });

    it("a different user signing in does not inherit the previous user's data", () => {
        markAuthenticated("alice", 1);
        m.set("st:u:1:draft", "alice");
        markUnauthorized();
        markAuthenticated("bob", 2);
        expect(m.has("st:u:1:draft")).toBe(false);
    });

    it("never exposes a user ID when login did not return one", () => {
        markAuthenticated("alice", 1);
        markAuthenticated("alice");
        expect(currentUserId.value).toBeNull();
    });
});
