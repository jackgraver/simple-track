import { nextTick, onBeforeUnmount, ref } from "vue";
import type { Ref } from "vue";

const PICKED_CHECK_MS = 1500;

/**
 * Tracks which list row was just picked so it can show a temporary check, and
 * announces the pick to screen readers. `scrollTarget` is the scroll container
 * holding rows marked with `data-key`.
 */
export function usePickedFlash(scrollTarget: Ref<HTMLElement | null>) {
    const pickedKey = ref<string | null>(null);
    const announcement = ref("");
    let timer: ReturnType<typeof setTimeout> | undefined;
    let unmounted = false;

    function flash(key: string, name: string) {
        pickedKey.value = key;
        clearTimeout(timer);
        timer = setTimeout(() => (pickedKey.value = null), PICKED_CHECK_MS);
        // Clear first so a repeat pick of the same row is announced again.
        announcement.value = "";
        nextTick(() => (announcement.value = name));
        nextTick(() => scrollTarget.value?.querySelector(`[data-key="${key}"]`)?.scrollIntoView({ block: "nearest" }));
    }

    onBeforeUnmount(() => {
        unmounted = true;
        clearTimeout(timer);
    });

    return { pickedKey, announcement, flash, isUnmounted: () => unmounted };
}
