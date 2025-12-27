import type { Exercise, LoggedExercise } from "~/types/workout";
import { useWorkoutLogsPrevious } from "../queries/useWorkoutLogs";
import { useLogExercise, useAddExerciseToWorkout, useRemoveExerciseFromWorkout } from "../queries/useWorkoutMutations";
import { computed } from "vue";

export type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

export type LoggedSetWithStatus = {
    weight: number;
    reps: number;
    weight_setup: string;
    status: 'pending' | 'success' | 'error';
    id: number | null;
    error: string | null;
    tempId: string;
};

export function useWorkoutStore(offset: number = 0) {
    const workoutLogsQuery = useWorkoutLogsPrevious(offset);
    const logExerciseMutation = useLogExercise(offset);
    const addExerciseMutation = useAddExerciseToWorkout(offset);
    const removeExerciseMutation = useRemoveExerciseFromWorkout(offset);

    const log = computed<ExerciseGroup[]>(() => {
        return workoutLogsQuery.data.value?.previous_exercises ?? [];
    });

    const data = computed(() => workoutLogsQuery.data.value);
    const pending = computed(() => workoutLogsQuery.isPending.value);
    const error = computed(() => workoutLogsQuery.error.value);

    const logExercise = async (
        exercise: LoggedExercise,
        type: "logged" | "previous",
    ): Promise<LoggedExercise | null> => {
        try {
            const result = await logExerciseMutation.mutateAsync({
                exercise,
                type,
            });
            return result;
        } catch (error) {
            console.error("Error logging exercise:", error);
            return null;
        }
    };

    const addExerciseToWorkout = async (exerciseId: number): Promise<void> => {
        try {
            await addExerciseMutation.mutateAsync(exerciseId);
        } catch (error) {
            console.error("Error adding exercise:", error);
            throw error;
        }
    };

    const removeExerciseFromWorkout = async (exerciseId: number): Promise<void> => {
        try {
            await removeExerciseMutation.mutateAsync(exerciseId);
        } catch (error) {
            console.error("Error removing exercise:", error);
            throw error;
        }
    };

    const getExerciseByIndex = (index: number): ExerciseGroup | null => {
        return log.value[index] || null;
    };

    const getExerciseIndexById = (exerciseId: number): number | null => {
        const index = log.value.findIndex(
            (eg) => eg.planned?.ID === exerciseId || eg.logged?.exercise_id === exerciseId
        );
        return index >= 0 ? index : null;
    };

    return {
        log,
        data,
        pending,
        error,
        logExercise,
        addExerciseToWorkout,
        removeExerciseFromWorkout,
        getExerciseByIndex,
        getExerciseIndexById,
    };
}
