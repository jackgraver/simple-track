<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import { Button } from "~/shared/ui";
import LoggingHeader from "./LoggingHeader.vue";
import LoggedSetsList from "./LoggedSetsList.vue";

defineProps<{
    exerciseName: string;
    setNumber: number;
    currentWeight: number;
    session: ExerciseLoggingSessionViewModel;
    finishing?: boolean;
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
        <LoggingHeader :title="exerciseName" :subtitle="`Set ${setNumber}`" @back="back">
            <template #right>
                <span class="font-medium tabular-nums text-fg">{{ isActive ? displayText : "0:00" }}</span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col items-center gap-3 border-b border-line pb-6 text-center">
            <p class="m-0 text-sm font-semibold tracking-wide text-fg-muted uppercase">Rest</p>
            <p class="m-0 text-display font-semibold text-fg tabular-nums">{{ isActive ? displayText : "0:00" }}</p>
            <p class="m-0 text-fg-muted">Time since your last set.</p>
        </div>
        <Button variant="secondary" block @click="emit('edit-setup')">
            <span class="flex w-full items-baseline justify-between gap-2">
                <span class="flex items-baseline gap-2">
                    <span class="text-xs font-semibold tracking-wide text-fg-muted uppercase">Weight</span>
                    <span class="tabular-nums">{{ currentWeight }} lbs</span>
                </span>
                <span class="text-sm font-normal text-fg-muted">Change</span>
            </span>
        </Button>
        <LoggedSetsList
            :logged-sets="session.loggedSets"
            @retry="session.retrySet"
            @delete="session.deleteSet"
            @edit="(index) => emit('edit', index)"
        />
        <div class="flex flex-col gap-3">
            <Button variant="primary" size="lg" block @click="nextSet">Next set</Button>
            <Button variant="secondary" block :loading="finishing" @click="finish">Finish exercise</Button>
            <Button variant="ghost" block @click="emit('progression')">View progression</Button>
        </div>
    </div>
</template>
