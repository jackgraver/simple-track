<script setup lang="ts">
import LoggingHeader from "./LoggingHeader.vue";

defineProps<{
    exerciseName: string;
    expectedWeight: number;
    previousReps: number | null;
    notes: string;
    cues: string;
    machineSetupNote: string;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "continue"): void;
}>();
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" @back="emit('back')" />
        <div class="flex flex-col gap-5">
            <section class="flex flex-col gap-1 border-b border-zinc-700 pb-5">
                <p class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500">
                    Set the machine to
                </p>
                <p class="m-0 mt-2 text-4xl font-semibold tabular-nums">
                    {{ expectedWeight }} lbs
                </p>
            </section>
            <section class="flex flex-col gap-1 border-b border-zinc-700 pb-5">
                <p class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500">
                    Previous session
                </p>
                <p class="m-0 mt-2 text-lg text-zinc-200">
                    {{ previousReps != null ? `${previousReps} reps` : "No previous reps recorded" }}
                </p>
            </section>
            <section class="flex flex-col gap-4">
                <div>
                    <p class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500">
                        Machine setup
                    </p>
                    <p class="m-0 mt-2 text-zinc-300">{{ machineSetupNote }}</p>
                </div>
                <div v-if="notes">
                    <p class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500">
                        Notes
                    </p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-zinc-300">{{ notes }}</p>
                </div>
                <div v-if="cues">
                    <p class="m-0 text-xs font-semibold uppercase tracking-wide text-zinc-500">
                        Cues
                    </p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-zinc-300">{{ cues }}</p>
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
