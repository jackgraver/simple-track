<script lang="ts" setup>
import type { Meal, MealItem, MealPlanDay } from "~/types/models";
import { computed } from "vue";
import { useApiFetch } from "~/composables/useApiFetch";

const props = defineProps<{
    confirmMeal: (meal: Meal) => void;
}>();

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

const confirm = ref(-1);

const emit = defineEmits<{ (e: "selectMeal", meal: Meal): void }>();

function submitMeal(meal: Meal) {
    confirm.value = meal.ID;
    emit("selectMeal", meal);
}
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
        <button
            v-for="dayMeal in expectedMeals"
            :key="dayMeal.ID"
            @click="
                confirm === dayMeal.meal.ID
                    ? confirmMeal(dayMeal.meal)
                    : submitMeal(dayMeal.meal)
            "
            :class="confirm === dayMeal.meal.ID ? 'confirm-meal' : ''"
        >
            {{ dayMeal.meal.name }}
        </button>
    </div>
</template>

<style scoped>
button.confirm-meal {
    background-color: rgb(84, 231, 39);
    color: white;
}
</style>
