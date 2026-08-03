<script setup lang="ts">
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import LoggingHeader from "./LoggingHeader.vue";
import NumberSlider from "./NumberSlider.vue";

const props = defineProps<{
    exerciseName: string;
    session: ExerciseLoggingSessionViewModel;
    previousReps: number | null;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "logged"): void;
}>();

const logSet = async () => {
    if (await props.session.addNextSet()) emit("logged");
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" @back="emit('back')">
            <template #right>
                <span class="text-sm text-zinc-500">Set {{ session.currentSetNumber }}</span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col gap-5">
            <NumberSlider
                label="Reps"
                :model-value="session.currentReps"
                :marker-value="previousReps ?? undefined"
                marker-label="Previous"
                integer-only
                :max="50"
                @update:model-value="session.commitRepsFromInput"
            />
        </div>
        <button
            class="w-full rounded-md bg-green-700 px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-green-600"
            type="button"
            @click="logSet"
        >
            Log set
        </button>
    </div>
</template>
