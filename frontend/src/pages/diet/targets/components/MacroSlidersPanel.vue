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
        legend: {
            labels: { color: '#a1a1aa' },
        },
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
        <label class="flex flex-col gap-1">
            <span class="text-sm font-medium text-zinc-300">Calories (anchor)</span>
            <input
                :value="sliders.calorieTarget.value"
                type="number"
                min="0"
                step="1"
                class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                @input="
                    sliders.setCalorieTarget(
                        Number(($event.target as HTMLInputElement).value) || 0,
                    )
                "
            />
        </label>
        <p
            v-if="sliders.caloriesAdjustedByLocks.value"
            class="text-sm text-amber-300/90"
        >
            Calorie target was updated to match locked macros.
        </p>
        <div class="flex flex-wrap gap-2">
            <button
                v-for="p in presets"
                :key="p.id"
                type="button"
                class="rounded-md border border-zinc-600 px-2 py-1 text-xs font-medium text-zinc-300 hover:border-amber-600/60 hover:text-zinc-100"
                @click="sliders.applyPresetMacros(p.id)"
            >
                {{ p.label }}
            </button>
        </div>
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
            @update:grams="sliders.setFatG"
            @toggle-lock="sliders.lockFat.value = !sliders.lockFat.value"
        />
        <label class="flex flex-col gap-1">
            <span class="text-sm font-medium text-zinc-300">Fiber (g)</span>
            <span class="text-xs text-zinc-500"
                >Suggested ~{{ sliders.suggestedFiber.value }} g (14g / 1000 kcal)</span
            >
            <input
                :value="sliders.fiberG.value"
                type="number"
                min="0"
                step="0.1"
                class="rounded-md border border-zinc-600 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                @input="onFiberInput"
            />
        </label>
        <div class="h-56">
            <Chart type="doughnut" :data="chartData" :options="chartOptions" class="h-full" />
        </div>
    </div>
</template>
