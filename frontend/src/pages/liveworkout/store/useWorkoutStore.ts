import type { Exercise, LoggedExercise, WorkoutLog } from "~/types/workout";
import { useAPIGet, useAPIPost } from "~/composables/useApiFetch";
import { ref, watch, toRaw } from "vue";
import { toast } from "~/composables/toast/useToast";

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

// Singleton store instance
let storeInstance: ReturnType<typeof createWorkoutStore> | null = null;

const createWorkoutStore = () => {
    const { data, pending, error } = useAPIGet<{
        day: WorkoutLog;
        previous_exercises: ExerciseGroup[];
    }>(`workout/logs/previous?offset=-1`);

    const log = ref<ExerciseGroup[]>(
        data.value?.previous_exercises ?? [],
    );

    watch(() => data.value, (newData) => {
        if (newData) {
            log.value = newData.previous_exercises ?? [];
        }
    }, { immediate: true });

    const logExercise = async (
        exercise: LoggedExercise,
        type: "logged" | "previous",
    ): Promise<LoggedExercise | null> => {
        const rawExercise = toRaw(exercise);
        rawExercise.sets = toRaw(rawExercise.sets).filter(
            (set) => !(set.reps === 0 && set.weight === 0),
        );
        rawExercise.workout_log_id =
            data.value?.day.ID ?? rawExercise.workout_log_id;

        const { response, error: apiError } = await useAPIPost<{
            exercise: LoggedExercise;
        }>(
            `workout/exercises/log`,
            "POST",
            {
                exercise: rawExercise,
                type: type,
            },
            {},
            false,
        );

        if (apiError) {
            console.error(apiError);
            toast.push(apiError.message, "error");
            return null;
        }

        return response?.exercise || null;
    };

    const addExerciseToWorkout = async (exerciseId: number) => {
        const { response, error: apiError } = await useAPIPost<{
            exercise: LoggedExercise;
        }>(`workout/exercises/add`, "POST", {
            exercise_id: exerciseId,
        });

        if (apiError) {
            console.error(apiError);
            toast.push(apiError.message, "error");
            return;
        }

        if (response?.exercise) {
            const newExerciseGroup: ExerciseGroup = {
                planned: response.exercise.exercise!,
                logged: response.exercise,
                previous: {} as LoggedExercise,
            };
            log.value.push(newExerciseGroup);
            toast.push(`Added ${response.exercise.exercise?.name}`, "success");
        }
    };

    const removeExerciseFromWorkout = async (index: number) => {
        const exerciseGroup = log.value[index];
        if (!exerciseGroup) return;

        if (exerciseGroup.logged && exerciseGroup.logged.ID > 0) {
            const exerciseId = exerciseGroup.logged.exercise_id;
            if (!exerciseId) {
                toast.push("Cannot remove exercise: ID not found", "error");
                return;
            }

            const { error: apiError } = await useAPIPost(`workout/exercises/remove`, "DELETE", {
                exercise_id: exerciseId,
            });

            if (apiError) {
                console.error(apiError);
                toast.push(apiError.message, "error");
                return;
            }
        }

        log.value.splice(index, 1);
        toast.push("Exercise removed", "success");
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
};

export const useWorkoutStore = () => {
    if (!storeInstance) {
        storeInstance = createWorkoutStore();
    }
    return storeInstance;
};

