<script setup lang="ts">
import type { Exercise } from "~/types/workout";
import SearchList from "~/components/SearchList.vue";

type ProgressionEntry = {
    date: string;
    weight: number;
    reps: number;
};

const selectedExercise = ref<Exercise | null>(null);
const progressionEndpoint = ref<string>("");


const {
    data: progressionData,
    pending: loading,
    error: fetchError,
} = useAPIGet<{ progression: ProgressionEntry[] }>(progressionEndpoint);

const progression = computed(() => {
    return progressionData.value?.progression ?? [];
});

const error = computed(() => {
    return fetchError.value?.message || null;
});

const selectExercise = async (exercise: Exercise): Promise<boolean> => {
    selectedExercise.value = exercise;
    progressionEndpoint.value = `workout/exercise/progression/${exercise.ID}`;
    return true;
};

const formatProgressionDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    const month = date.toLocaleString("en-US", { month: "short" });
    const day = date.getDate();
    return `${month} ${day}`;
};
</script>

<template>
    <div class="container">
        <h1>Exercise Progression</h1>
        <div class="content">
            <div class="exercise-selector">
                <h2>Select Exercise</h2>
                <SearchList
                    route="workout/exercises/all"
                    :on-select="selectExercise"
                />
            </div>
            <div class="progression-display">
                <div v-if="selectedExercise" class="exercise-header">
                    <h2>{{ selectedExercise.name }}:</h2>
                </div>
                <div v-if="loading" class="loading">Loading...</div>
                <div v-else-if="error" class="error">{{ error }}</div>
                <div v-else-if="progression.length === 0 && selectedExercise" class="no-data">
                    No progression data available for this exercise.
                </div>
                <div v-else-if="progression.length > 0" class="progression-list">
                    <div
                        v-for="(entry, index) in progression"
                        :key="index"
                        class="progression-entry"
                    >
                        {{ formatProgressionDate(entry.date) }} - {{ entry.weight }} lbs for {{ entry.reps }} reps
                    </div>
                </div>
            </div>
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

.content {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
}

.exercise-selector,
.progression-display {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.exercise-header h2 {
    margin: 0;
}

.progression-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.progression-entry {
    font-family: monospace;
    padding: 0.5rem;
    background-color: rgb(48, 48, 48);
    border-radius: 4px;
}

.loading,
.error,
.no-data {
    padding: 1rem;
    text-align: center;
}

.error {
    color: rgb(255, 100, 100);
}

@media (max-width: 767px) {
    .content {
        grid-template-columns: 1fr;
    }
}
</style>

