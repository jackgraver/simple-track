import { createSlideOver } from "~/shared/ui";

/** Matches the panel's `w-72 max-w-[85vw]`. */
export function drawerWidth(): number {
    return Math.min(288, window.innerWidth * 0.85);
}

// One drawer for the whole app (mounted in App.vue). Any header can open it, and the edge-swipe
// gesture (useDrawerGesture) drives it while a finger is down.
const panel = createSlideOver(drawerWidth);

export function useNavDrawer() {
    return { panel, open: panel.open, openDrawer: () => panel.settle(true) };
}
