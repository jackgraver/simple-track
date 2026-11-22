<script setup lang="ts">
import Chart from "~/shared/chart/Chart.vue";
import { computed, ref, toRef } from "vue";
import { EmptyState, ErrorState, SegmentedControl, Skeleton, useDelayedFlag } from "~/shared/ui";
import { LineChart } from "lucide-vue-next";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import {
    useProgressionCharts,
    roundWeightHalfLb,
    type ProgressionEntry,
    type ProgressionRange,
} from "~/pages/gym/progression/useProgressionCharts";

const props = defineProps<{
    exerciseId: number;
    lastTimeSets?: { weight: number; reps: number }[];
    initialRange?: ProgressionRange;
}>();

type ViewMode = "chart" | "text";

const range = ref<ProgressionRange>(props.initialRange ?? "1m");
const viewMode = ref<ViewMode>("chart");
const rangeLabel = computed(
    () =>
        ({
            "1m": "last month",
            "3m": "last 3 months",
            "6m": "last 6 months",
            "1y": "last year",
            all: "this exercise",
        })[range.value],
);

const viewOptions: { value: ViewMode; label: string }[] = [
    { value: "chart", label: "Graph" },
    { value: "text", label: "List" },
];
const rangeOptions: { value: ProgressionRange; label: string }[] = [
    { value: "1m", label: "1 mo" },
    { value: "3m", label: "3 mo" },
    { value: "6m", label: "6 mo" },
    { value: "1y", label: "1 yr" },
    { value: "all", label: "All" },
];

const exerciseId = toRef(props, "exerciseId");

const {
    data: progressionPayload,
    isPending: loading,
    error: fetchError,
    refetch,
} = useQuery({
    queryKey: computed(() => [
        "workout",
        "exercises",
        "progression",
        exerciseId.value,
    ]),
    queryFn: async () => {
        const id = exerciseId.value;
        const res = await apiClient.get<{ progression: ProgressionEntry[] }>(
            `/workout/exercises/progression/${id}`,
        );
        return res.data;
    },
    enabled: computed(() => Number.isFinite(exerciseId.value)),
});

const progression = computed(() => progressionPayload.value?.progression ?? []);

const showSkeleton = useDelayedFlag(() => loading.value);

const {
    progressionChartData,
    progressionChartOptions,
    hasProgressionChartData,
    progressionDayGroups,
    hasProgressionData,
} = useProgressionCharts(progression, range);

function formatSetKey(set: { weight: number; reps: number }): string {
    return `${roundWeightHalfLb(Number(set.weight))}x${Number(set.reps)}`;
}

const lastTimeDayKey = computed(() => {
    const targetSets = props.lastTimeSets;
    if (!targetSets?.length) return null;
    const targetKey = targetSets.map(formatSetKey).join("|");
    for (const day of progressionDayGroups.value) {
        if (day.sets.map(formatSetKey).join("|") === targetKey) {
            return day.dayKey;
        }
    }
    return null;
});
</script>

<template>
    <div class="flex min-w-0 flex-col gap-3">
        <ErrorState v-if="fetchError" :error="fetchError" title="Couldn't load progression" @retry="refetch()" />
        <template v-else-if="loading">
            <Skeleton v-if="showSkeleton" shape="block" height="16rem" />
        </template>
        <template v-else>
            <div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                <div class="lg:w-80"><SegmentedControl v-model="range" label="Range" :options="rangeOptions" /></div>
                <div class="lg:w-40"><SegmentedControl v-model="viewMode" label="View" :options="viewOptions" /></div>
            </div>
            <EmptyState
                v-if="!hasProgressionData"
                :icon="LineChart"
                :title="`No data in the ${rangeLabel}`"
                description="Log sets for this exercise to see progress."
            />
            <div v-else-if="viewMode === 'chart'" class="flex min-h-0 min-w-0 flex-col gap-2">
                <div v-if="hasProgressionChartData" class="h-64 min-h-64 w-full min-w-0 lg:h-80 lg:min-h-80">
                    <Chart type="line" :data="progressionChartData" :options="progressionChartOptions" class="h-full w-full" />
                </div>
                <p v-else class="m-0 text-sm text-fg-muted">No chart data in the {{ rangeLabel }}.</p>
            </div>
            <div v-else class="flex max-h-80 flex-col gap-3 overflow-y-auto text-sm">
                <div v-for="day in progressionDayGroups" :key="day.dayKey" class="flex flex-col gap-1">
                    <p class="m-0 font-medium text-fg">
                        {{ day.dateLabel }}<span v-if="day.dayKey === lastTimeDayKey" class="font-normal text-fg-muted"> (last time)</span>
                    </p>
                    <ul class="m-0 list-none space-y-0.5 pl-0">
                        <li v-for="(set, index) in day.sets" :key="`${day.dayKey}-${index}`" class="text-fg-muted tabular-nums">
                            - Set {{ index + 1 }}: {{ set.weight }}x{{ set.reps }}
                        </li>
                    </ul>
                </div>
            </div>
        </template>
    </div>
</template>
