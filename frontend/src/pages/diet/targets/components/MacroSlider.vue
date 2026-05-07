<script setup lang="ts">
import { Lock, Unlock } from 'lucide-vue-next';
import { computed } from 'vue';

const props = defineProps<{
    label: string;
    grams: number;
    maxGrams: number;
    pct: number;
    locked: boolean;
    trackBgClass: string;
    markerGrams?: number[];
}>();

const emit = defineEmits<{
    'update:grams': [v: number];
    toggleLock: [];
}>();

const maxSafe = computed(() => Math.max(1, props.maxGrams));

function onRangeInput(e: Event) {
    emit('update:grams', Number((e.target as HTMLInputElement).value) || 0);
}

function onNumberInput(e: Event) {
    emit('update:grams', Number((e.target as HTMLInputElement).value) || 0);
}

const markerStyles = computed(() => {
    const max = maxSafe.value;
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
                <span class="text-xs tabular-nums text-zinc-500"
                    >{{ Math.round(pct * 10) / 10 }}% kcal</span
                >
                <button
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
        <div class="relative pt-1 pb-1">
            <div
                class="pointer-events-none absolute inset-x-0 top-2 h-2 rounded-full bg-zinc-800"
            />
            <div
                class="pointer-events-none absolute top-2 h-2 rounded-l-full"
                :class="trackBgClass"
                :style="{ width: `${(grams / maxSafe) * 100}%` }"
            />
            <span
                v-for="(m, i) in markerStyles"
                :key="i"
                class="pointer-events-none absolute top-1 h-4 w-0.5 -translate-x-1/2 bg-zinc-100/80"
                :style="{ left: m.left }"
            />
            <input
                type="range"
                class="macro-slider-input relative z-10 h-2 w-full cursor-pointer appearance-none bg-transparent"
                :min="0"
                :max="maxSafe"
                step="0.1"
                :value="grams"
                @input="onRangeInput"
            />
        </div>
        <input
            type="number"
            min="0"
            :max="maxGrams"
            step="0.1"
            :value="grams"
            class="max-w-28 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1 text-sm text-zinc-100"
            @input="onNumberInput"
        />
    </div>
</template>

<style scoped>
.macro-slider-input::-webkit-slider-thumb {
    appearance: none;
    height: 14px;
    width: 14px;
    border-radius: 9999px;
    background: #fafafa;
    border: 2px solid #52525b;
}
.macro-slider-input::-moz-range-thumb {
    height: 14px;
    width: 14px;
    border-radius: 9999px;
    background: #fafafa;
    border: 2px solid #52525b;
}
</style>
