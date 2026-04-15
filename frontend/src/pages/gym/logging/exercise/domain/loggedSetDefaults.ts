export type LoggedSetForDefaults = {
    weight: number;
    reps: number;
    weight_setup?: string;
};

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
