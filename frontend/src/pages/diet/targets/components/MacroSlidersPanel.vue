<script setup lang="ts">
import { computed } from 'vue';
import Chart from 'primevue/chart';
import type { MacroPresetId, MacroSlidersApi } from '../useMacroSliders';
import MacroSlider from './MacroSlider.vue';

const props = defineProps<{
    sliders: MacroSlidersApi;
    weightLbs: number;
}>();

const maxP = computed(() =>
    props.sliders.maxProteinG(),
);
const maxC = computed(() =>
    props.sliders.maxCarbsG(),
);
const maxF = computed(() =>
    props.sliders.maxFatG(),
);
const minP = computed(() => props.sliders.minProteinG());
const minC = computed(() => props.sliders.minCarbsG());
const minF = computed(() => props.sliders.minFatG());
const displayMaxP = computed(() => Math.max(1, props.sliders.calorieTarget.value / 4));
const displayMaxC = computed(() => Math.max(1, props.sliders.calorieTarget.value / 4));
const displayMaxF = computed(() => Math.max(1, props.sliders.calorieTarget.value / 9));
const lockedCount = computed(() =>
    [
        props.sliders.lockProtein.value,
        props.sliders.lockCarbs.value,
        props.sliders.lockFat.value,
    ].filter(Boolean).length,
);
const isFixed = (min: number, max: number) => Math.abs(min - max) < 0.05;
const proteinReadonly = computed(() =>
    lockedCount.value === 2 &&
    !props.sliders.lockProtein.value &&
    isFixed(minP.value, maxP.value),
);
const carbsReadonly = computed(() =>
    lockedCount.value === 2 &&
    !props.sliders.lockCarbs.value &&
    isFixed(minC.value, maxC.value),
);
const fatReadonly = computed(() =>
    lockedCount.value === 2 &&
    !props.sliders.lockFat.value &&
    isFixed(minF.value, maxF.value),
);

const proteinFromWeight = computed(() =>
    props.sliders.proteinGramsFromGPerLb(
        props.weightLbs,
        props.sliders.proteinGPerLb.value,
    ),
);

const proteinWeightHint = computed(() => {
    if (props.weightLbs <= 0) return 'Log body weight to use g/lb';
    return `${proteinFromWeight.value}g at ${props.weightLbs} lbs`;
});

const proteinMarkers = computed(() => {
    const w = props.weightLbs;
    if (w <= 0) return [];
    const mult = props.sliders.proteinGPerLb.value;
    const markers = new Set([
        props.sliders.proteinGramsFromGPerLb(w, 0.8),
        props.sliders.proteinGramsFromGPerLb(w, 1.0),
        props.sliders.proteinGramsFromGPerLb(w, mult),
    ]);
    return [...markers].filter((g) => g > 0).sort((a, b) => a - b);
});

const carbsMarkers = computed(() => {
    const cal = props.sliders.calorieTarget.value;
    if (cal <= 0) return [];
    return [
        Math.round((cal * 0.45) / 4),
        Math.round((cal * 0.65) / 4),
    ];
});

const fatMarkers = computed(() => {
    const cal = props.sliders.calorieTarget.value;
    if (cal <= 0) return [];
    return [
        Math.round((cal * 0.20) / 9),
        Math.round((cal * 0.35) / 9),
    ];
});

const presets = computed(() =>
    (['balanced', 'high_protein', 'low_carb'] as MacroPresetId[]).map((id) => ({
        id,
        label: props.sliders.presetLabels[id],
    })),
);

const chartData = computed(() => {
    const p = props.sliders.proteinG.value;
    const c = props.sliders.carbsG.value;
    const f = props.sliders.fatG.value;
    const pc = 4 * p;
    const cc = 4 * c;
    const fc = 9 * f;
    return {
        labels: ['Protein', 'Carbs', 'Fat'],
        datasets: [
            {
                data: [pc, cc, fc],
                backgroundColor: ['#60a5fa', '#ef4444', '#a855f7'],
                borderWidth: 0,
            },
        ],
    };
});

const chartOptions = computed(() => ({
    plugins: {
        legend: { display: false },
    },
    maintainAspectRatio: false,
}));

function onFiberInput(e: Event) {
    props.sliders.fiberG.value =
        Number((e.target as HTMLInputElement).value) || 0;
}

function onProteinMultiplierInput(e: Event) {
    props.sliders.proteinGPerLb.value =
        Number((e.target as HTMLInputElement).value) || 0;
}

function applyProteinFromWeight() {
    props.sliders.applyProteinFromBodyWeight(props.weightLbs);
}
</script>

