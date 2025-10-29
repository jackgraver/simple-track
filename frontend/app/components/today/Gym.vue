<script setup lang="ts">
import type { LoggedExercise, WorkoutLog } from "~/types/workout";
import { Info } from "lucide-vue-next";

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    exercises: LoggedExercise[];
}>(`workout/previous`);

const day = data.value?.day;
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div class="title-row">
            <h2>{{ day?.workout_plan?.name }} Day</h2>
            <button>Live Workout</button>
        </div>
        <div class="workout-grid">
            <template v-for="exercise in data?.exercises" :key="exercise.ID">
                <TodayGymCard :exercise="exercise" />
            </template>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
}

.workout-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.75rem;
    width: 100%;
}

@media (min-width: calc(5 * 220px)) {
    .workout-grid {
        grid-template-columns: repeat(4, 1fr);
    }
}

.title-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
}

.title-row h2 {
    flex: 1;
    margin-bottom: 0;
}

.title-row button {
    margin: 0px;
    border-radius: 4px;
    font-size: large;
    padding: 6px 12px;
    font-weight: bold;
    text-decoration: none;
    font-size: large;
    padding: 6px 16px;
    align-self: end;
}
</style>
