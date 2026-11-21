<script setup lang="ts">
import { Lock, Unlock } from "lucide-vue-next";
import { computed } from "vue";
import { IconButton, Input } from "~/shared/ui";

const props = defineProps<{
    label: string;
    /** Short secondary text beside the label, e.g. "1.01 g/lb · 29%". */
    hint?: string;
    /** Unit shown after the number input ("g", "kcal"). */
    unit: string;
    grams: number;
    minGrams?: number;
    maxGrams: number;
    displayMaxGrams: number;
    locked?: boolean;
    hideLock?: boolean;
    /** Full token class for the filled part, e.g. `bg-macro-protein/70`. */
    trackBgClass: string;
    markerGrams?: number[];
}>();

const emit = defineEmits<{
    "update:grams": [v: number];
    toggleLock: [];
}>();

const minSafe = computed(() => Math.max(0, props.minGrams ?? 0));
const maxSafe = computed(() => Math.max(minSafe.value, props.maxGrams, props.grams, 0));
const displayMax = computed(() => Math.max(1, props.displayMaxGrams));
const trackFillPct = computed(() => Math.min(100, Math.max(0, (props.grams / displayMax.value) * 100)));
const clampGrams = (v: number) => Math.min(maxSafe.value, Math.max(minSafe.value, v));

function onRangeInput(e: Event) {
    if (props.locked) return;
    emit("update:grams", clampGrams(Number((e.target as HTMLInputElement).value) || 0));
}

const gramsModel = computed<number | string>({
    get: () => props.grams,
    set: (v) => {
        if (props.locked) return;
        emit("update:grams", clampGrams(Number(v) || 0));
    },
});

const markerStyles = computed(() => {
    const max = displayMax.value;
    return (props.markerGrams ?? []).filter((g) => g > 0 && g <= max).map((g) => ({ left: `${(g / max) * 100}%` }));
});
</script>

<template>
    <div class="flex flex-col gap-1">
        <div class="flex min-h-8 items-center justify-between gap-2">
            <div class="flex min-w-0 items-baseline gap-2">
                <span class="text-sm font-medium text-fg">{{ label }}</span>
                <span v-if="hint" class="truncate text-xs text-fg-muted tabular-nums">{{ hint }}</span>
            </div>
            <IconButton
                v-if="!hideLock"
                :label="`${locked ? 'Unlock' : 'Lock'} ${label}`"
                size="sm"
                :aria-pressed="!!locked"
                @click="emit('toggleLock')"
            >
                <Lock v-if="locked" class="size-4 text-accent" />
                <Unlock v-else class="size-4" />
            </IconButton>
        </div>
        <!-- A real range input sits invisibly on top (touch, keyboard, ARIA); the visible track,
             markers and thumb are drawn beneath it. -->
        <div class="relative h-touch lg:h-6">
            <!-- eslint-disable-next-line vue/no-restricted-html-elements -- invisible native range input over the drawn track and markers -->
            <input
                type="range"
                class="peer absolute inset-0 z-raised m-0 h-full w-full appearance-none bg-transparent opacity-0"
                :class="locked ? 'cursor-not-allowed' : 'cursor-pointer'"
                :aria-label="label"
                min="0"
                :max="displayMax"
                step="1"
                :value="grams"
                :disabled="locked"
                @input="onRangeInput"
            />
            <div class="pointer-events-none absolute inset-x-0 top-1/2 h-2 -translate-y-1/2 rounded-full bg-surface-3"></div>
            <div
                class="pointer-events-none absolute top-1/2 left-0 h-2 -translate-y-1/2 rounded-l-full"
                :class="trackBgClass"
                :style="{ width: `${trackFillPct}%` }"
            ></div>
            <span
                v-for="(m, i) in markerStyles"
                :key="i"
                class="pointer-events-none absolute top-1/2 h-4 w-0.5 -translate-x-1/2 -translate-y-1/2 bg-fg/80"
                :style="{ left: m.left }"
            ></span>
            <span
                class="pointer-events-none absolute top-1/2 size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-line-strong bg-fg peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-accent"
                :class="locked ? 'opacity-60' : ''"
                :style="{ left: `${trackFillPct}%` }"
            ></span>
        </div>
        <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
            <div class="flex items-center gap-2">
                <div class="w-24">
                    <Input
                        v-model="gramsModel"
                        type="number"
                        inputmode="decimal"
                        :disabled="locked"
                        :aria-label="`${label} (${unit})`"
                    />
                </div>
                <span class="text-sm text-fg-muted">{{ unit }}</span>
            </div>
            <slot />
        </div>
    </div>
</template>
