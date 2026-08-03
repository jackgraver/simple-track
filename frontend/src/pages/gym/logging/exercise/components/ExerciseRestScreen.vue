<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import LoggingHeader from "./LoggingHeader.vue";

defineProps<{
    exerciseName: string;
    setNumber: number;
    session: ExerciseLoggingSessionViewModel;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "next"): void;
    (event: "finish"): void;
}>();

const { displayText, isActive, clear } = useGlobalRestTimer();

const nextSet = () => {
    clear();
    emit("next");
};

const finish = () => {
    clear();
    emit("finish");
};

const back = () => {
    clear();
    emit("back");
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" @back="back">
            <template #right>
                <span class="text-sm text-zinc-500">Set {{ setNumber }}</span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col items-center gap-3 border-b border-zinc-700 pb-6 text-center">
            <p class="m-0 text-sm font-semibold uppercase tracking-wide text-zinc-500">Rest</p>
            <p class="m-0 text-6xl font-semibold tabular-nums">{{ isActive ? displayText : "Ready" }}</p>
            <p class="m-0 text-zinc-400">Take your time, then continue when ready.</p>
        </div>
        <div class="flex flex-col gap-3">
            <label class="text-[0.9rem] font-medium text-[rgb(150,150,150)]" for="exercise-notes">
                Notes
            </label>
            <textarea
                id="exercise-notes"
                class="min-h-24 resize-y rounded border border-zinc-700 bg-zinc-900 px-3 py-3 text-inherit outline-none transition-colors placeholder:text-zinc-600 focus:border-zinc-500"
                :value="session.notes"
                placeholder="Add any notes about this exercise..."
                rows="3"
                @input="session.updateNotes(($event.target as HTMLTextAreaElement).value)"
            />
        </div>
        <div class="flex flex-col gap-3 sm:flex-row">
            <button
                class="flex-1 rounded-md bg-secondBg px-4 py-3 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg"
                type="button"
                @click="nextSet"
            >
                Next set
            </button>
            <button
                class="flex-1 rounded-md bg-green-700 px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-green-600"
                type="button"
                @click="finish"
            >
                Finish exercise
            </button>
        </div>
    </div>
</template>
