<script setup lang="ts">
import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { Plus, Trash } from "lucide-vue-next";
type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/previous?offset=0`);

const log = ref<LoggedExercise[]>(
    data?.value?.previous_exercises.map((e) => e.previous) ?? [],
);

const addSet = (exercise: LoggedExercise) => {
    if (!exercise) return;

    const weight = exercise?.sets[0]?.weight ?? 0;
    const reps = exercise?.sets[0]?.reps ?? 0;

    exercise.sets.push({
        logged_exercise_id: 0,
        reps: reps,
        weight: weight,
        ID: 0,
        created_at: "",
        updated_at: "",
    });
};

const removeSet = (
    exercise: LoggedExercise,
    set: LoggedExercise["sets"][0],
) => {
    if (!exercise) return;

    const index = exercise.sets.indexOf(set);
    exercise.sets.splice(index, 1);
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <article v-for="log in log">
            <h1>{{ log.exercise.name }}</h1>
            <template v-for="(set, i) in log.sets">
                <div class="set">
                    <div class="set-input">
                        <label v-if="i === 0">Weight</label>
                        <input type="number" v-model="set.weight" />
                    </div>
                    <div class="set-input">
                        <label v-if="i === 0">Reps</label>
                        <input type="number" v-model="set.reps" />
                    </div>
                    <button class="delete-button" @click="removeSet(log, set)">
                        <Trash />
                    </button>
                </div>
            </template>
            <button @click="addSet(log)"><Plus /></button>
        </article>
        <button><NuxtLink to="/">Finish</NuxtLink></button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.container article {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
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

.set {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
}

.set-input {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
</style>
