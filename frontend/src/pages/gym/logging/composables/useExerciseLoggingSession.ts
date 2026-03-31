import type { ComputedRef, Ref } from "vue";
import { computed, reactive, ref, watch } from "vue";
import type { Router } from "vue-router";
import type { ExerciseGroup, LoggedSetWithStatus } from "../store/useWorkoutStore";
import type { LoggedExercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import {
    buildAllSetsForSave,
    buildExerciseToLog,
    mergeSavedExerciseIntoLoggedSets,
    markPendingSetsAsExerciseError,
} from "../domain/exerciseLoggingPayload";

/** Plain shape for templates (refs unwrapped via reactive session object). */
export type ExerciseLoggingSessionViewModel = {
    exerciseGroup: ExerciseGroup | null;
    currentSetNumber: number;
    loggedSets: LoggedSetWithStatus[];
    currentWeight: number;
    currentReps: number;
    currentWeightSetup: string;
    weightEditMode: boolean;
    repsEditMode: boolean;
    weightInputValue: string;
    repsInputValue: string;
    notes: string;
    restTimerStorageKey: string;
    restTimerStartToken: number;
    restTimerClearToken: number;
    restTimerDurationMs: number;
    incrementWeight: () => void;
    decrementWeight: () => void;
    incrementReps: () => void;
    decrementReps: () => void;
    enterWeightEditMode: () => void;
    exitWeightEditMode: () => void;
    enterRepsEditMode: () => void;
    exitRepsEditMode: () => void;
    addNextSet: () => Promise<void>;
    finishLogging: () => Promise<void>;
    retrySet: (setIndex: number) => Promise<void>;
    deleteSet: (setIndex: number) => Promise<void>;
    editSet: (setIndex: number) => void;
    goBackToList: () => void;
    updateNotes: (value: string) => void;
    updateWeightSetup: (value: string) => void;
    updateWeightInputValue: (value: string) => void;
    updateRepsInputValue: (value: string) => void;
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
        query: offset.value === 0 ? {} : { offset: String(offset.value) },
    });

    const currentWeight = ref(0);
    const currentReps = ref(0);
    const currentWeightSetup = ref("");
    const currentSetNumber = ref(1);

    const loggedSets = ref<LoggedSetWithStatus[]>([]);
    let tempIdCounter = 0;

    const weightEditMode = ref(false);
    const repsEditMode = ref(false);
    const weightInputValue = ref("");
    const repsInputValue = ref("");
    const draftDirty = ref(false);
    const notes = ref("");
    const restTimerStartToken = ref(0);
    const restTimerClearToken = ref(0);
    const restTimerDurationMs = 2 * 60 * 1000;

    const restTimerStorageKey = computed(() => {
        const group = exerciseGroup.value;
        const exerciseIdentity =
            group?.planned?.ID ??
            group?.logged?.exercise_id ??
            group?.logged?.exercise?.ID ??
            0;

        return `gym-rest-timer:v1:day:${dayId.value}:exercise:${exerciseIdentity}`;
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

        if (
            group.logged &&
            group.logged.sets &&
            Array.isArray(group.logged.sets) &&
            group.logged.sets.length > 0
        ) {
            loggedSets.value = group.logged.sets.map((set) => ({
                weight: set.weight,
                reps: set.reps,
                weight_setup: set.weight_setup || "",
                status: "success" as const,
                id: set.ID,
                error: null,
                tempId: `existing-${set.ID}`,
            }));
            currentSetNumber.value = loggedSets.value.length + 1;

            const lastSet = group.logged.sets[group.logged.sets.length - 1];
            if (lastSet) {
                currentWeight.value = lastSet.weight;
                currentReps.value = lastSet.reps;
                currentWeightSetup.value = lastSet.weight_setup || "";
            } else {
                currentWeight.value = 0;
                currentReps.value = 0;
                currentWeightSetup.value = "";
            }

            notes.value = group.logged.notes || "";
        } else {
            loggedSets.value = [];
            currentSetNumber.value = 1;

            if (
                group.previous &&
                group.previous.sets &&
                Array.isArray(group.previous.sets) &&
                group.previous.sets.length > 0
            ) {
                const firstSet = group.previous.sets[0];
                currentWeight.value = firstSet ? firstSet.weight : 0;
                currentReps.value = firstSet ? firstSet.reps : 0;
                currentWeightSetup.value = firstSet
                    ? firstSet.weight_setup || ""
                    : "";
            } else {
                currentWeight.value = 0;
                currentReps.value = 0;
                currentWeightSetup.value = "";
            }

            notes.value = group.previous?.notes || "";
        }

        weightEditMode.value = false;
        repsEditMode.value = false;
        draftDirty.value = false;
    };

    watch(
        [() => exerciseGroup.value, () => pending.value],
        ([group, isPending]) => {
            if (!isPending && group) {
                initializeExercise();
            } else if (!isPending && !group) {
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
                    weight_setup: currentWeightSetup.value,
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

    const incrementWeight = () => {
        draftDirty.value = true;
        currentWeight.value = (currentWeight.value || 0) + 2.5;
    };

    const decrementWeight = () => {
        draftDirty.value = true;
        currentWeight.value = Math.max(0, (currentWeight.value || 0) - 2.5);
    };

    const incrementReps = () => {
        draftDirty.value = true;
        currentReps.value = (currentReps.value || 0) + 1;
    };

    const decrementReps = () => {
        draftDirty.value = true;
        currentReps.value = Math.max(0, (currentReps.value || 0) - 1);
    };

    const enterWeightEditMode = () => {
        weightEditMode.value = true;
        weightInputValue.value = currentWeight.value.toString();
    };

    const exitWeightEditMode = () => {
        const trimmedValue = weightInputValue.value.trim();
        if (trimmedValue !== "") {
            const n = Number(trimmedValue);
            if (Number.isFinite(n) && n >= 0) {
                currentWeight.value = n;
                draftDirty.value = true;
            }
        }
        weightEditMode.value = false;
    };

    const enterRepsEditMode = () => {
        repsEditMode.value = true;
        repsInputValue.value = currentReps.value.toString();
    };

    const exitRepsEditMode = () => {
        const trimmedValue = repsInputValue.value.trim();
        if (trimmedValue !== "") {
            const n = Number(trimmedValue);
            if (Number.isInteger(n) && n >= 0) {
                currentReps.value = n;
                draftDirty.value = true;
            }
        }
        repsEditMode.value = false;
    };

    const addNextSet = async () => {
        if (currentReps.value <= 0) {
            toast.push("Enter reps before logging the set", "error");
            return;
        }
        restTimerStartToken.value += 1;

        const tempId = `temp-${Date.now()}-${tempIdCounter++}`;
        const newSet: LoggedSetWithStatus = {
            weight: currentWeight.value,
            reps: currentReps.value,
            weight_setup: currentWeightSetup.value,
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

        restTimerClearToken.value += 1;
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
        currentReps.value = set.reps;
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
    };

    const updateWeightSetup = (value: string) => {
        currentWeightSetup.value = value;
        draftDirty.value = true;
    };

    const updateWeightInputValue = (value: string) => {
        weightInputValue.value = value;
    };

    const updateRepsInputValue = (value: string) => {
        repsInputValue.value = value;
    };

    return reactive({
        exerciseGroup,
        currentSetNumber,
        loggedSets,
        currentWeight,
        currentReps,
        currentWeightSetup,
        weightEditMode,
        repsEditMode,
        weightInputValue,
        repsInputValue,
        notes,
        restTimerStorageKey,
        restTimerStartToken,
        restTimerClearToken,
        restTimerDurationMs,
        incrementWeight,
        decrementWeight,
        incrementReps,
        decrementReps,
        enterWeightEditMode,
        exitWeightEditMode,
        enterRepsEditMode,
        exitRepsEditMode,
        addNextSet,
        finishLogging,
        retrySet,
        deleteSet,
        editSet,
        goBackToList,
        updateNotes,
        updateWeightSetup,
        updateWeightInputValue,
        updateRepsInputValue,
    }) as ExerciseLoggingSessionViewModel;
}
