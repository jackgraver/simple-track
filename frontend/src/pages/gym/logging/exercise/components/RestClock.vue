<script setup lang="ts">
import { computed } from "vue";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { restStage, type RestStage } from "../domain/liveLogging";

// Counts up from the last logged set (any exercise). Red until 1:00, orange until 1:30, yellow until 2:00, green after.
// Just the time, no bar or stage words (owner, 2026-11-23).
const { elapsedMs, isActive, displayText } = useGlobalRestTimer();

const stage = computed(() => restStage(elapsedMs.value, isActive.value));

const textTone: Record<RestStage, string> = {
    idle: "text-fg-subtle",
    wait: "text-danger",
    near: "text-rest-wait",
    soon: "text-rest-near",
    ready: "text-success",
};
</script>

<template>
    <section class="flex flex-col items-center gap-1 pt-1 text-center" aria-label="Time since last set">
        <p
            class="m-0 text-7xl leading-none font-semibold tracking-tight tabular-nums transition-colors duration-500"
            :class="textTone[stage]"
            role="timer"
            aria-live="off"
        >
            {{ isActive ? displayText : "0:00" }}
        </p>
    </section>
</template>
