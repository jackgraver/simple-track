<script setup lang="ts" generic="T extends { key: string; isSelected: boolean }">
// A week of day tiles on a carousel track: the previous and next weeks sit just past either edge,
// so a swipe drags real days into view and a week change slides the whole row as one. One
// selected-day outline glides between columns, also across a week change. Pages render each tile
// through the `day` slot; selection styling comes from the outline, so tiles only need a
// transparent border and a hover state.
import { animate, motion, useMotionValue, type PanInfo } from "motion-v";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { carouselSpring, SWIPE_DISTANCE, SWIPE_VELOCITY } from "./motion";

const props = defineProps<{
    /** Seven days, Sunday first. `key` must be the day's `YYYY-MM-DD` date (it orders weeks). */
    days: T[];
    /** The weeks either side, shown while swiping and sliding. Without them that side is empty. */
    prevDays?: T[];
    nextDays?: T[];
    label: string;
}>();
const emit = defineEmits<{ prev: []; next: [] }>();
defineSlots<{ day(props: { day: T }): unknown }>();

const weekKey = computed(() => props.days[0]?.key ?? "");
const panels = computed(() =>
    [
        { pos: "prev" as const, days: props.prevDays },
        { pos: "current" as const, days: props.days },
        { pos: "next" as const, days: props.nextDays },
    ].filter((p): p is { pos: "prev" | "current" | "next"; days: T[] } => !!p.days?.length),
);
// Neighbours sit one week + the gap-1 between tiles away, so the row reads as continuous.
const panelClass = {
    prev: "absolute inset-y-0 right-full mr-1 w-full",
    current: "relative",
    next: "absolute inset-y-0 left-full ml-1 w-full",
};

const root = ref<HTMLElement | null>(null);
// motion components don't expose their element through a template ref, so find the track by attribute.
const trackEl = () => root.value?.querySelector<HTMLElement>("[data-track]") ?? null;
const GAP_PX = 4;
const step = () => (trackEl()?.offsetWidth ?? 0) + GAP_PX;

// The track's offset. When the week changes, the new current week is the panel that was beside
// it: shift the track by one week so nothing moves on screen (keeping any drag offset), then
// spring back to 0. Works the same after an arrow, a swipe, or a tap mid-slide.
const x = useMotionValue(0);
let settle: ReturnType<typeof animate> | null = null;
const slideTo0 = () => {
    settle?.stop();
    settle = animate(x, 0, carouselSpring);
};
watch(weekKey, (next, prev) => {
    if (!next || !prev || next === prev) return;
    x.jump(x.get() + (next > prev ? 1 : -1) * step());
    slideTo0();
});

// The outline follows the selected tile's layout box (offsets ignore transforms, so it targets
// where the tile lands, not where it is mid-slide). Measured, so any grid or gap works.
const outline = ref<{ x: number; width: number } | null>(null);
const measure = () => {
    const track = trackEl();
    const tile = track?.querySelector<HTMLElement>(`[data-week="${weekKey.value}"] [data-selected]`);
    outline.value = track && tile ? { x: track.offsetLeft + tile.offsetLeft, width: tile.offsetWidth } : null;
};
const selectedKey = computed(() => props.days.find((d) => d.isSelected)?.key ?? "");
watch([weekKey, selectedKey], () => void nextTick(measure));

let resize: ResizeObserver | null = null;
onMounted(() => {
    measure();
    resize = new ResizeObserver(measure);
    if (root.value) resize.observe(root.value);
});
onBeforeUnmount(() => {
    resize?.disconnect();
    settle?.stop();
    stopMirror?.();
});

// Swipe: the track follows the finger. Far or fast enough commits to the next week (the page
// changes `days`, and the watcher above slides from where the finger left it); otherwise, or if
// the page doesn't move, it springs back. A drag ends with a click on the tile under the finger;
// swallow it so a swipe never selects a day.
// While dragging, the outline rides along with its tile (outlineShift mirrors the track); after
// release it eases back to its column, which is the new week's selected day if the swipe committed.
const outlineShift = useMotionValue(0);
let stopMirror: (() => void) | null = null;
let dragged = false;
const onDragStart = () => {
    settle?.stop();
    dragged = true;
    stopMirror = x.on("change", (v) => outlineShift.set(v));
};
const onDragEnd = (_e: PointerEvent, info: PanInfo) => {
    setTimeout(() => (dragged = false));
    stopMirror?.();
    animate(outlineShift, 0, carouselSpring);
    const far = Math.abs(info.offset.x) > step() * SWIPE_DISTANCE;
    const fast = Math.abs(info.velocity.x) > SWIPE_VELOCITY;
    const dir = info.offset.x < 0 ? "next" : "prev";
    if (!(far || fast) || !(dir === "next" ? props.nextDays : props.prevDays)?.length) {
        slideTo0();
        return;
    }
    const from = weekKey.value;
    if (dir === "next") emit("next");
    else emit("prev");
    // The page navigates asynchronously; if the week never changes, don't leave the track stranded.
    setTimeout(() => {
        if (weekKey.value === from) slideTo0();
    }, 400);
};
const swallowClickAfterDrag = (e: MouseEvent) => {
    if (!dragged) return;
    e.preventDefault();
    e.stopPropagation();
};
</script>

<template>
    <!--
        The clip hides weeks past the edge; -mx-1/px-1 leave room for the tiles' focus rings.
        Tiles are usually links, which the browser would start dragging natively (taking the pointer
        away from the swipe), so native drags are cancelled.
    -->
    <div
        ref="root"
        class="relative -mx-1 overflow-x-clip px-1"
        @click.capture="swallowClickAfterDrag"
        @dragstart.prevent
    >
        <!-- Painted under the tiles: the track is positioned and comes later. The outer layer carries
             the drag offset, the inner one the column. -->
        <motion.div
            v-if="outline"
            aria-hidden="true"
            class="pointer-events-none absolute inset-0"
            :style="{ x: outlineShift }"
        >
            <motion.div
                class="absolute inset-y-0 left-0 rounded-control border border-accent bg-accent-soft"
                :initial="false"
                :animate="{ x: outline.x, width: outline.width }"
                :transition="carouselSpring"
            />
        </motion.div>
        <motion.div
            data-track
            class="relative touch-pan-y"
            :style="{ x }"
            drag="x"
            :drag-momentum="false"
            drag-direction-lock
            @drag-start="onDragStart"
            @drag-end="onDragEnd"
        >
            <div
                v-for="p in panels"
                :key="p.days[0]!.key"
                :data-week="p.days[0]!.key"
                :class="panelClass[p.pos]"
                :aria-hidden="p.pos === 'current' ? undefined : 'true'"
                :inert="p.pos !== 'current'"
            >
                <ol class="m-0 grid list-none grid-cols-7 gap-1 p-0" :aria-label="p.pos === 'current' ? label : undefined">
                    <li
                        v-for="d in p.days"
                        :key="d.key"
                        class="min-w-0"
                        :data-selected="(p.pos === 'current' && d.isSelected) || undefined"
                    >
                        <slot name="day" :day="d" />
                    </li>
                </ol>
            </div>
        </motion.div>
    </div>
</template>
