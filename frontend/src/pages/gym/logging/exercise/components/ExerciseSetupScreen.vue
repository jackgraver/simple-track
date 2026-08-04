<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import LoggingHeader from "./LoggingHeader.vue";

defineProps<{
    exerciseName: string;
    expectedWeight: number;
    previousReps: number[];
    notes: string;
    cues: string;
    weightIncrease: number | null;
    tracksWeightSetup: boolean;
    weightSetupSummary: string;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "continue"): void;
    (event: "undo-weight-increase"): void;
    (event: "edit-weight-setup"): void;
}>();

const {
    displayText,
    exerciseName: restExerciseName,
    isActive,
} = useGlobalRestTimer();
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" @back="emit('back')">
            <template v-if="isActive" #right>
                <span class="block text-xs text-zinc-500">
                    Rest since {{ restExerciseName || "last set" }}
                </span>
                <span class="block font-medium tabular-nums text-zinc-200">
                    {{ displayText }}
                </span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col gap-5">
            <section class="flex flex-col gap-1 border-b border-zinc-700 pb-5">
                <p
                    class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                >
                    Set the machine to
                </p>
                <div class="mt-2 flex items-start gap-2">
                    <p class="m-0 text-4xl font-semibold tabular-nums">
                        {{ expectedWeight }} lbs
                    </p>
                    <button
                        v-if="weightIncrease != null"
                        type="button"
                        class="group relative top-0.5 min-w-12 rounded-full bg-green-950 px-2 py-1 text-xs font-semibold tabular-nums text-green-400 transition-colors hover:bg-green-900 focus-visible:bg-green-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500"
                        aria-label="Undo automatic weight increase"
                        title="Rep target met last session. Click to undo."
                        @click="emit('undo-weight-increase')"
                    >
                        <span
                            class="group-hover:hidden group-focus-visible:hidden"
                            >+{{ weightIncrease }}</span
                        >
                        <span
                            class="hidden group-hover:inline group-focus-visible:inline"
                            >Undo</span
                        >
                    </button>
                </div>
                <button
                    v-if="tracksWeightSetup"
                    type="button"
                    class="mt-4 w-full rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-left text-zinc-300 transition-colors hover:border-zinc-500 hover:bg-zinc-800"
                    @click="emit('edit-weight-setup')"
                >
                    <span
                        class="mr-2 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                    >
                        Weight setup
                    </span>
                    {{ weightSetupSummary }}
                </button>
            </section>
            <section class="flex flex-col gap-1 border-b border-zinc-700 pb-5">
                <p
                    class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                >
                    Previous session
                </p>
                <p class="m-0 mt-2 text-lg text-zinc-200">
                    {{
                        previousReps.length
                            ? previousReps.join(", ")
                            : "No previous reps recorded"
                    }}
                </p>
            </section>
            <section class="flex flex-col gap-4">
                <div v-if="notes">
                    <p
                        class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                    >
                        Notes
                    </p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-zinc-300">
                        {{ notes }}
                    </p>
                </div>
                <div v-if="cues">
                    <p
                        class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500"
                    >
                        Cues
                    </p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-zinc-300">
                        {{ cues }}
                    </p>
                </div>
            </section>
        </div>
        <button
            class="w-full rounded-md bg-secondBg px-4 py-3 text-sm font-semibold text-textPrimary transition-colors hover:bg-thirdBg"
            type="button"
            @click="emit('continue')"
        >
            Start logging
        </button>
    </div>
</template>
