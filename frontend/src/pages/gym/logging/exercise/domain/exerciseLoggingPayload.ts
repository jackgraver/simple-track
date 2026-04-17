import type { LoggedExercise, LoggedSet } from "~/types/workout";
import type { ExerciseGroup, LoggedSetWithStatus } from "../../store/useWorkoutStore";

export function setsMatchLocalToSaved(
    local: Pick<LoggedSetWithStatus, "weight" | "reps" | "weight_setup">,
    saved: Pick<LoggedSet, "weight" | "reps" | "weight_setup">,
): boolean {
    return (
        Math.abs(local.weight - saved.weight) < 0.01 &&
        local.reps === saved.reps &&
        local.weight_setup === (saved.weight_setup || "")
    );
}

/** Hydrate the session list from API `LoggedSet` rows when opening an exercise with prior saves. */
export function loggedSetsFromServer(
    sets: readonly LoggedSet[] | undefined,
): LoggedSetWithStatus[] {
    if (!sets?.length) return [];
    return sets.map((s, i) => ({
        weight: s.weight,
        reps: s.reps,
        weight_setup: s.weight_setup || "",
        status: "success" as const,
        id: s.ID > 0 ? s.ID : null,
        error: null,
        tempId: s.ID > 0 ? `saved-${s.ID}` : `saved-row-${i}`,
    }));
}

/** Combines committed rows with an optional in-progress row for a save request. */
export function buildAllSetsForSave(
    loggedSets: LoggedSetWithStatus[],
    draft: {
        weight: number;
        reps: number;
        weight_setup: string;
        tempId: string;
    } | null,
): LoggedSetWithStatus[] {
    const all = [...loggedSets];
    if (
        draft &&
        draft.reps > 0
    ) {
        all.push({
            weight: draft.weight,
            reps: draft.reps,
            weight_setup: draft.weight_setup,
            status: "pending",
            id: null,
            error: null,
            tempId: draft.tempId,
        });
    }
    return all;
}

export function buildExerciseToLog(
    group: ExerciseGroup,
    allSets: LoggedSetWithStatus[],
    notes: string,
    dayId: number,
): { exercise: LoggedExercise; logType: "logged" | "previous" } | null {
    const filtered = allSets.filter(
        (set) => !(set.reps === 0 && set.weight === 0),
    );
    if (filtered.length === 0) return null;

    const exerciseId =
        group.logged?.exercise_id ||
        group.logged?.exercise?.ID ||
        group.previous?.exercise_id ||
        group.previous?.exercise?.ID ||
        group.planned?.ID ||
        0;

    if (!exerciseId) return null;

    let exerciseToLog: LoggedExercise;
    if (group.logged && group.logged.ID > 0) {
        exerciseToLog = {
            ...group.logged,
            exercise_id: exerciseId,
        };
    } else if (group.previous) {
        exerciseToLog = {
            ...group.previous,
            ID: 0,
            workout_log_id: dayId,
            exercise_id: exerciseId,
            exercise: group.previous.exercise || group.planned,
            sets: [],
        };
    } else {
        exerciseToLog = {
            ID: 0,
            workout_log_id: dayId,
            exercise_id: exerciseId,
            exercise: group.planned,
            sets: [],
            notes: "",
            created_at: "",
            updated_at: "",
        };
    }

    exerciseToLog.sets = filtered.map((set) => ({
        logged_exercise_id: exerciseToLog.ID,
        reps: set.reps,
        weight: set.weight,
        weight_setup: set.weight_setup || "",
        ID: set.id || 0,
        created_at: "",
        updated_at: "",
    }));

    exerciseToLog.notes = notes;
    exerciseToLog.workout_log_id = dayId || exerciseToLog.workout_log_id;

    const logType: "logged" | "previous" =
        group.logged && group.logged.ID > 0 ? "logged" : "previous";

    return { exercise: exerciseToLog, logType };
}

/** Updates pending/success rows from server response; mutates `loggedSets` in place. */
export function mergeSavedExerciseIntoLoggedSets(
    loggedSets: LoggedSetWithStatus[],
    savedExercise: LoggedExercise,
): void {
    if (!savedExercise.sets || !Array.isArray(savedExercise.sets)) return;

    const pendingSets = loggedSets.filter((s) => s.status === "pending");

    savedExercise.sets.forEach((savedSet) => {
        if (!savedSet) return;

        const pendingSet = pendingSets.find((ps) =>
            setsMatchLocalToSaved(ps, savedSet),
        );

        if (pendingSet) {
            const setIndex = loggedSets.findIndex(
                (s) => s.tempId === pendingSet.tempId,
            );
            if (setIndex !== -1 && savedSet.ID) {
                const setToUpdate = loggedSets[setIndex];
                if (setToUpdate) {
                    setToUpdate.status = "success";
                    setToUpdate.id = savedSet.ID;
                    setToUpdate.error = null;
                }
            }
        } else if (savedSet.ID) {
            const existingSet = loggedSets.find((s) => s.id === savedSet.ID);
            if (existingSet) {
                existingSet.status = "success";
                existingSet.error = null;
            }
        }
    });

    pendingSets.forEach((ps) => {
        const wasMatched = savedExercise.sets.some(
            (savedSet) =>
                savedSet && setsMatchLocalToSaved(ps, savedSet),
        );
        if (!wasMatched && ps.status === "pending") {
            ps.status = "error";
            ps.error = "Failed to save set";
        }
    });
}

export function markPendingSetsAsExerciseError(
    loggedSets: LoggedSetWithStatus[],
): void {
    loggedSets.forEach((set) => {
        if (set.status === "pending") {
            set.status = "error";
            set.error = "Failed to save exercise";
        }
    });
}
