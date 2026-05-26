import { computed, ref } from 'vue';

export type MacroPresetId = 'balanced' | 'high_protein' | 'low_carb';

const PRESET_CAL_PCTS: Record<MacroPresetId, { p: number; c: number; f: number }> = {
    balanced: { p: 30, c: 40, f: 30 },
    high_protein: { p: 40, c: 35, f: 25 },
    low_carb: { p: 35, c: 20, f: 45 },
};

const EPS = 0.05;

export type MacroDriver = 'p' | 'c' | 'f';

export function maxGramsForMacro(
    macro: MacroDriver,
    calorieTarget: number,
    P: number,
    C: number,
    F: number,
    lockP: boolean,
    lockC: boolean,
    lockF: boolean,
): number {
    const T = calorieTarget;
    if (T <= 0) return 0;
    if (macro === 'p') {
        if (lockC && lockF) return Math.max(0, (T - 4 * C - 9 * F) / 4);
        if (lockC) return Math.max(0, (T - 4 * C) / 4);
        if (lockF) return Math.max(0, (T - 9 * F) / 4);
        return T / 4;
    }
    if (macro === 'c') {
        if (lockP && lockF) return Math.max(0, (T - 4 * P - 9 * F) / 4);
        if (lockP) return Math.max(0, (T - 4 * P) / 4);
        if (lockF) return Math.max(0, (T - 9 * F) / 4);
        return T / 4;
    }
    if (lockP && lockC) return Math.max(0, (T - 4 * P - 4 * C) / 9);
    if (lockP) return Math.max(0, (T - 4 * P) / 9);
    if (lockC) return Math.max(0, (T - 4 * C) / 9);
    return T / 9;
}

function splitRemainingCalories(
    rem: number,
    unlocked: MacroDriver[],
    p: number,
    c: number,
    f: number,
): { p: number; c: number; f: number } {
    if (rem <= EPS || unlocked.length === 0) {
        return { p, c, f };
    }
    if (unlocked.length === 1) {
        const m = unlocked[0];
        if (m === 'p') return { p: rem / 4, c, f };
        if (m === 'c') return { p, c: rem / 4, f };
        return { p, c, f: rem / 9 };
    }
    let pK = unlocked.includes('p') ? 4 * p : 0;
    let cK = unlocked.includes('c') ? 4 * c : 0;
    let fK = unlocked.includes('f') ? 9 * f : 0;
    const sumK = pK + cK + fK;
    if (sumK <= EPS) {
        const share = rem / unlocked.length;
        return {
            p: unlocked.includes('p') ? share / 4 : p,
            c: unlocked.includes('c') ? share / 4 : c,
            f: unlocked.includes('f') ? share / 9 : f,
        };
    }
    const pK2 = unlocked.includes('p') ? rem * (pK / sumK) : 0;
    const cK2 = unlocked.includes('c') ? rem * (cK / sumK) : 0;
    const fK2 = unlocked.includes('f') ? rem * (fK / sumK) : 0;
    return {
        p: unlocked.includes('p') ? pK2 / 4 : p,
        c: unlocked.includes('c') ? cK2 / 4 : c,
        f: unlocked.includes('f') ? fK2 / 9 : f,
    };
}

