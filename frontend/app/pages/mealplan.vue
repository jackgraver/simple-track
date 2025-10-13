<script lang="ts" setup>
import type { Day } from "~/types/models";
import MacrosDate from "~/components/MacrosDate.vue";

const monthOffset = ref(0);

const data = ref<{ today: string; days: Day[] } | null>(null);
const pending = ref(false);
const error = ref<Error | null>(null);

async function fetchMealPlan() {
    pending.value = true;
    error.value = null;
    try {
        const res = await useApiFetch<{ today: string; days: Day[] }>(
            "mealplan/month",
            {
                query: {
                    monthoffset: monthOffset.value,
                },
            },
        );
        console.log(data.value);
        data.value = res.data.value ?? null;
        console.log(data.value);
    } catch (err) {
        error.value = err as Error;
    } finally {
        pending.value = false;
    }
}

function setCurrentMonth() {
    console.log("?");
    monthOffset.value = 0;
    fetchMealPlan();
}

function nextMonth() {
    monthOffset.value += 1;
    fetchMealPlan();
}

function prevMonth() {
    monthOffset.value -= 1;
    fetchMealPlan();
}

onMounted(() => {
    setCurrentMonth();
});
</script>

<template>
    <h1>Meal Plan</h1>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <Calendar
        v-else
        :fetchURL="'mealplan/month'"
        :display-component="MacrosDate"
    />
</template>
