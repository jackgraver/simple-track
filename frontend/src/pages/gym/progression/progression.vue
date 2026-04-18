<script setup lang="ts">
import type { Exercise } from "~/types/workout";
import Chart from "primevue/chart";
import Select from "primevue/select";
import { computed, ref, watch } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";
import {
    normalizeExercisesListPayload,
    useWorkoutExercisesAllQuery,
} from "~/api/workout/queries";
import {
    useProgressionCharts,
    type ProgressionEntry,
    type ProgressionRange,
} from "~/pages/gym/progression/useProgressionCharts";

const exerciseId = ref<number | null>(null);
const range = ref<ProgressionRange>("6m");

const rangeOptions: { id: ProgressionRange; label: string }[] = [
    { id: "3m", label: "3 mo" },
    { id: "6m", label: "6 mo" },
    { id: "1y", label: "1 yr" },
    { id: "all", label: "All" },
];

const { data: exercisesPayload, isPending: exercisesLoading } =
    useWorkoutExercisesAllQuery();

const exercises = computed(() =>
    normalizeExercisesListPayload(exercisesPayload.value),
);

const exerciseOptions = computed(() =>
    [...exercises.value]
        .sort((a, b) =>
            a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
        )
        .map((e: Exercise) => ({ label: e.name, value: e.ID })),
);

const selectedExercise = computed(() => {
    const id = exerciseId.value;
    if (id == null) return null;
    return exercises.value.find((e) => e.ID === id) ?? null;
});

watch(
    exercises,
    (list) => {
        if (list.length === 0) {
            exerciseId.value = null;
            return;
        }
        const cur = exerciseId.value;
        if (cur == null || !list.some((e) => e.ID === cur)) {
            exerciseId.value = list[0]!.ID;
        }
    },
    { immediate: true },
);

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
        if (id == null) throw new Error("No exercise");
        const res = await apiClient.get<{ progression: ProgressionEntry[] }>(
            `/workout/exercises/progression/${id}`,
        );
        return res.data;
    },
    enabled: computed(() => exerciseId.value != null),
});

const progression = computed(() => {
    return progressionPayload.value?.progression ?? [];
});

const error = computed(() => {
    return fetchError.value?.message || null;
});

const {
    progressionChartData,
    progressionChartOptions,
    hasProgressionChartData,
} = useProgressionCharts(progression, range);
</script>

<template>
    <div class="flex flex-col gap-6 p-4">
        <h1 class="m-0 text-2xl font-semibold text-textPrimary">
            Exercise Progression
        </h1>
        <div class="flex min-w-0 flex-col gap-2">
            <label
                class="text-sm font-medium text-textPrimary"
                for="progression-exercise"
                >Exercise</label
            >
            <Select
                id="progression-exercise"
                v-model="exerciseId"
                :options="exerciseOptions"
                option-label="label"
                option-value="value"
                placeholder="Choose an exercise"
                class="w-full max-w-2xl"
                :loading="exercisesLoading"
            />
        </div>
        <div v-if="selectedExercise" class="flex min-w-0 flex-col gap-4">
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
            <template v-else-if="progression.length === 0">
                <p class="m-0 py-4 text-center text-sm text-textSecondary">
                    No progression data available for this exercise.
                </p>
            </template>
            <template v-else>
                <div class="flex min-w-0 flex-col gap-2">
                    <p class="m-0 text-sm text-textSecondary">Range</p>
                    <div class="flex flex-wrap gap-1">
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
                </div>
                <div class="flex min-h-0 min-w-0 flex-col gap-2">
                    <h2 class="m-0 text-base font-semibold text-textPrimary">
                        Top set and volume
                    </h2>
                    <div
                        v-if="hasProgressionChartData"
                        class="h-80 min-h-80 w-full min-w-0 lg:h-112 lg:min-h-112"
                    >
                        <Chart
                            type="line"
                            :data="progressionChartData"
                            :options="progressionChartOptions"
                            class="h-full w-full"
                        />
                    </div>
                    <p v-else class="m-0 text-sm text-textSecondary">
                        No data in this range.
                    </p>
                </div>
            </template>
        </div>
    </div>
</template>
