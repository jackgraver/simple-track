<script setup lang="ts">
import { useWorkoutStore } from "./store/useWorkoutStore";
import ExerciseListView from "./components/ExerciseListView.vue";
import { useRouter } from "vue-router";
import { computed } from "vue";
import { toast } from "~/composables/toast/useToast";

const router = useRouter();
const { log, pending, error, data, addExerciseToWorkout, removeExerciseFromWorkout } = useWorkoutStore(0);

const workoutName = computed(() => data.value?.day.workout_plan?.name || "");

const selectExercise = (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;
    
    const exerciseId = exerciseGroup.planned?.ID || exerciseGroup.logged?.exercise_id;
    if (!exerciseId) return;
    
    router.push(`/liveworkout/log/${exerciseId}`);
};

const handleAddExercise = async (exerciseId: number) => {
    try {
        await addExerciseToWorkout(exerciseId);
        toast.push("Exercise added", "success");
    } catch (err: any) {
        toast.push(err.message || "Failed to add exercise", "error");
    }
};

const handleRemoveExercise = async (index: number) => {
    const exerciseGroup = log.value[index];
    if (!exerciseGroup) return;

    const exerciseId = exerciseGroup.logged?.exercise_id || exerciseGroup.planned?.ID;
    if (!exerciseId) {
        toast.push("Cannot remove exercise: ID not found", "error");
        return;
    }

    try {
        await removeExerciseFromWorkout(exerciseId);
        toast.push("Exercise removed", "success");
    } catch (err: any) {
        toast.push(err.message || "Failed to remove exercise", "error");
    }
};
</script>

<template>
    <div v-if="pending" class="container">
        <div>Loading...</div>
    </div>
    <div v-else-if="error" class="container">
        <div>Error: {{ error.message }}</div>
    </div>
    <div v-else class="container">
        <ExerciseListView
            :exercises="log"
            :workoutName="workoutName"
            @select-exercise="selectExercise"
            @add-exercise="handleAddExercise"
            @remove-exercise="handleRemoveExercise"
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
