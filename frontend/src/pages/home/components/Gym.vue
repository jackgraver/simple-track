<script setup lang="ts">
import { computed } from "vue";
import type {
    LoggedExercise,
    LoggedSet,
} from "~/types/workout";
import { useWorkoutLogsPrevious } from "../queries/useWorkoutLogsPrevious";
import GymCard from "./GymCard.vue";

const props = defineProps<{
    dateOffset: number;
}>();

const { data, isPending: pending, error } = useWorkoutLogsPrevious(props.dateOffset);

const day = computed(() => data.value?.day);
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div class="title-row">
            <h2>{{ day?.workout_plan?.name }} Day</h2>
            <button>
                <RouterLink :to="'/liveworkout'">
                    Live Workout
                </RouterLink>
            </button>
        </div>
        <div class="workout-grid">
            <template
                v-for="exercise in data?.previous_exercises"
                :key="exercise.planned.ID"
            >
                <template v-if="exercise.logged">
                    <GymCard
                        :exercise="exercise.logged"
                        :previous="
                            exercise.previous.exercise
                                ? exercise.previous
                                : null
                        "
                        :planned="true"
                    />
                </template>
                <template v-else>
                    <GymCard
                        :exercise="{
                            ID: 0,
                            created_at: '',
                            updated_at: '',
                            workout_log_id: day?.ID ?? 0,
                            exercise_id: exercise.planned.ID,
                            sets: [] as LoggedSet[],
                            exercise: exercise.planned,
                            notes: '',
                            percent_change: 0,
                        }"
                        :previous="
                            exercise.previous.exercise
                                ? exercise.previous
                                : null
                        "
                        :planned="true"
                    />
                </template>
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

.link {
    text-decoration: none;
}

.link:visited {
    color: white;
}
</style>
