<script setup lang="ts">
import { computed } from 'vue';
import Chart from 'primevue/chart';
import type { MacroPresetId, MacroSlidersApi } from '../useMacroSliders';
import MacroSlider from './MacroSlider.vue';

const props = defineProps<{
    sliders: MacroSlidersApi;
    weightLbs: number;
}>();

const maxP = computed(() => Math.max(1, props.sliders.calorieTarget.value / 4));
const maxC = computed(() => Math.max(1, props.sliders.calorieTarget.value / 4));
const maxF = computed(() => Math.max(1, props.sliders.calorieTarget.value / 9));

const proteinMarkers = computed(() => {
    const w = props.weightLbs;
    if (w <= 0) return [];
    return [
        props.sliders.proteinGramsFromGPerLb(w, 0.8),
        props.sliders.proteinGramsFromGPerLb(w, 1.0),
    ];
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
        <div class="grid grid-cols-1 gap-4 md:grid-cols-[1fr_10rem]">
            <div class="flex flex-col gap-4">
                <MacroSlider
                    label="Protein (g)"
                    :grams="sliders.proteinG.value"
                    :max-grams="maxP"
                    :pct="sliders.proteinCalPct.value"
                    :locked="sliders.lockProtein.value"
                    track-bg-class="bg-[#60a5fa]/70"
                    :marker-grams="proteinMarkers"
                    @update:grams="sliders.setProteinG"
                    @toggle-lock="sliders.lockProtein.value = !sliders.lockProtein.value"
                />
                <MacroSlider
                    label="Carbs (g)"
                    :grams="sliders.carbsG.value"
                    :max-grams="maxC"
                    :pct="sliders.carbsCalPct.value"
                    :locked="sliders.lockCarbs.value"
                    track-bg-class="bg-[#ef4444]/70"
                    :marker-grams="carbsMarkers"
                    @update:grams="sliders.setCarbsG"
                    @toggle-lock="sliders.lockCarbs.value = !sliders.lockCarbs.value"
                />
                <MacroSlider
                    label="Fat (g)"
                    :grams="sliders.fatG.value"
                    :max-grams="maxF"
                    :pct="sliders.fatCalPct.value"
                    :locked="sliders.lockFat.value"
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
