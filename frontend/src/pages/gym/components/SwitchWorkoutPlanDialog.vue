<script setup lang="ts">
import { offsetToDate } from "~/api/workout/dates";
import { computed } from "vue";
import axios from "axios";
import {
    useWorkoutPlansAll,
    useWorkoutPrograms,
    useSwitchWorkoutPlan,
} from "~/api/workout/queries";
import { Field, Select } from "~/shared/ui";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    dayOffset: number;
    currentPlanId: number | null;
    onResolve?: (value: boolean) => void;
}>();

const plansQuery = useWorkoutPlansAll();
const programsQuery = useWorkoutPrograms();
const switchPlanMutation = useSwitchWorkoutPlan();

const planGroups = computed(() => {
    const plans = plansQuery.data.value?.plans ?? [];
    const programs = programsQuery.data.value?.programs ?? [];
    const groupedPlanIds = new Set<number>();
    const groups = programs
        .map((program) => {
            const programPlans = plans.filter(
                (plan) => plan.workout_program_id === program.ID,
            );
            for (const plan of programPlans) groupedPlanIds.add(plan.ID);
            return { id: program.ID, name: program.name, plans: programPlans };
        })
        .filter((group) => group.plans.length > 0);
    const ungroupedPlans = plans.filter((plan) => !groupedPlanIds.has(plan.ID));
    if (ungroupedPlans.length) {
        groups.push({ id: 0, name: "Other plans", plans: ungroupedPlans });
    }
    return groups;
});

const planSelectValue = computed(() => {
    const id = props.currentPlanId;
    return id == null ? "" : String(id);
});

const switchingPlan = computed(() => switchPlanMutation.isPending.value);

const planOptions = computed(() => [
    { label: "No plan", value: "" },
    ...planGroups.value.flatMap((group) =>
        group.plans.map((plan) => ({ label: plan.name, value: String(plan.ID), group: group.name })),
    ),
]);

const handleSwitchPlan = async (v: string) => {
    const planId = v === "" ? null : Number(v);
    if (planId === props.currentPlanId) return;
    try {
        await switchPlanMutation.mutateAsync({ date: offsetToDate(props.dayOffset), planId });
        toast.push("Workout plan updated", "success");
        props.onResolve?.(true);
    } catch (err: unknown) {
        let msg = "Failed to switch plan";
        if (
            axios.isAxiosError(err) &&
            err.response?.data &&
            typeof err.response.data === "object" &&
            "error" in err.response.data
        ) {
            const e0 = (err.response.data as { error?: string }).error;
            if (e0) msg = e0;
        } else if (err instanceof Error) {
            msg = err.message;
        }
        toast.push(msg, "error");
    }
};
</script>
<template>
    <div class="flex flex-col gap-2">
        <Field v-if="planGroups.length" label="Today's plan">
            <Select
                :model-value="planSelectValue"
                :options="planOptions"
                :disabled="switchingPlan"
                @update:model-value="handleSwitchPlan"
            />
        </Field>
        <p v-else class="m-0 text-sm text-fg-muted">
            No plans yet.
            <router-link :to="{ name: 'gym-plans' }" class="text-fg underline underline-offset-2 hover:no-underline">Create one</router-link>
        </p>
    </div>
</template>
