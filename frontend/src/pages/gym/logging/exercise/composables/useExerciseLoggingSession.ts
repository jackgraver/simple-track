import type { ComputedRef, Ref } from "vue";
import { computed, reactive, ref, watch } from "vue";
import type { Router } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import { buildLoggingListQuery } from "../../composables/useLoggingRouteContext";
import type {
    ExerciseGroup,
    LoggedSetWithStatus,
} from "../../store/useWorkoutStore";
import { toast } from "~/composables/toast/useToast";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { useWebStorageJsonSync } from "~/composables/useWebStorageJsonSync";
import { newClientId } from "~/api/ids";
import { offsetToDate } from "~/api/workout/dates";
import {
    findFailedSetVars,
    notesSavePending,
    useDeleteSet,
    useLogSet,
    useSaveExerciseNotes,
    useSetIntents,
    useUpdateSet,
} from "~/api/workout/queries";
import { deriveSetRows } from "~/api/workout/setsCache";
import {
    DEFAULT_REPS_AFTER_WEIGHT_CHANGE,
    initializeWeightAndReps,
    previousRepsForSetAtWeight,
} from "../domain/loggedSetDefaults";
import { weightSetupForExercise } from "../domain/weightSetup";
import {
    getWeightProgression,
    type WeightProgression,
    WEIGHT_PROGRESSION_INCREASE_LBS,
} from "../domain/weightProgression";
import { useGymWeightPrefs } from "./useGymWeightPrefs";

type StoredDraft = {
    weight: number;
    reps: number;
    weightSetup: string;
    notes: string;
    dirty: boolean;
    editingSetIndex: number | null;
};

/** Plain shape for templates (refs unwrapped via reactive session object). */
export type ExerciseLoggingSessionViewModel = {
    exerciseGroup: ExerciseGroup | null;
    currentSetNumber: number;
    /** Derived from the query cache + pending/failed set mutations; never a local copy. */
    loggedSets: LoggedSetWithStatus[];
    currentWeight: number;
    currentReps: number;
    currentWeightSetup: string;
    notes: string;
    weightProgression: WeightProgression | null;
    editingSetIndex: number | null;
    weightIncrement: number;
    stepWeight: (direction: "plus" | "minus") => void;
    stepReps: (direction: "plus" | "minus") => void;
    commitWeightFromInput: (value: number) => void;
    commitRepsFromInput: (value: number) => void;
    addNextSet: () => boolean;
    finishLogging: () => boolean;
    retrySet: (setIndex: number) => void;
    deleteSet: (setIndex: number) => void;
    editSet: (setIndex: number) => void;
    cancelEditingSet: () => void;
    goBackToList: () => void;
    updateNotes: (value: string) => void;
    updateWeightSetup: (value: string) => void;
    undoWeightProgression: () => void;
    setWeightIncrement: (value: number) => void;
};

/**
 * The lift → log → rest loop for one exercise. Sets are written per set and optimistically
 * (optimistic-ui G2): the row appears on tap, the rest clock starts on tap, and a failed save stays
 * in the list marked "Didn't save" with an inline retry. The only local state is the draft (weight,
 * reps, setup, notes, which set is being edited).
 */
