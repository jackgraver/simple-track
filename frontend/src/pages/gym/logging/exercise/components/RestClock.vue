<script setup lang="ts">
import { computed } from "vue";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { REST_BAR_MAX_MS, REST_READY_MS, restBarPercent, restStage, type RestStage } from "../domain/liveLogging";

// Counts up from the last logged set (any exercise). Colour says "not yet" until 2:00, then "good to go".
const { elapsedMs, isActive, displayText } = useGlobalRestTimer();

const stage = computed(() => restStage(elapsedMs.value, isActive.value));
const percent = computed(() => restBarPercent(elapsedMs.value));
const readyAt = `${(REST_READY_MS / REST_BAR_MAX_MS) * 100}%`;

const textTone: Record<RestStage, string> = {
    idle: "text-fg-subtle",
    wait: "text-rest-wait",
    near: "text-rest-near",
    ready: "text-success",
};
const fillTone: Record<RestStage, string> = {
    idle: "bg-line",
    wait: "bg-rest-wait",
    near: "bg-rest-near",
    ready: "bg-success",
};
// Never colour alone: the word changes with the stage.
const stateLabel: Record<RestStage, string> = {
    idle: "Starts when you log a set",
    wait: "",
    near: "Almost 2:00",
    ready: "Past 2:00",
};
</script>

<template>
    <section class="flex flex-col items-center gap-1 pt-1 text-center" aria-label="Time since last set">
        <!-- Stage word above the time. Fixed height, so it coming and going never moves the page. -->
        <p class="m-0 min-h-5 text-sm font-medium transition-colors duration-500" :class="textTone[stage]">
            {{ stateLabel[stage] }}
        </p>
        <p
            class="m-0 text-7xl leading-none short:text-6xl font-semibold tracking-tight tabular-nums transition-colors duration-500"
            :class="textTone[stage]"
            role="timer"
            aria-live="off"
        >
            {{ isActive ? displayText : "0:00" }}
        </p>
        <div class="relative mt-3 short:mt-2 h-1.5 w-full rounded-full bg-surface-3" aria-hidden="true">
            <div
                class="absolute inset-y-0 left-0 rounded-full transition-[width,background-color] duration-500"
                :class="fillTone[stage]"
                :style="{ width: `${percent}%` }"
            />
            <div class="absolute -inset-y-1 w-0.5 -translate-x-1/2 rounded-full bg-fg" :style="{ left: readyAt }" />
        </div>
        <div class="relative h-4 w-full text-2xs text-fg-subtle tabular-nums" aria-hidden="true">
            <span class="absolute left-0">0:00</span>
            <span class="absolute -translate-x-1/2" :style="{ left: readyAt }">2:00</span>
            <span class="absolute right-0">3:00</span>
        </div>
    </section>
</template>
