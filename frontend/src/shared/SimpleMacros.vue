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
    <div class="macros">
        <span class="macro calories" :style="{ fontSize: props.fontSize }"
            >{{ fmt(props.calories) }}C</span
        >
        <span class="macro protein" :style="{ fontSize: props.fontSize }"
            >{{ fmt(props.protein) }}P</span
        >
        <span class="macro carbs" :style="{ fontSize: props.fontSize }"
            >{{ fmt(props.carbs) }}C</span
        >
        <span class="macro fat" :style="{ fontSize: props.fontSize }"
            >{{ fmt(props.fat) }}F</span
        >
        <span class="macro fiber" :style="{ fontSize: props.fontSize }"
            >{{ fmt(props.fiber) }}Fi</span
        >
    </div>
</template>

<style scoped>
.macros {
    display: flex;
    gap: 0.6rem;
    font-weight: 500;
    margin-top: 0.25rem;
}

.macro {
    font-size: 0.9rem;
}

.calories {
    color: var(--macro-calories, #fb923c);
}
.protein {
    color: var(--macro-protein, #60a5fa);
}
.carbs {
    color: var(--macro-carbs, #f87171);
}
.fat {
    color: var(--macro-fat, #c084fc);
}
.fiber {
    color: var(--macro-fiber, #34d399);
}
</style>
