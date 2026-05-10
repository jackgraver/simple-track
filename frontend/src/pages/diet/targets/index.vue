<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { updatePlanMacros } from "~/api/diet/api";
import { fetchProfile, fetchWeightLogs, saveProfile } from "~/api/tracking/api";
import { trackingKeys } from "~/api/tracking/keys";
import type { ActivityLevel } from "~/api/tracking/types";
import { toast } from "~/composables/toast/useToast";
import { homeKeys } from "~/pages/home/queries/keys";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import BodyProfileCard from "./components/BodyProfileCard.vue";
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
const { data: profileRow, isPending: profilePending } = useQuery({
    queryKey: trackingKeys.profile,
    queryFn: fetchProfile,
});
const { data: weightLogs } = useQuery({
    queryKey: trackingKeys.weight,
    queryFn: () => fetchWeightLogs(1),
});
const latestLogWeightLbs = computed(() => {
    const first = weightLogs.value?.[0];
    return first != null && first.weight_lbs > 0 ? first.weight_lbs : null;
});
const weightLbs = ref(0);
const heightIn = ref(70);
const age = ref(30);
const sex = ref<"male" | "female">("male");
const activityLevel = ref<ActivityLevel>("moderately_active");
watch(
    profileRow,
    (p) => {
        if (!p) return;
        heightIn.value = p.height_in;
        age.value = p.age;
        sex.value = p.sex;
        activityLevel.value = p.activity_level;
    },
    { immediate: true },
);
const sliders = useMacroSliders();
watch(
    () => dayData.value?.day.plan,
    (plan) => {
        if (!plan) return;
        sliders.calorieTarget.value = plan.calories;
        sliders.proteinG.value = plan.protein;
        sliders.carbsG.value = plan.carbs;
        sliders.fatG.value = plan.fat ?? 0;
        sliders.fiberG.value = plan.fiber;
    },
    { immediate: true },
);
const { bmr, tdee, multiplier } = useTDEE({
    weightLbs: () => weightLbs.value,
    heightIn: () => heightIn.value,
    age: () => age.value,
    sex: () => sex.value,
    activity: () => activityLevel.value,
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
const profileValid = computed(
    () => heightIn.value > 0 && age.value > 0 && weightLbs.value > 0,
);
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
const profileMutation = useMutation({
    mutationFn: () =>
        saveProfile({
            height_in: heightIn.value,
            age: age.value,
            sex: sex.value,
            activity_level: activityLevel.value,
        }),
    onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: trackingKeys.profile });
        toast.push("Profile saved", "success");
    },
    onError: () => {
        toast.push("Could not save profile", "error");
    },
});
const saving = computed(() => saveMutation.isPending.value);
const savingProfile = computed(() => profileMutation.isPending.value);
const isPending = computed(() => dayPending.value || profilePending.value);
const submit = () => saveMutation.mutate();
const saveProfileAction = () => profileMutation.mutate();
const goBack = () => {
    router.push({ name: "diet" });
};
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
            <div class="grid grid-cols-1 items-start gap-5 md:grid-cols-2">
                <section
                    class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4"
                >
                    <div class="flex items-center justify-between">
                        <h2 class="m-0 text-sm font-medium text-zinc-200">
                            Body Profile
                        </h2>
                        <button
                            type="button"
                            class="rounded-md border border-zinc-600 bg-zinc-800 px-3 py-1.5 text-xs font-medium text-zinc-100 hover:bg-zinc-700 disabled:opacity-50"
                            :disabled="savingProfile || !profileValid"
                            @click="saveProfileAction"
                        >
                            {{ savingProfile ? "Saving…" : "Save profile" }}
                        </button>
                    </div>
                    <BodyProfileCard
                        :weight-lbs="weightLbs"
                        :height-in="heightIn"
                        :age="age"
                        :sex="sex"
                        :activity-level="activityLevel"
                        :latest-log-weight-lbs="latestLogWeightLbs"
                        @update:weight-lbs="weightLbs = $event"
                        @update:height-in="heightIn = $event"
                        @update:age="age = $event"
                        @update:sex="sex = $event"
                        @update:activity-level="activityLevel = $event"
                    />
                    <p v-if="!profileValid" class="m-0 text-xs text-zinc-500">
                        Add weight, height, and age for TDEE.
                    </p>
                </section>
                <section
                    class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4"
                >
                    <h2 class="m-0 text-sm font-medium text-zinc-200">
                        TDEE Estimate
                    </h2>
                    <TDEECalculator
                        :bmr="bmr"
                        :tdee="tdee"
                        :multiplier="multiplier"
                        @apply-calories="applyTdeeCalories"
                    />
                </section>
            </div>
            <section
                class="flex flex-col gap-3 rounded-lg border border-zinc-700 bg-zinc-950/40 p-4"
            >
                <h2 class="m-0 text-sm font-medium text-zinc-200">Macros</h2>
                <MacroSlidersPanel :sliders="sliders" :weight-lbs="weightLbs" />
            </section>
            <form @submit.prevent="submit">
                <button
                    type="submit"
                    class="w-full rounded-md border border-amber-700/50 bg-amber-950/40 px-4 py-2.5 text-sm font-semibold text-amber-100 transition-colors hover:bg-amber-950/70 disabled:opacity-50"
                    :disabled="saving || planId == null || !macrosValid"
                >
                    {{ saving ? "Saving…" : "Save targets" }}
                </button>
            </form>
        </template>
    </div>
</template>
