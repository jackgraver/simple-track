<script setup lang="ts">
import { computed } from "vue";
import type { MacroSlidersApi } from "../useMacroSliders";
import MacroSlider from "./MacroSlider.vue";
import MacroToWeightRatio from "./MacroToWeightRatio.vue";

const props = defineProps<{
    sliders: MacroSlidersApi;
    weightLbs: number;
}>();

const maxP = computed(() => props.sliders.maxProteinG());
const maxC = computed(() => props.sliders.maxCarbsG());
const maxF = computed(() => props.sliders.maxFatG());
const minP = computed(() => props.sliders.minProteinG());
const minC = computed(() => props.sliders.minCarbsG());
const minF = computed(() => props.sliders.minFatG());
const displayMaxP = computed(() => Math.max(1, maxP.value));
const displayMaxC = computed(() => Math.max(1, maxC.value));
const displayMaxF = computed(() => Math.max(1, maxF.value));

const proteinFromWeight = computed(() =>
    props.sliders.proteinGramsFromGPerLb(
        props.weightLbs,
        props.sliders.proteinGPerLb.value,
    ),
);

const fatFromWeight = computed(() =>
    props.sliders.fatGramsFromGPerLb(
        props.weightLbs,
        props.sliders.fatGPerLb.value,
    ),
);

const proteinWeightHint = computed(() => {
    if (props.weightLbs <= 0) return "Log body weight to use g/lb";
    return `${proteinFromWeight.value}g at ${props.weightLbs} lbs`;
});

const fatWeightHint = computed(() => {
    if (props.weightLbs <= 0) return "Log body weight to use g/lb";
    return `${fatFromWeight.value}g at ${props.weightLbs} lbs`;
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
    const suggested = props.sliders.suggestedCarbsG.value;
    const markers = new Set<number>([suggested]);
    if (suggested > 0) {
        markers.add(Math.max(0, Math.round(suggested / 5) * 5));
        markers.add(Math.max(0, Math.ceil(suggested / 5) * 5));
    }
    return [...markers].filter((g) => g > 0).sort((a, b) => a - b);
});

const fatMarkers = computed(() => {
    const weight = props.weightLbs;
    const markers = new Set<number>();
    if (weight > 0) {
        markers.add(props.sliders.fatGramsFromGPerLb(weight, 0.3));
        markers.add(props.sliders.fatGramsFromGPerLb(weight, 0.4));
        markers.add(fatFromWeight.value);
    }
    return [...markers].filter((g) => g > 0).sort((a, b) => a - b);
});

const macroCalories = computed(() =>
    Math.round(props.sliders.macroCaloriesTotal.value),
);

const calorieDeltaLabel = computed(() => {
    const delta = props.sliders.calorieDelta.value;
    if (delta === 0) return "matches target";
    return delta > 0 ? `+${delta} kcal` : `${delta} kcal`;
});

const calorieDeltaClass = computed(() => {
    const delta = props.sliders.calorieDelta.value;
    if (delta === 0) return "text-emerald-400/90";
    if (Math.abs(delta) <= 20) return "text-amber-300/90";
    return "text-red-400/90";
});

const carbsNeedsFill = computed(
    () => props.sliders.carbsG.value !== props.sliders.suggestedCarbsG.value,
);

const chartData = computed(() => {
    const p = props.sliders.proteinG.value;
    const c = props.sliders.carbsG.value;
    const f = props.sliders.fatG.value;
    return {
        labels: ["Protein", "Carbs", "Fat"],
        datasets: [
            {
                data: [4 * p, 4 * c, 9 * f],
                backgroundColor: ["#60a5fa", "#ef4444", "#a855f7"],
                borderWidth: 0,
            },
        ],
    };
});

const chartOptions = computed(() => ({
    plugins: { legend: { display: false } },
    maintainAspectRatio: false,
}));

function onFiberInput(e: Event) {
    props.sliders.setFiberG(Number((e.target as HTMLInputElement).value) || 0);
}

function onProteinMultiplierInput(e: Event) {
    props.sliders.proteinGPerLb.value =
        Number((e.target as HTMLInputElement).value) || 0;
}

function onFatMultiplierInput(e: Event) {
    props.sliders.fatGPerLb.value =
        Number((e.target as HTMLInputElement).value) || 0;
}

function applyProteinFromWeight() {
    props.sliders.applyProteinFromBodyWeight(props.weightLbs);
}

function applyFatFromWeight() {
    props.sliders.applyFatFromBodyWeight(props.weightLbs);
}

function fillRemainingCarbs() {
    props.sliders.fillCarbsFromRemaining();
}
</script>

<template>
    <div class="flex flex-col gap-4">
        <MacroSlider
            label="Calorie target"
            :grams="sliders.calorieTarget.value"
            :min-grams="0"
            :max-grams="5000"
            :display-max-grams="5000"
            track-bg-class="bg-amber-500/70"
            hide-lock
            :stat-text="calorieDeltaLabel"
            :stat-class="calorieDeltaClass"
            @update:grams="sliders.setCalorieTarget"
        />
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
                        track-bg-class="bg-[#60a5fa]/70"
                        :marker-grams="proteinMarkers"
                        @update:grams="sliders.setProteinG"
                        @toggle-lock="
                            sliders.lockProtein.value =
                                !sliders.lockProtein.value
                        "
                    />
                    <MacroToWeightRatio
                        label="Protein"
                        :weight="weightLbs"
                        @set-ratio="applyProteinFromWeight"
                    />
                </div>

                <div class="flex flex-col gap-2">
                    <MacroSlider
                        label="Fat (g)"
                        :grams="sliders.fatG.value"
                        :min-grams="minF"
                        :max-grams="maxF"
                        :display-max-grams="displayMaxF"
                        :pct="sliders.fatCalPct.value"
                        :locked="sliders.lockFat.value"
                        track-bg-class="bg-[#a855f7]/70"
                        :marker-grams="fatMarkers"
                        @update:grams="sliders.setFatG"
                        @toggle-lock="
                            sliders.lockFat.value = !sliders.lockFat.value
                        "
                    />
                    <MacroToWeightRatio
                        label="Fat"
                        :weight="weightLbs"
                        @set-ratio="applyFatFromWeight"
                    />
                </div>
                <div class="flex flex-col gap-2">
                    <MacroSlider
                        label="Carbs (g)"
                        :grams="sliders.carbsG.value"
                        :min-grams="minC"
                        :max-grams="maxC"
                        :display-max-grams="displayMaxC"
                        :pct="sliders.carbsCalPct.value"
                        :locked="sliders.lockCarbs.value"
                        track-bg-class="bg-[#ef4444]/70"
                        :marker-grams="carbsMarkers"
                        @update:grams="sliders.setCarbsG"
                        @toggle-lock="
                            sliders.lockCarbs.value = !sliders.lockCarbs.value
                        "
                    />
                    <span class="text-xs text-zinc-500">Fiber</span>
                    <input
                        :value="sliders.fiberG.value"
                        type="number"
                        min="0"
                        :max="sliders.carbsG.value"
                        step="1"
                        class="max-w-24 rounded-md border border-zinc-600 bg-zinc-900 px-2 py-1 text-sm text-zinc-100 outline-none ring-amber-700/40 focus:border-amber-600 focus:ring-2"
                        @input="onFiberInput"
                    />
                </div>
            </div>
        </div>
    </div>
</template>
