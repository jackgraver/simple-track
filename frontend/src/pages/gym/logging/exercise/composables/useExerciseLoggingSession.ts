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
    findFailedSetWrites,
    notesSavePending,
    useDeleteSet,
    useLogSet,
    useSaveExerciseNotes,
    useSetIntents,
    useUpdateSet,
} from "~/api/workout/queries";
import { deriveSetRows, planSetRetry, type SetFields } from "~/api/workout/setsCache";
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
    /** Before the first set: go up to the earned weight, or stay at last session's. Reversible. */
    chooseWeightProgression: (choice: "up" | "stay") => void;
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

    /**
     * The day this page reads and writes. A computed over the route offset only, so it is fixed when
     * the page opens (it stays on that day past midnight, like the day query's key). Each action
     * copies it into its mutation variables, and the server reads it as its own local calendar day.
     */
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
            // Coming back before set 1 with one of the two weights (go up / stay): keep the choice.
            const p = weightProgression.value;
            const stillChoosing =
                p != null && loggedSets.value.length === 0 && [p.previousWeight, p.nextWeight].includes(currentWeight.value);
            if (!stillChoosing) weightProgression.value = null;
            if (editingSetIndex.value != null && !loggedSets.value[editingSetIndex.value]) {
                editingSetIndex.value = null;
            }
        }
        draftStorage.setSaveEnabled(true);
    };

    // Once the day has loaded, a missing group means a bad id or the exercise left the log (e.g. an
    // add from the list that failed and rolled back; its toast says so): back to the list, replacing
    // this entry so Back doesn't return to a page that no longer exists.
    watch(
        [() => exerciseGroup.value, () => pending.value],
        ([group, isPending]) => {
            if (!isPending && !group) void router.replace(loggingRoute());
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

    // An exercise added from the list opens before its history is in the cache (it arrives when
    // the day refetches after the add). Fill the draft again then, unless logging has started.
    watch(
        () => exerciseGroup.value?.previous,
        (previous, before) => {
            if (!previous || before || sessionKey.value == null) return;
            if (draftDirty.value || editingSetIndex.value != null || loggedSets.value.length > 0) return;
            initializeExercise();
        },
    );

    /**
     * Save the notes when they differ from the cached exercise's (no-op otherwise). `afterSet`: a set
     * was just logged, so the save may go out even though the cache doesn't show a logged exercise
     * yet (the create's patch lands a tick later). It queues behind that create in the exercise's
     * scope, and reads the saved exercise when it is sent.
     */
    const syncNotes = (day: string, exerciseId: number, afterSet = false) => {
        const saved = exerciseGroup.value?.logged?.notes ?? "";
        const vars = { date: day, exercise_id: exerciseId, notes: notes.value };
        if (notes.value === saved || notesSavePending(qc, vars)) return;
        // Without a logged exercise there is nothing to attach notes to yet; the first set does it.
        if (!afterSet && !exerciseGroup.value?.logged && loggedSets.value.length === 0) return;
        saveNotes.mutate(vars);
    };

    /**
     * Change the weight. Before the first set the reps follow last session's set 1 at that weight;
     * once this session has a set (or one is being fixed) the slider holds this session's reps and
     * a weight change leaves it where it is.
     */
    const updateWeightAndDefaultReps = (value: number) => {
        if (value === currentWeight.value) return;
        draftDirty.value = true;
        currentWeight.value = value;
        if (editingSetIndex.value != null || loggedSets.value.length > 0) return;
        currentReps.value = normalizeReps(
            previousRepsForSetAtWeight(
                exerciseGroup.value?.previous?.sets,
                currentSetNumber.value - 1,
                value,
            ) ?? DEFAULT_REPS_AFTER_WEIGHT_CHANGE,
        );
    };

    /**
     * Point the draft at this session's last set (or `last`), after editing a set. The next set
     * starts from the previous set of this session; last session's reps live in the "Last time"
     * marker, not the slider.
     */
    const resumeFromLastSet = (last?: SetFields) => {
        const row = last ?? loggedSets.value[loggedSets.value.length - 1];
        if (!row) return;
        currentWeight.value = row.weight;
        currentReps.value = normalizeReps(row.reps);
        currentWeightSetup.value = row.weight_setup;
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
            saveEditedSet(row, day, exerciseId, fields);
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
        syncNotes(day, exerciseId, true);

        draftDirty.value = false;
        editingSetIndex.value = null;
        // A new set leaves the draft at what was just logged. After fixing a set, go back to this
        // session's last set (the edited values if that was the last one: rows update a tick later).
        if (editingIndex != null) resumeFromLastSet(editingIndex === count - 1 ? fields : undefined);
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

    /** Re-send a failed row: its create (with the newest values) if that failed, else its edit. */
    const runRetry = (key: string, fields?: SetFields): boolean => {
        const plan = planSetRetry(findFailedSetWrites(qc, key), fields);
        if (plan?.kind === "create") {
            if (plan.dismissUpdate) updateSet.dismiss(plan.dismissUpdate);
            logSet.retry(plan.vars);
        } else if (plan?.kind === "update") {
            updateSet.retry(plan.vars);
        }
        return plan != null;
    };

    /**
     * Save an edit to any row. Saved or still being created: a PATCH by client_id, which waits in the
     * exercise's scope for the create (offline, both wait for the connection). Failed: re-send the
     * failed write with the new values instead, since an edit behind a failed create would fail too.
     */
    const saveEditedSet = (row: LoggedSetWithStatus, day: string, exerciseId: number, fields: SetFields) => {
        if (row.status === "error" && runRetry(row.key, fields)) return;
        updateSet.mutate({ client_id: row.key, date: day, exercise_id: exerciseId, ...fields });
    };

    const retrySet = (setIndex: number) => {
        const row = loggedSets.value[setIndex];
        if (!row || row.status !== "error") return;
        runRetry(row.key);
    };

    /**
     * Delete any row, saved or not. The server delete is idempotent (by client_id), so it is safe
     * even if a failed create did reach the server. `useDeleteSet` dismisses the row's failed writes
     * only once the delete succeeds: that also covers a create that fails while the delete waits
     * behind it, and a failed delete leaves an unsaved row marked "Didn't save".
     */
    const deleteSet = (setIndex: number) => {
        const row = loggedSets.value[setIndex];
        const exerciseId = exerciseIdentity.value;
        if (!row || !exerciseId) return;
        if (editingSetIndex.value === setIndex) editingSetIndex.value = null;
        else if (editingSetIndex.value != null && editingSetIndex.value > setIndex) editingSetIndex.value--;
        deleteSetMutation.mutate({ client_id: row.key, date: date.value, exercise_id: exerciseId });
    };

    const editSet = (setIndex: number) => {
        const set = loggedSets.value[setIndex];
        if (!set) return;

        currentWeight.value = set.weight;
        currentReps.value = normalizeReps(set.reps);
        currentWeightSetup.value = set.weight_setup;
        draftDirty.value = true;

        editingSetIndex.value = setIndex;
    };

    const cancelEditingSet = () => {
        if (editingSetIndex.value == null) return;
        editingSetIndex.value = null;
        resumeFromLastSet();
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

    // Keeps `weightProgression` so the choice can be flipped back until the first set is logged.
    const chooseWeightProgression = (choice: "up" | "stay") => {
        const p = weightProgression.value;
        if (!p || loggedSets.value.length > 0) return;
        updateWeightAndDefaultReps(choice === "up" ? p.nextWeight : p.previousWeight);
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
        chooseWeightProgression,
        setWeightIncrement,
    }) as ExerciseLoggingSessionViewModel;
}