export function useExerciseLoggingSession(options: {
    exerciseGroup: ComputedRef<ExerciseGroup | null>;
    pending: Ref<boolean>;
    offset: ComputedRef<number>;
    router: Router;
}): ExerciseLoggingSessionViewModel {
    const { exerciseGroup, pending, offset, router } = options;
    const loggingRoute = () => ({
        name: "logging" as const,
        query: buildLoggingListQuery(offset.value),
    });

    const qc = useQueryClient();
    const logSet = useLogSet();
    const updateSet = useUpdateSet();
    const deleteSetMutation = useDeleteSet();
    const saveNotes = useSaveExerciseNotes();
    const intents = useSetIntents();

    /** The day this page writes to. Read once per action, so a write keeps its day. */
    const date = computed(() => offsetToDate(offset.value));

    const currentWeight = ref(0);
    const currentReps = ref(0);
    const currentWeightSetup = ref("");
    const weightProgression = ref<WeightProgression | null>(null);
    const editingSetIndex = ref<number | null>(null);
    const { weightIncrement, setWeightIncrement } = useGymWeightPrefs();

    const draftDirty = ref(false);
    const notes = ref("");
    const globalTimer = useGlobalRestTimer();
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

    const loggedSets = computed<LoggedSetWithStatus[]>(() =>
        deriveSetRows(exerciseGroup.value?.logged?.sets, intents.value, {
            date: date.value,
            exercise_id: exerciseIdentity.value,
        }),
    );

    const currentSetNumber = computed(() =>
        editingSetIndex.value != null
            ? editingSetIndex.value + 1
            : loggedSets.value.length + 1,
    );

    const exerciseDisplayName = computed(() => {
        const group = exerciseGroup.value;
        return group?.planned?.name ?? group?.logged?.exercise?.name ?? "";
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
        () =>
            `gym-draft:v2:date:${date.value}:exercise:${exerciseIdentity.value}`,
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
            editingSetIndex,
        ],
        getSnapshot: () => ({
            weight: currentWeight.value,
            reps: currentReps.value,
            weightSetup: currentWeightSetup.value,
            notes: notes.value,
            dirty: draftDirty.value,
            editingSetIndex: editingSetIndex.value,
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
            editingSetIndex.value =
                typeof parsed.editingSetIndex === "number" &&
                Number.isInteger(parsed.editingSetIndex) &&
                parsed.editingSetIndex >= 0
                    ? parsed.editingSetIndex
                    : null;
            return true;
        },
    });

    /** Fill the draft for a newly opened exercise. Cache updates afterwards (optimistic patches,
     *  reconciles, refetches) only change the rows, never the draft. */
    const initializeExercise = () => {
        const group = exerciseGroup.value;
        if (!group) return;

        notes.value = group.logged?.notes ?? group.previous?.notes ?? "";
        editingSetIndex.value = null;

        const { weight, reps, weightSetup } = initializeWeightAndReps(
            group.logged?.sets,
            group.previous?.sets,
        );
        weightProgression.value =
            (group.logged?.sets?.length ?? 0) === 0
                ? getWeightProgression(
                      group.previous?.sets,
                      group.previous?.exercise?.rep_rollover,
                      weight,
                      WEIGHT_PROGRESSION_INCREASE_LBS,
                  )
                : null;
        currentWeight.value = weightProgression.value?.nextWeight ?? weight;
        currentReps.value =
            currentWeight.value === weight
                ? normalizeReps(reps)
                : DEFAULT_REPS_AFTER_WEIGHT_CHANGE;
        currentWeightSetup.value = weightSetup;

        draftDirty.value = false;

        if (draftStorage.restore()) {
            weightProgression.value = null;
            if (
                editingSetIndex.value != null &&
                loggedSets.value[editingSetIndex.value]?.status !== "success"
            ) {
                editingSetIndex.value = null;
            }
        }
        draftStorage.setSaveEnabled(true);
    };

    // Once the day has loaded, a missing group means a bad id (or the exercise left the log):
    // back to the list.
    watch(
        [() => exerciseGroup.value, () => pending.value],
        ([group, isPending]) => {
            if (!isPending && !group) router.push(loggingRoute());
        },
        { immediate: true },
    );

    // Initialise once per (day, exercise), not on every cache change.
    const sessionKey = computed(() =>
        pending.value || !exerciseGroup.value
            ? null
            : `${date.value}:${exerciseIdentity.value}`,
    );
    watch(
        sessionKey,
        (key) => {
            if (key) initializeExercise();
        },
        { immediate: true },
    );

    /** Save the notes when they differ from the cached exercise's (no-op otherwise). */
    const syncNotes = (day: string, exerciseId: number) => {
        const saved = exerciseGroup.value?.logged?.notes ?? "";
        const vars = { date: day, exercise_id: exerciseId, notes: notes.value };
        if (notes.value === saved || notesSavePending(qc, vars)) return;
        // Without a logged exercise there is nothing to attach notes to yet; the first set does it.
        if (!exerciseGroup.value?.logged && loggedSets.value.length === 0) return;
        saveNotes.mutate(vars);
    };

    const updateWeightAndDefaultReps = (value: number) => {
        if (value === currentWeight.value) return;
        draftDirty.value = true;
        currentWeight.value = value;
        currentReps.value = normalizeReps(
            previousRepsForSetAtWeight(
                exerciseGroup.value?.previous?.sets,
                currentSetNumber.value - 1,
                value,
            ) ?? DEFAULT_REPS_AFTER_WEIGHT_CHANGE,
        );
    };

    /** Point the reps slider at last session's reps for set `index` (0-based), when there is one. */
    const resetRepsForSet = (index: number) => {
        const previous = previousRepsForSetAtWeight(
            exerciseGroup.value?.previous?.sets,
            index,
            currentWeight.value,
        );
        if (previous != null) currentReps.value = normalizeReps(previous);
    };

    const stepWeight = (direction: "plus" | "minus") => {
        if (direction === "plus") {
            updateWeightAndDefaultReps(
                (currentWeight.value || 0) + weightIncrement.value,
            );
        } else {
            updateWeightAndDefaultReps(
                Math.max(0, (currentWeight.value || 0) - weightIncrement.value),
            );
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
        updateWeightAndDefaultReps(value);
    };

    const commitRepsFromInput = (value: number) => {
        draftDirty.value = true;
        currentReps.value = normalizeReps(value);
    };

    /**
     * Log the next set, or save the set being edited. Returns at once: the row is in the list
     * before the request leaves, and the rest clock starts on tap (not on save).
     */
    const addNextSet = (): boolean => {
        const reps = normalizeReps(currentReps.value);
        currentReps.value = reps;
        if (reps <= 0) {
            toast.push("Enter reps before logging the set", "error");
            return false;
        }
        const exerciseId = exerciseIdentity.value;
        if (!exerciseId) return false;
        const day = date.value;
        const fields = {
            weight: currentWeight.value,
            reps,
            weight_setup: weightSetupForSave(),
        };
        const count = loggedSets.value.length;

        const editingIndex = editingSetIndex.value;
        if (editingIndex != null) {
            const row = loggedSets.value[editingIndex];
            if (!row) return false;
            updateSet.mutate({ client_id: row.key, date: day, exercise_id: exerciseId, ...fields });
            // A fix to an earlier set isn't a new set: the rest clock keeps running.
        } else {
            logSet.mutate({
                client_id: newClientId(),
                date: day,
                exercise_id: exerciseId,
                created_at: new Date().toISOString(),
                ...fields,
            });
            globalTimer.start(exerciseDisplayName.value);
        }
        syncNotes(day, exerciseId);

        draftDirty.value = false;
        editingSetIndex.value = null;
        // The next set is index `count` (+1 if we just added one); the rows update a tick later.
        resetRepsForSet(editingIndex != null ? count : count + 1);
        return true;
    };

    /**
     * Sets are saved as they're logged, so finishing only saves changed notes and drops the draft.
     * The slider position is never logged here. The caller decides where to go.
     */
    const finishLogging = (): boolean => {
        const rows = loggedSets.value;
        if (rows.length === 0) {
            toast.push("Log at least one set first", "error");
            return false;
        }
        const unsaved = rows.filter((r) => r.status === "error").length;
        if (unsaved > 0) {
            toast.push(
                unsaved === 1
                    ? "One set didn't save. Retry or delete it first."
                    : `${unsaved} sets didn't save. Retry or delete them first.`,
                "error",
            );
            return false;
        }

        editingSetIndex.value = null;
        if (exerciseIdentity.value) syncNotes(date.value, exerciseIdentity.value);
        draftDirty.value = false;
        draftStorage.clear();
        return true;
    };

    const retrySet = (setIndex: number) => {
        const row = loggedSets.value[setIndex];
        if (!row || row.status !== "error") return;
        const failed = findFailedSetVars(qc, row.key);
        if (failed?.kind === "create") logSet.retry(failed.vars);
        else if (failed?.kind === "update") updateSet.retry(failed.vars);
    };

    /** Delete any row, saved or not. A failed write is dropped first; the server delete is
     *  idempotent (by client_id), so it is safe even if the failed create did reach the server. */
    const deleteSet = (setIndex: number) => {
        const row = loggedSets.value[setIndex];
        const exerciseId = exerciseIdentity.value;
        if (!row || !exerciseId) return;
        const failed = findFailedSetVars(qc, row.key);
        if (failed?.kind === "create") logSet.dismiss(failed.vars);
        else if (failed?.kind === "update") updateSet.dismiss(failed.vars);
        if (editingSetIndex.value === setIndex) editingSetIndex.value = null;
        else if (editingSetIndex.value != null && editingSetIndex.value > setIndex) editingSetIndex.value--;
        deleteSetMutation.mutate({ client_id: row.key, date: date.value, exercise_id: exerciseId });
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

        editingSetIndex.value = setIndex;
    };

    const cancelEditingSet = () => {
        if (editingSetIndex.value == null) return;
        editingSetIndex.value = null;
        const last = loggedSets.value[loggedSets.value.length - 1];
        if (last) currentWeight.value = last.weight;
        resetRepsForSet(loggedSets.value.length);
    };

    const goBackToList = () => {
        router.push(loggingRoute());
    };

    const updateNotes = (value: string) => {
        notes.value = value;
        draftDirty.value = true;
        if (exerciseIdentity.value) syncNotes(date.value, exerciseIdentity.value);
    };

    const updateWeightSetup = (value: string) => {
        currentWeightSetup.value = value;
        draftDirty.value = true;
    };

    const undoWeightProgression = () => {
        if (!weightProgression.value) return;
        updateWeightAndDefaultReps(weightProgression.value.previousWeight);
        weightProgression.value = null;
    };

    return reactive({
        exerciseGroup,
        currentSetNumber,
        loggedSets,
        currentWeight,
        currentReps,
        currentWeightSetup,
        notes,
        weightProgression,
        editingSetIndex,
        weightIncrement,
        stepWeight,
        stepReps,
        commitWeightFromInput,
        commitRepsFromInput,
        addNextSet,
        finishLogging,
        retrySet,
        deleteSet,
        editSet,
        cancelEditingSet,
        goBackToList,
        updateNotes,
        updateWeightSetup,
        undoWeightProgression,
        setWeightIncrement,
    }) as ExerciseLoggingSessionViewModel;
}
