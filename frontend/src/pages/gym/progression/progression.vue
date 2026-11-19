<script setup lang="ts">
import type { Exercise } from "~/types/workout";
import Dropdown from "~/shared/input/Dropdown.vue";
import type { DropdownOption } from "~/shared/input/Dropdown.vue";
import ExerciseProgressionChartPanel from "~/pages/gym/progression/ExerciseProgressionChartPanel.vue";
import { computed, ref, watch } from "vue";
import { Dumbbell } from "lucide-vue-next";
import { EmptyState, ErrorState, PageHeader } from "~/shared/ui";
import {
    normalizeExercisesListPayload,
    useWorkoutExercisesAllQuery,
} from "~/api/workout/queries";

const exerciseId = ref<number | null>(null);

const {
    data: exercisesPayload,
    isPending: exercisesLoading,
    isError: exercisesError,
    error: exercisesErr,
    refetch,
} = useWorkoutExercisesAllQuery();

const exercises = computed(() =>
    normalizeExercisesListPayload(exercisesPayload.value),
);

const exerciseOptions = computed<DropdownOption[]>(() =>
    [...exercises.value]
        .sort((a, b) =>
            a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
        )
        .map((e: Exercise) => ({ label: e.name, value: e.ID })),
);

const onExerciseSelect = (option: DropdownOption) => {
    exerciseId.value = option.value as number;
};

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
</script>

<template>
    <div class="flex flex-col gap-6 py-2 lg:py-4">
        <PageHeader title="Exercise progression" />
        <ErrorState
            v-if="exercisesError"
            :error="exercisesErr"
            title="Couldn't load your exercises"
            @retry="refetch()"
        />
        <EmptyState
            v-else-if="!exercisesLoading && exercises.length === 0"
            :icon="Dumbbell"
            title="No exercises yet"
            description="Create exercises in a workout plan to track progression."
        />
        <template v-else>
            <div class="flex min-w-0 max-w-2xl flex-col gap-2">
                <span class="text-xs font-medium text-fg-muted">Exercise</span>
                <p v-if="selectedExercise" class="m-0 text-base font-medium text-fg">{{ selectedExercise.name }}</p>
                <Dropdown
                    :options="exerciseOptions"
                    placeholder="Search exercises…"
                    :loading="exercisesLoading && exerciseOptions.length === 0"
                    :on-select="onExerciseSelect"
                />
            </div>
            <ExerciseProgressionChartPanel v-if="exerciseId != null" :exercise-id="exerciseId" />
        </template>
    </div>
</template>
