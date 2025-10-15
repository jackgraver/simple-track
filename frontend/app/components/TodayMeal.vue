<script setup lang="ts">
import type { Day, Meal } from "~/types/models";

const { data, pending, error } = useApiFetch<{
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
}>(`mealplan/today`);

console.log("data?", data);
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
        <h1>Today</h1>
        <div
            v-if="data"
            class="macro-fill calories"
            :style="{
                width: `${Math.min(100, (data?.totalCalories ?? 0 / data?.day.plan.calories) * 100)}%`,
            }"
        >
            {{ data?.totalCalories ?? 0 }}
        </div>
        <div
            v-if="data"
            class="macro-fill protein"
            :style="{
                width: `${Math.min(100, (data?.totalProtein ?? 0 / data?.day.plan.protein) * 100)}%`,
            }"
        >
            {{ data?.totalProtein ?? 0 }}
        </div>
        <div
            v-if="data"
            class="macro-fill fiber"
            :style="{
                width: `${Math.min(100, (data?.totalFiber ?? 0 / data?.day.plan.fiber) * 100)}%`,
            }"
        >
            {{ data?.totalFiber ?? 0 }}
        </div>
    </div>
</template>

<style scoped>
.macro-fill {
    height: 100%;
    color: #ffffff;
    font-weight: bold;
    text-align: right;
    padding: 0 8px;
    white-space: nowrap;
    line-height: 20px;
    border-radius: 4px 0 0 4px;
}

.calories {
    background-color: orange;
}
.protein {
    background-color: blue;
}
.fiber {
    background-color: green;
}
</style>
