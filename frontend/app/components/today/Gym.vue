<script setup lang="ts">
import type { WorkoutLog } from "~/types/workout";

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
}>(`workout/previous`);

const day = data.value?.day;
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div class="title-row">
            <h2>{{ day?.workout_plan?.name }} Day</h2>
            <button>Log Workout</button>
        </div>
        <div class="workout-grid">
            <article
                class="workout-card"
                v-for="exercise in day?.exercises"
                :key="exercise.ID"
            >
                <header class="card-header">
                    <h3>{{ exercise.name }}</h3>
                    {{ exercise.sets[exercise.sets.length - 1]?.reps ?? "X" }} x
                    {{ exercise.sets[exercise.sets.length - 1]?.weight ?? "X" }}
                </header>
                <span>Use 2 45 plates</span>
                <span class="info">Time to increase!</span>
                <span class="fact">Up 5% since last workout</span>
            </article>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    /* Fill the width provided by the page wrapper */
    width: 100%;
}

.workout-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 0.75rem;
    width: 100%;
}

.title-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
}

.title-row h2 {
    flex: 1;
}

.title-row button {
    margin-top: 6px;
    border-radius: 4px;
    font-size: large;
    padding: 6px 12px;
    font-weight: bold;
    text-decoration: none;
    font-size: large;
    padding: 6px 16px;
}

.workout-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid #333;
    border-radius: 0.5rem;
    background: #1a1a1a;
    color: #fff;
    min-height: 120px;
}

.workout-card span {
    font-size: small;
}
/*
.workout-card h2 {
    margin-top: 0;
    margin-bottom: 0.5rem;
    width: 100%;
} */

.card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
}

.card-header h3 {
    margin: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.workout-card .info {
    color: rgb(206, 206, 48);
}

.workout-card .fact {
    color: rgb(63, 197, 46);
}

.sets {
    display: grid;
    grid-auto-rows: min-content;
    row-gap: 0.25rem;
    padding: 0;
    margin: 0;
    list-style: none;
}

.set {
    display: inline-grid;
    grid-auto-flow: column;
    column-gap: 0.25rem;
    justify-content: start;
    align-items: baseline;
    color: #ddd;
}

.times {
    opacity: 0.7;
}

.card-footer {
    min-height: 0.5rem; /* reserved space for future indicators */
}

.badge {
    display: inline-block;
    font-size: 0.75rem;
    padding: 0.1rem 0.35rem;
    border-radius: 0.25rem;
}

.badge-inc {
    background: rgba(0, 200, 83, 0.15);
    color: #6ee7b7;
    border: 1px solid rgba(0, 200, 83, 0.25);
}
</style>
