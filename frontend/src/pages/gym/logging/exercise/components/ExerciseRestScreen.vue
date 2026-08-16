<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import LoggingHeader from "./LoggingHeader.vue";
import LoggedSetsList from "./LoggedSetsList.vue";

defineProps<{
    exerciseName: string;
    setNumber: number;
    currentWeight: number;
    session: ExerciseLoggingSessionViewModel;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "next"): void;
    (event: "finish"): void;
    (event: "progression"): void;
    (event: "edit", index: number): void;
    (event: "edit-setup"): void;
}>();

const { displayText, isActive } = useGlobalRestTimer();

const nextSet = () => {
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
        <LoggingHeader
            :title="exerciseName"
            :subtitle="`Set ${setNumber}`"
            @back="back"
        >
            <template #right>
                <span class="font-medium tabular-nums text-zinc-200">
                    {{ isActive ? displayText : "0:00" }}
                </span>
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
        <button
            type="button"
            class="flex w-full items-center justify-between gap-3 rounded-md border border-zinc-700 bg-zinc-900 px-3 py-3 text-left transition-colors hover:border-zinc-500 hover:bg-zinc-800"
            @click="emit('edit-setup')"
        >
            <span>
                <span
                    class="mr-2 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                    >Weight</span
                >
                <span class="tabular-nums text-zinc-200"
                    >{{ currentWeight }} lbs</span
                >
            </span>
            <span class="text-sm text-zinc-400">Change</span>
        </button>
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
        <button
            class="w-full rounded-md border border-zinc-700 px-4 py-3 text-sm font-semibold text-textSecondary transition-colors hover:border-zinc-500 hover:bg-firstBg hover:text-textPrimary"
            type="button"
            @click="emit('progression')"
        >
            View progression
        </button>
    </div>
</template>
