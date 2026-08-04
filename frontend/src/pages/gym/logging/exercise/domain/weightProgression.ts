export const WEIGHT_PROGRESSION_INCREASE_LBS = 5;

export type WeightProgression = {
    previousWeight: number;
    nextWeight: number;
    increase: number;
};

export function getWeightProgression(
    previousSets: readonly { weight: number; reps: number }[] | undefined,
    repRollover: number | undefined,
    currentWeight: number,
): WeightProgression | null {
    if (
        typeof repRollover !== "number" ||
        !Number.isFinite(repRollover) ||
        !Number.isFinite(currentWeight) ||
        currentWeight <= 0
    ) {
        return null;
    }
    const setsAtCurrentWeight = (previousSets ?? []).filter(
        (set) =>
            Number(set.weight) === currentWeight &&
            Number.isFinite(Number(set.reps)),
    );
    if (
        setsAtCurrentWeight.length === 0 ||
        !setsAtCurrentWeight.every((set) => Number(set.reps) >= repRollover)
    ) {
        return null;
    }
    return {
        previousWeight: currentWeight,
        nextWeight: currentWeight + WEIGHT_PROGRESSION_INCREASE_LBS,
        increase: WEIGHT_PROGRESSION_INCREASE_LBS,
    };
}
