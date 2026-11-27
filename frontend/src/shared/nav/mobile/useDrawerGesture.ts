import { onScopeDispose, type Ref } from "vue";
import { drawerTuning } from "~/shared/ui";
import { useNavDrawer } from "./useNavDrawer";

const EDGE = 24; // px from the left edge where a swipe can start
const LOCK = 8; // px of travel before we decide the gesture is horizontal

/**
 * Swipe from the left edge to reveal the nav drawer (it follows the finger), swipe left anywhere
 * on the open drawer or its scrim to hide it. A panel that is still sliding can be caught by
 * touching it. On release it springs open or closed from where it is, carrying the finger's speed.
 * Touch only, below `lg` only.
 */
export function useDrawerGesture(isDesktop: Ref<boolean>) {
    const { panel } = useNavDrawer();

    let tracking = false;
    let locked = false;
    let fromClosed = false; // the panel was fully hidden: its edge tracks the finger
    let startOffset = 0;
    let startX = 0;
    let startY = 0;
    let lastX = 0;
    let lastT = 0;
    let velocity = 0; // px per s, positive = rightwards

    function onStart(e: TouchEvent) {
        if (isDesktop.value || e.touches.length !== 1) return;
        const t = e.touches[0]!;
        const w = panel.width();
        const shown = panel.present.value ? w + panel.x.get() : 0; // visible px of the panel
        if (!panel.open.value) {
            // Closed or sliding out: the edge, or the part of the panel still on screen. Never
            // under another overlay (sheet, dialog).
            if (t.clientX > Math.max(EDGE, shown)) return;
            if (!panel.present.value && document.querySelector('[role="dialog"]')) return;
        }
        tracking = true;
        locked = false;
        fromClosed = shown <= 0;
        startOffset = shown - w;
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
        const w = panel.width();
        if (!locked) {
            if (Math.abs(dx) < LOCK && Math.abs(dy) < LOCK) return;
            // Vertical (scrolling), or pushing a settled panel further than it goes: not ours.
            const pastOpen = dx > 0 && startOffset >= -0.5;
            const pastClosed = dx < 0 && startOffset <= -w + 0.5;
            if (Math.abs(dy) > Math.abs(dx) || pastOpen || pastClosed) {
                tracking = false;
                return;
            }
            locked = true;
        }
        if (e.cancelable) e.preventDefault(); // we own the gesture: no page scroll or overscroll-nav
        const dt = Math.max(1, e.timeStamp - lastT);
        velocity = ((t.clientX - lastX) / dt) * 1000;
        lastX = t.clientX;
        lastT = e.timeStamp;
        panel.follow(fromClosed ? -w + t.clientX : startOffset + dx);
    }

    function onEnd(e: TouchEvent) {
        if (!tracking) return;
        tracking = false;
        if (!locked) return;
        const w = panel.width();
        // A finger that rested before lifting is not a fling: ignore the last move's speed.
        const v = e.timeStamp - lastT < 80 ? velocity : 0;
        const flinging = Math.abs(v) > drawerTuning.flingVelocity;
        const settleOpen = flinging ? v > 0 : w + panel.x.get() > w * drawerTuning.openShare;
        panel.settle(settleOpen, v);
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
