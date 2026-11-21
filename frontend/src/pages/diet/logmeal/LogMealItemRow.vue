<script setup lang="ts">
import type { Food, MealItem } from "~/types/diet";
import { MoreVertical } from "lucide-vue-next";
import { computed } from "vue";
import { itemServingAmount } from "./logmealItemFormat";
import { Checkbox, IconButton, Menu, NumberField } from "~/shared/ui";
import type { MenuItem } from "~/shared/ui/Menu.vue";
import { amountStep, formatKcal, itemMacros, shareOfMeal } from "./domain/mealCard";
import MacroValues from "./components/MacroValues.vue";

const props = defineProps<{
    item: MealItem;
    rowIndex: number;
    selected: boolean;
    /** Show the selection checkbox (Select mode, for grouping). */
    selecting?: boolean;
    /** Meal kcal, for this item's share bar. */
    mealCalories: number;
}>();

const emit = defineEmits<{
    toggleSelect: [index: number];
    remove: [index: number];
    setItemAmount: [index: number, amount: number];
    swapVariant: [index: number, variant: Food];
}>();

const name = computed(() => props.item.food?.name ?? "");
const macros = computed(() => itemMacros(props.item));
const kcal = computed(() => macros.value.calories);
const share = computed(() => shareOfMeal(kcal.value, props.mealCalories));
const step = computed(() => amountStep(props.item));

const menuItems = computed<MenuItem[]>(() => [
    { key: "half", label: "Halve amount" },
    { key: "double", label: "Double amount" },
    ...(props.item.food?.variants ?? []).map((v) => ({
        key: `v-${v.ID}`,
        label: `Swap to ${v.name} · ${Math.round(v.calories)} kcal`,
    })),
    { key: "remove", label: "Remove", tone: "danger" as const },
]);

function onMenu(key: string) {
    if (key === "remove") return emit("remove", props.rowIndex);
    if (key === "half") return emit("setItemAmount", props.rowIndex, props.item.amount / 2);
    if (key === "double") return emit("setItemAmount", props.rowIndex, props.item.amount * 2);
    const v = props.item.food?.variants?.find((x) => `v-${x.ID}` === key);
    if (v) emit("swapVariant", props.rowIndex, v);
}

// Quantity is shown in serving units (g, pieces); the stored amount is in servings. 0 removes.
const qty = computed<number | null>({
    get: () => itemServingAmount(props.item),
    set: (n) => {
        if (n == null || !props.item.food) return;
        if (n <= 0) {
            emit("remove", props.rowIndex);
            return;
        }
        emit("setItemAmount", props.rowIndex, n / (props.item.food.serving_amount || 1));
    },
});
</script>

<template>
    <li
        class="flex gap-3 px-4 py-3 transition-colors duration-100"
        :class="selected ? 'bg-accent-soft' : ''"
    >
        <Checkbox
            v-if="selecting"
            class="self-start"
            :model-value="selected"
            :label="`Select ${name}`"
            @update:model-value="emit('toggleSelect', rowIndex)"
        >
            <span class="sr-only">Select {{ name }}</span>
        </Checkbox>
        <div class="flex min-w-0 flex-1 flex-col gap-2">
            <div class="flex min-h-touch items-center gap-2 lg:min-h-row">
                <span class="min-w-0 flex-1 truncate text-base font-medium text-fg lg:text-sm" :title="name">{{ name }}</span>
                <span class="shrink-0 text-sm text-fg-muted tabular-nums">
                    <span class="font-medium text-fg">{{ formatKcal(kcal) }}</span> kcal
                </span>
                <Menu :items="menuItems" @select="onMenu">
                    <IconButton :label="`Options for ${name}`" size="sm">
                        <MoreVertical class="size-5" />
                    </IconButton>
                </Menu>
            </div>
            <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
                <NumberField
                    v-model="qty"
                    class="w-48 shrink-0"
                    :label="`${name} quantity`"
                    :min="0"
                    :step="step"
                    :decimals="2"
                    :unit="item.food?.serving_type === 'g' ? 'g' : undefined"
                />
                <div class="flex min-w-0 flex-1 flex-col gap-0.5">
                    <MacroValues :macros="macros" />
                    <span class="text-xs text-fg-muted tabular-nums">{{ share }}% of meal kcal</span>
                </div>
            </div>
        </div>
    </li>
</template>
