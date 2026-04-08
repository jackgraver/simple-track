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

const dateLabel = computed(() => {
    const d = data.value?.date;
    return d ? formatDateLong(d) : "";
});

const splitLabel = computed(
    () => data.value?.workout_plan?.name ?? "No split assigned",
);

const switchingPlan = computed(() => switchPlanMutation.isPending.value);

const loggingRoute = computed(() => ({
    name: "logging",
    query: dayOffset.value === 0 ? {} : { offset: String(dayOffset.value) },
}));
</script>

<template>
    <div class="gym">
        <template v-if="isGymHome">
            <div v-if="isPending" class="state">Loading...</div>
            <div v-else-if="isError" class="state">
                Error: {{ error?.message ?? "Failed to load" }}
            </div>
            <template v-else>
                <div class="flex items-center gap-2 pt-2">
                    <router-link :to="{ name: 'gym-plans' }">
                        <p class="bg-firstBg hover:bg-secondBg rounded-md p-2">
                            Manage Plans
                        </p>
                    </router-link>
                    <router-link :to="{ name: 'progression' }">
                        <p class="bg-firstBg hover:bg-secondBg rounded-md p-2">
                            Progression
                        </p>
                    </router-link>
                </div>
                <div class="date-nav">
                    <button
                        type="button"
                        class="date-nav-button"
                        @click="goToPreviousDay"
                    >
                        <ChevronLeftIcon />
                    </button>
                    <p class="date">{{ dateLabel }}</p>
                    <button
                        type="button"
                        class="date-nav-button"
                        @click="goToNextDay"
                    >
                        <ChevronRightIcon />
                    </button>
                </div>
                <div
                    v-if="planOptions.length"
                    class="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3"
                >
                    <label
                        for="gym-plan-switch"
                        class="text-sm font-medium text-zinc-400"
                        >Today's plan</label
                    >
                    <select
                        id="gym-plan-switch"
                        class="w-full max-w-md rounded border border-zinc-600 bg-zinc-900 px-2 py-1.5 text-sm text-inherit disabled:opacity-50 sm:w-auto"
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
                </div>
                <h1 class="split">{{ splitLabel }}</h1>
                <WorkoutActivityGraph />
                <router-link :to="loggingRoute" class="gym-cta"
                    >Log workout</router-link
                >
            </template>
        </template>
        <router-view v-else />
    </div>
</template>

<style scoped>
.gym {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 44rem;
}
.state {
    color: #aaa;
}
.date-nav {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}
.date {
    margin: 0;
    font-size: 0.95rem;
    color: #b0b0b0;
}
.date-nav-button {
    padding: 0.35rem 0.7rem;
    border: 1px solid #666;
    border-radius: 6px;
    background: #2d2d2d;
    color: #fff;
    font: inherit;
    cursor: pointer;
}
.date-nav-button:hover {
    background: #3a3a3a;
}
.split {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
}
.gym-cta {
    display: inline-block;
    margin-top: 0.5rem;
    padding: 10px 18px;
    border-radius: 6px;
    background: #3a3a3a;
    color: #fff;
    text-decoration: none;
    font-weight: 500;
    border: 1px solid #666;
}
.gym-cta:hover {
    background: #4a4a4a;
}
</style>
