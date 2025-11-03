<script setup lang="ts">
import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { Plus } from "lucide-vue-next";
type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/previous?offset=0`);
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <article v-for="prev in data?.previous_exercises">
            <h1>{{ prev.planned.name }}</h1>
            <template v-for="(set, i) in prev.previous.sets">
                <label v-if="i === 0">Weight</label>
                <input type="number" v-model="set.weight" />
                <label v-if="i === 0">Reps</label>
                <input type="number" v-model="set.reps" />
            </template>
            <button><Plus /></button>
        </article>
        <button><NuxtLink to="/">Finish</NuxtLink></button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 1rem;
    width: 75%;
    max-width: 1200px;
    margin: 0 auto;
    min-height: 100vh;
}

.container article {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    width: 100%;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    padding: 1rem;
}

.container article header {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    align-items: center;
}

.container article footer {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    align-items: center;
}

.container article h1 {
    margin: 0;
}
</style>
