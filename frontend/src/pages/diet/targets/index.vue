<script setup lang="ts">
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { computed, ref, watch } from "vue";
import { updatePlanMacros } from "~/api/diet/api";
import { dietKeys } from "~/api/diet/keys";
import { useDietLogsToday, usePlans, useRecentDietDays } from "~/api/diet/queries";
import { latestWeightLbs, useProfile, useSaveProfile, useWeightLogs } from "~/api/tracking/queries";
import type { ActivityLevel } from "~/api/tracking/types";
import { toast } from "~/composables/toast/useToast";
import {
    Badge,
    Button,
    Card,
    EmptyState,
    ErrorState,
    List,
    ListRow,
    PageHeader,
    Section,
    Skeleton,
    SkeletonList,
    useDelayedFlag,
    useIsDesktop,
} from "~/shared/ui";
import CaloriesStep from "./components/CaloriesStep.vue";
import SplitStep from "./components/SplitStep.vue";
import {
    fmtInt,
    fmtSince,
    intakeSummary,
    observedMaintenance,
    planHistory,
    type Targets,
    targetsEqual,
    weightTrendLbPerWeek,
} from "./domain/targets";
import { useMacroSliders } from "./useMacroSliders";
import { useTDEE } from "./useTDEE";

const queryClient = useQueryClient();
const isDesktop = useIsDesktop();

const { data: dayData, isPending: dayPending, error: loadError, refetch: refetchDay } = useDietLogsToday(0);
const { data: profile, isPending: profilePending } = useProfile();
const { data: weightLogs, isPending: weightPending } = useWeightLogs();
const recentDays = useRecentDietDays(7);
const plansQuery = usePlans();

const plan = computed(() => dayData.value?.day.plan ?? null);
const planId = computed(() => plan.value?.ID || null);
const weightLbs = computed(() => latestWeightLbs(weightLogs.value) ?? 0);

// ---- editor state: sliders start from the saved plan; "dirty" compares against it.
const sliders = useMacroSliders();
const baseline = computed<Targets | null>(() =>
    plan.value
        ? {
              calories: plan.value.calories,
              protein: plan.value.protein,
              carbs: plan.value.carbs,
              fat: plan.value.fat ?? 0,
              fiber: plan.value.fiber,
          }
        : null,
);
function resetToSaved() {
    if (baseline.value) sliders.setMacroTargets(baseline.value);
}
watch(baseline, resetToSaved, { immediate: true });

const current = computed<Targets>(() => ({
    calories: sliders.calorieTarget.value,
    protein: sliders.proteinG.value,
    carbs: sliders.carbsG.value,
    fat: sliders.fatG.value,
    fiber: sliders.fiberG.value,
}));
const dirty = computed(() => baseline.value != null && !targetsEqual(current.value, baseline.value));
const valid = computed(() => Object.values(current.value).every((n) => Number.isFinite(n) && n >= 0));

// ---- calories evidence: formula (profile + weight), recent intake, weight trend.
const activity = ref<ActivityLevel>("moderately_active");
watch(
    () => profile.value?.activity_level,
    (a) => {
        if (a) activity.value = a;
    },
    { immediate: true },
);
const saveProfile = useSaveProfile();
watch(activity, (a) => {
    const p = profile.value;
    if (!p || p.activity_level === a) return;
    saveProfile.mutate(
        { height_in: p.height_in, age: p.age, sex: p.sex, activity_level: a },
        {
            onError: () => {
                toast.push("Could not save activity level", "error");
                activity.value = p.activity_level;
            },
        },
    );
});

const { bmr, tdee } = useTDEE({
    weightLbs: () => weightLbs.value,
    heightIn: () => profile.value?.height_in ?? 0,
    age: () => profile.value?.age ?? 0,
    sex: () => profile.value?.sex ?? "male",
    activity: () => activity.value,
});

const intakeError = computed(() => recentDays.value.some((q) => q.isError));
const intake = computed(() => {
    const qs = recentDays.value;
    if (qs.some((q) => !q.data)) return null;
    return intakeSummary(qs.map((q) => q.data!));
});
function retryIntake() {
    for (const q of recentDays.value) if (q.isError) void q.refetch();
}
const trend = computed(() => weightTrendLbPerWeek(weightLogs.value));
const observed = computed(() => (intake.value ? observedMaintenance(intake.value, trend.value) : null));

// ---- history
const plans = computed(() => planHistory(plansQuery.data.value ?? []));

