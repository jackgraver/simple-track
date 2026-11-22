<script setup lang="ts">
const props = withDefaults(
    defineProps<{
        calories: number;
        protein: number;
        carbs: number;
        fat: number;
        fiber: number;
        fontSize?: string;
        decimals?: "two" | "none";
    }>(),
    { decimals: "two" },
);

function round2(n: number) {
    return Math.round((n + Number.EPSILON) * 100) / 100;
}

function fmt(n: number) {
    return props.decimals === "none" ? Math.round(n) : round2(n);
}
</script>

<template>
    <div class="mt-1 flex gap-2.5 font-medium">
        <span class="text-sm text-macro-calories" :style="{ fontSize: props.fontSize }">{{ fmt(props.calories) }}C</span>
        <span class="text-sm text-macro-protein" :style="{ fontSize: props.fontSize }">{{ fmt(props.protein) }}P</span>
        <span class="text-sm text-macro-carbs" :style="{ fontSize: props.fontSize }">{{ fmt(props.carbs) }}C</span>
        <span class="text-sm text-macro-fat" :style="{ fontSize: props.fontSize }">{{ fmt(props.fat) }}F</span>
        <span class="text-sm text-macro-fiber" :style="{ fontSize: props.fontSize }">{{ fmt(props.fiber) }}Fi</span>
    </div>
</template>
