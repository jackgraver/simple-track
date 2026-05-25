<script setup lang="ts">
import { computed } from "vue";
import axios from "axios";
import {
    useWorkoutPlansAll,
    useSwitchWorkoutPlan,
} from "~/api/workout/queries";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    dayOffset: number;
    currentPlanId: number | null;
    onResolve?: (value: boolean) => void;
}>();

const plansQuery = useWorkoutPlansAll();
const switchPlanMutation = useSwitchWorkoutPlan(() => props.dayOffset);

const planOptions = computed(() =>
    (plansQuery.data.value?.plans ?? []).map((p) => ({
        id: p.ID,
        name: p.name,
    })),
);

const planSelectValue = computed(() => {
    const id = props.currentPlanId;
    return id == null ? "" : String(id);
});

const switchingPlan = computed(() => switchPlanMutation.isPending.value);

const handleSwitchPlan = async (e: Event) => {
    const v = (e.target as HTMLSelectElement).value;
    const planId = v === "" ? null : Number(v);
    if (planId === props.currentPlanId) return;
    try {
        await switchPlanMutation.mutateAsync(planId);
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
        <label
            for="gym-plan-switch"
            class="text-xs font-medium uppercase tracking-wide text-textSecondary"
            >Today's plan</label
        >
        <select
            v-if="planOptions.length"
            id="gym-plan-switch"
            class="w-full rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm text-textPrimary transition-colors hover:bg-secondBg disabled:opacity-50"
            :value="planSelectValue"
            :disabled="switchingPlan"
            @change="handleSwitchPlan"
        >
            <option value="">No plan</option>
            <option
                v-for="p in planOptions"
                :key="p.id"
                :value="String(p.id)"
            >
                {{ p.name }}
            </option>
        </select>
        <p v-else class="m-0 text-sm text-textSecondary">
            No plans yet.
            <router-link
                :to="{ name: 'gym-plans' }"
                class="text-textPrimary underline underline-offset-2 hover:no-underline"
                >Create one</router-link
            >
        </p>
    </div>
</template>
