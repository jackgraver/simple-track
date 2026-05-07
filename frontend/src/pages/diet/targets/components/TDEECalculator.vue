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
    <div class="rounded-lg border border-zinc-600 bg-zinc-900/60 p-4">
        <div v-if="bmr != null && tdee != null" class="flex flex-col gap-3">
            <div class="text-sm text-zinc-400">
                <p class="m-0">BMR: {{ Math.round(bmr) }} kcal/day</p>
                <p class="m-0">Activity × {{ multiplier }}</p>
                <p class="m-0 font-medium text-zinc-200">
                    TDEE: {{ Math.round(tdee) }} kcal/day
                </p>
            </div>
            <div>
                <p class="mb-2 text-sm font-medium text-zinc-300">Goal</p>
                <div class="mb-2 flex flex-wrap gap-2">
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
                <p class="mb-2 text-sm tabular-nums text-zinc-200">
                    Target: {{ goalCalories ?? '—' }} kcal/day
                </p>
                <button
                    type="button"
                    class="rounded-md border border-amber-700/50 bg-amber-950/30 px-3 py-2 text-sm font-medium text-amber-100 hover:bg-amber-950/50"
                    :disabled="goalCalories == null"
                    @click="apply"
                >
                    Use as calorie target
                </button>
            </div>
        </div>
        <p v-else class="m-0 text-sm text-zinc-500">
            Fill body profile and weight to estimate TDEE.
        </p>
    </div>
</template>
