<script lang="ts" setup>
import type { Meal, MealItem, Day } from "~/types/models";
import { computed } from "vue";
import { useApiFetch } from "~/composables/useApiFetch";

const props = defineProps<{
    ID: number;
    name: string;
    items: Partial<MealItem>[];
    confirmMeal: (meal: Meal) => void;
}>();

const emit = defineEmits<{
    (e: "update:ID", value: number): void;
    (e: "update:name", value: string): void;
    (e: "update:items", value: Partial<MealItem>[]): void;
}>();

const confirm = ref(-1);

const { data, pending, error } = useApiFetch<{
    date: string;
    days: Day[];
}>("mealplan/today");

function confirmMealInner(meal: Meal) {
    confirm.value = meal.ID;
    emit("update:ID", meal.ID);
    emit("update:name", meal.name);
    emit("update:items", meal.items);
    props.confirmMeal(meal);
}

function submitMeal(meal: Meal) {
    console.log("submit", meal);
    const payload = {
        meal_id: meal.ID,
        name: meal.name,
        items: meal.items,
    };
    console.log("payload", payload);
    const { data, error } = useApiFetch<Meal>("mealplan/meal/log", {
        method: "POST",
        body: payload,
    });
    if (error) {
        console.log("error", error);
    }
    if (data) {
        console.log("data", data);
    }
}
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else>
        <button
            v-for="dayMeal in data?.days[0]?.plannedMeals"
            :key="dayMeal.ID"
            @click="
                confirm === dayMeal.meal.ID
                    ? submitMeal(dayMeal.meal)
                    : confirmMealInner(dayMeal.meal)
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
