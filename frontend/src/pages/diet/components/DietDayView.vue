<script setup lang="ts">
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import LoggedMealsDisplay from "./LoggedMealsDisplay.vue";
import PlannedMealsDisplay from "./PlannedMealsDisplay.vue";

const props = defineProps<{
    dateOffset: number;
}>();

const {
    data,
    isPending: pending,
    error,
} = useDietLogsToday(() => props.dateOffset);
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="flex w-full flex-col gap-4">
        <div v-if="data" class="w-full">
            <MacroBars
                :totalCalories="data?.totalCalories ?? 0"
                :totalProtein="data?.totalProtein ?? 0"
                :totalFiber="data?.totalFiber ?? 0"
                :totalCarbs="data?.totalCarbs ?? 0"
                :plannedCalories="data?.day.plan.calories ?? 0"
                :plannedProtein="data?.day.plan.protein ?? 0"
                :plannedFiber="data?.day.plan.fiber ?? 0"
                :plannedCarbs="data?.day.plan.carbs ?? 0"
            />
            <div class="flex w-full flex-col gap-8 sm:flex-row sm:gap-6">
                <LoggedMealsDisplay :date-offset="dateOffset" />
                <PlannedMealsDisplay :date-offset="dateOffset" />
            </div>
        </div>
    </div>
</template>
