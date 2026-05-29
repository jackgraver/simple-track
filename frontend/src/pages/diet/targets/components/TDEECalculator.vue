<script setup lang="ts">
import { computed, ref } from 'vue';
import {
    CALORIE_GOAL_LABELS,
    calorieTargetFromTdee,
    type CalorieGoalId,
} from '../useTDEE';

const props = defineProps<{
    bmr: number | null;
    tdee: number | null;
    multiplier: number;
}>();

const emit = defineEmits<{ applyCalories: [n: number] }>();

const goal = ref<CalorieGoalId>('maintain');

const goalChoices = computed(() =>
    (Object.keys(CALORIE_GOAL_LABELS) as CalorieGoalId[]).map((id) => ({
        id,
        label: CALORIE_GOAL_LABELS[id],
    })),
);

const goalCalories = computed(() => {
    if (props.tdee == null) return null;
    return calorieTargetFromTdee(props.tdee, goal.value);
});

function apply() {
    const c = goalCalories.value;
    if (c != null) emit('applyCalories', c);
}
</script>

<template>
    <div class="flex flex-col gap-3">
        <template v-if="bmr != null && tdee != null">
            <div class="flex flex-wrap items-center justify-center gap-x-2 gap-y-1 rounded-md bg-zinc-900/60 px-3 py-2.5 text-sm">
                <span class="text-zinc-500">BMR</span>
                <span class="tabular-nums font-medium text-zinc-200">{{ Math.round(bmr) }}</span>
                <span class="text-zinc-500">× {{ multiplier }}</span>
                <span class="text-zinc-600">=</span>
                <span class="text-zinc-500">TDEE</span>
                <span class="tabular-nums font-semibold text-zinc-100">{{ Math.round(tdee) }}</span>
            </div>
            <div class="flex flex-col gap-2">
                <span class="text-xs font-medium text-zinc-400">Goal</span>
                <div class="flex flex-wrap gap-1.5">
                    <button
                        v-for="g in goalChoices"
                        :key="g.id"
                        type="button"
                        class="rounded-md border px-2 py-1 text-xs font-medium"
                        :class="
                            goal === g.id
                                ? 'border-amber-600 bg-amber-950/50 text-amber-100'
                                : 'border-zinc-600 text-zinc-400 hover:border-zinc-500'
                        "
                        @click="goal = g.id"
                    >
                        {{ g.label }}
                    </button>
                </div>
            </div>
            <div class="flex items-center justify-between gap-3 rounded-md bg-zinc-900/60 px-3 py-2">
                <span class="text-sm tabular-nums text-zinc-200">
                    {{ goalCalories ?? '—' }} kcal/day
                </span>
                <button
                    type="button"
                    class="shrink-0 rounded-md border border-amber-700/50 bg-amber-950/30 px-3 py-1.5 text-xs font-medium text-amber-100 hover:bg-amber-950/50 disabled:opacity-50"
                    :disabled="goalCalories == null"
                    @click="apply"
                >
                    Apply to macros
                </button>
            </div>
        </template>
        <p v-else class="m-0 text-sm text-zinc-500">
            Fill body profile and weight to estimate TDEE.
        </p>
    </div>
</template>
