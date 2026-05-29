<script setup lang="ts">
import { Lock, Unlock } from "lucide-vue-next";
import { computed } from "vue";

const props = defineProps<{
    label: string;
    grams: number;
    minGrams?: number;
    maxGrams: number;
    displayMaxGrams: number;
    pct?: number;
    locked?: boolean;
    readonly?: boolean;
    trackBgClass: string;
    markerGrams?: number[];
    hideLock?: boolean;
    statText?: string;
    statClass?: string;
}>();

const emit = defineEmits<{
    "update:grams": [v: number];
    toggleLock: [];
}>();

const minSafe = computed(() => Math.max(0, props.minGrams ?? 0));
const maxSafe = computed(() =>
    Math.max(minSafe.value, props.maxGrams, props.grams, 0),
);
const displayMax = computed(() => Math.max(1, props.displayMaxGrams));
const inputDisabled = computed(() => props.locked || props.readonly);
const trackFillPct = computed(() =>
    Math.min(100, Math.max(0, (props.grams / displayMax.value) * 100)),
);
const clampGrams = (v: number) =>
    Math.min(maxSafe.value, Math.max(minSafe.value, v));

function onRangeInput(e: Event) {
    if (inputDisabled.value) return;
    emit(
        "update:grams",
        clampGrams(Number((e.target as HTMLInputElement).value) || 0),
    );
}

function onNumberInput(e: Event) {
    if (inputDisabled.value) return;
    emit(
        "update:grams",
        clampGrams(Number((e.target as HTMLInputElement).value) || 0),
    );
}

const markerStyles = computed(() => {
    const max = displayMax.value;
    const markers = props.markerGrams ?? [];
    return markers
        .filter((g) => g > 0 && g <= max)
        .map((g) => ({
            left: `${(g / max) * 100}%`,
        }));
});
</script>

<template>
    <div class="flex flex-col gap-1">
        <div class="flex flex-wrap items-center justify-between gap-2">
            <span class="text-sm font-medium text-zinc-300">{{ label }}</span>
            <div class="flex items-center gap-2">
                <span
                    v-if="readonly"
                    class="rounded-full border border-zinc-700 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-zinc-500"
                >
                    Fixed
                </span>
                <span
                    v-if="statText !== undefined"
                    class="text-xs font-medium"
                    :class="statClass"
                    >{{ statText }}</span
                >
                <span v-else class="text-xs tabular-nums text-zinc-500"
                    >{{ Math.round((pct ?? 0) * 10) / 10 }}% kcal</span
                >
                <button
                    v-if="!hideLock"
                    type="button"
                    class="rounded p-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
                    :aria-pressed="locked"
                    @click="emit('toggleLock')"
                >
                    <Lock v-if="locked" class="size-4" />
                    <Unlock v-else class="size-4" />
                </button>
            </div>
        </div>
        <div class="relative h-5">
            <div
                class="pointer-events-none absolute inset-x-0 top-1/2 h-2 -translate-y-1/2 rounded-full bg-zinc-800"
            />
            <div
                class="pointer-events-none absolute left-0 top-1/2 h-2 -translate-y-1/2 rounded-l-full"
                :class="trackBgClass"
                :style="{ width: `${trackFillPct}%` }"
            />
            <span
                v-for="(m, i) in markerStyles"
                :key="i"
                class="pointer-events-none absolute top-1/2 h-4 w-0.5 -translate-x-1/2 -translate-y-1/2 bg-zinc-100/80"
                :style="{ left: m.left }"
            />
            <span
                class="pointer-events-none absolute top-1/2 z-20 size-3.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-zinc-600 bg-zinc-50"
                :class="inputDisabled ? 'opacity-60' : ''"
                :style="{ left: `${trackFillPct}%` }"
            />
            <input
                type="range"
                class="absolute inset-x-0 top-0 z-10 h-5 w-full appearance-none bg-transparent opacity-0"
                :class="inputDisabled ? 'cursor-not-allowed' : 'cursor-pointer'"
                min="0"
                :max="displayMax"
                step="1"
                :value="grams"
                :disabled="inputDisabled"
                @input="onRangeInput"
            />
        </div>
        <input
            type="number"
            :min="minSafe"
            :max="maxSafe"
            step="1"
            :value="grams"
            :disabled="inputDisabled"
            class="max-w-28 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 disabled:cursor-not-allowed disabled:opacity-60"
            @input="onNumberInput"
        />
    </div>
</template>

<style scoped>
.macro-slider-input::-webkit-slider-runnable-track {
    height: 8px;
    background: transparent;
}
.macro-slider-input::-webkit-slider-thumb {
    appearance: none;
    height: 14px;
    width: 14px;
    border-radius: 9999px;
    background: #fafafa;
    border: 2px solid #52525b;
}
.macro-slider-input::-moz-range-track,
.macro-slider-input::-moz-range-progress {
    height: 8px;
    background: transparent;
}
.macro-slider-input::-moz-range-thumb {
    height: 14px;
    width: 14px;
    border-radius: 9999px;
    background: #fafafa;
    border: 2px solid #52525b;
}
</style>
