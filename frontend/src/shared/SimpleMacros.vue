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
    color: var(--color-macro-calories);
}
.protein {
    color: var(--color-macro-protein);
}
.carbs {
    color: var(--color-macro-carbs);
}
.fat {
    color: var(--color-macro-fat);
}
.fiber {
    color: var(--color-macro-fiber);
}
</style>
