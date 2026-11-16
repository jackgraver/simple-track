<script setup lang="ts">
import { computed } from "vue";
import MealCard from "./MealCard.vue";
import { useDietDayMealHandlers } from "./useDietDayMealHandlers";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, logPlannedMeal, logMeal, deleteLoggedMeal } =
    useDietDayMealHandlers(() => props.dateOffset);

const loggedMeals = computed(() => data.value?.day.loggedMeals ?? []);
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
        <div class="flex h-9 items-center justify-between gap-2">
            <div class="flex min-w-0 flex-1 items-center gap-2">
                <h2
                    class="m-0 min-w-0 flex-1 text-base font-semibold leading-none"
                >
                    Logged ({{ loggedMeals.length }})
                </h2>
            </div>
        </div>
        <template v-if="data">
            <span v-if="loggedMeals.length === 0" class="text-zinc-500"
                >Nothing logged yet.</span
            >
            <MealCard
                v-for="log in loggedMeals"
                :key="log.ID"
                :meal="log.meal"
                :on-log-planned="logPlannedMeal"
                :on-log-edited="logMeal"
                :on-delete="deleteLoggedMeal"
                type="logged"
            />
        </template>
    </div>
</template>
