import { ref, watch } from "vue";

// One drawer for the whole app (mounted in App.vue). Any header can open it, and the edge-swipe
// gesture (useDrawerGesture) drives `dragX` while a finger is down.
const open = ref(false);
/** Panel offset from fully open, in px: 0 = open, -width = off screen. */
const dragX = ref(0);
const dragging = ref(false);
/** The current open was started by a swipe, so the panel is already in place: skip the slide-in. */
const openedBySwipe = ref(false);

/** Matches the panel's `w-72 max-w-[85vw]`. */
export function drawerWidth(): number {
    return Math.min(288, window.innerWidth * 0.85);
}

// Reset after the exit animation (150ms) so it can start from where the finger let go.
let resetTimer: ReturnType<typeof setTimeout> | undefined;
watch(open, (isOpen) => {
    clearTimeout(resetTimer);
    if (isOpen) return;
    resetTimer = setTimeout(() => {
        dragX.value = 0;
        dragging.value = false;
        openedBySwipe.value = false;
    }, 200);
});

export function useNavDrawer() {
    return {
        open,
        dragX,
        dragging,
        openedBySwipe,
        openDrawer: () => {
            openedBySwipe.value = false;
            open.value = true;
        },
    };
}
