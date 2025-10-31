<script setup lang="ts">
import type {
    Exercise,
    LoggedExercise,
    LoggedSet,
    WorkoutLog,
} from "~/types/workout";
import { Info } from "lucide-vue-next";

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    exercises: LoggedExercise[];
}>(`workout/previous`);

const day = data.value?.day;

// IDs of already logged exercises
const loggedExerciseIds = computed(() =>
    day?.exercises.map((e) => e.exercise?.ID),
);

// Filter out duplicates from the planned list
const unloggedExercises = computed(
    () =>
        day?.workout_plan?.exercises.filter(
            (planEx: Exercise) =>
                !loggedExerciseIds?.value?.includes(planEx.ID),
        ) ?? [],
);
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
            <template
                v-if="data?.exercises"
                v-for="exercise in data.exercises"
                :key="exercise.ID"
            >
                <TodayGymCard :exercise="exercise" :planned="false" />
            </template>

            <!-- planned but not yet logged -->
            <template
                v-if="unloggedExercises.length"
                v-for="exercise in unloggedExercises"
                :key="exercise.ID"
            >
                <TodayGymCard
                    :exercise="{
                        ID: 0,
                        created_at: '',
                        updated_at: '',
                        workout_log_id: day?.ID ?? 0,
                        exercise_id: exercise.ID,
                        sets: [] as LoggedSet[],
                        exercise: exercise,
                        weight_setup: '',
                        percent_change: 0,
                    }"
                    :planned="true"
                />
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
