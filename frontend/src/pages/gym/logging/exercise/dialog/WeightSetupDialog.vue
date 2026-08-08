<script setup lang="ts">
import { computed, ref } from "vue";
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

function changePlate(plate: number, delta: number) {
    const next = Math.max(0, plateCount(plate) + delta);
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
        <div class="flex flex-col gap-2">
            <label
                for="weight-setup-text"
                class="text-sm font-medium text-[rgb(150,150,150)]"
                >Setup text</label
            >
            <input
                id="weight-setup-text"
                v-model="setupText"
                type="text"
                placeholder="e.g. belt, 2×45 + 10, or notes"
                class="w-full min-w-0 rounded-md border border-[rgb(56,56,56)] bg-[rgb(27,27,27)] px-3 py-2 text-sm outline-none focus:border-[rgb(100,100,100)]"
                @input="syncFromTextInput"
            />
            <p
                v-if="parseFailed"
                class="m-0 text-xs text-amber-300"
            >
                Free text — use the plate picker below to build plate notation.
            </p>
        </div>
        <div class="flex flex-col gap-2">
            <span class="text-sm font-medium text-[rgb(150,150,150)]">{{
                props.loadType === "plate_loaded_with_bar"
                    ? "Plates per side (45 lb bar included)"
                    : props.loadType === "plate_loaded_total"
                      ? "Total plates (machine load)"
                    : "Plates per side (bar not counted)"
            }}</span>
            <ul class="m-0 flex w-full min-w-0 list-none flex-col gap-1.5 p-0">
                <li
                    v-for="plate in STANDARD_PLATE_LBS"
                    :key="plate"
                    class="flex items-center justify-between gap-2 rounded-md border border-[rgb(56,56,56)] bg-[rgb(27,27,27)] px-2 py-1.5"
                >
                    <span class="min-w-12 text-sm">{{ plate }} lb</span>
                    <div class="flex items-center gap-2">
                        <button
                            type="button"
                            class="flex h-8 w-8 items-center justify-center rounded-md border border-[rgb(56,56,56)] bg-[rgb(35,35,35)] text-lg leading-none hover:bg-[rgb(50,50,50)] disabled:opacity-40"
                            :disabled="plateCount(plate) === 0"
                            aria-label="Remove plate"
                            @click="changePlate(plate, -1)"
                        >
                            −
                        </button>
                        <span
                            class="w-6 text-center text-sm tabular-nums"
                            >{{ plateCount(plate) }}</span
                        >
                        <button
                            type="button"
                            class="flex h-8 w-8 items-center justify-center rounded-md border border-[rgb(56,56,56)] bg-[rgb(35,35,35)] text-lg leading-none hover:bg-[rgb(50,50,50)]"
                            aria-label="Add plate"
                            @click="changePlate(plate, 1)"
                        >
                            +
                        </button>
                    </div>
                </li>
            </ul>
        </div>
        <div
            class="flex flex-col gap-1 rounded-md border border-[rgb(56,56,56)] bg-[rgb(24,24,24)] px-3 py-2.5"
        >
            <div
                v-if="!parseFailed"
                class="flex flex-wrap items-baseline justify-between gap-2"
            >
                <span class="text-sm text-[rgb(150,150,150)]">Total load</span>
                <span class="text-lg font-medium tabular-nums"
                    >{{ totalLbs }} lbs</span
                >
            </div>
            <p
                v-else-if="setupText.trim()"
                class="m-0 text-sm text-[rgb(180,180,180)]"
            >
                {{ setupText.trim() }}
            </p>
            <p
                v-else-if="!hasPlates && props.loadType === 'plate_loaded_with_bar'"
                class="m-0 text-sm text-[rgb(120,120,120)]"
            >
                Bar only (45 lbs)
            </p>
            <p
                v-else-if="!hasPlates"
                class="m-0 text-sm text-[rgb(120,120,120)]"
            >
                No plates
            </p>
            <p
                v-if="targetWeightLbs > 0 && mismatchDelta != null"
                class="m-0 text-sm text-amber-300"
            >
                Logged weight is {{ targetWeightLbs }} lbs ({{
                    mismatchDelta > 0 ? "+" : ""
                }}{{ mismatchDelta }} lbs vs plates)
            </p>
        </div>
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <button
                type="button"
                class="flex-1 rounded-md border border-[rgb(56,56,56)] bg-[rgb(35,35,35)] px-4 py-2 text-sm hover:bg-[rgb(50,50,50)]"
                @click="save(false)"
            >
                Save setup
            </button>
            <button
                type="button"
                class="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-500 disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="parseFailed || totalLbs <= 0"
                @click="save(true)"
            >
                Save & set weight to {{ totalLbs }} lbs
            </button>
        </div>
    </div>
</template>
