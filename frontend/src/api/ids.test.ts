import { afterEach, describe, expect, it, vi } from "vitest";

import { newClientId } from "./ids";

const V4 = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

afterEach(() => vi.unstubAllGlobals());

describe("newClientId", () => {
    it("uses crypto.randomUUID when available", () => {
        const randomUUID = vi.fn(() => "11111111-1111-4111-8111-111111111111");
        vi.stubGlobal("crypto", { randomUUID });
        expect(newClientId()).toBe("11111111-1111-4111-8111-111111111111");
        expect(randomUUID).toHaveBeenCalledOnce();
    });

    it("falls back to getRandomValues without randomUUID (plain-http LAN)", () => {
        const getRandomValues = vi.fn(<T,>(a: T) => {
            (a as Uint8Array).fill(0xff);
            return a;
        });
        vi.stubGlobal("crypto", { getRandomValues });
        const id = newClientId();
        expect(getRandomValues).toHaveBeenCalledOnce();
        expect(id).toMatch(V4);
        // Version and variant bits are forced even when every random byte is 0xff.
        expect(id).toBe("ffffffff-ffff-4fff-bfff-ffffffffffff");
    });

    it("falls back to Math.random without crypto at all", () => {
        vi.stubGlobal("crypto", undefined);
        const ids = new Set(Array.from({ length: 200 }, newClientId));
        expect(ids.size).toBe(200);
        for (const id of ids) expect(id).toMatch(V4);
    });

    it("returns valid unique v4 ids with the real crypto", () => {
        const a = newClientId();
        const b = newClientId();
        expect(a).toMatch(V4);
        expect(a).not.toBe(b);
    });
});
