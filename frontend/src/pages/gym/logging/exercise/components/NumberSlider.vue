<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";

const TICK_WIDTH_PX = 44;

export type NumberSliderMarker = {
    value: number;
    label?: string;
    color?: "green" | "orange";
    labelPosition?: "top" | "bottom";
};

const props = withDefaults(
    defineProps<{
        label: string;
        hint?: string;
        error?: string;
        step?: number;
        min?: number;
        max?: number;
        integerOnly?: boolean;
        markerValue?: number;
        markerLabel?: string;
        markers?: NumberSliderMarker[];
    }>(),
    { step: 1, min: 0, max: 100, integerOnly: false },
);

const model = defineModel<number>({ required: true });

const scrollEl = ref<HTMLElement | null>(null);
const padPx = ref(0);
const scrollLeft = ref(0);
const syncingScroll = ref(false);
const scrollDrivenUpdate = ref(false);

const displayError = computed(() => props.error ?? "");

const hasHintSlot = computed(() => props.hint !== undefined);
const hintVisible = computed(() => Boolean(props.hint?.trim()));

const tickValues = computed(() => {
    const step = props.step > 0 ? props.step : 1;
    const count = Math.max(0, Math.round((props.max - props.min) / step));
    const values: number[] = [];
    for (let i = 0; i <= count; i++) {
        const raw = props.min + i * step;
        const v = props.integerOnly
            ? Math.round(raw)
            : Math.round(raw * 100) / 100;
        values.push(v);
    }
    return values;
});

const formatValue = (n: number) => {
    if (props.integerOnly || Number.isInteger(n)) {
        return String(Math.round(n));
    }
    const s = n.toFixed(1);
    return s.endsWith(".0") ? s.slice(0, -2) : s;
};

const displayValue = computed(() => formatValue(model.value || 0));

const valueToIndex = (value: number) => {
    const step = props.step > 0 ? props.step : 1;
    const idx = Math.round((value - props.min) / step);
    return Math.min(Math.max(idx, 0), tickValues.value.length - 1);
};

const resolvedMarkers = computed(() => {
    const list: NumberSliderMarker[] = [...(props.markers ?? [])];
    if (
        props.markerValue != null &&
        Number.isFinite(props.markerValue) &&
        !list.some((marker) => marker.value === props.markerValue)
    ) {
        list.push({
            value: props.markerValue,
            label: props.markerLabel,
        });
    }
    return list;
});

const visibleMarkers = computed(() => {
    const markers = [];
    for (const marker of resolvedMarkers.value) {
        if (!Number.isFinite(marker.value)) continue;
        if (marker.value < props.min || marker.value > props.max) continue;
        const index = valueToIndex(marker.value);
        markers.push({
            index,
            label: marker.label?.trim() ?? "",
            color: marker.color ?? "green",
            labelPosition: marker.labelPosition ?? "top",
        });
    }
    return markers.map((marker) => ({
        ...marker,
        leftPx:
            padPx.value +
            marker.index * TICK_WIDTH_PX +
            TICK_WIDTH_PX / 2,
    }));
});

const hasTopMarkerLabel = computed(() =>
    visibleMarkers.value.some(
        (marker) => marker.label && marker.labelPosition === "top",
    ),
);

const hasBottomMarkerLabel = computed(() =>
    visibleMarkers.value.some(
        (marker) => marker.label && marker.labelPosition === "bottom",
    ),
);

const indexToScrollLeft = (index: number) => index * TICK_WIDTH_PX;

// While we scroll the strip ourselves (a tap or a model change), scroll events
// must not write the value: a smooth scroll passes through other ticks and
// outlasts any single frame. The sync ends when the strip reaches its target,
// on `scrollend`, when the user grabs the strip, or after a fallback timeout.
const SYNC_FALLBACK_MS = 1000;
let syncTarget: number | null = null;
let syncTimer: ReturnType<typeof setTimeout> | undefined;

const endSync = () => {
    syncTarget = null;
    clearTimeout(syncTimer);
    syncingScroll.value = false;
    const el = scrollEl.value;
    if (el) scrollLeft.value = el.scrollLeft;
};

