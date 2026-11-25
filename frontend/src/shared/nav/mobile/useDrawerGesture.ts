import { onScopeDispose, type Ref } from "vue";
import { drawerWidth, useNavDrawer } from "./useNavDrawer";

const EDGE = 24; // px from the left edge where a swipe can start
const LOCK = 8; // px of travel before we decide the gesture is horizontal

/**
 * Swipe from the left edge to reveal the nav drawer (it follows the finger), swipe left anywhere
 * on the open drawer or its scrim to hide it. Touch only, below `lg` only.
 */
export function useDrawerGesture(isDesktop: Ref<boolean>) {
    const { open, dragX, dragging, openedBySwipe } = useNavDrawer();

    let tracking = false;
    let revealing = false; // started closed: the panel is being pulled out
    let locked = false;
    let startX = 0;
    let startY = 0;
    let lastX = 0;
    let lastT = 0;
    let velocity = 0; // px per ms, positive = rightwards

    function onStart(e: TouchEvent) {
        if (isDesktop.value || e.touches.length !== 1) return;
        const t = e.touches[0]!;
        if (open.value) {
            revealing = false;
        } else {
            // Edge only, and never under another overlay (sheet, dialog).
            if (t.clientX > EDGE || document.querySelector('[role="dialog"]')) return;
            revealing = true;
        }
        tracking = true;
        locked = false;
        startX = lastX = t.clientX;
        startY = t.clientY;
        lastT = e.timeStamp;
        velocity = 0;
    }

    function onMove(e: TouchEvent) {
        if (!tracking) return;
        const t = e.touches[0]!;
        const dx = t.clientX - startX;
        const dy = t.clientY - startY;
        if (!locked) {
            if (Math.abs(dx) < LOCK && Math.abs(dy) < LOCK) return;
            // Vertical (scrolling) or the wrong direction: not ours.
            if (Math.abs(dy) > Math.abs(dx) || (revealing ? dx < 0 : dx > 0)) {
                tracking = false;
                return;
            }
            locked = true;
            dragging.value = true;
            if (revealing) {
                openedBySwipe.value = true;
                open.value = true;
            }
        }
        if (e.cancelable) e.preventDefault(); // we own the gesture: no page scroll or overscroll-nav
        const dt = Math.max(1, e.timeStamp - lastT);
        velocity = (t.clientX - lastX) / dt;
        lastX = t.clientX;
        lastT = e.timeStamp;
        const w = drawerWidth();
        dragX.value = revealing ? Math.min(0, Math.max(-w, -w + t.clientX)) : Math.min(0, dx);
    }

    function onEnd(e: TouchEvent) {
        if (!tracking) return;
        tracking = false;
        if (!locked) return;
        dragging.value = false;
        const w = drawerWidth();
        const shown = w + dragX.value; // visible px of the panel
        // A finger that rested before lifting is not a fling: ignore the last move's speed.
        const flinging = e.timeStamp - lastT < 80 && Math.abs(velocity) > 0.4;
        const settleOpen = flinging ? velocity > 0 : shown > w * 0.5;
        if (settleOpen) dragX.value = 0;
        else open.value = false; // exit animation starts from the current offset
    }

    // Non-passive so preventDefault works once we've claimed a horizontal swipe.
    document.addEventListener("touchstart", onStart, { passive: true });
    document.addEventListener("touchmove", onMove, { passive: false });
    document.addEventListener("touchend", onEnd);
    document.addEventListener("touchcancel", onEnd);
    onScopeDispose(() => {
        document.removeEventListener("touchstart", onStart);
        document.removeEventListener("touchmove", onMove);
        document.removeEventListener("touchend", onEnd);
        document.removeEventListener("touchcancel", onEnd);
    });
}
