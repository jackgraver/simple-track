<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { WorkoutPlan, Exercise } from "~/types/workout";
import SearchList from "~/shared/SearchList.vue";
import { toast } from "~/composables/toast/useToast";
import { useAddRoutineExercise, useCreateExercise } from "~/api/workout/routines";

const props = defineProps<{
    plan: WorkoutPlan;
    onResolve?: (result: boolean) => void;
    onCancel?: () => void;
}>();

const addToPlan = useAddRoutineExercise();
const createExerciseMutation = useCreateExercise();

const addExerciseToPlan = async (exercise: Exercise): Promise<boolean> => {
    try {
        await addToPlan.mutateAsync({ planId: props.plan.ID, exerciseId: exercise.ID });
        toast.push(`Added ${exercise.name} to ${props.plan.name}`, "success");
        props.onResolve?.(true);
        return true;
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to add exercise: " + message, "error");
        return false;
    }
};

const createExercise = async (name: string) => {
    if (!name.trim()) return false;

    try {
        // Create and add in one step: typing a name that isn't in the library means "this one".
        const created = await createExerciseMutation.mutateAsync({ name: name.trim(), rep_rollover: 10, cues: "" });
        return addExerciseToPlan(created);
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to create exercise: " + message, "error");
        return false;
    }
};
</script>

<template>
    <div class="flex max-h-[60vh] min-h-72 flex-col">
        <SearchList
            route="workout/exercises/all"
            :onSelect="addExerciseToPlan"
            :onCreate="createExercise"
            create-label="Create and add"
            :prefilter="plan.exercises.map((e) => e.ID)"
            auto-focus
        />
    </div>
</template>
