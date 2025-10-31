<script setup lang="ts">
import type {
    Exercise,
    LoggedExercise,
    LoggedSet,
    WorkoutLog,
} from "~/types/workout";
import { Info } from "lucide-vue-next";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const props = defineProps<{
    dateOffset: number;
}>();

const { data, pending, error } = useAPIGet<{
    day: WorkoutLog;
    previous_exercises: ExerciseGroup[];
}>(`workout/previous?offset=${props.dateOffset}`);

const day = data.value?.day;
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div class="title-row">
            <h2>{{ day?.workout_plan?.name }} Day {{ dateOffset }}</h2>
            <button>Live Workout</button>
        </div>
        <div class="workout-grid">
            <template
                v-for="exercise in data?.previous_exercises"
                :key="exercise.planned.ID"
            >
                <template v-if="exercise.logged">
                    <TodayGymCard
                        :exercise="exercise.logged"
                        :previous="exercise.previous"
                        :planned="true"
                    />
                </template>
                <template v-else>
                    <TodayGymCard
                        :exercise="{
                            ID: 0,
                            created_at: '',
                            updated_at: '',
                            workout_log_id: day?.ID ?? 0,
                            exercise_id: exercise.planned.ID,
                            sets: [] as LoggedSet[],
                            exercise: exercise.planned,
                            weight_setup: '',
                            percent_change: 0,
                        }"
                        :previous="exercise.previous"
                        :planned="true"
                    />
                </template>
            </template>
            <!-- <template
                v-if="data?.exercises"
                v-for="exercise in data.exercises"
                :key="exercise.ID"
            >
                <TodayGymCard :exercise="exercise" :planned="false" />
            </template>
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
                        sets: [
                            {
                                ID: 0,
                                created_at: '',
                                updated_at: '',
                                logged_exercise_id: 0,
                                reps: 0,
                                weight: 0,
                            },
                        ] as LoggedSet[],
                        exercise: exercise,
                        weight_setup: '',
                        percent_change: 0,
                    }"
                    :planned="true"
                />
            </template> -->
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
