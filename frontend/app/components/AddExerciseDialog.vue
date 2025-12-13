<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import SearchList from "~/components/SearchList.vue";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    plan: WorkoutPlan;
    onResolve?: (result?: any) => void;
}>();

const { refresh: refreshExercises } = useAPIGet<{ exercises: Exercise[] }>("workout/exercises/all");

const addExerciseToPlan = async (exercise: Exercise): Promise<boolean> => {
    console.log("addExerciseToPlan", exercise);
    const { response, error } = await useAPIPost(
        `workout/plans/${props.plan.ID}/exercises/add`,
        "POST",
        { exercise_id: exercise.ID },
    );
    console.log(response);

    if (error) {
        toast.push("Failed to add exercise: " + error.message, "error");
        return false;
    }

    toast.push(`Added ${exercise.name} to ${props.plan.name}`, "success");
    props.onResolve?.(null);
    return true;
};

const createExercise = async (name: string) => {
    if (!name.trim()) return false;

    const { response, error } = await useAPIPost<{ exercise: Exercise }>(
        "workout/exercises/create",
        "POST",
        { name: name.trim(), rep_rollover: 10 },
    );

    if (error) {
        toast.push("Failed to create exercise: " + error.message, "error");
        return false;
    }

    toast.push(`Created exercise: ${name}`, "success");
    await refreshExercises();
    return true;
};
</script>

<template>
    <div class="add-exercise-dialog">
        <SearchList
            route="workout/exercises/all"
            :onSelect="addExerciseToPlan"
            :onCreate="createExercise"
            :prefilter="plan.exercises.map((e) => e.ID)"
        />
    </div>
</template>

<style scoped>
.add-exercise-dialog {
    display: flex;
    flex-direction: column;
    min-height: 300px;
    max-height: 60vh;
}
</style>

