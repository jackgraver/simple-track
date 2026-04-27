<script setup lang="ts">
const props = defineProps<{
    calories: number;
    protein: number;
    carbs: number;
    fiber: number;
    fontSize?: string;
}>();

function round2(n: number) {
    return Math.round((n + Number.EPSILON) * 100) / 100;
}
</script>

<template>
    <div class="macros">
        <span class="macro calories" :style="{ fontSize: props.fontSize }"
            >{{ round2(props.calories) }}C</span
        >
        <span class="macro protein" :style="{ fontSize: props.fontSize }"
            >{{ round2(props.protein) }}P</span
        >
        <span class="macro carbs" :style="{ fontSize: props.fontSize }"
            >{{ round2(props.carbs) }}C</span
        >
        <span class="macro fiber" :style="{ fontSize: props.fontSize }"
            >{{ round2(props.fiber) }}F</span
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
    color: #facc15;
}
.protein {
    color: #60a5fa;
}
.carbs {
    color: red;
}
.fiber {
    color: #34d399;
}
</style>
