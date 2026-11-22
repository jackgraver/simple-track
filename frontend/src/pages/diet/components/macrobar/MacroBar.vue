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
const width = computed(() => calcWidth(displayTotal.value, props.planned));
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col gap-1">
        <div
            class="flex items-baseline justify-between gap-2 text-xs font-semibold tracking-wide text-fg-muted uppercase"
        >
            <span>{{ typeLabels[type] }}</span>
            <span
                v-if="percentage"
                class="font-medium tabular-nums normal-case"
                :class="percentageClass"
            >
                {{ percentage }}
            </span>
        </div>
        <div
            class="text-base font-semibold whitespace-nowrap text-fg tabular-nums"
            :class="
                determineOverflow(
                    Math.round(displayTotal),
                    Math.round(props.planned),
                    props.indicateOverflow,
                )
            "
        >
            {{
                formatInt(displayTotal) +
                " / " +
                formatInt(props.planned) +
                (valueSuffix ?? "")
            }}
        </div>
        <div
            class="h-1.5 overflow-hidden rounded-full bg-surface-3"
            role="progressbar"
            :aria-label="typeLabels[type]"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-valuenow="Math.round(width)"
        >
            <div
                class="h-full rounded-full"
                :class="macroFillClass[type]"
                :style="{ width: `${width}%` }"
            ></div>
        </div>
    </div>
</template>