export function redistributeMacros(
    calorieTarget: number,
    P: number,
    C: number,
    F: number,
    lockP: boolean,
    lockC: boolean,
    lockF: boolean,
    driver: MacroDriver,
    driverVal: number,
): { P: number; C: number; F: number; calorieTarget: number; warning: boolean } {
    const T = Math.max(0, calorieTarget);
    const driverLocked =
        (driver === 'p' && lockP) ||
        (driver === 'c' && lockC) ||
        (driver === 'f' && lockF);
    if (driverLocked) {
        return { P, C, F, calorieTarget: T, warning: false };
    }

    let p = Math.max(0, P);
    let c = Math.max(0, C);
    let f = Math.max(0, F);
    const lockedK =
        (lockP ? 4 * p : 0) + (lockC ? 4 * c : 0) + (lockF ? 9 * f : 0);
    if (lockedK > T + EPS) {
        return { P: p, C: c, F: f, calorieTarget: T, warning: true };
    }

    const unlocked: MacroDriver[] = [];
    if (!lockP) unlocked.push('p');
    if (!lockC) unlocked.push('c');
    if (!lockF) unlocked.push('f');
    const otherUnlocked = unlocked.filter((m) => m !== driver);
    const availableForDriver = Math.max(0, T - lockedK);

    if (driver === 'p') {
        p = Math.min(Math.max(0, driverVal), availableForDriver / 4);
    } else if (driver === 'c') {
        c = Math.min(Math.max(0, driverVal), availableForDriver / 4);
    } else {
        f = Math.min(Math.max(0, driverVal), availableForDriver / 9);
    }

    if (otherUnlocked.length === 0) {
        const rem = T - lockedK;
        if (driver === 'p') p = rem / 4;
        if (driver === 'c') c = rem / 4;
        if (driver === 'f') f = rem / 9;
        return { P: p, C: c, F: f, calorieTarget: T, warning: false };
    }

    const driverK =
        driver === 'p' ? 4 * p : driver === 'c' ? 4 * c : 9 * f;
    const split = splitRemainingCalories(
        T - lockedK - driverK,
        otherUnlocked,
        p,
        c,
        f,
    );
    return { P: split.p, C: split.c, F: split.f, calorieTarget: T, warning: false };
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
    const fiberG = ref(initial?.fiber ?? 0);
    const lockProtein = ref(false);
    const lockCarbs = ref(false);
    const lockFat = ref(false);
    const proteinGPerLb = ref(0.8);
    const caloriesAdjustedByLocks = ref(false);
    const applyResult = (r: {
        P: number;
        C: number;
        F: number;
        calorieTarget: number;
        warning: boolean;
    }) => {
        proteinG.value = roundG(r.P);
        carbsG.value = roundG(r.C);
        fatG.value = roundG(r.F);
        calorieTarget.value = Math.round(r.calorieTarget);
        caloriesAdjustedByLocks.value = r.warning;
    };
    function roundG(g: number) {
        return Math.round(g * 10) / 10;
    }
    function setProteinG(v: number) {
        applyResult(
            redistributeMacros(
                calorieTarget.value,
                proteinG.value,
                carbsG.value,
                fatG.value,
                lockProtein.value,
                lockCarbs.value,
                lockFat.value,
                'p',
                v,
            ),
        );
    }
    function setCarbsG(v: number) {
        applyResult(
            redistributeMacros(
                calorieTarget.value,
                proteinG.value,
                carbsG.value,
                fatG.value,
                lockProtein.value,
                lockCarbs.value,
                lockFat.value,
                'c',
                v,
            ),
        );
    }
    function setFatG(v: number) {
        applyResult(
            redistributeMacros(
                calorieTarget.value,
                proteinG.value,
                carbsG.value,
                fatG.value,
                lockProtein.value,
                lockCarbs.value,
                lockFat.value,
                'f',
                v,
            ),
        );
    }
    function setCalorieTarget(v: number) {
        const T = Math.max(0, Math.round(v));
        const lp = lockProtein.value;
        const lc = lockCarbs.value;
        const lf = lockFat.value;
        if (!lp && !lc && !lf) {
            const oldK =
                4 * proteinG.value + 4 * carbsG.value + 9 * fatG.value;
            if (oldK <= EPS) {
                applyPresetMacros('balanced', T);
                caloriesAdjustedByLocks.value = false;
                return;
            }
            const factor = T / oldK;
            proteinG.value = roundG(proteinG.value * factor);
            carbsG.value = roundG(carbsG.value * factor);
            fatG.value = roundG(fatG.value * factor);
            calorieTarget.value = T;
            caloriesAdjustedByLocks.value = false;
            return;
        }
        calorieTarget.value = T;
        if (lp && lc && lf) {
            caloriesAdjustedByLocks.value =
                Math.abs(macroCaloriesTotal.value - T) > EPS;
            return;
        }
        let p = proteinG.value;
        let c = carbsG.value;
        let f = fatG.value;
        const lockedK =
            (lp ? 4 * p : 0) + (lc ? 4 * c : 0) + (lf ? 9 * f : 0);
        const rem = T - lockedK;
        if (rem < -EPS) {
            caloriesAdjustedByLocks.value = true;
            return;
        }
        const unlocked: MacroDriver[] = [];
        if (!lp) unlocked.push('p');
        if (!lc) unlocked.push('c');
        if (!lf) unlocked.push('f');
        const split = splitRemainingCalories(rem, unlocked, p, c, f);
        if (!lp) p = split.p;
        if (!lc) c = split.c;
        if (!lf) f = split.f;
        proteinG.value = roundG(p);
        carbsG.value = roundG(c);
        fatG.value = roundG(f);
        caloriesAdjustedByLocks.value = false;
    }
    function applyPresetMacros(preset: MacroPresetId, T?: number) {
        const cal = T ?? calorieTarget.value;
        const { p, c, f } = PRESET_CAL_PCTS[preset];
        const t = Math.max(0, cal);
        calorieTarget.value = Math.round(t);
        const lockedK =
            (lockProtein.value ? 4 * proteinG.value : 0) +
            (lockCarbs.value ? 4 * carbsG.value : 0) +
            (lockFat.value ? 9 * fatG.value : 0);
        if (lockedK > t + EPS) {
            caloriesAdjustedByLocks.value = true;
            return;
        }
        const unlocked = [
            { macro: 'p' as const, pct: p },
            { macro: 'c' as const, pct: c },
            { macro: 'f' as const, pct: f },
        ].filter(({ macro }) => {
            if (macro === 'p') return !lockProtein.value;
            if (macro === 'c') return !lockCarbs.value;
            return !lockFat.value;
        });
        if (unlocked.length === 0) {
            caloriesAdjustedByLocks.value =
                Math.abs(macroCaloriesTotal.value - t) > EPS;
            return;
        }
        const pctTotal = unlocked.reduce((sum, item) => sum + item.pct, 0);
        const rem = t - lockedK;
        for (const item of unlocked) {
            const share = pctTotal > EPS ? item.pct / pctTotal : 1 / unlocked.length;
            if (item.macro === 'p') proteinG.value = roundG((rem * share) / 4);
            if (item.macro === 'c') carbsG.value = roundG((rem * share) / 4);
            if (item.macro === 'f') fatG.value = roundG((rem * share) / 9);
        }
        caloriesAdjustedByLocks.value = false;
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
    const macroCaloriesTotal = computed(
        () => 4 * proteinG.value + 4 * carbsG.value + 9 * fatG.value,
    );
    const suggestedFiber = computed(() =>
        Math.max(0, Math.round((14 * calorieTarget.value) / 1000)),
    );
    function proteinGramsFromGPerLb(lbs: number, mult: number) {
        if (lbs <= 0) return 0;
        return Math.round(lbs * mult * 10) / 10;
    }
    function applyProteinFromBodyWeight(weightLbs: number) {
        if (weightLbs <= 0) return;
        const grams = proteinGramsFromGPerLb(weightLbs, proteinGPerLb.value);
        applyResult(
            redistributeMacros(
                calorieTarget.value,
                proteinG.value,
                carbsG.value,
                fatG.value,
                false,
                lockCarbs.value,
                lockFat.value,
                'p',
                grams,
            ),
        );
    }
    function maxProteinG() {
        return maxGramsForMacro(
            'p',
            calorieTarget.value,
            proteinG.value,
            carbsG.value,
            fatG.value,
            lockProtein.value,
            lockCarbs.value,
            lockFat.value,
        );
    }
    function maxCarbsG() {
        return maxGramsForMacro(
            'c',
            calorieTarget.value,
            proteinG.value,
            carbsG.value,
            fatG.value,
            lockProtein.value,
            lockCarbs.value,
            lockFat.value,
        );
    }
    function maxFatG() {
        return maxGramsForMacro(
            'f',
            calorieTarget.value,
            proteinG.value,
            carbsG.value,
            fatG.value,
            lockProtein.value,
            lockCarbs.value,
            lockFat.value,
        );
    }
    function minGramsForMacro(macro: MacroDriver) {
        const unlockedCount = [
            !lockProtein.value,
            !lockCarbs.value,
            !lockFat.value,
        ].filter(Boolean).length;
        if (unlockedCount !== 1) return 0;
        if (macro === 'p' && lockProtein.value) return 0;
        if (macro === 'c' && lockCarbs.value) return 0;
        if (macro === 'f' && lockFat.value) return 0;
        return maxGramsForMacro(
            macro,
            calorieTarget.value,
            proteinG.value,
            carbsG.value,
            fatG.value,
            lockProtein.value,
            lockCarbs.value,
            lockFat.value,
        );
    }
    function minProteinG() {
        return minGramsForMacro('p');
    }
    function minCarbsG() {
        return minGramsForMacro('c');
    }
    function minFatG() {
        return minGramsForMacro('f');
    }
    const lockedMacroCalories = computed(() => {
        let k = 0;
        if (lockProtein.value) k += 4 * proteinG.value;
        if (lockCarbs.value) k += 4 * carbsG.value;
        if (lockFat.value) k += 9 * fatG.value;
        return k;
    });
    const lockConflict = computed(
        () =>
            lockedMacroCalories.value >
            calorieTarget.value + EPS,
    );
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
        caloriesAdjustedByLocks,
        lockConflict,
        lockedMacroCalories,
        proteinCalPct,
        carbsCalPct,
        fatCalPct,
        macroCaloriesTotal,
        suggestedFiber,
        setProteinG,
        setCarbsG,
        setFatG,
        setCalorieTarget,
        applyPresetMacros,
        proteinGramsFromGPerLb,
        applyProteinFromBodyWeight,
        minProteinG,
        minCarbsG,
        minFatG,
        maxProteinG,
        maxCarbsG,
        maxFatG,
        presetLabels: {
            balanced: 'Balanced',
            high_protein: 'High protein',
            low_carb: 'Low carb',
        } as Record<MacroPresetId, string>,
    };
}

export type MacroSlidersApi = ReturnType<typeof useMacroSliders>;
