<script setup lang="ts">
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { computed, watch } from "vue";
import { useRouter } from "vue-router";
import { updatePlanMacros } from "~/api/diet/api";
import {
    latestWeightLbs,
    useProfile,
    useWeightLogs,
} from "~/api/tracking/queries";
import { toast } from "~/composables/toast/useToast";
import { dietKeys } from "~/api/diet/keys";
import { useDietLogsToday } from "~/api/diet/queries";
import { Button, Card, ErrorState, PageHeader, Section, Skeleton, useDelayedFlag } from "~/shared/ui";
import MacroSlidersPanel from "./components/MacroSlidersPanel.vue";
import TDEECalculator from "./components/TDEECalculator.vue";
import { useMacroSliders } from "./useMacroSliders";
import { useTDEE } from "./useTDEE";

const router = useRouter();
const queryClient = useQueryClient();
const {
    data: dayData,
    isPending: dayPending,
    error: loadError,
    refetch: refetchDay,
} = useDietLogsToday(0);
const { data: profile, isPending: profilePending } = useProfile();
const { data: weightLogs, isPending: weightPending } = useWeightLogs();
const latestLogWeightLbs = computed(() => latestWeightLbs(weightLogs.value));
const weightLbs = computed(() => latestLogWeightLbs.value ?? 0);
const heightIn = computed(() => profile.value?.height_in ?? 0);
const heightFtIn = computed(() => {
    const total = Math.round(heightIn.value);
    return `${Math.floor(total / 12)}'${total % 12}"`;
});
const age = computed(() => profile.value?.age ?? 0);
const sex = computed(() => profile.value?.sex ?? "male");

const sliders = useMacroSliders();
watch(
    () => dayData.value?.day.plan,
    (plan) => {
        if (!plan) return;
        sliders.setMacroTargets({
            calories: plan.calories,
            protein: plan.protein,
            carbs: plan.carbs,
            fat: plan.fat ?? 0,
            fiber: plan.fiber,
        });
    },
    { immediate: true },
);
const { bmr, tdee, multiplier } = useTDEE({
    weightLbs: () => weightLbs.value,
    heightIn: () => heightIn.value,
    age: () => age.value,
    sex: () => sex.value,
    activity: () => "moderately_active",
});
const planId = computed(() => dayData.value?.day.plan.ID);
const macrosValid = computed(() => {
    const nums = [
        sliders.calorieTarget.value,
        sliders.proteinG.value,
        sliders.carbsG.value,
        sliders.fatG.value,
        sliders.fiberG.value,
    ];
    return nums.every(
        (n) => typeof n === "number" && !Number.isNaN(n) && n >= 0,
    );
});
const saveMutation = useMutation({
    mutationFn: async () => {
        const id = planId.value;
        if (id == null) throw new Error("No diet plan loaded");
        return updatePlanMacros(id, {
            calories: sliders.calorieTarget.value,
            protein: sliders.proteinG.value,
            fiber: sliders.fiberG.value,
            carbs: sliders.carbsG.value,
            fat: sliders.fatG.value,
        });
    },
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: dietKeys.days.all });
        toast.push("Macro targets saved", "success");
        router.push({ name: "diet" });
    },
    onError: () => {
        toast.push("Could not save macro targets", "error");
    },
});
const saving = computed(() => saveMutation.isPending.value);
const isPending = computed(
    () => dayPending.value || profilePending.value || weightPending.value,
);
const showSkeleton = useDelayedFlag(isPending);
const submit = () => saveMutation.mutate();
function applyTdeeCalories(n: number) {
    sliders.setCalorieTarget(n);
}
</script>

<template>
    <div class="mx-auto flex w-full max-w-3xl flex-col gap-6 pt-2 pb-above-tabbar lg:pb-10">
        <PageHeader title="Macro targets">
            <template v-if="dayData?.day.plan?.name" #meta>{{
                dayData.day.plan.name
            }}</template>
        </PageHeader>
        <ErrorState
            v-if="loadError"
            title="Couldn't load your targets"
            :error="loadError"
            @retry="refetchDay()"
        />
        <div v-else-if="isPending" class="flex flex-col gap-6">
            <template v-if="showSkeleton">
                <Skeleton shape="block" height="12rem" />
                <Skeleton shape="block" height="20rem" />
            </template>
        </div>
        <template v-else>
            <Section title="TDEE estimate">
                <Card>
                    <div class="flex flex-col gap-4">
                        <p class="m-0 text-xs text-fg-muted">
                            {{ weightLbs }} lb · {{ heightFtIn }} · {{ age }} yr ·
                            {{ sex }}
                        </p>
                        <TDEECalculator
                            :bmr="bmr"
                            :tdee="tdee"
                            :multiplier="multiplier"
                            @apply-calories="applyTdeeCalories"
                        />
                    </div>
                </Card>
            </Section>
            <Section title="Macros">
                <Card>
                    <MacroSlidersPanel
                        :sliders="sliders"
                        :weight-lbs="weightLbs"
                    />
                </Card>
            </Section>
            <div>
                <Button
                    variant="primary"
                    size="lg"
                    class="w-full lg:w-auto"
                    :loading="saving"
                    :disabled="planId == null || !macrosValid"
                    @click="submit"
                >
                    Save targets
                </Button>
            </div>
        </template>
    </div>
</template>
