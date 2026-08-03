import type { ComputedRef, Ref } from "vue";
import { computed, reactive, ref, watch } from "vue";
import type { Router } from "vue-router";
import { buildLoggingListQuery } from "../../composables/useLoggingRouteContext";
import type { ExerciseGroup, LoggedSetWithStatus } from "../../store/useWorkoutStore";
import type { LoggedExercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import {
    buildAllSetsForSave,
    buildExerciseToLog,
    loggedSetsFromServer,
    mergeSavedExerciseIntoLoggedSets,
    markPendingSetsAsExerciseError,
} from "../domain/exerciseLoggingPayload";
import { initializeWeightAndReps } from "../domain/loggedSetDefaults";
import { weightSetupForExercise } from "../domain/weightSetup";

type StoredDraft = {
    weight: number;
    reps: number;
    weightSetup: string;
    notes: string;
    dirty: boolean;
};

/** Plain shape for templates (refs unwrapped via reactive session object). */
export type ExerciseLoggingSessionViewModel = {
    exerciseGroup: ExerciseGroup | null;
    currentSetNumber: number;
    loggedSets: LoggedSetWithStatus[];
    currentWeight: number;
    currentReps: number;
    currentWeightSetup: string;
    notes: string;
    stepWeight: (direction: "plus" | "minus") => void;
    stepReps: (direction: "plus" | "minus") => void;
    commitWeightFromInput: (value: number) => void;
    commitRepsFromInput: (value: number) => void;
    addNextSet: () => Promise<boolean>;
    finishLogging: () => Promise<void>;
    retrySet: (setIndex: number) => Promise<void>;
    deleteSet: (setIndex: number) => Promise<void>;
    editSet: (setIndex: number) => void;
    goBackToList: () => void;
    updateNotes: (value: string) => void;
    updateWeightSetup: (value: string) => void;
};

type LogExerciseFn = (
    exercise: LoggedExercise,
    type: "logged" | "previous",
) => Promise<LoggedExercise | null>;
type DeleteLoggedSetFn = (setId: number) => Promise<void>;

export function useExerciseLoggingSession(options: {
    exerciseGroup: ComputedRef<ExerciseGroup | null>;
    pending: Ref<boolean>;
    dayId: ComputedRef<number>;
    offset: ComputedRef<number>;
    logExercise: LogExerciseFn;
    deleteLoggedSet: DeleteLoggedSetFn;
    router: Router;
}): ExerciseLoggingSessionViewModel {
    const {
        exerciseGroup,
        pending,
        dayId,
        offset,
        logExercise,
        deleteLoggedSet,
        router,
    } = options;
    const loggingRoute = () => ({
        name: "logging" as const,
        query: buildLoggingListQuery(offset.value),
    });

    const currentWeight = ref(0);
    const currentReps = ref(0);
    const currentWeightSetup = ref("");
    const currentSetNumber = ref(1);

    const loggedSets = ref<LoggedSetWithStatus[]>([]);
    let tempIdCounter = 0;

    const draftDirty = ref(false);
    const notes = ref("");
    const globalTimer = useGlobalRestTimer();
    const REST_DURATION_MS = 2 * 60 * 1000;
    const MAX_REPS = 50;

    const normalizeReps = (value: number) => {
        if (!Number.isFinite(value)) return 0;
        return Math.min(MAX_REPS, Math.max(0, Math.round(value)));
    };

    const exerciseIdentity = computed(() => {
        const group = exerciseGroup.value;
        return (
            group?.planned?.ID ??
            group?.logged?.exercise_id ??
            group?.logged?.exercise?.ID ??
            0
        );
    });

    const exerciseDisplayName = computed(() => {
        const group = exerciseGroup.value;
        return (
            group?.planned?.name ??
            group?.logged?.exercise?.name ??
            ""
        );
    });

    const exerciseLoadType = computed(
        () =>
            exerciseGroup.value?.planned?.load_type ??
            exerciseGroup.value?.logged?.exercise?.load_type ??
            "plate_loaded_with_bar",
    );

    const weightSetupForSave = () =>
        weightSetupForExercise(
            exerciseLoadType.value,
            currentWeightSetup.value,
        );

    const draftStorageKey = computed(
        () => `gym-draft:v1:day:${dayId.value}:exercise:${exerciseIdentity.value}`,
    );

    const draftStorage = useWebStorageJsonSync<StoredDraft>({
        storage: window.sessionStorage,
        key: draftStorageKey,
        watchSources: [
            currentWeight,
            currentReps,
            currentWeightSetup,
            notes,
            draftDirty,
        ],
        getSnapshot: () => ({
            weight: currentWeight.value,
            reps: currentReps.value,
            weightSetup: currentWeightSetup.value,
            notes: notes.value,
            dirty: draftDirty.value,
        }),
        canPersist: () => exerciseIdentity.value !== 0,
        tryRestore: (parsed, { remove }) => {
            if (typeof parsed.weight !== "number") return false;
            if (parsed.dirty !== true) {
                remove();
                return false;
            }
            currentWeight.value = parsed.weight;
            currentReps.value = normalizeReps(parsed.reps ?? 0);
            currentWeightSetup.value = parsed.weightSetup ?? "";
            if (typeof parsed.notes === "string" && parsed.notes.length > 0) {
                notes.value = parsed.notes;
            }
            draftDirty.value = true;
            return true;
        },
    });

    const hasDraftSet = () =>
        currentWeight.value > 0 || currentReps.value > 0;

    const initializeExercise = () => {
        if (pending.value) return;

        const group = exerciseGroup.value;
        if (!group) {
            router.push(loggingRoute());
            return;
        }

        loggedSets.value = loggedSetsFromServer(group.logged?.sets);
        notes.value = group.logged?.notes ?? group.previous?.notes ?? "";
        currentSetNumber.value = loggedSets.value.length + 1;

        const { weight, reps, weightSetup } = initializeWeightAndReps(group.logged?.sets, group.previous?.sets);
        currentWeight.value = weight;
        currentReps.value = normalizeReps(reps);
        currentWeightSetup.value = weightSetup;

        draftDirty.value = false;

        draftStorage.restore();
        draftStorage.setSaveEnabled(true);
    };

    // React to async load + route param resolution: pending until data arrives;
    // exerciseGroup null when id is invalid or missing from the log. Then hydrate
    // or send the user back to the day’s exercise list.
    watch(
        [() => exerciseGroup.value, () => pending.value],
        ([group, isPending]) => {
            if (isPending) return;
            if (group) {
                initializeExercise();
            } else {
                router.push(loggingRoute());
            }
        },
        { immediate: true },
    );

    const saveCurrentExercise = async (includeDraft: boolean): Promise<boolean> => {
        const group = exerciseGroup.value;
        if (!group) return false;

        const draftTempId = `temp-current-${Date.now()}`;
        const draft =
            includeDraft && currentReps.value > 0
                ? {
                    weight: currentWeight.value,
                    reps: currentReps.value,
                    weight_setup: weightSetupForSave(),
                    tempId: draftTempId,
                }
                : null;

        const allSets = buildAllSetsForSave(loggedSets.value, draft);
        if (allSets.length === 0) return false;

        const payload = buildExerciseToLog(
            group,
            allSets,
            notes.value,
            dayId.value,
        );
        if (!payload) return false;

        const savedExercise = await logExercise(
            payload.exercise,
            payload.logType,
        );

        if (
            savedExercise &&
            savedExercise.sets &&
            Array.isArray(savedExercise.sets)
        ) {
            mergeSavedExerciseIntoLoggedSets(loggedSets.value, savedExercise);
            group.logged = savedExercise;
            return true;
        }

        markPendingSetsAsExerciseError(loggedSets.value);
        return false;
    };

    const stepWeight = (direction: "plus" | "minus") => {
        draftDirty.value = true;
        if (direction === "plus") {
            currentWeight.value = (currentWeight.value || 0) + 2.5;
        } else {
            currentWeight.value = Math.max(0, (currentWeight.value || 0) - 2.5);
        }
    };

    const stepReps = (direction: "plus" | "minus") => {
        draftDirty.value = true;
        if (direction === "plus") {
            currentReps.value = normalizeReps((currentReps.value || 0) + 1);
        } else {
            currentReps.value = normalizeReps((currentReps.value || 0) - 1);
        }
    };

    const commitWeightFromInput = (value: number) => {
        draftDirty.value = true;
        currentWeight.value = value;
    };

    const commitRepsFromInput = (value: number) => {
        draftDirty.value = true;
        currentReps.value = normalizeReps(value);
    };

    const addNextSet = async (): Promise<boolean> => {
        const reps = normalizeReps(currentReps.value);
        currentReps.value = reps;
        if (reps <= 0) {
            toast.push("Enter reps before logging the set", "error");
            return false;
        }
        globalTimer.start(REST_DURATION_MS, exerciseDisplayName.value);

        const tempId = `temp-${Date.now()}-${tempIdCounter++}`;
        const newSet: LoggedSetWithStatus = {
            weight: currentWeight.value,
            reps: currentReps.value,
            weight_setup: weightSetupForSave(),
            status: "pending",
            id: null,
            error: null,
            tempId,
        };

        loggedSets.value.push(newSet);
        currentSetNumber.value++;

        const success = await saveCurrentExercise(false);
        if (success) {
            draftDirty.value = false;
        }
        return success;
    };

    const finishLogging = async () => {
        if (loggedSets.value.length === 0 && !hasDraftSet()) {
            toast.push("Please log at least one set", "error");
            return;
        }

        if (currentWeight.value > 0 && currentReps.value === 0) {
            toast.push("Add reps or clear the draft set before finishing", "error");
            return;
        }

        const shouldIncludeDraft =
            currentReps.value > 0 &&
            (loggedSets.value.length === 0 || draftDirty.value);
        const success = await saveCurrentExercise(shouldIncludeDraft);

        if (!success) {
            toast.push("Failed to save exercise. Please try again.", "error");
            return;
        }

        draftStorage.clear();
        router.push(loggingRoute());
    };

    const retrySet = async (setIndex: number) => {
        const set = loggedSets.value[setIndex];
        if (!set || set.status !== "error") return;

        set.status = "pending";
        set.error = null;
        await saveCurrentExercise(false);
    };

    const deleteSet = async (setIndex: number) => {
        const set = loggedSets.value[setIndex];
        if (!set) return;

        if (set.status !== "success" || !set.id) {
            toast.push("Can only delete sets that have been saved", "error");
            return;
        }

        const setId = set.id;
        set.status = "pending";

        try {
            await deleteLoggedSet(setId);
            const updatedIndex = loggedSets.value.findIndex(
                (loggedSet) => loggedSet.id === setId,
            );
            if (updatedIndex !== -1) {
                loggedSets.value.splice(updatedIndex, 1);
            }
            currentSetNumber.value = loggedSets.value.length + 1;
        } catch {
            const currentSet = loggedSets.value.find(
                (loggedSet) => loggedSet.id === setId,
            );
            if (currentSet) {
                currentSet.status = "success";
            }
            toast.push("Failed to delete set. Please try again.", "error");
        }
    };

    const editSet = (setIndex: number) => {
        const set = loggedSets.value[setIndex];
        if (!set) return;

        if (set.status !== "success") {
            toast.push("Can only edit sets that have been saved", "error");
            return;
        }

        currentWeight.value = set.weight;
        currentReps.value = normalizeReps(set.reps);
        currentWeightSetup.value = set.weight_setup;
        draftDirty.value = true;

        loggedSets.value.splice(setIndex, 1);
        currentSetNumber.value = loggedSets.value.length + 1;
    };

    const goBackToList = () => {
        router.push(loggingRoute());
    };

    const updateNotes = (value: string) => {
        notes.value = value;
        draftDirty.value = true;
    };

    const updateWeightSetup = (value: string) => {
        currentWeightSetup.value = value;
        draftDirty.value = true;
    };

    return reactive({
        exerciseGroup,
        currentSetNumber,
        loggedSets,
        currentWeight,
        currentReps,
        currentWeightSetup,
        notes,
        stepWeight,
        stepReps,
        commitWeightFromInput,
        commitRepsFromInput,
        addNextSet,
        finishLogging,
        retrySet,
        deleteSet,
        editSet,
        goBackToList,
        updateNotes,
        updateWeightSetup,
    }) as ExerciseLoggingSessionViewModel;
}
