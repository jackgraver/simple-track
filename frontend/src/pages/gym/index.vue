<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useWorkoutLogToday } from "~/api/workout/queries";
import WorkoutActivityGraph from "~/pages/gym/WorkoutActivityGraph.vue";
import SwitchWorkoutPlanDialog from "~/pages/gym/components/SwitchWorkoutPlanDialog.vue";
import ModuleTitleBar from "~/shared/ModuleTitleBar.vue";
import { dialogManager } from "~/composables/dialog/useDialog";
import { Pencil } from "lucide-vue-next";
import { Button, ErrorState, IconButton, Skeleton, useDelayedFlag } from "~/shared/ui";

const route = useRoute();
const isGymHome = computed(() => route.name === "gym");
const dayOffset = computed(() => {
    const raw = route.query.offset;
    const value = typeof raw === "string" ? Number.parseInt(raw, 10) : 0;
    return Number.isNaN(value) ? 0 : value;
});
const { data, isError, error, isPending, refetch } = useWorkoutLogToday(dayOffset, isGymHome);
const showSkeleton = useDelayedFlag(() => isPending.value);

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

</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <template v-if="isGymHome">
            <ErrorState v-if="isError" :error="error" title="Couldn't load today's workout" @retry="refetch()" />
            <template v-else>
                <ModuleTitleBar title="Gym" :links="[]" />
                <section class="flex items-center justify-between gap-2">
                    <Skeleton v-if="isPending && showSkeleton" width="10rem" height="1.75rem" />
                    <h1 v-else class="m-0 text-lg font-semibold text-fg">{{ data?.workout_plan?.name }} Day</h1>
                    <IconButton label="Switch workout plan" variant="secondary" @click="openSwitchPlanDialog()">
                        <Pencil />
                    </IconButton>
                </section>
                <Button variant="primary" size="lg" block :to="loggingRoute">Log workout</Button>
                <WorkoutActivityGraph />
            </template>
        </template>
        <router-view v-else />
    </div>
</template>
