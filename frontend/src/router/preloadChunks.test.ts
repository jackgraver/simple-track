import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { defineComponent } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";

import { preloadRouteChunks } from "./preloadChunks";

const Loaded = defineComponent({ render: () => null });

function setup() {
    const lazyA = vi.fn(() => Promise.resolve({ default: Loaded }));
    const lazyB = vi.fn(() => Promise.resolve({ default: Loaded }));
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [
            { path: "/a", name: "a", component: lazyA },
            { path: "/b", name: "b", component: lazyB },
            { path: "/c", name: "c", component: Loaded },
        ],
    });
    return { router, lazyA, lazyB };
}

describe("preloadRouteChunks", () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => {
        vi.useRealTimers();
        vi.unstubAllGlobals();
    });

    it("loads only the named lazy routes, once idle", async () => {
        const { router, lazyA, lazyB } = setup();
        preloadRouteChunks(router, ["a", "c", "missing"]);
        expect(lazyA).not.toHaveBeenCalled();
        await vi.runAllTimersAsync();
        expect(lazyA).toHaveBeenCalledOnce();
        expect(lazyB).not.toHaveBeenCalled();
    });

    it("skips Save-Data connections", async () => {
        vi.stubGlobal("navigator", { connection: { saveData: true } });
        const { router, lazyA } = setup();
        preloadRouteChunks(router, ["a"]);
        await vi.runAllTimersAsync();
        expect(lazyA).not.toHaveBeenCalled();
    });

    it("swallows a failed chunk load", async () => {
        const { router } = setup();
        const failing = vi.fn(() => Promise.reject(new Error("offline")));
        router.addRoute({ path: "/d", name: "d", component: failing });
        preloadRouteChunks(router, ["d"]);
        await vi.runAllTimersAsync();
        expect(failing).toHaveBeenCalledOnce();
    });
});
