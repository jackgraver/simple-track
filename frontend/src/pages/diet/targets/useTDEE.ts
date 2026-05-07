import { computed, type MaybeRefOrGetter, toValue } from 'vue';
import type { ActivityLevel } from '~/api/tracking/types';

export const ACTIVITY_MULTIPLIERS: Record<ActivityLevel, number> = {
    sedentary: 1.2,
    lightly_active: 1.375,
    moderately_active: 1.55,
    very_active: 1.725,
    extra_active: 1.9,
};

export type CalorieGoalId =
    | 'lose_agg'
    | 'lose_mod'
    | 'maintain'
    | 'gain_mod'
    | 'gain_agg';

export const GOAL_DELTAS: Record<CalorieGoalId, number> = {
    lose_agg: -500,
    lose_mod: -250,
    maintain: 0,
    gain_mod: 250,
    gain_agg: 500,
};

export const CALORIE_GOAL_LABELS: Record<CalorieGoalId, string> = {
    lose_agg: 'Lose (aggressive)',
    lose_mod: 'Lose (moderate)',
    maintain: 'Maintain',
    gain_mod: 'Gain (moderate)',
    gain_agg: 'Gain (aggressive)',
};

export function lbsToKg(lbs: number): number {
    return lbs * 0.45359237;
}

export function inchesToCm(inches: number): number {
    return inches * 2.54;
}

export function mifflinStJeorBmr(input: {
    weightKg: number;
    heightCm: number;
    age: number;
    sex: 'male' | 'female';
}): number {
    const { weightKg, heightCm, age, sex } = input;
    const base = 10 * weightKg + 6.25 * heightCm - 5 * age;
    return sex === 'male' ? base + 5 : base - 161;
}

export function tdeeFromBmr(bmr: number, activity: ActivityLevel): number {
    const m = ACTIVITY_MULTIPLIERS[activity];
    return bmr * m;
}

export function calorieTargetFromTdee(tdee: number, goal: CalorieGoalId): number {
    const t = Math.round(tdee + GOAL_DELTAS[goal]);
    return Math.max(500, t);
}

export function useTDEE(input: {
    weightLbs: MaybeRefOrGetter<number>;
    heightIn: MaybeRefOrGetter<number>;
    age: MaybeRefOrGetter<number>;
    sex: MaybeRefOrGetter<'male' | 'female'>;
    activity: MaybeRefOrGetter<ActivityLevel>;
}) {
    const bmr = computed(() => {
        const w = toValue(input.weightLbs);
        const h = toValue(input.heightIn);
        const a = toValue(input.age);
        const s = toValue(input.sex);
        if (w <= 0 || h <= 0 || a <= 0) return null;
        return mifflinStJeorBmr({
            weightKg: lbsToKg(w),
            heightCm: inchesToCm(h),
            age: a,
            sex: s,
        });
    });
    const tdee = computed(() => {
        const b = bmr.value;
        if (b == null) return null;
        const act = toValue(input.activity);
        return tdeeFromBmr(b, act);
    });
    const multiplier = computed(() => {
        const act = toValue(input.activity);
        return ACTIVITY_MULTIPLIERS[act];
    });
    return { bmr, tdee, multiplier };
}
