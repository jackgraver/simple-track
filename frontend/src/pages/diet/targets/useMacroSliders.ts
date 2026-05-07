import { computed, ref } from 'vue';

export type MacroPresetId = 'balanced' | 'high_protein' | 'low_carb';

const PRESET_CAL_PCTS: Record<MacroPresetId, { p: number; c: number; f: number }> = {
    balanced: { p: 30, c: 40, f: 30 },
    high_protein: { p: 40, c: 35, f: 25 },
    low_carb: { p: 35, c: 20, f: 45 },
};

const EPS = 0.05;

export type MacroDriver = 'p' | 'c' | 'f';

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
    let p = Math.max(0, driver === 'p' ? driverVal : P);
    let c = Math.max(0, driver === 'c' ? driverVal : C);
    let f = Math.max(0, driver === 'f' ? driverVal : F);
    let T = calorieTarget;
    let warning = false;
    const remP = () => T - 4 * p;
    const remC = () => T - 4 * c;
    const remF = () => T - 9 * f;
    const splitCF = (rem: number) => {
        if (lockC && lockF) {
            const need = 4 * c + 9 * f;
            if (Math.abs(rem - need) > EPS) {
                T = 4 * p + 4 * c + 9 * f;
                warning = true;
            }
            return;
        }
        if (lockC) {
            f = (rem - 4 * c) / 9;
            if (f < 0) f = 0;
            return;
        }
        if (lockF) {
            c = (rem - 9 * f) / 4;
            if (c < 0) c = 0;
            return;
        }
        const cK = 4 * c;
        const fK = 9 * f;
        const sumK = cK + fK;
        if (sumK <= EPS) {
            const half = rem / 2;
            c = half / 4;
            f = half / 9;
        } else {
            const cK2 = rem * (cK / sumK);
            const fK2 = rem * (fK / sumK);
            c = cK2 / 4;
            f = fK2 / 9;
        }
    };
    const splitPF = (rem: number) => {
        if (lockP && lockF) {
            const need = 4 * p + 9 * f;
            if (Math.abs(rem - need) > EPS) {
                T = 4 * p + 4 * c + 9 * f;
                warning = true;
            }
            return;
        }
        if (lockP) {
            f = (rem - 4 * p) / 9;
            if (f < 0) f = 0;
            return;
        }
        if (lockF) {
            p = (rem - 9 * f) / 4;
            if (p < 0) p = 0;
            return;
        }
        const pK = 4 * p;
        const fK = 9 * f;
        const sumK = pK + fK;
        if (sumK <= EPS) {
            const half = rem / 2;
            p = half / 4;
            f = half / 9;
        } else {
            const pK2 = rem * (pK / sumK);
            const fK2 = rem * (fK / sumK);
            p = pK2 / 4;
            f = fK2 / 9;
        }
    };
    const splitPC = (rem: number) => {
        if (lockP && lockC) {
            const need = 4 * p + 4 * c;
            if (Math.abs(rem - need) > EPS) {
                T = 4 * p + 4 * c + 9 * f;
                warning = true;
            }
            return;
        }
        if (lockP) {
            c = (rem - 4 * p) / 4;
            if (c < 0) c = 0;
            return;
        }
        if (lockC) {
            p = (rem - 4 * c) / 4;
            if (p < 0) p = 0;
            return;
        }
        const pK = 4 * p;
        const cK = 4 * c;
        const sumK = pK + cK;
        if (sumK <= EPS) {
            const half = rem / 2;
            p = half / 4;
            c = half / 4;
        } else {
            const pK2 = rem * (pK / sumK);
            const cK2 = rem * (cK / sumK);
            p = pK2 / 4;
            c = cK2 / 4;
        }
    };
    if (driver === 'p') {
        splitCF(remP());
    } else if (driver === 'c') {
        splitPF(remC());
    } else {
        splitPC(remF());
    }
    return { P: p, C: c, F: f, calorieTarget: Math.round(T * 10) / 10, warning };
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
        const T = Math.max(0, v);
        const oldK = 4 * proteinG.value + 4 * carbsG.value + 9 * fatG.value;
        if (oldK <= EPS) {
            applyPresetMacros('balanced', T);
            caloriesAdjustedByLocks.value = false;
            return;
        }
        const factor = T / oldK;
        proteinG.value = roundG(proteinG.value * factor);
        carbsG.value = roundG(carbsG.value * factor);
        fatG.value = roundG(fatG.value * factor);
        calorieTarget.value = Math.round(T);
        caloriesAdjustedByLocks.value = false;
    }
    function applyPresetMacros(preset: MacroPresetId, T?: number) {
        const cal = T ?? calorieTarget.value;
        const { p, c, f } = PRESET_CAL_PCTS[preset];
        const t = Math.max(0, cal);
        proteinG.value = roundG((t * (p / 100)) / 4);
        carbsG.value = roundG((t * (c / 100)) / 4);
        fatG.value = roundG((t * (f / 100)) / 9);
        calorieTarget.value = Math.round(t);
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
    return {
        calorieTarget,
        proteinG,
        carbsG,
        fatG,
        fiberG,
        lockProtein,
        lockCarbs,
        lockFat,
        caloriesAdjustedByLocks,
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
        presetLabels: {
            balanced: 'Balanced',
            high_protein: 'High protein',
            low_carb: 'Low carb',
        } as Record<MacroPresetId, string>,
    };
}

export type MacroSlidersApi = ReturnType<typeof useMacroSliders>;
