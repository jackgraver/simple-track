<script lang="ts" setup>
import type { MealPlanDay } from "~/types/models";
import { computed } from "vue";
import { useApiFetch } from "~/composables/useApiFetch";

const { data, pending, error } = useApiFetch<{
    date: string;
    days: MealPlanDay[];
}>("mealplan/today");

const expectedMeals = computed(() => {
    if (!data.value?.days || data.value.days.length === 0) return [];
    return (
        data?.value?.days[0]?.meals?.filter(
            (m: any) => m.status === "expected",
        ) || []
    );
});
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
        <button v-for="meal in expectedMeals" :key="meal.ID" @click="$emit('set-meal', meal)">
            {{ meal.meal.name }}
        </button>
    </div>
</template>
