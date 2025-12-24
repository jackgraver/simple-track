<script setup lang="ts">
import { useWorkoutStore } from "./store/useWorkoutStore";
import ExerciseListView from "./components/ExerciseListView.vue";
import { useRouter } from "vue-router";

const router = useRouter();
const { log, pending, error, addExerciseToWorkout, removeExerciseFromWorkout } = useWorkoutStore();

const selectExercise = (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;
    
    const exerciseId = exerciseGroup.planned?.ID || exerciseGroup.logged?.exercise_id;
    if (!exerciseId) return;
    
    router.push(`/liveworkout/log/${exerciseId}`);
};
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <ExerciseListView
            :exercises="log"
            @select-exercise="selectExercise"
            @add-exercise="addExerciseToWorkout($event)"
            @remove-exercise="removeExerciseFromWorkout"
        />
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 800px;
    margin: 0 auto;
}

@media (max-width: 767px) {
    .container {
        padding: 0.5rem;
        max-width: 100%;
    }
}
</style>

