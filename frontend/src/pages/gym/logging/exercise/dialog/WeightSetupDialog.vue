<script setup lang="ts">
import { computed, ref } from "vue";
import { Button, Field, Input, NumberField } from "~/shared/ui";
import type { ExerciseLoadType } from "~/types/workout";
import {
    computeExerciseLoadLbs,
    formatExerciseWeightSetup,
    parseWeightSetup,
    perSideLoadLbs,
    STANDARD_PLATE_LBS,
    weightSetupMismatchLbs,
    type PlateCountsPerSide,
    type WeightSetupDialogResult,
} from "../domain/weightSetup";

const props = defineProps<{
    currentSetup: string;
    targetWeightLbs: number;
    loadType: ExerciseLoadType;
    onResolve?: (value: WeightSetupDialogResult) => void;
}>();

const initial = parseWeightSetup(props.currentSetup);
const setupText = ref(
    initial.parsed
        ? formatExerciseWeightSetup(props.loadType, initial.platesPerSide)
        : props.currentSetup,
);
const counts = ref<PlateCountsPerSide>({ ...initial.platesPerSide });
const parseFailed = ref(!initial.parsed && props.currentSetup.trim().length > 0);

function refreshSetupText() {
    setupText.value = formatExerciseWeightSetup(props.loadType, counts.value);
    parseFailed.value = false;
}

const totalLbs = computed(() => {
    if (parseFailed.value) return 0;
    return computeExerciseLoadLbs(props.loadType, counts.value);
});

const mismatchDelta = computed(() =>
    !parseFailed.value
        ? weightSetupMismatchLbs(totalLbs.value, props.targetWeightLbs)
        : null,
);

const hasPlates = computed(() => perSideLoadLbs(counts.value) > 0);

function plateCount(plate: number): number {
    return counts.value[plate] ?? 0;
}

function setPlate(plate: number, value: number | null | undefined) {
    const next = Math.max(0, Math.round(value ?? 0));
    if (next === 0) {
        const { [plate]: _, ...rest } = counts.value;
        counts.value = rest;
    } else {
        counts.value = { ...counts.value, [plate]: next };
    }
    refreshSetupText();
}

function syncFromTextInput() {
    const parsed = parseWeightSetup(setupText.value);
    if (parsed.parsed) {
        counts.value = { ...parsed.platesPerSide };
        parseFailed.value = false;
        return;
    }
    parseFailed.value = setupText.value.trim().length > 0;
    if (!setupText.value.trim()) {
        counts.value = {};
    }
}

function save(syncWeight: boolean) {
    const parsed = parseWeightSetup(setupText.value);
    const weightSetup = parsed.parsed
        ? formatExerciseWeightSetup(props.loadType, parsed.platesPerSide)
        : setupText.value.trim();
    props.onResolve?.({
        weightSetup,
        totalLbs: totalLbs.value,
        syncWeight,
    });
}
</script>

<template>
    <div class="flex w-full min-w-0 flex-col gap-4">
        <Field label="Setup text">
            <Input
                v-model="setupText"
                placeholder="e.g. belt, 2×45 + 10, or notes"
                @update:model-value="syncFromTextInput"
            />
        </Field>
        <p v-if="parseFailed" class="m-0 -mt-2 text-xs text-warn">
            Free text — use the plate picker below to build plate notation.
        </p>
        <section class="flex flex-col gap-2">
            <h3 class="m-0 text-xs font-medium text-fg-muted">
                {{
                    props.loadType === "plate_loaded_with_bar"
                        ? "Plates per side (45 lb bar included)"
                        : props.loadType === "plate_loaded_total"
                          ? "Total plates (machine load)"
                          : "Plates per side (bar not counted)"
                }}
            </h3>
            <ul class="m-0 flex w-full min-w-0 list-none flex-col gap-2 p-0">
                <li v-for="plate in STANDARD_PLATE_LBS" :key="plate" class="flex items-center justify-between gap-3">
                    <span class="min-w-14 text-base text-fg tabular-nums lg:text-sm">{{ plate }} lb</span>
                    <div class="w-44">
                        <NumberField
                            :model-value="plateCount(plate)"
                            :label="`${plate} lb plates`"
                            :min="0"
                            :step="1"
                            :decimals="0"
                            @update:model-value="setPlate(plate, $event)"
                        />
                    </div>
                </li>
            </ul>
        </section>
        <div class="flex flex-col gap-1 rounded-card border border-line bg-surface-2 px-3 py-2.5">
            <div v-if="!parseFailed" class="flex flex-wrap items-baseline justify-between gap-2">
                <span class="text-sm text-fg-muted">Total load</span>
                <span class="text-lg font-medium text-fg tabular-nums">{{ totalLbs }} lbs</span>
            </div>
            <p v-else-if="setupText.trim()" class="m-0 text-sm text-fg-muted">{{ setupText.trim() }}</p>
            <p v-else-if="!hasPlates && props.loadType === 'plate_loaded_with_bar'" class="m-0 text-sm text-fg-subtle">
                Bar only (45 lbs)
            </p>
            <p v-else-if="!hasPlates" class="m-0 text-sm text-fg-subtle">No plates</p>
            <p v-if="targetWeightLbs > 0 && mismatchDelta != null" class="m-0 text-sm text-warn">
                Logged weight is {{ targetWeightLbs }} lbs ({{ mismatchDelta > 0 ? "+" : "" }}{{ mismatchDelta }} lbs vs
                plates)
            </p>
        </div>
        <div class="flex flex-col gap-2 lg:flex-row lg:justify-end">
            <Button variant="secondary" @click="save(false)">Save setup</Button>
            <Button variant="primary" :disabled="parseFailed || totalLbs <= 0" @click="save(true)">
                Save & set weight to {{ totalLbs }} lbs
            </Button>
        </div>
    </div>
</template>
