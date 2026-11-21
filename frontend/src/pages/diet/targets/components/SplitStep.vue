<script setup lang="ts">
import { computed } from "vue";
import { Button, Card, Field, Input } from "~/shared/ui";
import { calorieGapLabel, fmtInt, gramsPerLb, type MacroKey, splitSegments } from "../domain/targets";
import type { MacroSlidersApi } from "../useMacroSliders";
import MacroSlider from "./MacroSlider.vue";

const props = defineProps<{ sliders: MacroSlidersApi; weightLbs: number }>();

const PROTEIN_PRESETS = [0.8, 1.0, 1.2];
const FAT_PRESETS = [0.3, 0.4, 0.5];

const s = props.sliders;
const segments = computed(() =>
    splitSegments({ protein: s.proteinG.value, carbs: s.carbsG.value, fat: s.fatG.value }),
);
const pct = (k: MacroKey) => Math.round(segments.value.find((x) => x.key === k)?.pct ?? 0);
const LABEL: Record<MacroKey, string> = { protein: "Protein", carbs: "Carbs", fat: "Fat" };
const BAR: Record<MacroKey, string> = { protein: "bg-macro-protein", carbs: "bg-macro-carbs", fat: "bg-macro-fat" };

const gap = computed(() => s.calorieDelta.value);
const gapClass = computed(() => (gap.value === 0 ? "text-success" : Math.abs(gap.value) <= 50 ? "text-fg-muted" : "text-warn"));

function hint(grams: number, k: MacroKey) {
    const perLb = gramsPerLb(grams, props.weightLbs);
    return [perLb != null ? `${perLb} g/lb` : null, `${pct(k)}%`].filter(Boolean).join(" · ");
}

const presetGrams = (ratios: number[]) =>
    props.weightLbs > 0 ? ratios.map((r) => Math.round(r * props.weightLbs)) : [];

function applyProtein(ratio: number) {
    s.proteinGPerLb.value = ratio;
    s.applyProteinFromBodyWeight(props.weightLbs);
}
function applyFat(ratio: number) {
    s.fatGPerLb.value = ratio;
    s.applyFatFromBodyWeight(props.weightLbs);
}

const fiberModel = computed<number | string>({
    get: () => s.fiberG.value,
    set: (v) => s.setFiberG(Number(v) || 0),
});
</script>

<template>
    <Card>
        <div class="flex flex-col gap-4">
            <div class="flex flex-col gap-2">
                <p class="m-0 flex flex-wrap items-baseline justify-between gap-x-3 text-sm">
                    <span class="text-fg-muted">
                        Macros add up to
                        <span class="font-medium text-fg tabular-nums">{{ fmtInt(sliders.macroCaloriesTotal.value) }}</span>
                        of {{ fmtInt(sliders.calorieTarget.value) }} kcal
                    </span>
                    <span class="font-medium tabular-nums" :class="gapClass">{{ calorieGapLabel(gap) }}</span>
                </p>
                <div class="flex h-3 w-full gap-0.5 overflow-hidden rounded-full bg-surface-3" aria-hidden="true">
                    <div
                        v-for="seg in segments"
                        :key="seg.key"
                        class="h-full transition-[width] duration-150"
                        :class="BAR[seg.key]"
                        :style="{ width: `${seg.pct}%` }"
                    ></div>
                </div>
                <ul class="m-0 flex list-none justify-between gap-2 p-0 text-xs text-fg-muted">
                    <li v-for="seg in segments" :key="seg.key" class="flex items-center gap-1.5 tabular-nums">
                        <span aria-hidden="true" class="size-2 rounded-full" :class="BAR[seg.key]"></span>
                        {{ LABEL[seg.key] }} {{ Math.round(seg.pct) }}%
                    </li>
                </ul>
            </div>

            <MacroSlider
                label="Protein"
                unit="g"
                :hint="hint(sliders.proteinG.value, 'protein')"
                :grams="sliders.proteinG.value"
                :max-grams="sliders.maxProteinG()"
                :display-max-grams="Math.max(1, sliders.maxProteinG())"
                :locked="sliders.lockProtein.value"
                track-bg-class="bg-macro-protein/70"
                :marker-grams="presetGrams(PROTEIN_PRESETS)"
                @update:grams="sliders.setProteinG"
                @toggle-lock="sliders.lockProtein.value = !sliders.lockProtein.value"
            >
                <div v-if="weightLbs > 0" class="flex gap-1">
                    <Button
                        v-for="r in PROTEIN_PRESETS"
                        :key="r"
                        size="sm"
                        variant="ghost"
                        :disabled="sliders.lockProtein.value"
                        @click="applyProtein(r)"
                        >{{ r.toFixed(1) }} g/lb</Button
                    >
                </div>
            </MacroSlider>

            <MacroSlider
                label="Fat"
                unit="g"
                :hint="hint(sliders.fatG.value, 'fat')"
                :grams="sliders.fatG.value"
                :max-grams="sliders.maxFatG()"
                :display-max-grams="Math.max(1, sliders.maxFatG())"
                :locked="sliders.lockFat.value"
                track-bg-class="bg-macro-fat/70"
                :marker-grams="presetGrams(FAT_PRESETS)"
                @update:grams="sliders.setFatG"
                @toggle-lock="sliders.lockFat.value = !sliders.lockFat.value"
            >
                <div v-if="weightLbs > 0" class="flex gap-1">
                    <Button
                        v-for="r in FAT_PRESETS"
                        :key="r"
                        size="sm"
                        variant="ghost"
                        :disabled="sliders.lockFat.value"
                        @click="applyFat(r)"
                        >{{ r.toFixed(1) }} g/lb</Button
                    >
                </div>
            </MacroSlider>

            <MacroSlider
                label="Carbs"
                unit="g"
                :hint="`${pct('carbs')}% · ${sliders.lockCarbs.value ? 'locked' : 'fills the rest'}`"
                :grams="sliders.carbsG.value"
                :max-grams="sliders.maxCarbsG()"
                :display-max-grams="Math.max(1, sliders.maxCarbsG())"
                :locked="sliders.lockCarbs.value"
                track-bg-class="bg-macro-carbs/70"
                :marker-grams="[sliders.suggestedCarbsG.value]"
                @update:grams="sliders.setCarbsG"
                @toggle-lock="sliders.lockCarbs.value = !sliders.lockCarbs.value"
            >
                <Button
                    v-if="sliders.carbsG.value !== sliders.suggestedCarbsG.value && !sliders.lockCarbs.value"
                    size="sm"
                    variant="ghost"
                    @click="sliders.fillCarbsFromRemaining()"
                    >Fill to {{ sliders.suggestedCarbsG.value }} g</Button
                >
            </MacroSlider>

            <div class="flex items-end gap-3 border-t border-line pt-4">
                <div class="w-24">
                    <Field label="Fiber (g)">
                        <Input v-model="fiberModel" type="number" inputmode="decimal" />
                    </Field>
                </div>
                <p class="m-0 pb-3 text-xs text-fg-muted lg:pb-2">
                    ~{{ sliders.suggestedFiber.value }} g suggested (14 g per 1,000 kcal)
                </p>
            </div>
        </div>
    </Card>
</template>
