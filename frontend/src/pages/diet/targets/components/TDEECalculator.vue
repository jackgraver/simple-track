<script setup lang="ts">
import { computed, ref } from "vue";
import { Button, Field, Select } from "~/shared/ui";
import {
    CALORIE_GOAL_LABELS,
    calorieTargetFromTdee,
    type CalorieGoalId,
} from "../useTDEE";

const props = defineProps<{
    bmr: number | null;
    tdee: number | null;
    multiplier: number;
}>();

const emit = defineEmits<{ applyCalories: [n: number] }>();

const goal = ref<CalorieGoalId>("maintain");

const goalChoices = computed(() =>
    (Object.keys(CALORIE_GOAL_LABELS) as CalorieGoalId[]).map((id) => ({
        value: id,
        label: CALORIE_GOAL_LABELS[id],
    })),
);

const goalCalories = computed(() => {
    if (props.tdee == null) return null;
    return calorieTargetFromTdee(props.tdee, goal.value);
});

function apply() {
    const c = goalCalories.value;
    if (c != null) emit("applyCalories", c);
}
</script>

<template>
    <div class="flex flex-col gap-4">
        <template v-if="bmr != null && tdee != null">
            <div
                class="flex flex-wrap items-center justify-center gap-x-2 gap-y-1 rounded-control bg-surface-2 px-3 py-3 text-sm"
            >
                <span class="text-fg-muted">BMR</span>
                <span class="font-medium text-fg tabular-nums">{{ Math.round(bmr) }}</span>
                <span class="text-fg-muted">× {{ multiplier }}</span>
                <span class="text-fg-subtle">=</span>
                <span class="text-fg-muted">TDEE</span>
                <span class="font-semibold text-fg tabular-nums">{{ Math.round(tdee) }}</span>
            </div>
            <Field label="Goal">
                <Select v-model="goal" :options="goalChoices" />
            </Field>
            <div
                class="flex items-center justify-between gap-3 rounded-control bg-surface-2 px-3 py-2"
            >
                <span class="text-sm text-fg tabular-nums">
                    {{ goalCalories ?? "—" }} kcal/day
                </span>
                <Button
                    size="sm"
                    :disabled="goalCalories == null"
                    @click="apply"
                >
                    Apply to macros
                </Button>
            </div>
        </template>
        <p v-else class="m-0 text-sm text-fg-muted">
            Fill body profile and weight to estimate TDEE.
        </p>
    </div>
</template>