const scrollToIndex = (index: number, smooth = false) => {
    const el = scrollEl.value;
    if (!el) return;
    const target = Math.min(
        indexToScrollLeft(index),
        Math.max(0, el.scrollWidth - el.clientWidth),
    );
    syncTarget = target;
    syncingScroll.value = true;
    clearTimeout(syncTimer);
    el.scrollTo({ left: target, behavior: smooth ? "smooth" : "instant" });
    if (!smooth || Math.abs(el.scrollLeft - target) < 1) {
        requestAnimationFrame(endSync);
        return;
    }
    syncTimer = setTimeout(endSync, SYNC_FALLBACK_MS);
};

const onUserScrollStart = () => {
    if (syncTarget !== null) endSync();
};

const onScrollEnd = () => {
    if (syncTarget !== null) endSync();
};

const updatePad = () => {
    const el = scrollEl.value;
    if (!el) return;
    padPx.value = Math.max(0, el.clientWidth / 2 - TICK_WIDTH_PX / 2);
};

let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
    updatePad();
    const el = scrollEl.value;
    if (el && typeof ResizeObserver !== "undefined") {
        resizeObserver = new ResizeObserver(updatePad);
        resizeObserver.observe(el);
    }
    nextTick(() => {
        scrollToIndex(valueToIndex(model.value || 0));
        scrollLeft.value = scrollEl.value?.scrollLeft ?? 0;
    });
});

onUnmounted(() => {
    resizeObserver?.disconnect();
    clearTimeout(syncTimer);
});

// The side pads follow the strip's width. When the left pad grows, scroll anchoring moves
// scrollLeft with the content and snapping then lands a few ticks off, which a later scroll event
// would write as the value (180 -> 180.9). Re-pin to the model whenever the pad changes (resize,
// rotation). The pads are `transition-none`: the reduced-motion rule gives every element a 1ms
// transition, which delayed their first width by a frame and caused exactly this on load.
watch(padPx, () => {
    nextTick(() => scrollToIndex(valueToIndex(model.value || 0)));
});

watch(
    () => [props.min, props.max, props.step],
    () => {
        nextTick(() => scrollToIndex(valueToIndex(model.value || 0)));
    },
);

watch(
    () => model.value,
    (value) => {
        if (syncingScroll.value || scrollDrivenUpdate.value) return;
        nextTick(() => scrollToIndex(valueToIndex(value || 0)));
    },
);

const isMajorTick = (value: number, index: number) => {
    const step = props.step > 0 ? props.step : 1;
    if (props.integerOnly) {
        return index % 5 === 0;
    }
    if (step >= 2.5) {
        return value % 10 === 0;
    }
    if (step < 1) {
        return Number.isInteger(value);
    }
    return index % 5 === 0;
};

const tickLabel = (value: number, index: number) => {
    return isMajorTick(value, index) ? formatValue(value) : "";
};

const onScroll = () => {
    const el = scrollEl.value;
    if (!el) return;
    scrollLeft.value = el.scrollLeft;
    if (syncTarget !== null) {
        if (Math.abs(el.scrollLeft - syncTarget) < 1) endSync();
        return;
    }
    const index = Math.min(
        Math.max(Math.round(el.scrollLeft / TICK_WIDTH_PX), 0),
        tickValues.value.length - 1,
    );
    const next = tickValues.value[index] ?? props.min;
    if (next !== model.value) {
        scrollDrivenUpdate.value = true;
        model.value = next;
        requestAnimationFrame(() => {
            scrollDrivenUpdate.value = false;
        });
    }
};

const onTickClick = (index: number) => {
    scrollToIndex(index, true);
    const next = tickValues.value[index] ?? props.min;
    if (next !== model.value) {
        scrollDrivenUpdate.value = true;
        model.value = next;
        requestAnimationFrame(() => {
            scrollDrivenUpdate.value = false;
        });
    }
};
</script>

