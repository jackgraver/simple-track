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
import { homeKeys } from "~/pages/home/queries/keys";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
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
        queryClient.invalidateQueries({ queryKey: homeKeys.diet.all });
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
const submit = () => saveMutation.mutate();
function applyTdeeCalories(n: number) {
    sliders.setCalorieTarget(n);
}
</script>

<template>
    <div class="flex flex-col gap-5 pb-8 pt-2">
        <div class="flex items-baseline gap-3">
            <h1 class="m-0 text-lg font-semibold text-textPrimary">
                Macro Targets
            </h1>
            <span v-if="dayData?.day.plan?.name" class="text-sm text-zinc-500">
                {{ dayData.day.plan.name }}
            </span>
        </div>
        <div v-if="isPending" class="text-zinc-500">Loading…</div>
        <div v-else-if="loadError" class="text-red-400">
            {{ loadError.message }}
        </div>
        <template v-else>
            <section
                class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4"
            >
                <div
                    class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1"
                >
                    <h2 class="m-0 text-sm font-medium text-zinc-200">
                        TDEE estimate
                    </h2>
                    <span class="text-xs text-zinc-500"
                        >{{ weightLbs }} lb · {{ heightFtIn }} · {{ age }} yr ·
                        {{ sex }}</span
                    >
                </div>
                <TDEECalculator
                    :bmr="bmr"
                    :tdee="tdee"
                    :multiplier="multiplier"
                    @apply-calories="applyTdeeCalories"
                />
            </section>
            <section
                class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4"
            >
                <h2 class="m-0 text-sm font-medium text-zinc-200">Macros</h2>
                <MacroSlidersPanel :sliders="sliders" :weight-lbs="weightLbs" />
            </section>
            <button
                type="button"
                class="w-full rounded-md border border-amber-700/50 bg-amber-950/40 px-4 py-2.5 text-sm font-semibold text-amber-100 transition-colors hover:bg-amber-950/70 disabled:opacity-50"
                :disabled="saving || planId == null || !macrosValid"
                @click="submit"
            >
                {{ saving ? "Saving…" : "Save targets" }}
            </button>
        </template>
    </div>
</template>
