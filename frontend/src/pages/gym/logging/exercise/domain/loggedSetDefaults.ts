import type { LoggedSet } from "~/types/workout";
import { platesOfSet, type PlateCounts } from "./plates";

export const DEFAULT_REPS_AFTER_WEIGHT_CHANGE = 6;

export type LoggedSetForDefaults = {
    weight: number;
    reps: number;
    weight_setup?: string;
    plate_setup?: LoggedSet["plate_setup"];
};

/** Seeds the draft from this session's last set, else last session's best. `plates` is that set's
 *  plates (stored, or parsed from its text), null when it has none. */
export function initializeWeightAndReps(
    loggedSets: readonly LoggedSet[] | undefined,
    previousSets: readonly LoggedSet[] | undefined,
): { weight: number; reps: number; weightSetup: string; notes: string; plates: PlateCounts | null } {
    let seed: LoggedSetForDefaults | null = null;
    if (loggedSets && loggedSets.length > 0) {
        seed = loggedSets[loggedSets.length - 1];
    } else if (previousSets) {
        seed = pickBestLoggedSet(previousSets);
    }
    return {
        weight: seed?.weight ?? 0,
        reps: seed?.reps ?? 0,
        weightSetup: seed?.weight_setup || "",
        notes: "",
        plates: seed ? platesOfSet({ weight_setup: seed.weight_setup ?? "", plate_setup: seed.plate_setup }) : null,
    };
}

export function previousRepsForSetAtWeight(
    sets: readonly LoggedSetForDefaults[] | undefined,
    setIndex: number,
    weight: number,
): number | null {
    const set = sets?.[setIndex];
    if (!set || set.weight !== weight || !Number.isFinite(set.reps))
        return null;
    return set.reps;
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
export function defaultsFromLoggedSets(sets: readonly LoggedSetForDefaults[]): {
    weight: number;
    reps: number;
    weight_setup: string;
} {
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
