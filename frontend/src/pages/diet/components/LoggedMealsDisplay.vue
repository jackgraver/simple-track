<script setup lang="ts">
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal, editLogMeal } =
    useDietDayMealHandlers(() => props.dateOffset);
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
        <h2 class="mb-0 text-lg font-semibold">Logged</h2>
        <template v-if="data">
            <span
                v-if="data.day.loggedMeals.length === 0"
                class="text-zinc-500"
                >Nothing logged yet.</span
            >
            <MealCard
                v-for="log in data.day.loggedMeals"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                :on-edit="editLogMeal"
                type="logged"
            />
        </template>
    </div>
</template>
