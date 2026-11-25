<script setup lang="ts">
import { computed } from "vue";
import { plateStack, type PlateCounts } from "../domain/plates";

// One side of the bar, drawn: a sleeve with the plates on it, heaviest nearest the collar, sized
// by weight so 45s and 10s read apart at a glance. Decorative; the text next to it says the same.
const props = defineProps<{ plates: PlateCounts; size?: "sm" | "md" }>();

const HEIGHTS: Record<string, [sm: string, md: string]> = {
    "45": ["h-6", "h-10"],
    "35": ["h-5", "h-9"],
    "25": ["h-4", "h-8"],
    "10": ["h-3.5", "h-6"],
    "5": ["h-3", "h-5"],
    "2.5": ["h-2.5", "h-4"],
};

const stack = computed(() => {
    const md = props.size === "md";
    return plateStack(props.plates).map((lbs, i) => {
        const h = HEIGHTS[String(lbs)] ?? ["h-3", "h-5"];
        return { key: `${lbs}-${i}`, cls: md ? h[1] : h[0] };
    });
});
</script>

<template>
    <span class="flex items-center" aria-hidden="true">
        <span class="rounded-control-inner bg-line-strong" :class="size === 'md' ? 'h-3 w-2' : 'h-2 w-1.5'" />
        <span class="flex items-center gap-px">
            <span
                v-for="p in stack"
                :key="p.key"
                class="rounded-control-inner bg-primary"
                :class="[p.cls, size === 'md' ? 'w-2' : 'w-1.5']"
            />
        </span>
        <span class="bg-line-strong" :class="size === 'md' ? 'h-1.5 w-6' : 'h-1 w-4'" />
    </span>
</template>
