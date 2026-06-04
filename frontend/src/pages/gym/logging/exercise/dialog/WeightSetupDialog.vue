<script setup lang="ts">
import { computed, ref } from "vue";
import {
    BAR_WEIGHT_OPTIONS,
    computeBarbellTotalLbs,
    formatWeightSetup,
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
    onResolve?: (value: WeightSetupDialogResult) => void;
}>();

const initial = parseWeightSetup(props.currentSetup);
const barLbs = ref(initial.barLbs);
const counts = ref<PlateCountsPerSide>({ ...initial.platesPerSide });
const parseFailed = ref(!initial.parsed && props.currentSetup.trim().length > 0);

const totalLbs = computed(() =>
    computeBarbellTotalLbs(barLbs.value, counts.value),
);

const formattedSetup = computed(() =>
    formatWeightSetup(barLbs.value, counts.value),
);

const mismatchDelta = computed(() =>
    weightSetupMismatchLbs(totalLbs.value, props.targetWeightLbs),
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
}

function save(syncWeight: boolean) {
    props.onResolve?.({
        weightSetup: formattedSetup.value,
        totalLbs: totalLbs.value,
        syncWeight,
    });
}
</script>

<template>
    <div class="flex flex-col gap-4">
        <p
            v-if="parseFailed"
            class="m-0 rounded-md border border-amber-900/60 bg-amber-950/40 px-3 py-2 text-sm text-amber-200"
        >
            Could not read the current setup. Plate counts start empty; saving
            replaces the old text.
        </p>
        <div class="flex flex-col gap-2">
            <span class="text-sm font-medium text-[rgb(150,150,150)]">Bar</span>
            <div class="flex flex-row gap-2">
                <button
                    v-for="bar in BAR_WEIGHT_OPTIONS"
                    :key="bar"
                    type="button"
                    class="flex-1 rounded-md border px-3 py-2 text-sm transition-colors"
                    :class="
                        barLbs === bar
                            ? 'border-[rgb(100,100,100)] bg-[rgb(40,40,40)]'
                            : 'border-[rgb(56,56,56)] bg-[rgb(27,27,27)] hover:bg-[rgb(35,35,35)]'
                    "
                    @click="barLbs = bar"
                >
                    {{ bar }} lb
                </button>
            </div>
        </div>
        <div class="flex flex-col gap-2">
            <span class="text-sm font-medium text-[rgb(150,150,150)]"
                >Plates per side</span
            >
            <ul class="m-0 flex list-none flex-col gap-1.5 p-0">
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
            <div class="flex flex-wrap items-baseline justify-between gap-2">
                <span class="text-sm text-[rgb(150,150,150)]">Total on bar</span>
                <span class="text-lg font-medium tabular-nums"
                    >{{ totalLbs }} lbs</span
                >
            </div>
            <p
                v-if="formattedSetup"
                class="m-0 text-sm text-[rgb(180,180,180)]"
            >
                {{ formattedSetup }}
            </p>
            <p v-else-if="!hasPlates" class="m-0 text-sm text-[rgb(120,120,120)]">
                Bar only ({{ barLbs }} lbs)
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
        <div class="flex flex-col gap-2 sm:flex-row">
            <button
                type="button"
                class="flex-1 rounded-md border border-[rgb(56,56,56)] bg-[rgb(35,35,35)] px-4 py-2 text-sm hover:bg-[rgb(50,50,50)]"
                @click="save(false)"
            >
                Save setup
            </button>
            <button
                v-if="totalLbs > 0"
                type="button"
                class="flex-1 rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-500"
                @click="save(true)"
            >
                Save & set weight to {{ totalLbs }} lbs
            </button>
        </div>
    </div>
</template>
