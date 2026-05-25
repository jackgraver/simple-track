<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useWorkoutLogToday } from "~/api/workout/queries";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";
import SwitchWorkoutPlanDialog from "~/pages/gym/components/SwitchWorkoutPlanDialog.vue";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { Pencil } from "lucide-vue-next";

const route = useRoute();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});
const { data, isError, error } = useWorkoutLogToday(dayOffset, isGymHome);

const currentPlanId = computed(() => data.value?.workout_plan_id ?? null);

const openSwitchPlanDialog = () => {
    void dialogManager.custom<boolean>({
        title: "Switch workout plan",
        component: SwitchWorkoutPlanDialog,
        componentProps: {
            dayOffset: dayOffset.value,
            currentPlanId: currentPlanId.value,
        },
    });
};

const loggingRoute = computed(() => ({
    name: "logging",
    query: dayOffset.value === 0 ? {} : { offset: String(dayOffset.value) },
}));

const gymNavLinks = [
    { name: "Manage plans", to: "gym-plans" },
    { name: "Progression", to: "progression" },
    { name: "Weight", to: "gym-weight", offset: true },
    { name: "Steps", to: "gym-steps", offset: true },
] as const;
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <template v-if="isGymHome">
            <div v-if="isError" class="text-sm text-(--color-cf-red)">
                Error: {{ error?.message ?? "Failed to load" }}
            </div>
            <template v-else>
                <ModuleTitleBar
                    title="Gym"
                    :day-offset="dayOffset"
                    :links="[...gymNavLinks]"
                />
                <section class="flex flex-col gap-2">
                    <div class="flex items-center justify-between">
                        <h1 class="text-lg font-semibold text-textPrimary">
                            {{ data?.workout_plan?.name }} Day
                        </h1>
                        <button
                            type="button"
                            class="flex items-center gap-1 rounded-md border px-2 py-1 m-0! text-sm font-medium text-zinc-200 transition-colors hover:bg-secondBg"
                            @click="openSwitchPlanDialog()"
                        >
                            <Pencil :size="15" />
                        </button>
                    </div>
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
