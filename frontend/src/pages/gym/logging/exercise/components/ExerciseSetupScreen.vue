<script setup lang="ts">
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { Badge, Button } from "~/shared/ui";
import LoggingHeader from "./LoggingHeader.vue";
import NumberSlider from "./NumberSlider.vue";

defineProps<{
    exerciseName: string;
    setNumber: number;
    weight: number;
    weightStep: number;
    previousSets: { weight: number; reps: number }[];
    notes: string;
    cues: string;
    weightIncrease: number | null;
    tracksWeightSetup: boolean;
    weightSetupSummary: string;
    continueLabel: string;
}>();

const emit = defineEmits<{
    (event: "back"): void;
    (event: "continue"): void;
    (event: "undo-weight-increase"): void;
    (event: "edit-weight-setup"): void;
    (event: "update:weight", value: number): void;
}>();

const { displayText, isActive } = useGlobalRestTimer();
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <LoggingHeader :title="exerciseName" :subtitle="`Set ${setNumber}`" @back="emit('back')">
            <template #right>
                <span class="font-medium tabular-nums text-fg">{{ isActive ? displayText : "0:00" }}</span>
            </template>
        </LoggingHeader>
        <div class="flex flex-col gap-5">
            <section class="flex flex-col gap-1 border-b border-line pb-5">
                <div class="flex items-center justify-between gap-2">
                    <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Set the machine to</p>
                    <Button
                        v-if="weightIncrease != null"
                        variant="ghost"
                        size="sm"
                        aria-label="Undo automatic weight increase"
                        title="Rep target met last session. Tap to undo."
                        @click="emit('undo-weight-increase')"
                    >
                        <Badge tone="success">+{{ weightIncrease }}</Badge>
                        Undo
                    </Button>
                </div>
                <NumberSlider
                    label=""
                    :model-value="weight"
                    :step="weightStep"
                    :max="500"
                    @update:model-value="emit('update:weight', $event)"
                />
                <Button v-if="tracksWeightSetup" variant="secondary" block class="mt-2" @click="emit('edit-weight-setup')">
                    <span class="flex w-full min-w-0 items-baseline gap-2">
                        <span class="text-xs font-semibold tracking-wide text-fg-muted uppercase">Weight setup</span>
                        <span class="truncate font-normal">{{ weightSetupSummary }}</span>
                    </span>
                </Button>
            </section>
            <section class="flex flex-col gap-1 border-b border-line pb-5">
                <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Previous session</p>
                <p class="m-0 mt-2 text-lg text-fg tabular-nums">
                    {{
                        previousSets.length
                            ? previousSets.map((set) => `${set.weight} × ${set.reps}`).join(", ")
                            : "No previous sets recorded"
                    }}
                </p>
            </section>
            <section v-if="notes || cues" class="flex flex-col gap-4">
                <div v-if="notes">
                    <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Notes</p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-fg-muted">{{ notes }}</p>
                </div>
                <div v-if="cues">
                    <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Cues</p>
                    <p class="m-0 mt-2 whitespace-pre-wrap text-fg-muted">{{ cues }}</p>
                </div>
            </section>
        </div>
        <Button variant="primary" size="lg" block @click="emit('continue')">{{ continueLabel }}</Button>
    </div>
</template>
