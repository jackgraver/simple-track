<script setup lang="ts">
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/utils/axios";
import type { Plan } from "~/types/diet";

const { data, isPending, error } = useQuery({
    queryKey: ["diet", "plans", "plan", "all"],
    queryFn: async () => {
        const res = await apiClient.get<{ plans: Plan[] }>("/diet/plans/plan/all");
        return res.data;
    },
});
</script>

<template>
    <div class="container">
        <div class="section">
            <h1>Manage plans (macros and planned meals)</h1>
            <div v-if="isPending">Loading...</div>
            <div v-else-if="error">Error: {{ (error as Error).message }}</div>
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
