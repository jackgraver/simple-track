<script setup lang="ts">
// Draws the hover and active backgrounds of a list of links as two pills that slide between items,
// instead of each item painting its own. Items are the descendants marked `data-highlight`; the
// active one is the item matching `active` (aria-current by default). Items need `relative` so they
// paint above the pills, and no background of their own.
//
// The hover pill follows a mouse (not touch: there's no hover there); it slides between items,
// across small gaps too, and fades when the pointer leaves the list. Pressing an item with a mouse
// shrinks it and its pill a touch, for a click feel. Timings are tunable (motion.ts, `?tune`).
import { animate, motion, useMotionValue } from "motion-v";
import { onBeforeUnmount, onMounted, ref } from "vue";
import { highlight, highlightSpring, pressTransition, reduceMotion } from "./motion";

const props = withDefaults(defineProps<{ as?: string; active?: string }>(), {
    as: "div",
    active: '[aria-current="page"]',
});

const root = ref<HTMLElement | null>(null);

function pill() {
    const style = {
        x: useMotionValue(0),
        y: useMotionValue(0),
        width: useMotionValue(0),
        height: useMotionValue(0),
        opacity: useMotionValue(0),
        scale: useMotionValue(1),
    };
    return { style, item: null as HTMLElement | null };
}
const activePill = pill();
const hoverPill = pill();

/** The item's layout box within the root (offsets ignore the press scale). */
function boxOf(el: HTMLElement) {
    let x = 0;
    let y = 0;
    for (let n: HTMLElement | null = el; n && n !== root.value; n = n.offsetParent as HTMLElement | null) {
        x += n.offsetLeft;
        y += n.offsetTop;
    }
    // The root itself is the offsetParent of the items only if it's positioned (it is: `relative`).
    return { x, y, width: el.offsetWidth, height: el.offsetHeight };
}

function moveTo(p: ReturnType<typeof pill>, el: HTMLElement, slide: boolean) {
    p.item = el;
    const box = boxOf(el);
    for (const k of ["x", "y", "width", "height"] as const) {
        if (slide) animate(p.style[k], box[k], highlightSpring());
        else p.style[k].jump(box[k]);
    }
}

// Active pill: follows aria-current (the router updates it), jumping on first show.
function syncActive(slide = true) {
    const el = root.value?.querySelector<HTMLElement>(`[data-highlight]${props.active}`) ?? null;
    if (!el) {
        activePill.item = null;
        activePill.style.opacity.jump(0);
        return;
    }
    const wasShown = activePill.item !== null;
    moveTo(activePill, el, slide && wasShown);
    activePill.style.opacity.jump(1);
}

// Hover pill.
let hideTimer: ReturnType<typeof setTimeout> | undefined;
const itemFrom = (e: Event) => (e.target as HTMLElement | null)?.closest<HTMLElement>("[data-highlight]") ?? null;
function onPointerOver(e: PointerEvent) {
    if (e.pointerType !== "mouse") return;
    const el = itemFrom(e);
    if (!el) {
        // Over a gap: hide after a beat, so crossing to the next group still slides.
        clearTimeout(hideTimer);
        hideTimer = setTimeout(hideHover, 120);
        return;
    }
    clearTimeout(hideTimer);
    const shown = hoverPill.item !== null;
    moveTo(hoverPill, el, shown);
    if (!shown) animate(hoverPill.style.opacity, 1, { duration: reduceMotion() ? 0 : 0.1 });
}
function hideHover() {
    clearTimeout(hideTimer);
    if (!hoverPill.item) return;
    hoverPill.item = null;
    animate(hoverPill.style.opacity, 0, { duration: reduceMotion() ? 0 : 0.12 });
}

// Press: shrink the item and whichever pills sit under it; release (anywhere) restores them.
let pressed: HTMLElement | null = null;
function onPointerDown(e: PointerEvent) {
    if (e.pointerType === "touch" || e.button !== 0) return;
    const el = itemFrom(e);
    if (!el) return;
    pressed = el;
    setPressed(highlight.pressScale);
    window.addEventListener("pointerup", release, { once: true });
    window.addEventListener("pointercancel", release, { once: true });
}
function setPressed(scale: number) {
    if (!pressed) return;
    animate(pressed, { scale }, pressTransition());
    for (const p of [hoverPill, activePill]) if (p.item === pressed) animate(p.style.scale, scale, pressTransition());
}
function release() {
    setPressed(1);
    pressed = null;
    // The click may have moved aria-current to another item: reset any pill left shrunk.
    for (const p of [hoverPill, activePill]) if (p.style.scale.get() !== 1) animate(p.style.scale, 1, pressTransition());
}

let mutations: MutationObserver | null = null;
let resize: ResizeObserver | null = null;
onMounted(() => {
    syncActive(false);
    mutations = new MutationObserver(() => syncActive());
    mutations.observe(root.value!, { subtree: true, attributes: true, attributeFilter: ["aria-current"], childList: true });
    resize = new ResizeObserver(() => {
        syncActive(false);
        if (hoverPill.item) moveTo(hoverPill, hoverPill.item, false);
    });
    resize.observe(root.value!);
});
onBeforeUnmount(() => {
    mutations?.disconnect();
    resize?.disconnect();
    clearTimeout(hideTimer);
    window.removeEventListener("pointerup", release);
    window.removeEventListener("pointercancel", release);
});
</script>

<template>
    <component
        :is="as"
        ref="root"
        class="relative"
        @pointerover="onPointerOver"
        @pointerleave="hideHover"
        @pointerdown="onPointerDown"
    >
        <motion.div
            aria-hidden="true"
            class="pointer-events-none absolute top-0 left-0 rounded-control bg-surface-2"
            :style="hoverPill.style"
        />
        <motion.div
            aria-hidden="true"
            class="pointer-events-none absolute top-0 left-0 rounded-control bg-surface-3"
            :style="activePill.style"
        />
        <slot />
    </component>
</template>
