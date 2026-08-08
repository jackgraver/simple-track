<script setup lang="ts">
import Chart from "primevue/chart";
import { computed, ref, toRef } from "vue";
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

const viewOptions: { id: ViewMode; label: string }[] = [
    { id: "chart", label: "Graph" },
    { id: "text", label: "List" },
];
const rangeOptions: { id: ProgressionRange; label: string }[] = [
    { id: "1m", label: "1 mo" },
    { id: "3m", label: "3 mo" },
    { id: "6m", label: "6 mo" },
    { id: "1y", label: "1 yr" },
    { id: "all", label: "All" },
];

const exerciseId = toRef(props, "exerciseId");

const {
    data: progressionPayload,
    isPending: loading,
    error: fetchError,
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

const error = computed(() => fetchError.value?.message || null);

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
        <div
            v-if="loading"
            class="py-4 text-center text-sm text-textSecondary"
        >
            Loading…
        </div>
        <div
            v-else-if="error"
            class="py-4 text-center text-sm text-(--color-cf-red)"
        >
            {{ error }}
        </div>
        <template v-else-if="!hasProgressionData">
            <p class="m-0 py-4 text-center text-sm text-textSecondary">
                No data in the {{ rangeLabel }}.
            </p>
        </template>
        <template v-else>
            <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex flex-wrap items-center gap-1">
                    <span class="mr-1 text-xs text-textSecondary">Range</span>
                    <button
                        v-for="opt in rangeOptions"
                        :key="opt.id"
                        type="button"
                        class="rounded px-2 py-0.5 text-xs transition-colors"
                        :class="
                            range === opt.id
                                ? 'bg-secondBg text-textPrimary'
                                : 'text-textSecondary hover:bg-firstBg hover:text-textPrimary'
                        "
                        @click="range = opt.id"
                    >
                        {{ opt.label }}
                    </button>
                </div>
                <div class="flex flex-wrap gap-1">
                    <button
                        v-for="opt in viewOptions"
                        :key="opt.id"
                        type="button"
                        class="rounded px-2 py-0.5 text-xs transition-colors"
                        :class="
                            viewMode === opt.id
                                ? 'bg-secondBg text-textPrimary'
                                : 'text-textSecondary hover:bg-firstBg hover:text-textPrimary'
                        "
                        @click="viewMode = opt.id"
                    >
                        {{ opt.label }}
                    </button>
                </div>
            </div>
            <div
                v-if="viewMode === 'chart'"
                class="flex min-h-0 min-w-0 flex-col gap-2"
            >
                <div
                    v-if="hasProgressionChartData"
                    class="h-64 min-h-64 w-full min-w-0 lg:h-80 lg:min-h-80"
                >
                    <Chart
                        type="line"
                        :data="progressionChartData"
                        :options="progressionChartOptions"
                        class="h-full w-full"
                    />
                </div>
                <p v-else class="m-0 text-sm text-textSecondary">
                    No chart data in the {{ rangeLabel }}.
                </p>
            </div>
            <div
                v-else
                class="flex max-h-80 flex-col gap-3 overflow-y-auto text-sm"
            >
                <div
                    v-for="day in progressionDayGroups"
                    :key="day.dayKey"
                    class="flex flex-col gap-1"
                >
                    <p class="m-0 font-medium text-textPrimary">
                        {{ day.dateLabel }}<span
                            v-if="day.dayKey === lastTimeDayKey"
                            class="font-normal text-textSecondary"
                        >
                            (last time)</span
                        >
                    </p>
                    <ul class="m-0 list-none space-y-0.5 pl-0">
                        <li
                            v-for="(set, index) in day.sets"
                            :key="`${day.dayKey}-${index}`"
                            class="text-textSecondary"
                        >
                            - Set {{ index + 1 }}: {{ set.weight }}x{{
                                set.reps
                            }}
                        </li>
                    </ul>
                </div>
            </div>
        </template>
    </div>
</template>
