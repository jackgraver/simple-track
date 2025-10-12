<script lang="ts" setup>
import type { Day } from "~/types/models";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import MacrosDate from "~/components/MacrosDate.vue";

const {
    data: mealPlan,
    pending,
    error,
} = useApiFetch<{
    today: string;
    days: Day[];
}>("mealplan/month");

const displayDayDialog = async (day: Day) => {
    dialogManager.confirm({
        title: "Modify " + formatDate(day.date),
        message: "Are you sure you want to modify this meal?",
        confirmText: "Modify",
        cancelText: "Cancel",
    });
};
</script>

<template>
    <h1>Meal Plan</h1>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <Calendar
        v-else
        :days="mealPlan?.days ?? []"
        :display-component="MacrosDate"
    />
</template>

<style scoped></style>