// ---- save: stays on the page; the refetched plan becomes the new baseline.
const saveMutation = useMutation({
    mutationFn: () => {
        if (planId.value == null) throw new Error("No diet plan loaded");
        return updatePlanMacros(planId.value, {
            calories: current.value.calories,
            protein: current.value.protein,
            carbs: current.value.carbs,
            fat: current.value.fat,
            fiber: current.value.fiber,
        });
    },
    onSuccess: async () => {
        await Promise.all([
            queryClient.invalidateQueries({ queryKey: dietKeys.days.all }),
            queryClient.invalidateQueries({ queryKey: dietKeys.plans.all }),
        ]);
        toast.push("Targets saved", "success");
    },
    onError: () => toast.push("Could not save targets", "error"),
});
const saving = computed(() => saveMutation.isPending.value);
const canSave = computed(() => dirty.value && valid.value && planId.value != null);
const save = () => saveMutation.mutate();

const isPending = computed(() => dayPending.value || profilePending.value || weightPending.value);
const showSkeleton = useDelayedFlag(isPending);
</script>

<template>
    <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 pt-2 pb-above-tabbar lg:pb-10">
        <PageHeader title="Macro targets">
            <template v-if="plan" #meta>
                {{ plan.name }}<template v-if="plan.effective_from"> · since {{ fmtSince(plan.effective_from) }}</template>
            </template>
            <template v-if="plan && (dirty || isDesktop)" #actions>
                <Badge v-if="dirty" tone="warn">Unsaved changes</Badge>
                <Button v-if="dirty" variant="ghost" size="sm" :disabled="saving" @click="resetToSaved">Reset</Button>
                <Button v-if="isDesktop" variant="primary" :loading="saving" :disabled="!canSave" @click="save">
                    Save targets
                </Button>
            </template>
        </PageHeader>

        <ErrorState v-if="loadError" title="Couldn't load your targets" :error="loadError" @retry="refetchDay()" />
        <div v-else-if="isPending" class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <template v-if="showSkeleton">
                <Skeleton shape="block" height="20rem" />
                <Skeleton shape="block" height="28rem" />
            </template>
        </div>
        <Card v-else-if="!plan || planId == null">
            <EmptyState title="No diet plan yet" description="Targets belong to a diet plan. Log a meal to create today's plan." />
        </Card>
        <template v-else>
            <div class="grid grid-cols-1 items-start gap-6 lg:grid-cols-2">
                <Section title="1 · Calories">
                    <template #actions>
                        <span class="text-base font-semibold text-fg tabular-nums">
                            {{ fmtInt(sliders.calorieTarget.value) }} kcal
                        </span>
                    </template>
                    <CaloriesStep
                        v-model:activity="activity"
                        :sliders="sliders"
                        :bmr="bmr"
                        :formula-tdee="tdee"
                        :observed="observed"
                        :intake="intake"
                        :intake-error="intakeError"
                        :trend="trend"
                        :weight-lbs="weightLbs"
                        :has-profile="!!profile"
                        :activity-saving="saveProfile.isPending.value"
                        @retry-intake="retryIntake"
                    />
                </Section>
                <Section title="2 · Split">
                    <SplitStep :sliders="sliders" :weight-lbs="weightLbs" />
                </Section>
            </div>

            <Section title="History">
                <ErrorState
                    v-if="plansQuery.error.value"
                    title="Couldn't load plan history"
                    :error="plansQuery.error.value"
                    @retry="plansQuery.refetch()"
                />
                <Card v-else-if="plansQuery.isPending.value" padding="none">
                    <SkeletonList :rows="2" />
                </Card>
                <Card v-else-if="plans.length === 0">
                    <EmptyState title="No saved plans yet" />
                </Card>
                <Card v-else padding="none">
                    <List>
                        <ListRow v-for="p in plans" :key="p.ID">
                            {{ p.name }}
                            <template #subtitle>
                                {{ p.effective_from ? `since ${fmtSince(p.effective_from)}` : "no start date" }}
                            </template>
                            <template #trailing>
                                <Badge v-if="p.ID === planId" tone="accent">Current</Badge>
                                <span>
                                    {{ fmtInt(p.calories) }} kcal · P{{ Math.round(p.protein) }} C{{
                                        Math.round(p.carbs)
                                    }}
                                    F{{ Math.round(p.fat ?? 0) }}
                                </span>
                            </template>
                        </ListRow>
                    </List>
                </Card>
            </Section>

            <Button
                v-if="!isDesktop"
                variant="primary"
                size="lg"
                block
                :loading="saving"
                :disabled="!canSave"
                @click="save"
            >
                {{ dirty ? "Save targets" : "Targets saved" }}
            </Button>
        </template>
    </div>
</template>
