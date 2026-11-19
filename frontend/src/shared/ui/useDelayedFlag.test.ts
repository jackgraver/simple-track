import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, nextTick, ref } from "vue";
import { useDelayedFlag } from "./useDelayedFlag";

describe("useDelayedFlag", () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it("stays false for a load that finishes inside the delay", async () => {
        const pending = ref(true);
        const flag = effectScope().run(() => useDelayedFlag(pending, 150))!;
        vi.advanceTimersByTime(100);
        pending.value = false;
        await nextTick();
        vi.advanceTimersByTime(200);
        expect(flag.value).toBe(false);
    });

    it("turns true after the delay and false as soon as the source clears", async () => {
        const pending = ref(true);
        const flag = effectScope().run(() => useDelayedFlag(pending, 150))!;
        expect(flag.value).toBe(false);
        vi.advanceTimersByTime(150);
        expect(flag.value).toBe(true);
        pending.value = false;
        await nextTick();
        expect(flag.value).toBe(false);
    });
});