<template>
    <div class="flex flex-col gap-4">
        <div class="flex flex-wrap items-end gap-3">
            <label class="flex min-w-40 flex-1 flex-col gap-1">
                <span class="text-xs font-medium text-zinc-400">Calories (anchor)</span>
                <input
                    :value="sliders.calorieTarget.value"
                    type="number"
                    min="0"
                    step="1"
                    class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                    @input="
                        sliders.setCalorieTarget(
                            Number(($event.target as HTMLInputElement).value) || 0,
                        )
                    "
                />
            </label>
            <div class="flex gap-1.5 pb-0.5">
                <button
                    v-for="p in presets"
                    :key="p.id"
                    type="button"
                    class="rounded-md border border-zinc-600 px-2 py-1.5 text-xs font-medium text-zinc-300 hover:border-amber-600/60 hover:text-zinc-100"
                    @click="sliders.applyPresetMacros(p.id)"
                >
                    {{ p.label }}
                </button>
            </div>
        </div>
        <p
            v-if="sliders.caloriesAdjustedByLocks.value"
            class="m-0 text-xs text-amber-300/90"
        >
            Calorie target was updated to match locked macros.
        </p>
        <p
            v-if="sliders.lockConflict.value"
            class="m-0 text-xs text-red-400/90"
        >
            Locked macros ({{ Math.round(sliders.lockedMacroCalories.value) }} kcal)
            exceed the calorie target — raise calories or unlock a macro.
        </p>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-[1fr_10rem]">
            <div class="flex flex-col gap-4">
                <div class="flex flex-col gap-2">
                    <MacroSlider
                        label="Protein (g)"
                        :grams="sliders.proteinG.value"
                        :min-grams="minP"
                        :max-grams="maxP"
                        :display-max-grams="displayMaxP"
                        :pct="sliders.proteinCalPct.value"
                        :locked="sliders.lockProtein.value"
                        :readonly="proteinReadonly"
                        track-bg-class="bg-[#60a5fa]/70"
                        :marker-grams="proteinMarkers"
                        @update:grams="sliders.setProteinG"
                        @toggle-lock="sliders.lockProtein.value = !sliders.lockProtein.value"
                    />
                    <div class="flex flex-wrap items-end gap-2 pl-0.5">
                        <label class="flex min-w-28 flex-col gap-1">
                            <span class="text-xs font-medium text-zinc-500">Protein (g/lb)</span>
                            <input
                                :value="sliders.proteinGPerLb.value"
                                type="number"
                                min="0"
                                step="0.05"
                                class="max-w-24 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                                @input="onProteinMultiplierInput"
                            />
                        </label>
                        <button
                            type="button"
                            class="rounded-md border border-zinc-600 px-2 py-1 text-xs font-medium text-zinc-300 hover:border-amber-600/60 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-50"
                            :disabled="weightLbs <= 0"
                            @click="applyProteinFromWeight"
                        >
                            Apply {{ proteinFromWeight }}g
                        </button>
                        <span class="pb-1 text-xs text-zinc-600">{{ proteinWeightHint }}</span>
                    </div>
                </div>
                <MacroSlider
                    label="Carbs (g)"
                    :grams="sliders.carbsG.value"
                    :min-grams="minC"
                    :max-grams="maxC"
                    :display-max-grams="displayMaxC"
                    :pct="sliders.carbsCalPct.value"
                    :locked="sliders.lockCarbs.value"
                    :readonly="carbsReadonly"
                    track-bg-class="bg-[#ef4444]/70"
                    :marker-grams="carbsMarkers"
                    @update:grams="sliders.setCarbsG"
                    @toggle-lock="sliders.lockCarbs.value = !sliders.lockCarbs.value"
                />
                <MacroSlider
                    label="Fat (g)"
                    :grams="sliders.fatG.value"
                    :min-grams="minF"
                    :max-grams="maxF"
                    :display-max-grams="displayMaxF"
                    :pct="sliders.fatCalPct.value"
                    :locked="sliders.lockFat.value"
                    :readonly="fatReadonly"
                    track-bg-class="bg-[#a855f7]/70"
                    :marker-grams="fatMarkers"
                    @update:grams="sliders.setFatG"
                    @toggle-lock="sliders.lockFat.value = !sliders.lockFat.value"
                />
                <label class="flex flex-col gap-1">
                    <div class="flex items-baseline gap-2">
                        <span class="text-xs font-medium text-zinc-400">Fiber (g)</span>
                        <span class="text-xs text-zinc-600">~{{ sliders.suggestedFiber.value }}g suggested</span>
                    </div>
                    <input
                        :value="sliders.fiberG.value"
                        type="number"
                        min="0"
                        step="0.1"
                        class="max-w-28 rounded-md border border-zinc-600 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                        @input="onFiberInput"
                    />
                </label>
            </div>
            <div class="flex items-start justify-center pt-2 md:pt-0">
                <div class="h-40 w-40">
                    <Chart type="doughnut" :data="chartData" :options="chartOptions" class="h-full w-full" />
                </div>
            </div>
        </div>
    </div>
</template>
