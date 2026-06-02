<script setup lang="ts">
import type { Exercise } from "~/types/workout";
import Dropdown from "~/shared/input/Dropdown.vue";
import type { DropdownOption } from "~/shared/input/Dropdown.vue";
import ExerciseProgressionChartPanel from "~/pages/gym/progression/ExerciseProgressionChartPanel.vue";
import { computed, ref, watch } from "vue";
import {
    normalizeExercisesListPayload,
    useWorkoutExercisesAllQuery,
} from "~/api/workout/queries";

const exerciseId = ref<number | null>(null);

const { data: exercisesPayload, isPending: exercisesLoading } =
    useWorkoutExercisesAllQuery();

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
    <div class="flex flex-col gap-6 p-4">
        <h1 class="m-0 text-2xl font-semibold text-textPrimary">
            Exercise Progression
        </h1>
        <div class="flex min-w-0 max-w-2xl flex-col gap-2">
            <span class="text-sm font-medium text-textPrimary">Exercise</span>
            <p
                v-if="selectedExercise"
                class="m-0 text-base font-medium text-textPrimary"
            >
                {{ selectedExercise.name }}
            </p>
            <Dropdown
                :options="exerciseOptions"
                placeholder="Search exercises…"
                :loading="exercisesLoading && exerciseOptions.length === 0"
                :on-select="onExerciseSelect"
            />
        </div>
        <ExerciseProgressionChartPanel
            v-if="exerciseId != null"
            :exercise-id="exerciseId"
        />
    </div>
</template>
