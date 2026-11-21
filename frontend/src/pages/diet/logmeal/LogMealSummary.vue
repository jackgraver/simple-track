<script setup lang="ts">
// Receipt summary in two labelled parts: what this meal adds, and where the day lands once it's
// logged. The day part is omitted for saved meals (not tied to a day).
import { computed } from "vue";
import { AlertTriangle } from "lucide-vue-next";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import MacroValues from "./components/MacroValues.vue";
import { dayFit, formatKcal, type Macros } from "./domain/mealCard";

const props = defineProps<{
    meal: Macros;
    /** Day totals with this meal counted. */
    dayAfter?: Macros;
    plan?: { calories: number; protein: number; carbs: number; fat: number };
}>();

const kcalFit = computed(() => (props.dayAfter ? dayFit(props.dayAfter.calories, props.plan?.calories) : null));

const overline = "m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase";
</script>

<template>
    <div class="flex flex-col gap-4 lg:flex-row lg:gap-8">
        <section class="flex shrink-0 flex-col gap-1" aria-labelledby="sum-meal">
            <h2 id="sum-meal" :class="overline">This meal</h2>
            <p class="m-0 flex items-baseline gap-2">
                <span class="text-3xl font-semibold tracking-tight text-fg tabular-nums">{{ formatKcal(meal.calories) }}</span>
                <span class="text-sm text-fg-muted">kcal</span>
            </p>
            <MacroValues :macros="meal" long />
        </section>
        <section
            v-if="dayAfter && plan"
            class="flex min-w-0 flex-1 flex-col gap-2 border-t border-line pt-4 lg:border-t-0 lg:border-l lg:pt-0 lg:pl-8"
            aria-labelledby="sum-day"
        >
            <div class="flex flex-wrap items-baseline justify-between gap-x-3">
                <h2 id="sum-day" :class="overline">Today after logging</h2>
                <span
                    v-if="kcalFit"
                    class="flex items-center gap-1 text-sm tabular-nums"
                    :class="kcalFit.over ? 'font-medium text-warn' : 'text-fg-muted'"
                >
                    <template v-if="kcalFit.over">
                        <AlertTriangle aria-hidden="true" class="size-4 shrink-0" />
                        {{ formatKcal(-kcalFit.left) }} kcal over
                    </template>
                    <template v-else>{{ formatKcal(kcalFit.left) }} kcal left</template>
                </span>
            </div>
            <MacroBars
                two-by-two
                :total-calories="dayAfter.calories"
                :total-protein="dayAfter.protein"
                :total-fat="dayAfter.fat"
                :total-carbs="dayAfter.carbs"
                :planned-calories="plan.calories"
                :planned-protein="plan.protein"
                :planned-fat="plan.fat"
                :planned-carbs="plan.carbs"
            />
        </section>
    </div>
</template>
