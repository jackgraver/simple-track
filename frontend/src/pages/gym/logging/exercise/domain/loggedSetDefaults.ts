import { LoggedSet } from "~/types/workout";

export type LoggedSetForDefaults = {
    weight: number;
    reps: number;
    weight_setup?: string;
};


export function initializeWeightAndReps(loggedSets: readonly LoggedSet[] | undefined, previousSets: readonly LoggedSet[] | undefined): { weight: number; reps: number; weightSetup: string; notes: string } {
    let weight = 0;
    let reps = 0;
    let weightSetup = "";
    let notes = "";

    if (
        loggedSets &&
        loggedSets.length > 0
    ) {
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

/** Values to seed the current-set inputs when continuing from existing logged sets. */
export function defaultsFromLoggedSets(
    sets: readonly LoggedSetForDefaults[],
): { weight: number; reps: number; weight_setup: string } {
    const maxSet = pickBestLoggedSet(sets);
    if (maxSet) {
        return {
            weight: maxSet.weight,
            reps: maxSet.reps,
            weight_setup: maxSet.weight_setup || "",
        };
    }
    return { weight: 0, reps: 0, weight_setup: "" };
}
