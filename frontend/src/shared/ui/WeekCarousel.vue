<script setup lang="ts" generic="T extends { key: string; isSelected: boolean }">
// A week of day tiles that slides as one continuous row when the week changes, with one
// selected-day outline that glides between columns (also across a week change). Swiping the row
// sideways moves a week. Pages render each tile through the `day` slot; selection styling comes
// from the outline, so tiles only need a transparent border and a hover state.
import { AnimatePresence, motion, type PanInfo } from "motion-v";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { carouselSpring, SWIPE_DISTANCE, SWIPE_VELOCITY } from "./motion";

const props = defineProps<{
    /** Seven days, Sunday first. `key` must be the day's `YYYY-MM-DD` date (it orders weeks). */
    days: T[];
    label: string;
}>();
const emit = defineEmits<{ prev: []; next: [] }>();
defineSlots<{ day(props: { day: T }): unknown }>();

const weekKey = computed(() => props.days[0]?.key ?? "");

// 1: a later week comes in from the right. Read by the exiting week too (AnimatePresence `custom`),
// so a week leaves the way the new one pushes it.
const direction = ref<1 | -1>(1);
watch(weekKey, (next, prev) => {
    if (next && prev) direction.value = next > prev ? 1 : -1;
});

// Offsets in px from the measured strip, so the weeks stay exactly one week + gap apart.
const root = ref<HTMLElement | null>(null);
const stripWidth = ref(0);
const GAP_PX = 4; // gap-1 between tiles, also between weeks
// Where the outgoing week is on screen right now: dragged by a swipe, or mid-slide after quick
// taps. The new week starts one week behind it, so the row stays continuous with no gap.
const outgoingX = () => {
    const r = root.value;
    const old = r && [...r.querySelectorAll<HTMLElement>("[data-week]")].find((el) => el.dataset.week !== weekKey.value);
    if (!r || !old) return 0;
    return old.getBoundingClientRect().left - r.getBoundingClientRect().left - old.offsetLeft;
};
const variants = {
    enter: (dir: 1 | -1) => ({ x: outgoingX() + dir * (stripWidth.value + GAP_PX) }),
    center: { x: 0 },
    exit: (dir: 1 | -1) => ({ x: -dir * (stripWidth.value + GAP_PX) }),
};

// The outline follows the selected tile's layout box (offsets ignore transforms, so the entering
// week reports where its tile will land). Measured, not computed, so any grid or gap works.
const outline = ref<{ x: number; width: number } | null>(null);
const measure = () => {
    const week = root.value?.querySelector<HTMLElement>(`[data-week="${weekKey.value}"]`);
    const tile = week?.querySelector<HTMLElement>("[data-selected]");
    stripWidth.value = week?.offsetWidth ?? stripWidth.value;
    outline.value = week && tile ? { x: week.offsetLeft + tile.offsetLeft, width: tile.offsetWidth } : null;
};
const selectedKey = computed(() => props.days.find((d) => d.isSelected)?.key ?? "");
watch([weekKey, selectedKey], () => void nextTick(measure));

let resize: ResizeObserver | null = null;
onMounted(() => {
    measure();
    resize = new ResizeObserver(measure);
    if (root.value) resize.observe(root.value);
});
onBeforeUnmount(() => resize?.disconnect());

// Swipe: the row follows the finger and snaps back unless the swipe is far or fast enough.
// A drag ends with a click on the tile under the finger; swallow it so a swipe never selects a day.
let dragged = false;
const onDragStart = () => {
    dragged = true;
};
const onDragEnd = (_e: PointerEvent, info: PanInfo) => {
    const far = Math.abs(info.offset.x) > stripWidth.value * SWIPE_DISTANCE;
    const fast = Math.abs(info.velocity.x) > SWIPE_VELOCITY;
    if (far || fast) {
        if (info.offset.x < 0) emit("next");
        else emit("prev");
    }
    setTimeout(() => (dragged = false));
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
        class="relative -mx-1 grid overflow-x-clip px-1"
        @click.capture="swallowClickAfterDrag"
        @dragstart.prevent
    >
        <!-- Painted under the tiles: the week rows are positioned and come later. -->
        <motion.div
            v-if="outline"
            aria-hidden="true"
            class="pointer-events-none absolute inset-y-0 left-0 rounded-control border border-accent bg-accent-soft"
            :initial="false"
            :animate="{ x: outline.x, width: outline.width }"
            :transition="carouselSpring"
        />
        <AnimatePresence :initial="false" :custom="direction">
            <motion.div
                :key="weekKey"
                :data-week="weekKey"
                class="relative col-start-1 row-start-1 touch-pan-y"
                :custom="direction"
                :variants="variants"
                initial="enter"
                animate="center"
                exit="exit"
                :transition="carouselSpring"
                drag="x"
                drag-snap-to-origin
                drag-direction-lock
                @drag-start="onDragStart"
                @drag-end="onDragEnd"
            >
                <ol class="m-0 grid list-none grid-cols-7 gap-1 p-0" :aria-label="label">
                    <li v-for="d in days" :key="d.key" class="min-w-0" :data-selected="d.isSelected || undefined">
                        <slot name="day" :day="d" />
                    </li>
                </ol>
            </motion.div>
        </AnimatePresence>
    </div>
</template>
