<script setup lang="ts">
import { useMacroBarAnimation } from "./useMacroBarAnimation";
import {
    calcWidth,
    determineOverflow,
    formatInt,
    macroFillClass,
    type MacroBarNutrientType,
    typeLabels,
} from "./useMacroBarStyling";

const props = defineProps<{
    total: number;
    planned: number;
    type: MacroBarNutrientType;
    indicateOverflow?: boolean;
}>();

const { displayTotal } = useMacroBarAnimation(() => props.total ?? 0);
</script>

<template>
    <div class="flex min-w-0 flex-1 flex-col">
        <div
            class="mb-0.5 pl-0.5 text-[0.65rem] font-semibold uppercase leading-[1.15] tracking-[0.02em] text-textSecondary"
        >
            {{ typeLabels[type] }}
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
                        formatInt(displayTotal) + " / " + formatInt(props.planned)
                    }}</span
                >
            </div>
        </div>
    </div>
</template>
