import type { MachineSetting } from "~/types/workout";
import type { ExerciseGroup } from "../../store/useWorkoutStore";
import { formatSets, type WorkoutStep } from "../../domain/workoutSteps";
import { initializeWeightAndReps } from "./loggedSetDefaults";
import { getWeightProgression, WEIGHT_PROGRESSION_INCREASE_LBS } from "./weightProgression";
import { overloadStatus, visibleMachineSettings, type OverloadStatus } from "./liveLogging";

export type UpNextItem = {
    key: string;
    kind: WorkoutStep["kind"];
    name: string;
    /** Lifts: the weight the logging screen will open with (after any auto-increase). */
    weight: number | null;
    overload: OverloadStatus;
    /** "Last 100 lb × 10, 10" for lifts, the step detail otherwise. */
    detail: string | null;
    machineSettings: MachineSetting[];
    exerciseId: number | null;
};

const exerciseIdOf = (g: ExerciseGroup | undefined) => g?.planned?.ID ?? g?.logged?.exercise_id ?? null;

/** Same rule the logging screen applies on open: last session's final set weight, +5 if earned (see `sessionOutcome`). */
export function openingWeight(group: ExerciseGroup): number {
    const { weight } = initializeWeightAndReps(group.logged?.sets, group.previous?.sets);
    if (group.logged?.sets?.length) return weight;
    const progression = getWeightProgression(
        group.previous?.sets,
        group.previous?.exercise?.rep_rollover,
        WEIGHT_PROGRESSION_INCREASE_LBS,
    );
    return progression?.nextWeight ?? weight;
}

/**
 * What's left after finishing `currentExerciseId`: every unfinished step except the current one, in plan
 * order. Skipping one is just picking another; it stays in the list until it's done.
 */
export function buildUpNext(
    steps: readonly WorkoutStep[],
    log: readonly ExerciseGroup[],
    currentExerciseId: number | null,
): UpNextItem[] {
    const items: UpNextItem[] = [];
    for (const step of steps) {
        if (step.done) continue;
        const group = step.exerciseIndex != null ? log[step.exerciseIndex] : undefined;
        const exerciseId = exerciseIdOf(group);
        if (step.kind === "lift" && exerciseId != null && exerciseId === currentExerciseId) continue;
        if (step.kind === "lift" && group) {
            const weight = openingWeight(group);
            const rollover = group.planned?.rep_rollover ?? group.previous?.exercise?.rep_rollover;
            const last = formatSets(group.previous?.sets);
            items.push({
                key: step.key,
                kind: step.kind,
                name: step.name,
                weight: weight > 0 ? weight : null,
                overload: overloadStatus(group.previous?.sets, rollover, weight),
                detail: last ? `Last ${last}` : "No history",
                machineSettings: visibleMachineSettings(group.planned?.machine_settings),
                exerciseId,
            });
        } else {
            items.push({
                key: step.key,
                kind: step.kind,
                name: step.name,
                weight: null,
                overload: { kind: "none" },
                detail: step.detail,
                machineSettings: [],
                exerciseId: null,
            });
        }
    }
    return items;
}
