import type { Exercise, LoggedExercise } from "~/types/workout";
import {
    useWorkoutLogsPrevious,
    useAddExerciseToWorkout,
    useRemoveExerciseFromWorkout,
    useUpsertCardio,
    useUpsertMobilityPre,
    useUpsertMobilityPost,
} from "~/api/workout/queries";
import { userKey } from "~/utils/userStorage";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import { sortExerciseGroupsByLogOrder } from "./sortExerciseGroupsByLogOrder";
import { computed, ref, toValue, watch, type MaybeRefOrGetter } from "vue";
import { offsetToDate } from "~/api/workout/dates";
import type { SetRow } from "~/api/workout/setsCache";

const HIDDEN_EXERCISES_STORAGE_PREFIX = "workout:hidden-exercises:";

type HiddenExercisesSnapshot = { ids: number[] };

export type ExerciseGroup = {
    planned?: Exercise;
    logged?: LoggedExercise;
    previous?: LoggedExercise;
    max?: LoggedExercise;
};

/** A set row with its derived save state (cache + pending/failed mutations, see `deriveSetRows`). */
export type LoggedSetWithStatus = SetRow;

export function useWorkoutStore(offset: MaybeRefOrGetter<number> = 0) {
    const workoutLogsQuery = useWorkoutLogsPrevious(offset);
    /** The day a write targets, fixed when the action starts (not when it settles). */
    const currentDate = () => offsetToDate(toValue(offset));

    const addExerciseMutation = useAddExerciseToWorkout();
    const removeExerciseMutation = useRemoveExerciseFromWorkout();
    const upsertCardioMutation = useUpsertCardio();
    const upsertMobilityPreMutation = useUpsertMobilityPre();
    const upsertMobilityPostMutation = useUpsertMobilityPost();

    /** Catalog exercise IDs hidden from the log list; persisted per workout log in localStorage. */
    const skippedExerciseIds = ref<Set<number>>(new Set());

    const hiddenExercisesStorageKey = computed(() => {
        const id = workoutLogsQuery.data.value?.day?.ID;
        if (id == null) return "";
        return userKey(`${HIDDEN_EXERCISES_STORAGE_PREFIX}${id}`) ?? "";
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

    // All of these update the cache at once and report failures themselves (rollback + Retry toast).
    const addExerciseToWorkout = (exerciseId: number, exercise?: Exercise): void => {
        unhideExerciseFromSession(exerciseId);
        addExerciseMutation.mutate({ date: currentDate(), exerciseId, exercise });
    };

    const removeExerciseFromLog = (exerciseId: number): void => {
        const group = workoutLogsQuery.data.value?.planned_exercises?.find(
            (eg) => eg.planned?.ID === exerciseId || eg.logged?.exercise_id === exerciseId,
        );
        // Nothing logged: the server has nothing to delete (it answered 404), so just hide it here.
        if (!group?.logged) {
            hideExerciseLocally(exerciseId);
            return;
        }
        unhideExerciseFromSession(exerciseId);
        removeExerciseMutation.mutate({ date: currentDate(), exerciseId });
    };

    const saveCardio = (minutes: number, type?: string, notes?: string): void => {
        upsertCardioMutation.mutate({ date: currentDate(), minutes, type, notes });
    };

    const savePreMobility = (checked: string[]): void => {
        upsertMobilityPreMutation.mutate({ date: currentDate(), checked });
    };

    const savePostMobility = (checked: string[]): void => {
        upsertMobilityPostMutation.mutate({ date: currentDate(), checked });
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
        refetch: () => workoutLogsQuery.refetch(),
        addExerciseToWorkout,
        removeExerciseFromLog,
        hideExerciseLocally,
        saveCardio,
        savePreMobility,
        savePostMobility,
        getExerciseByIndex,
        getExerciseIndexById,
    };
}
