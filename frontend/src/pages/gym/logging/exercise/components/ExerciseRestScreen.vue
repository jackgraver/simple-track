<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import LoggingHeader from "./LoggingHeader.vue";
import LoggedSetsList from "./LoggedSetsList.vue";

defineProps<{
    exerciseName: string;
    setNumber: number;
    session: ExerciseLoggingSessionViewModel;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "next"): void;
    (event: "finish"): void;
    (event: "edit", index: number): void;
}>();

const { displayText, isActive, clear } = useGlobalRestTimer();

const nextSet = () => {
    clear();
    emit("next");
};

const finish = () => {
    emit("finish");
};

const back = () => {
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
        <div
            class="flex flex-col items-center gap-3 border-b border-zinc-700 pb-6 text-center"
        >
            <p
                class="m-0 text-sm font-semibold uppercase tracking-wide text-zinc-500"
            >
                Rest
            </p>
            <p class="m-0 text-6xl font-semibold tabular-nums">
                {{ isActive ? displayText : "0:00" }}
            </p>
            <p class="m-0 text-zinc-400">Time since your last set.</p>
        </div>
        <LoggedSetsList
            :logged-sets="session.loggedSets"
            @retry="session.retrySet"
            @delete="session.deleteSet"
            @edit="(index) => emit('edit', index)"
        />
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
