import type { Exercise, LoggedExercise } from "~/types/workout";
import {
    useWorkoutLogsPrevious,
    useLogExercise,
    useAddExerciseToWorkout,
    useRemoveExerciseFromWorkout,
    useDeleteLoggedSet,
    useUpsertCardio,
    useUpsertMobilityPre,
    useUpsertMobilityPost,
} from "~/api/workout/queries";
import { sortExerciseGroupsByLogOrder } from "./sortExerciseGroupsByLogOrder";
import { computed, type MaybeRefOrGetter } from "vue";

export type ExerciseGroup = {
    planned?: Exercise;
    logged?: LoggedExercise;
    previous?: LoggedExercise;
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

export function useWorkoutStore(offset: MaybeRefOrGetter<number> = 0) {
    const workoutLogsQuery = useWorkoutLogsPrevious(offset);
    const logExerciseMutation = useLogExercise(offset);
    const addExerciseMutation = useAddExerciseToWorkout(offset);
    const removeExerciseMutation = useRemoveExerciseFromWorkout(offset);
    const deleteLoggedSetMutation = useDeleteLoggedSet(offset);
    const upsertCardioMutation = useUpsertCardio(offset);
    const upsertMobilityPreMutation = useUpsertMobilityPre(offset);
    const upsertMobilityPostMutation = useUpsertMobilityPost(offset);

    const log = computed<ExerciseGroup[]>(() => {
        const raw = workoutLogsQuery.data.value?.planned_exercises ?? [];
        return sortExerciseGroupsByLogOrder(raw);
    });

    const plannedCardio = computed(
        () => workoutLogsQuery.data.value?.planned_cardio ?? null,
    );
    const loggedCardio = computed(
        () => workoutLogsQuery.data.value?.logged_cardio ?? null,
    );

    const plannedPreMobility = computed(
        () => workoutLogsQuery.data.value?.planned_pre_mobility ?? null,
    );
    const loggedPreMobility = computed(
        () => workoutLogsQuery.data.value?.logged_pre_mobility ?? null,
    );
    const plannedPostMobility = computed(
        () => workoutLogsQuery.data.value?.planned_post_mobility ?? null,
    );
    const loggedPostMobility = computed(
        () => workoutLogsQuery.data.value?.logged_post_mobility ?? null,
    );

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

    const deleteLoggedSet = async (setId: number): Promise<void> => {
        try {
            await deleteLoggedSetMutation.mutateAsync(setId);
        } catch (error) {
            console.error("Error deleting logged set:", error);
            throw error;
        }
    };

    const saveCardio = async (minutes: number, type?: string, notes?: string): Promise<void> => {
        await upsertCardioMutation.mutateAsync({ minutes, type, notes });
    };

    const savePreMobility = async (checked: string[]): Promise<void> => {
        await upsertMobilityPreMutation.mutateAsync(checked);
    };

    const savePostMobility = async (checked: string[]): Promise<void> => {
        await upsertMobilityPostMutation.mutateAsync(checked);
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
        plannedCardio,
        loggedCardio,
        plannedPreMobility,
        loggedPreMobility,
        plannedPostMobility,
        loggedPostMobility,
        data,
        pending,
        error,
        logExercise,
        addExerciseToWorkout,
        removeExerciseFromWorkout,
        deleteLoggedSet,
        saveCardio,
        savePreMobility,
        savePostMobility,
        getExerciseByIndex,
        getExerciseIndexById,
    };
}
