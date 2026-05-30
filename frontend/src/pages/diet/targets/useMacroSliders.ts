import { computed, ref } from "vue";

export type MacroDriver = "p" | "c" | "f";

const CALORIES_PER_GRAM: Record<MacroDriver, number> = {
    p: 4,
    c: 4,
    f: 9,
};

export function maxGramsForMacro(
    macro: MacroDriver,
    calorieTarget: number,
    currentGrams = 0,
): number {
    const targetMax = Math.max(0, calorieTarget) / CALORIES_PER_GRAM[macro];
    const fallbackMax = macro === "f" ? 150 : 300;
    return Math.max(currentGrams, targetMax, fallbackMax);
}

function macroCalories(p: number, c: number, f: number) {
    return 4 * p + 4 * c + 9 * f;
}

function roundGrams(g: number) {
    return Math.max(0, Math.round(g || 0));
}

export function useMacroSliders(initial?: {
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
}) {
    const calorieTarget = ref(initial?.calories ?? 2000);
    const proteinG = ref(initial?.protein ?? 0);
    const carbsG = ref(initial?.carbs ?? 0);
    const fatG = ref(initial?.fat ?? 0);
    const fiberG = ref(initial?.fiber ?? 38);
    const lockProtein = ref(false);
    const lockCarbs = ref(false);
    const lockFat = ref(false);
    const proteinGPerLb = ref(0.8);
    const fatGPerLb = ref(0.3);

    const remainingCaloriesForCarbs = computed(() =>
        Math.max(0, calorieTarget.value - 4 * proteinG.value - 9 * fatG.value),
    );

    const suggestedCarbsG = computed(() =>
        roundGrams(remainingCaloriesForCarbs.value / 4),
    );

    function fillCarbsFromRemaining() {
        if (lockCarbs.value) return;
        carbsG.value = suggestedCarbsG.value;
        setFiberG(fiberG.value);
    }

    function setProteinG(v: number) {
        proteinG.value = roundGrams(v);
        fillCarbsFromRemaining();
    }

    function setCarbsG(v: number) {
        carbsG.value = roundGrams(v);
        setFiberG(fiberG.value);
    }

    function setFatG(v: number) {
        fatG.value = roundGrams(v);
        fillCarbsFromRemaining();
    }

    function setFiberG(v: number) {
        fiberG.value = Math.min(carbsG.value, roundGrams(v));
    }

    function setMacroTargets(targets: {
        calories: number;
        protein: number;
        carbs: number;
        fat: number;
        fiber: number;
    }) {
        calorieTarget.value = Math.max(0, Math.round(targets.calories));
        proteinG.value = roundGrams(targets.protein);
        carbsG.value = roundGrams(targets.carbs);
        fatG.value = roundGrams(targets.fat);
        setFiberG(targets.fiber);
    }

    function setCalorieTarget(v: number) {
        calorieTarget.value = Math.max(0, Math.round(v));
        fillCarbsFromRemaining();
    }

    const proteinCalPct = computed(() => {
        const t = calorieTarget.value;
        if (t <= 0) return 0;
        return (4 * proteinG.value * 100) / t;
    });

    const carbsCalPct = computed(() => {
        const t = calorieTarget.value;
        if (t <= 0) return 0;
        return (4 * carbsG.value * 100) / t;
    });

    const fatCalPct = computed(() => {
        const t = calorieTarget.value;
        if (t <= 0) return 0;
        return (9 * fatG.value * 100) / t;
    });

    const macroCaloriesTotal = computed(() =>
        macroCalories(proteinG.value, carbsG.value, fatG.value),
    );

    const calorieDelta = computed(
        () => Math.round(macroCaloriesTotal.value) - calorieTarget.value,
    );

    const netCarbsG = computed(() => Math.max(0, carbsG.value - fiberG.value));

    const suggestedFiber = computed(() =>
        Math.max(0, Math.round((14 * calorieTarget.value) / 1000)),
    );

    function gramsFromGPerLb(lbs: number, mult: number) {
        if (lbs <= 0) return 0;
        return roundGrams(lbs * mult);
    }

    function proteinGramsFromGPerLb(lbs: number, mult: number) {
        return gramsFromGPerLb(lbs, mult);
    }

    function fatGramsFromGPerLb(lbs: number, mult: number) {
        return gramsFromGPerLb(lbs, mult);
    }

    function applyProteinFromBodyWeight(weightLbs: number) {
        if (weightLbs <= 0) return;
        const grams = proteinGramsFromGPerLb(weightLbs, proteinGPerLb.value);
        setProteinG(grams);
    }

    function applyFatFromBodyWeight(weightLbs: number) {
        if (weightLbs <= 0) return;
        const grams = fatGramsFromGPerLb(weightLbs, fatGPerLb.value);
        setFatG(grams);
    }

    function maxProteinG() {
        return maxGramsForMacro("p", calorieTarget.value, proteinG.value);
    }

    function maxCarbsG() {
        return maxGramsForMacro("c", calorieTarget.value, carbsG.value);
    }

    function maxFatG() {
        return maxGramsForMacro("f", calorieTarget.value, fatG.value);
    }

    function minProteinG() {
        return 0;
    }

    function minCarbsG() {
        return 0;
    }

    function minFatG() {
        return 0;
    }

    return {
        calorieTarget,
        proteinG,
        carbsG,
        fatG,
        fiberG,
        lockProtein,
        lockCarbs,
        lockFat,
        proteinGPerLb,
        fatGPerLb,
        proteinCalPct,
        carbsCalPct,
        fatCalPct,
        macroCaloriesTotal,
        calorieDelta,
        remainingCaloriesForCarbs,
        suggestedCarbsG,
        netCarbsG,
        suggestedFiber,
        setProteinG,
        setCarbsG,
        setFatG,
        setFiberG,
        setMacroTargets,
        setCalorieTarget,
        fillCarbsFromRemaining,
        proteinGramsFromGPerLb,
        fatGramsFromGPerLb,
        applyProteinFromBodyWeight,
        applyFatFromBodyWeight,
        minProteinG,
        minCarbsG,
        minFatG,
        maxProteinG,
        maxCarbsG,
        maxFatG,
    };
}

export type MacroSlidersApi = ReturnType<typeof useMacroSliders>;
