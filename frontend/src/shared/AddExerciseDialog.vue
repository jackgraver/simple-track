<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import SearchList from "~/shared/SearchList.vue";
import { toast } from "~/composables/toast/useToast";
import { useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "~/utils/axios";

const props = defineProps<{
    plan: WorkoutPlan;
    onResolve?: (result: boolean) => void;
    onCancel?: () => void;
}>();

const queryClient = useQueryClient();

const addExerciseToPlan = async (exercise: Exercise): Promise<boolean> => {
    try {
        await apiClient.post(`workout/plans/${props.plan.ID}/exercises/add`, {
            exercise_id: exercise.ID,
        });
        toast.push(`Added ${exercise.name} to ${props.plan.name}`, "success");
        props.onResolve?.(true);
        return true;
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to add exercise: " + message, "error");
        return false;
    }
};

const createExercise = async (name: string) => {
    if (!name.trim()) return false;

    try {
        await apiClient.post<{ exercise: Exercise }>("/workout/exercises", {
            name: name.trim(),
            rep_rollover: 10,
            cues: "",
        });
        toast.push(`Created exercise: ${name}`, "success");
        await queryClient.invalidateQueries({ queryKey: ["searchList"] });
        return true;
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to create exercise: " + message, "error");
        return false;
    }
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
