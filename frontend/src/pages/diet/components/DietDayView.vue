<script setup lang="ts">
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import LoggedMealsDisplay from "./LoggedMealsDisplay.vue";
import PlannedMealsDisplay from "./PlannedMealsDisplay.vue";

const props = defineProps<{
    dateOffset: number;
}>();

function getTotalNumber(
    pending: boolean,
    error: unknown,
    num: number | undefined,
): number {
    return pending || error ? 0 : (num ?? 0);
}

const {
    data,
    isPending: pending,
    error,
} = useDietLogsToday(() => props.dateOffset);
</script>

<template>
    <div class="flex w-full flex-col gap-4">
        <div class="w-full">
            <MacroBars
                :totalCalories="
                    getTotalNumber(pending, error, data?.totalCalories)
                "
                :totalProtein="
                    getTotalNumber(pending, error, data?.totalProtein)
                "
                :totalFiber="getTotalNumber(pending, error, data?.totalFiber)"
                :totalCarbs="getTotalNumber(pending, error, data?.totalCarbs)"
                :plannedCalories="
                    getTotalNumber(pending, error, data?.day.plan.calories)
                "
                :plannedProtein="
                    getTotalNumber(pending, error, data?.day.plan.protein)
                "
                :plannedFiber="
                    getTotalNumber(pending, error, data?.day.plan.fiber)
                "
                :plannedCarbs="
                    getTotalNumber(pending, error, data?.day.plan.carbs)
                "
            />

            <div class="flex w-full flex-col gap-8 sm:flex-row sm:gap-6">
                <LoggedMealsDisplay :date-offset="dateOffset" />
                <PlannedMealsDisplay :date-offset="dateOffset" />
            </div>
        </div>
    </div>
</template>
