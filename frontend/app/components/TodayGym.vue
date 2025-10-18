<script setup lang="ts">
import type { WorkoutLog } from "~/types/workout";

const { data, pending, error } = useAPIGet<WorkoutLog>(`workout/previous`);

const day = data.value;
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-if="data">
        <div v-for="exercise in day?.exercises" :key="exercise.ID">
            <h2>{{ exercise.name }}</h2>
            <p v-for="set in exercise.sets" :key="set.ID">
                {{ set.reps }} x {{ set.weight }}
            </p>
        </div>
    </div>
</template>

<style scoped></style>
