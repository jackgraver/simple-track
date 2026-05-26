<script setup lang="ts">
import { computed } from "vue";
import { useMacroBarAnimation } from "./useMacroBarAnimation";
import {
    calcWidth,
    determineOverflow,
    formatInt,
    formatPercentage,
    macroFillClass,
    percentageColorClass,
    calcPercentage,
    type MacroBarNutrientType,
    typeLabels,
} from "./useMacroBarStyling";

const props = defineProps<{
    total: number;
    planned: number;
    type: MacroBarNutrientType;
    indicateOverflow?: boolean;
    /** Appended after "total / planned" text (e.g. water unit label). */
    valueSuffix?: string;
}>();

const { displayTotal } = useMacroBarAnimation(() => props.total ?? 0);
const percentageValue = computed(() =>
    calcPercentage(props.total, props.planned),
);
const percentage = computed(() => formatPercentage(props.total, props.planned));
const percentageClass = computed(() =>
    percentageColorClass(percentageValue.value),
);
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col">
        <div
            class="mb-0.5 pl-0.5 text-[0.65rem] font-semibold uppercase leading-[1.15] tracking-[0.02em] text-textSecondary flex items-center gap-1"
        >
            <span> {{ typeLabels[type] }} </span>
            <span v-if="percentage" :class="percentageClass">
                {{ percentage }}
            </span>
        </div>
        <div class="flex-1 rounded border border-solid border-secondBg">
            <div
                class="h-full rounded text-center leading-5"
                :class="macroFillClass[type]"
                :style="{
                    width: `${calcWidth(displayTotal, props.planned)}%`,
                }"
            >
                <span
                    class="tabular-nums whitespace-nowrap px-2 font-bold text-white"
                    :class="
                        determineOverflow(
                            Math.round(displayTotal),
                            Math.round(props.planned),
                            props.indicateOverflow,
                        )
                    "
                    >{{
                        formatInt(displayTotal) +
                        " / " +
                        formatInt(props.planned) +
                        (valueSuffix ?? "")
                    }}</span
                >
            </div>
        </div>
    </div>
</template>

<style scoped>
.calories {
    background-color: var(--macro-calories);
}
.protein {
    background-color: var(--macro-protein);
}
.carbs {
    background-color: var(--macro-carbs);
}
.fat {
    background-color: var(--macro-fat);
}
.fiber {
    background-color: var(--macro-fiber);
}
.water {
    background-color: var(--macro-water);
}
</style>
