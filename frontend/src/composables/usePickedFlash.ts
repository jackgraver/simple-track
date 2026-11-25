import { nextTick, onBeforeUnmount, ref } from "vue";
import type { Ref } from "vue";

const PICKED_CHECK_MS = 1500;

export function usePickedFlash(scrollContainer: Ref<HTMLElement | null>) {
    const pickedKey = ref("");
    const ariaAnnouncement = ref("");
    let clearPickedTimer: ReturnType<typeof setTimeout>;
    let unmounted = false;

    function resetAnnouncementThen(name: string) {
        ariaAnnouncement.value = "";
        nextTick(() => (ariaAnnouncement.value = name));
    }

    function scrollRowIntoView(key: string) {
        nextTick(() => scrollContainer.value?.querySelector(`[data-key="${key}"]`)?.scrollIntoView({ block: "nearest" }));
    }

    function flash(key: string, name: string) {
        pickedKey.value = key;
        clearTimeout(clearPickedTimer);
        clearPickedTimer = setTimeout(() => (pickedKey.value = ""), PICKED_CHECK_MS);
        resetAnnouncementThen(name);
        scrollRowIntoView(key);
    }

    onBeforeUnmount(() => {
        unmounted = true;
        clearTimeout(clearPickedTimer);
    });

    return { isPicked: (key: string) => pickedKey.value === key, ariaAnnouncement, flash, isUnmounted: () => unmounted };
}
