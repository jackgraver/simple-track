<script setup lang="ts">
// Protein, carbs, fat and fiber as coloured grams. Calories are shown separately by the caller,
// so "C" here always means carbs. `long` spells the names out ("72 g protein").
import { computed } from "vue";
import type { Macros } from "../domain/mealCard";

const props = defineProps<{ macros: Macros; long?: boolean; hideFiber?: boolean }>();

const parts = computed(() =>
    [
        { key: "P", name: "protein", value: props.macros.protein, cls: "text-macro-protein" },
        { key: "C", name: "carbs", value: props.macros.carbs, cls: "text-macro-carbs" },
        { key: "F", name: "fat", value: props.macros.fat, cls: "text-macro-fat" },
        { key: "Fi", name: "fiber", value: props.macros.fiber, cls: "text-macro-fiber" },
    ].filter((p) => !(props.hideFiber && p.key === "Fi")),
);
</script>

<template>
    <span class="flex flex-wrap items-baseline gap-x-3 gap-y-0.5 text-sm tabular-nums">
        <span v-for="p in parts" :key="p.key" class="whitespace-nowrap">
            <span class="font-medium" :class="p.cls">{{ Math.round(p.value) }}</span>
            <template v-if="long"><span class="text-fg-muted"> g {{ p.name }}</span></template>
            <template v-else>
                <span class="text-fg-muted" aria-hidden="true">{{ p.key }}</span>
                <span class="sr-only"> g {{ p.name }}</span>
            </template>
        </span>
    </span>
</template>
