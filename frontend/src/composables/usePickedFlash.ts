import { nextTick, onBeforeUnmount, ref } from "vue";
import type { Ref } from "vue";

const PICKED_CHECK_MS = 1500;

export function usePickedFlash(scrollContainer: Ref<HTMLElement | null>) {
    const pickedKey = ref<string | null>(null);
    const announcement = ref("");
    let clearPickedTimer: ReturnType<typeof setTimeout> | undefined;
    let unmounted = false;

    function resetAnnouncementThen(name: string) {
        announcement.value = "";
        nextTick(() => (announcement.value = name));
    }

    function scrollRowIntoView(key: string) {
        nextTick(() => scrollContainer.value?.querySelector(`[data-key="${key}"]`)?.scrollIntoView({ block: "nearest" }));
    }

    function flash(key: string, name: string) {
        pickedKey.value = key;
        clearTimeout(clearPickedTimer);
        clearPickedTimer = setTimeout(() => (pickedKey.value = null), PICKED_CHECK_MS);
        resetAnnouncementThen(name);
        scrollRowIntoView(key);
    }

    onBeforeUnmount(() => {
        unmounted = true;
        clearTimeout(clearPickedTimer);
    });

    return { pickedKey, announcement, flash, isUnmounted: () => unmounted };
}
