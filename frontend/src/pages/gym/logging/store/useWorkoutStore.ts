import type { Exercise, LoggedExercise } from "~/types/workout";
import {
    useWorkoutLogsPrevious,
    useLogExercise,
    useAddExerciseToWorkout,
    useDeleteLoggedSet,
    useUpsertCardio,
    useUpsertMobilityPre,
    useUpsertMobilityPost,
} from "~/api/workout/queries";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import { sortExerciseGroupsByLogOrder } from "./sortExerciseGroupsByLogOrder";
import { computed, ref, watch, type MaybeRefOrGetter } from "vue";

const HIDDEN_EXERCISES_STORAGE_PREFIX = "simpletracker:workout:hidden-exercises:";

type HiddenExercisesSnapshot = { ids: number[] };

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
    const deleteLoggedSetMutation = useDeleteLoggedSet(offset);
    const upsertCardioMutation = useUpsertCardio(offset);
    const upsertMobilityPreMutation = useUpsertMobilityPre(offset);
    const upsertMobilityPostMutation = useUpsertMobilityPost(offset);

    /** Catalog exercise IDs hidden from the log list; persisted per workout log in localStorage. */
    const skippedExerciseIds = ref<Set<number>>(new Set());

    const hiddenExercisesStorageKey = computed(() => {
        const id = workoutLogsQuery.data.value?.day?.ID;
        if (id == null) return "";
        return `${HIDDEN_EXERCISES_STORAGE_PREFIX}${id}`;
    });

    const hiddenExercisesSync = useWebStorageJsonSync<HiddenExercisesSnapshot>({
        key: hiddenExercisesStorageKey,
        watchSources: [skippedExerciseIds],
        getSnapshot: () => ({ ids: [...skippedExerciseIds.value] }),
        canPersist: () => hiddenExercisesStorageKey.value !== "",
        tryRestore: (parsed, ctx) => {
            const raw = parsed as unknown;
            let ids: unknown;
            if (Array.isArray(raw)) {
                ids = raw;
            } else if (raw && typeof raw === "object" && "ids" in raw) {
                ids = (raw as { ids?: unknown }).ids;
            } else {
                ctx.remove();
                return false;
            }
            if (!Array.isArray(ids)) {
                ctx.remove();
                return false;
            }
            skippedExerciseIds.value = new Set(
                ids.filter(
                    (x): x is number =>
                        typeof x === "number" && Number.isFinite(x),
                ),
            );
            return true;
        },
    });

    watch(
        () => workoutLogsQuery.data.value?.day?.ID,
        (workoutLogId) => {
            hiddenExercisesSync.setSaveEnabled(false);
            skippedExerciseIds.value = new Set();
            if (workoutLogId == null) {
                hiddenExercisesSync.setSaveEnabled(true);
                return;
            }
            hiddenExercisesSync.restore();
            hiddenExercisesSync.setSaveEnabled(true);
        },
        { immediate: true },
    );

    const hideExerciseLocally = (exerciseId: number) => {
        const next = new Set(skippedExerciseIds.value);
        next.add(exerciseId);
        skippedExerciseIds.value = next;
    };

    const unhideExerciseFromSession = (exerciseId: number) => {
        if (!skippedExerciseIds.value.has(exerciseId)) return;
        const next = new Set(skippedExerciseIds.value);
        next.delete(exerciseId);
        skippedExerciseIds.value = next;
    };

    const log = computed<ExerciseGroup[]>(() => {
        const raw = workoutLogsQuery.data.value?.planned_exercises ?? [];
        const sorted = sortExerciseGroupsByLogOrder(raw);
        const skipped = skippedExerciseIds.value;
        if (skipped.size === 0) return sorted;
        return sorted.filter((eg) => {
            const id = eg.logged?.exercise_id ?? eg.planned?.ID;
            if (id == null) return true;
            return !skipped.has(id);
        });
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
            unhideExerciseFromSession(exerciseId);
        } catch (error) {
            console.error("Error adding exercise:", error);
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
        hideExerciseLocally,
        deleteLoggedSet,
        saveCardio,
        savePreMobility,
        savePostMobility,
        getExerciseByIndex,
        getExerciseIndexById,
    };
}
