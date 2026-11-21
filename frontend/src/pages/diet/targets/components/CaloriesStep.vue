<script setup lang="ts">
import { computed } from "vue";
import type { ActivityLevel } from "~/api/tracking/types";
import { Button, Card, SegmentedControl, Select, Skeleton } from "~/shared/ui";
import { fmtInt, fmtTrend, type IntakeSummary, MIN_LOGGED_DAYS } from "../domain/targets";
import type { MacroSlidersApi } from "../useMacroSliders";
import { ACTIVITY_MULTIPLIERS, calorieTargetFromTdee, type CalorieGoalId } from "../useTDEE";
import MacroSlider from "./MacroSlider.vue";

const props = defineProps<{
    sliders: MacroSlidersApi;
    bmr: number | null;
    formulaTdee: number | null;
    observed: number | null;
    intake: IntakeSummary | null;
    intakeError: boolean;
    trend: number | null;
    weightLbs: number;
    hasProfile: boolean;
    activitySaving: boolean;
}>();
const activity = defineModel<ActivityLevel>("activity", { required: true });
const emit = defineEmits<{ retryIntake: [] }>();

const ACTIVITY_LABELS: Record<ActivityLevel, string> = {
    sedentary: "Sedentary",
    lightly_active: "Lightly active",
    moderately_active: "Moderately active",
    very_active: "Very active",
    extra_active: "Extra active",
};
const activityOptions = (Object.keys(ACTIVITY_LABELS) as ActivityLevel[]).map((v) => ({
    value: v,
    label: `${ACTIVITY_LABELS[v]} ×${ACTIVITY_MULTIPLIERS[v]}`,
}));

// Goal chips adjust from observed maintenance when there is enough data, else from the formula.
const base = computed(() => props.observed ?? props.formulaTdee);
const baseLabel = computed(() =>
    props.observed != null ? "observed maintenance" : props.formulaTdee != null ? "formula TDEE" : null,
);
type GoalChoice = CalorieGoalId | "custom";
const goalOptions: { label: string; value: GoalChoice }[] = [
    { label: "−500", value: "lose_agg" },
    { label: "−250", value: "lose_mod" },
    { label: "±0", value: "maintain" },
    { label: "+250", value: "gain_mod" },
    { label: "+500", value: "gain_agg" },
];
const goal = computed<GoalChoice>({
    get: () => {
        const b = base.value;
        if (b == null) return "custom";
        const hit = goalOptions.find(
            (o) => o.value !== "custom" && calorieTargetFromTdee(b, o.value) === props.sliders.calorieTarget.value,
        );
        return hit?.value ?? "custom";
    },
    set: (g) => {
        if (g === "custom" || base.value == null) return;
        props.sliders.setCalorieTarget(calorieTargetFromTdee(base.value, g));
    },
});

const markers = computed(() =>
    [props.formulaTdee, props.observed].filter((n): n is number => n != null).map((n) => Math.round(n)),
);
</script>

<template>
    <Card>
        <div class="flex flex-col gap-4">
            <MacroSlider
                label="Daily calories"
                unit="kcal"
                :grams="sliders.calorieTarget.value"
                :min-grams="0"
                :max-grams="5000"
                :display-max-grams="5000"
                track-bg-class="bg-macro-calories/70"
                hide-lock
                :marker-grams="markers"
                @update:grams="sliders.setCalorieTarget"
            />
            <dl v-if="markers.length" class="m-0 flex flex-wrap gap-x-4 gap-y-1 text-xs text-fg-muted">
                <div v-if="formulaTdee != null" class="flex items-center gap-1.5">
                    <span aria-hidden="true" class="h-3 w-0.5 bg-fg/80"></span>
                    <dt>Formula</dt>
                    <dd class="m-0 font-medium text-fg tabular-nums">{{ fmtInt(formulaTdee) }}</dd>
                </div>
                <div v-if="observed != null" class="flex items-center gap-1.5">
                    <span aria-hidden="true" class="h-3 w-0.5 bg-fg/80"></span>
                    <dt>Observed</dt>
                    <dd class="m-0 font-medium text-fg tabular-nums">~{{ fmtInt(observed) }}</dd>
                </div>
            </dl>

            <div v-if="base != null" class="flex flex-col gap-2">
                <SegmentedControl v-model="goal" :options="goalOptions" label="Goal" />
                <p class="m-0 text-xs text-fg-muted">
                    kcal/day from {{ baseLabel }} ({{ fmtInt(base) }})
                </p>
            </div>

            <ul class="m-0 flex list-none flex-col gap-2 border-t border-line p-0 pt-4 text-sm text-fg-muted">
                <li class="flex flex-wrap items-center gap-x-2">
                    <span class="text-fg">Last 7 days</span>
                    <template v-if="intakeError">
                        <span>couldn't load</span>
                        <Button size="sm" variant="ghost" @click="emit('retryIntake')">Retry</Button>
                    </template>
                    <Skeleton v-else-if="!intake" shape="line" width="10rem" />
                    <span v-else-if="intake.loggedDays === 0">nothing logged</span>
                    <span v-else class="tabular-nums">
                        {{ fmtInt(intake.avgCalories ?? 0) }} kcal avg over {{ intake.loggedDays }}
                        {{ intake.loggedDays === 1 ? "day" : "days" }}
                        <template v-if="intake.loggedDays < MIN_LOGGED_DAYS">
                            · log {{ MIN_LOGGED_DAYS }}+ days to see observed maintenance</template
                        >
                    </span>
                </li>
                <li class="flex flex-wrap items-center gap-x-2">
                    <span class="text-fg">Weight</span>
                    <span v-if="weightLbs <= 0">no weigh-ins yet</span>
                    <span v-else class="tabular-nums">
                        {{ weightLbs }} lb{{ trend != null ? `, ${fmtTrend(trend)}` : " · log a few more weigh-ins for a trend" }}
                    </span>
                </li>
                <li v-if="hasProfile" class="flex flex-col gap-2 sm:flex-row sm:items-center">
                    <span class="flex flex-wrap gap-x-2">
                        <span class="text-fg">Formula</span>
                        <span v-if="bmr != null" class="tabular-nums">BMR {{ fmtInt(bmr) }} ×</span>
                        <span v-else>needs a weigh-in</span>
                    </span>
                    <div class="sm:w-56">
                        <Select
                            v-model="activity"
                            :options="activityOptions"
                            size="sm"
                            aria-label="Activity level"
                            :disabled="activitySaving"
                        />
                    </div>
                </li>
                <li v-else class="flex flex-wrap items-center gap-x-2">
                    <span class="text-fg">Formula</span>
                    <span>needs your height, age and sex</span>
                    <Button size="sm" variant="ghost" :to="{ name: 'settings-profile' }">Add body profile</Button>
                </li>
            </ul>
        </div>
    </Card>
</template>
