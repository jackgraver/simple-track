import { animate, motionValue } from "motion-v";
import { ref, watch } from "vue";
import { drawerSpring } from "./motion";

export type SlideOver = ReturnType<typeof createSlideOver>;

/**
 * State and motion for a panel that slides in from the left (Drawer.vue renders it). `x` is the
 * panel's offset in px: 0 = fully open, -width = hidden. Every move is a spring from wherever the
 * panel is, so it can be caught mid-slide and a released swipe carries its speed.
 *
 * - `open`: the logical state. Setting it (a button, Esc, a scrim tap) slides the panel there.
 * - `present`: mounted, i.e. open or still sliding out.
 * - `follow(px)`: a finger is driving the panel; `settle(open, velocity)` on release.
 */
export function createSlideOver(width: () => number) {
    const open = ref(false);
    const present = ref(false);
    const x = motionValue(0);
    let anim: ReturnType<typeof animate> | null = null;
    let target = false;

    function settle(toOpen: boolean, velocity = 0) {
        anim?.stop();
        target = toOpen;
        if (toOpen && !present.value) x.jump(-width());
        if (toOpen) present.value = true;
        open.value = toOpen;
        if (!toOpen && x.get() <= -width() + 1) present.value = false; // already off screen
        else anim = animate(x, toOpen ? 0 : -width(), drawerSpring(velocity));
    }

    // Unmount once the panel is off screen, without waiting for the spring's last sub-pixel tail:
    // until then Reka keeps the page modal (no taps reach it).
    x.on("change", (v) => {
        if (!open.value && present.value && v <= -width() + 1) present.value = false;
    });

    function follow(px: number) {
        anim?.stop();
        anim = null;
        target = true;
        present.value = true;
        open.value = true;
        x.jump(Math.min(0, Math.max(-width(), px)));
    }

    // Changes that didn't come through settle/follow (v-model, `open.value = false`).
    watch(open, (v) => v !== target && settle(v));

    return { open, present, x, width, settle, follow };
}
