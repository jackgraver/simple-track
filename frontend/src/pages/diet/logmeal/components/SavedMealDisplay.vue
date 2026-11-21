<script setup lang="ts">
import { computed } from "vue";
import type { SavedMeal } from "~/types/diet";
import MacroValues from "./MacroValues.vue";
import { formatKcal, itemsMacros } from "../domain/mealCard";

const props = defineProps<{ item: SavedMeal }>();

const totals = computed(() => itemsMacros(props.item.items ?? []));
const count = computed(() => props.item.items?.length ?? 0);
</script>

<template>
    <div class="flex min-w-0 flex-col items-start gap-1 whitespace-normal text-fg">
        <h3 class="m-0 text-base font-semibold">{{ item.name }}</h3>
        <span class="flex flex-wrap items-baseline gap-x-3 gap-y-0.5 text-sm text-fg-muted tabular-nums">
            <span><span class="font-medium text-fg">{{ formatKcal(totals.calories) }}</span> kcal</span>
            <MacroValues :macros="totals" hide-fiber />
            <span>{{ count }} {{ count === 1 ? "item" : "items" }}</span>
        </span>
    </div>
</template>
