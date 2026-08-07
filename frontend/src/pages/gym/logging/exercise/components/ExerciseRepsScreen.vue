<script setup lang="ts">
import { computed } from "vue";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import LoggingHeader from "./LoggingHeader.vue";
import NumberSlider from "./NumberSlider.vue";

const props = defineProps<{
    exerciseName: string;
    session: ExerciseLoggingSessionViewModel;
    previousRep: number | null;
    repRollover: number | undefined;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "logged"): void;
}>();

const sliderMarkers = computed(() => {
    const markers: { value: number; label: string }[] = [];
    if (props.previousRep != null && Number.isFinite(props.previousRep)) {
        markers.push({ value: props.previousRep, label: "Last time" });
    }
    if (props.repRollover != null && Number.isFinite(props.repRollover)) {
        markers.push({ value: props.repRollover, label: "Rollover" });
    }
    return markers;
});

const logSet = async () => {
    if (await props.session.addNextSet()) emit("logged");
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" @back="emit('back')">
            <template #right>
                <span class="text-sm text-zinc-500"
                    >Set {{ session.currentSetNumber }}</span
                >
            </template>
        </LoggingHeader>
        <div class="flex flex-col gap-5">
            <NumberSlider
                label="Reps"
                :model-value="session.currentReps"
                :markers="sliderMarkers"
                integer-only
                :max="50"
                @update:model-value="session.commitRepsFromInput"
            />
            <div class="flex flex-col gap-3">
                <label
                    class="text-[0.9rem] font-medium text-[rgb(150,150,150)]"
                    for="exercise-notes"
                >
                    Notes
                </label>
                <textarea
                    id="exercise-notes"
                    class="min-h-24 resize-y rounded border border-zinc-700 bg-zinc-900 px-3 py-3 text-inherit outline-none transition-colors placeholder:text-zinc-600 focus:border-zinc-500"
                    :value="session.notes"
                    placeholder="Add any notes about this exercise..."
                    rows="3"
                    @input="
                        session.updateNotes(
                            ($event.target as HTMLTextAreaElement).value,
                        )
                    "
                />
            </div>
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
