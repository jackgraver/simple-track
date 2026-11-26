export const WEIGHT_PROGRESSION_INCREASE_LBS = 5;

type SetLike = { weight: number; reps: number };

/**
 * What one session's sets mean for the next session. The one rule every screen uses:
 * - the next session starts at the final set's weight (after a drop, the dropped weight is the new
 *   starter and you work back up from it);
 * - the weight goes up only when every set was at that one weight and every set hit the rollover.
 */
export type SessionOutcome = {
    /** The final set's weight: where the next session starts. */
    startWeight: number;
    /** Every set was at startWeight. */
    sameWeight: boolean;
    /** An earlier set was heavier than the final one. */
    dropped: boolean;
    /** Sets that reached the rollover (null when no rollover is set). */
    hit: number | null;
    total: number;
    rollover: number | null;
    /** Every set at one weight and every set hit the rollover: go up next time. */
    earned: boolean;
};

const validRollover = (rollover: number | null | undefined): number | null =>
    typeof rollover === "number" && Number.isFinite(rollover) && rollover > 0 ? rollover : null;

export function sessionOutcome(
    sets: readonly SetLike[] | null | undefined,
    repRollover: number | null | undefined,
): SessionOutcome | null {
    const valid = (sets ?? []).filter((s) => Number.isFinite(Number(s.weight)) && Number.isFinite(Number(s.reps)));
    const last = valid[valid.length - 1];
    if (!last || !(Number(last.weight) > 0)) return null;
    const startWeight = Number(last.weight);
    const rollover = validRollover(repRollover);
    const sameWeight = valid.every((s) => Number(s.weight) === startWeight);
    const dropped = valid.some((s) => Number(s.weight) > startWeight);
    const hit = rollover == null ? null : valid.filter((s) => Number(s.reps) >= rollover).length;
    return {
        startWeight,
        sameWeight,
        dropped,
        hit,
        total: valid.length,
        rollover,
        earned: sameWeight && hit != null && hit === valid.length,
    };
}

export type WeightProgression = {
    previousWeight: number;
    nextWeight: number;
    increase: number;
};

/** The automatic increase for a session that opens from `previousSets`, or null when it wasn't earned. */
export function getWeightProgression(
    previousSets: readonly SetLike[] | undefined,
    repRollover: number | undefined,
    increase = WEIGHT_PROGRESSION_INCREASE_LBS,
): WeightProgression | null {
    if (!Number.isFinite(increase) || increase <= 0) return null;
    const outcome = sessionOutcome(previousSets, repRollover);
    if (!outcome?.earned) return null;
    return {
        previousWeight: outcome.startWeight,
        nextWeight: outcome.startWeight + increase,
        increase,
    };
}
