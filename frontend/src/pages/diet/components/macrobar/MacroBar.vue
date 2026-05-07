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
    /** Appended after "total / planned" text (e.g. water unit label). */
    valueSuffix?: string;
}>();

const { displayTotal } = useMacroBarAnimation(() => props.total ?? 0);
</script>

<template>
    <div class="@container flex min-w-0 flex-1 flex-col">
        <div
            class="mb-0.5 hidden pl-0.5 text-[clamp(0.45rem,4cqi+0.15rem,0.65rem)] font-semibold uppercase leading-[1.15] tracking-[0.02em] text-textSecondary @min-[2.75rem]:block"
        >
            {{ typeLabels[type] }}
        </div>
        <div
            class="flex-1 rounded border border-solid border-secondBg leading-5"
        >
            <div
                class="@container flex h-full min-w-0 items-center justify-center rounded"
                :class="macroFillClass[type]"
                :style="{
                    width: `${calcWidth(displayTotal, props.planned)}%`,
                }"
            >
                <span
                    class="hidden max-w-full px-0.5 font-bold whitespace-nowrap text-[clamp(0.5rem,3cqi+0.35rem,0.95rem)] tabular-nums text-white @min-[3.75rem]:inline"
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
