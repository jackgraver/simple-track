<script setup lang="ts">
import type { Plan } from "~/types/diet";

const { data, pending, error } = useAPIGet<{ plans: Plan[] }>(
    "mealplan/plan/all",
);
console.log(data);
</script>

<template>
    <div class="container">
        <div class="section">
            <h1>Manage plans (macros and planned meals)</h1>
            <div v-if="pending">Loading...</div>
            <div v-else-if="error">Error: {{ error.message }}</div>
            <div v-else>
                <h1>Todays Plan</h1>
                <p>...</p>
                <h1>All Plans</h1>
                <div v-for="plan in data?.plans" :key="plan.ID" class="plan">
                    <h2>
                        {{ plan.name }} (April - May) / (60 days) / (5 days
                        remaining)
                    </h2>
                    <p>{{ plan.calories }} calories</p>
                    <p>{{ plan.protein }} protein</p>
                    <p>{{ plan.fiber }} fiber</p>
                </div>
            </div>
        </div>
        <div class="section">
            <h1>Manage logged meals</h1>
        </div>
        <div class="section">
            <h1>Manage saved meals</h1>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
}

.section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    background-color: rgb(19, 19, 19);
    border: 1px solid rgb(39, 39, 39);
    border-radius: 0.5rem;
    padding: 1rem;
}
</style>

<!-- <script lang="ts" setup>
import type { Day } from "~/types/diet";
import MacrosDate from "~/components/MacrosDate.vue";

const monthOffset = ref(0);

const data = ref<{ today: string; days: Day[] } | null>(null);
const pending = ref(false);
const error = ref<Error | null>(null);

async function fetchMealPlan() {
    pending.value = true;
    error.value = null;
    try {
        const res = await useAPIGet<{ today: string; days: Day[] }>(
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
</template> -->