<template>
    <div class="flex flex-col gap-3">
        <label
            v-if="label.trim()"
            class="text-sm font-medium text-fg-muted"
            >{{ label }}</label
        >
        <p
            class="m-0 flex h-8 items-center justify-center text-display font-medium leading-none text-fg tabular-nums"
        >
            <span class="inline-block min-w-[5ch] text-center">{{
                displayValue
            }}</span>
        </p>
        <div
            :class="[
                hasTopMarkerLabel ? 'pt-4' : '',
                hasBottomMarkerLabel ? 'pb-4' : '',
            ]"
        >
            <div class="relative h-13">
                <div
                    v-for="marker in visibleMarkers"
                    :key="`${marker.index}-${marker.label}-${marker.labelPosition}`"
                    class="pointer-events-none absolute bottom-0 z-raised flex -translate-x-1/2 flex-col items-center justify-end gap-1"
                    :style="{
                        left: `${marker.leftPx - scrollLeft}px`,
                        width: `${TICK_WIDTH_PX}px`,
                    }"
                    aria-hidden="true"
                >
                    <span
                        v-if="
                            marker.label && marker.labelPosition === 'top'
                        "
                        class="text-2xs font-medium leading-none whitespace-nowrap"
                        :class="
                            marker.color === 'orange'
                                ? 'text-warn'
                                : 'text-success'
                        "
                    >
                        {{ marker.label }}
                    </span>
                    <span
                        class="block h-8 w-px rounded-full"
                        :class="
                            marker.color === 'orange'
                                ? 'bg-warn'
                                : 'bg-success'
                        "
                    />
                    <span class="h-3.5" aria-hidden="true">&nbsp;</span>
                    <span
                        v-if="
                            marker.label && marker.labelPosition === 'bottom'
                        "
                        class="absolute top-full mt-1 text-2xs font-medium leading-none whitespace-nowrap"
                        :class="
                            marker.color === 'orange'
                                ? 'text-warn'
                                : 'text-success'
                        "
                    >
                        {{ marker.label }}
                    </span>
                </div>
                <!-- Centre line after the markers: same z, so DOM order keeps it on top (a lined-up marker hides under it). -->
                <div
                    class="pointer-events-none absolute inset-y-0 left-1/2 z-raised w-0.5 -translate-x-1/2 rounded-full bg-accent"
                    aria-hidden="true"
                />
                <div
                    ref="scrollEl"
                    class="relative flex h-full touch-pan-x overflow-x-auto overflow-y-hidden overscroll-x-contain overscroll-y-none [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
                    style="scroll-snap-type: x mandatory"
                    @scroll.passive="onScroll"
                    @scrollend="onScrollEnd"
                    @pointerdown="onUserScrollStart"
                    @wheel.passive="onUserScrollStart"
                    @touchstart.passive="onUserScrollStart"
                >
                    <div
                        class="shrink-0 transition-none"
                        :style="{ width: `${padPx}px` }"
                        aria-hidden="true"
                    />
                    <!-- eslint-disable-next-line vue/no-restricted-html-elements -- slider tick, part of the drawn scale -->
                    <button
                        v-for="(tick, index) in tickValues"
                        :key="index"
                        type="button"
                        class="flex h-full shrink-0 touch-pan-x snap-center flex-col items-center justify-end gap-1 text-inherit focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent"
                        :style="{
                            width: `${TICK_WIDTH_PX}px`,
                            scrollSnapAlign: 'center',
                        }"
                        :aria-label="formatValue(tick)"
                        :aria-current="
                            index === valueToIndex(model || 0)
                                ? 'true'
                                : undefined
                        "
                        @click="onTickClick(index)"
                    >
                        <span
                            class="flex h-8 items-end justify-center"
                            aria-hidden="true"
                        >
                            <span
                                class="block w-px rounded-full bg-line-strong"
                                :class="
                                    index === valueToIndex(model || 0)
                                        ? 'h-8 bg-accent'
                                        : isMajorTick(tick, index)
                                          ? 'h-5'
                                          : 'h-3'
                                "
                            />
                        </span>
                        <span
                            class="h-3.5 text-2xs leading-none text-fg-subtle tabular-nums"
                            >{{ tickLabel(tick, index) || "\u00a0" }}</span
                        >
                    </button>
                    <div
                        class="shrink-0 transition-none"
                        :style="{ width: `${padPx}px` }"
                        aria-hidden="true"
                    />
                </div>
            </div>
        </div>
        <p
            v-if="hasHintSlot"
            class="m-0 -mt-1 min-h-5 text-xs leading-tight text-warn"
            :class="{ invisible: !hintVisible }"
            :aria-hidden="!hintVisible"
        >
            {{ hint?.trim() ? hint : "\u00a0" }}
        </p>
        <p
            v-if="displayError"
            class="m-0 text-sm leading-snug text-danger"
        >
            {{ displayError }}
        </p>
    </div>
</template>
