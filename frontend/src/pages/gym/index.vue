<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import axios from "axios";
import {
    useWorkoutLogToday,
    useWorkoutPlansAll,
    useSwitchWorkoutPlan,
} from "~/api/workout/queries";
import { formatDateLong } from "~/utils/dateUtil";
import { toast } from "~/composables/toast/useToast";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-vue-next";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";

const route = useRoute();
const router = useRouter();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});
const { data, isPending, isError, error } = useWorkoutLogToday(
    dayOffset,
    isGymHome,
);

const plansQuery = useWorkoutPlansAll();
const switchPlanMutation = useSwitchWorkoutPlan(dayOffset);

const planOptions = computed(() =>
    (plansQuery.data.value?.plans ?? []).map((p) => ({
        id: p.ID,
        name: p.name,
    })),
);

const currentPlanId = computed(() => data.value?.workout_plan_id ?? null);

const planSelectValue = computed(() => {
    const id = currentPlanId.value;
    return id == null ? "" : String(id);
});

const handleSwitchPlan = async (e: Event) => {
    const v = (e.target as HTMLSelectElement).value;
    const planId = v === "" ? null : Number(v);
    if (planId === currentPlanId.value) return;
    try {
        await switchPlanMutation.mutateAsync(planId);
        toast.push("Workout plan updated", "success");
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

const updateOffset = (nextOffset: number) => {
    const nextQuery = { ...route.query };
    if (nextOffset === 0) {
        delete nextQuery.offset;
    } else {
        nextQuery.offset = String(nextOffset);
    }

    router.push({
        name: "gym",
        query: nextQuery,
    });
};

const goToPreviousDay = () => {
    updateOffset(dayOffset.value + 1);
};

const goToNextDay = () => {
    updateOffset(dayOffset.value - 1);
};

const isToday = computed(() => dayOffset.value === 0);

const goToToday = () => {
    if (!isToday.value) updateOffset(0);
};

const dateLabel = computed(() => {
    const d = data.value?.date;
    return d ? formatDateLong(d) : "";
});

const switchingPlan = computed(() => switchPlanMutation.isPending.value);

const loggingRoute = computed(() => ({
    name: "logging",
    query: dayOffset.value === 0 ? {} : { offset: String(dayOffset.value) },
}));
</script>

<template>
    <div class="flex w-full flex-col gap-6 max-w-3xl mx-auto">
        <template v-if="isGymHome">
            <div v-if="isError" class="text-sm text-(--color-cf-red)">
                Error: {{ error?.message ?? "Failed to load" }}
            </div>
            <template v-else>
                <div
                    class="flex items-center justify-between gap-4 border-b border-(--color-border) pb-3"
                >
                    <h1 class="m-0 text-lg font-semibold text-textPrimary">
                        Gym
                    </h1>
                    <nav
                        class="flex items-center gap-3 text-sm text-textSecondary"
                    >
                        <router-link
                            :to="{ name: 'gym-plans' }"
                            class="transition-colors hover:text-textPrimary"
                            >Manage plans</router-link
                        >
                        <span aria-hidden="true" class="text-textSecondary/50"
                            >·</span
                        >
                        <router-link
                            :to="{ name: 'progression' }"
                            class="transition-colors hover:text-textPrimary"
                            >Progression</router-link
                        >
                        <span aria-hidden="true" class="text-textSecondary/50"
                            >·</span
                        >
                        <router-link
                            :to="{ name: 'gym-weight' }"
                            class="transition-colors hover:text-textPrimary"
                            >Weight</router-link
                        >
                        <span aria-hidden="true" class="text-textSecondary/50"
                            >·</span
                        >
                        <router-link
                            :to="{ name: 'gym-steps' }"
                            class="transition-colors hover:text-textPrimary"
                            >Steps</router-link
                        >
                    </nav>
                </div>
                <section class="flex flex-col gap-2">
                    <div class="flex items-center justify-between gap-3">
                        <button
                            type="button"
                            aria-label="Previous day"
                            class="rounded-md border border-(--color-border) bg-firstBg p-2 text-textPrimary transition-colors hover:bg-secondBg"
                            @click="goToPreviousDay"
                        >
                            <ChevronLeftIcon class="size-4" />
                        </button>
                        <p class="m-0 text-base font-medium text-textPrimary">
                            {{ dateLabel }}
                        </p>
                        <button
                            type="button"
                            aria-label="Next day"
                            class="rounded-md border border-(--color-border) bg-firstBg p-2 text-textPrimary transition-colors hover:bg-secondBg"
                            @click="goToNextDay"
                        >
                            <ChevronRightIcon class="size-4" />
                        </button>
                    </div>
                    <button
                        v-if="!isToday"
                        type="button"
                        class="self-center text-xs text-textSecondary underline-offset-2 transition-colors hover:text-textPrimary hover:underline"
                        @click="goToToday"
                    >
                        Jump to today
                    </button>
                </section>
                <section class="flex flex-col gap-2">
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
                </section>
                <router-link
                    :to="loggingRoute"
                    class="flex items-center justify-center rounded-md bg-secondBg px-4 py-3 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg"
                    >Log workout</router-link
                >
                <WorkoutActivityGraph />
            </template>
        </template>
        <router-view v-else />
    </div>
</template>
