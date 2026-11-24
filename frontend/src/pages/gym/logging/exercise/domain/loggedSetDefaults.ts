import type { LoggedSet } from "~/types/workout";

export const DEFAULT_REPS_AFTER_WEIGHT_CHANGE = 6;

export type LoggedSetForDefaults = {
    weight: number;
    reps: number;
    weight_setup?: string;
};

export function initializeWeightAndReps(
    loggedSets: readonly LoggedSet[] | undefined,
    previousSets: readonly LoggedSet[] | undefined,
): { weight: number; reps: number; weightSetup: string; notes: string } {
    let weight = 0;
    let reps = 0;
    let weightSetup = "";
    let notes = "";

    if (loggedSets && loggedSets.length > 0) {
        weight = loggedSets[loggedSets.length - 1].weight;
        reps = loggedSets[loggedSets.length - 1].reps;
        weightSetup = loggedSets[loggedSets.length - 1].weight_setup || "";
    } else if (previousSets) {
        const dPrev = defaultsFromLoggedSets(previousSets);
        weight = dPrev.weight;
        reps = dPrev.reps;
        weightSetup = dPrev.weight_setup;
    }

    return { weight, reps, weightSetup, notes };
}

/**
 * Last session's reps to beat for set `setIndex` at `weight`: the same set number if it was at that
 * weight, else the nearest set last time at that weight (after a drop, today's set 1 at the dropped
 * weight compares with last time's dropped set). Null when last time had no set at that weight.
 */
export function previousRepsForSetAtWeight(
    sets: readonly LoggedSetForDefaults[] | undefined,
    setIndex: number,
    weight: number,
): number | null {
    let reps: number | null = null;
    let bestDistance = Infinity;
    const list = sets ?? [];
    for (let i = 0; i < list.length; i++) {
        const set = list[i];
        if (set.weight !== weight || !Number.isFinite(set.reps)) continue;
        const distance = Math.abs(i - setIndex);
        if (distance < bestDistance) {
            reps = set.reps;
            bestDistance = distance;
        }
    }
    return reps;
}

/**
 * Picks the set with the highest weight; on a tie, the higher rep count wins.
 * If weight and reps both tie, the earlier set in the array wins.
 */
export function pickBestLoggedSet<T extends { weight: number; reps: number }>(
    sets: readonly T[],
): T | null {
    if (sets.length === 0) return null;
    return sets.reduce<T | null>((max, set) => {
        if (!max) return set;
        if (set.weight > max.weight) return set;
        if (set.weight === max.weight && set.reps > max.reps) return set;
        return max;
    }, null);
}

/**
 * Values to seed the current-set inputs from last session: the final set's weight (the next session
 * starts there, see `sessionOutcome`), with the best reps and the setup at that weight.
 */
export function defaultsFromLoggedSets(sets: readonly LoggedSetForDefaults[]): {
    weight: number;
    reps: number;
    weight_setup: string;
} {
    const last = sets[sets.length - 1];
    const best = last ? pickBestLoggedSet(sets.filter((s) => s.weight === last.weight)) : null;
    if (best) {
        return {
            weight: best.weight,
            reps: best.reps,
            weight_setup: best.weight_setup || "",
        };
    }
    return { weight: 0, reps: 0, weight_setup: "" };
}
