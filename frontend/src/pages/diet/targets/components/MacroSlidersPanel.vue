<script setup lang="ts">
import { computed } from "vue";
import type { MacroSlidersApi } from "../useMacroSliders";
import MacroSlider from "./MacroSlider.vue";
import MacroToWeightRatio from "./MacroToWeightRatio.vue";
import { Field, Input } from "~/shared/ui";

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

const fatFromWeight = computed(() =>
    props.sliders.fatGramsFromGPerLb(
        props.weightLbs,
        props.sliders.fatGPerLb.value,
    ),
);

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

const calorieDeltaLabel = computed(() => {
    const delta = props.sliders.calorieDelta.value;
    if (delta === 0) return "matches target";
    return delta > 0 ? `+${delta} kcal` : `${delta} kcal`;
});

const calorieDeltaClass = computed(() => {
    const delta = props.sliders.calorieDelta.value;
    if (delta === 0) return "text-success";
    if (Math.abs(delta) <= 20) return "text-warn";
    return "text-danger";
});

const fiberModel = computed<number | string>({
    get: () => props.sliders.fiberG.value,
    set: (v) => props.sliders.setFiberG(Number(v) || 0),
});

// The ratio typed in the row is what gets applied (it used to be ignored in favour of the stored 0.8 / 0.3).
function applyProteinFromWeight(ratio: number) {
    props.sliders.proteinGPerLb.value = ratio;
    props.sliders.applyProteinFromBodyWeight(props.weightLbs);
}

function applyFatFromWeight(ratio: number) {
    props.sliders.fatGPerLb.value = ratio;
    props.sliders.applyFatFromBodyWeight(props.weightLbs);
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
            track-bg-class="bg-macro-calories/70"
            hide-lock
            :stat-text="calorieDeltaLabel"
            :stat-class="calorieDeltaClass"
            @update:grams="sliders.setCalorieTarget"
        />
        <div class="flex flex-col gap-4">
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
                        track-bg-class="bg-macro-protein/70"
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
                        track-bg-class="bg-macro-fat/70"
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
                        track-bg-class="bg-macro-carbs/70"
                        :marker-grams="carbsMarkers"
                        @update:grams="sliders.setCarbsG"
                        @toggle-lock="
                            sliders.lockCarbs.value = !sliders.lockCarbs.value
                        "
                    />
                    <div class="max-w-32">
                        <Field label="Fiber (g)">
                            <Input
                                v-model="fiberModel"
                                type="number"
                                inputmode="decimal"
                            />
                        </Field>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
