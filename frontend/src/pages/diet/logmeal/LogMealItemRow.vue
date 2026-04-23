<script setup lang="ts">
import type { MealItem } from "~/types/diet";
import { Minus, Plus, Trash2 } from "lucide-vue-next";
import { formatNum, itemServingAmount } from "./logmealItemFormat";
import { mealItemsListGridClass } from "./mealItemsListGrid";

withDefaults(
    defineProps<{
        item: MealItem;
        rowIndex: number;
        selected: boolean;
        compactName?: boolean;
    }>(),
    { compactName: false },
);

const emit = defineEmits<{
    toggleSelect: [index: number];
    amountPlusMinus: [index: number, direction: "plus" | "minus"];
    remove: [index: number];
}>();
</script>
<template>
    <div
        :class="[
            mealItemsListGridClass,
            'gap-y-1 border-b border-secondBg py-2.5 transition-colors last:border-b-0',
            selected && 'bg-thirdBg/25',
        ]"
    >
        <input
            type="checkbox"
            class="h-4 w-4 place-self-center accent-thirdBg"
            :checked="selected"
            @change="emit('toggleSelect', rowIndex)"
        />
        <span
            class="min-w-0 truncate font-medium text-textPrimary"
            :class="compactName ? 'text-sm' : 'text-base'"
            :title="item.food?.name"
            >{{ item.food?.name ?? "" }}</span
        >
        <div class="flex items-center justify-center gap-1 tabular-nums">
            <button
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                type="button"
                @click="emit('amountPlusMinus', rowIndex, 'minus')"
            >
                <Minus :size="18" />
            </button>
            <span
                class="min-w-11 shrink-0 text-center text-sm text-textPrimary"
                >{{
                    formatNum(itemServingAmount(item))
                }}<span
                    v-if="item.food?.serving_type === 'g'"
                    class="text-textSecondary"
                    >g</span
                ></span
            >
            <button
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                type="button"
                @click="emit('amountPlusMinus', rowIndex, 'plus')"
            >
                <Plus :size="18" />
            </button>
        </div>
        <span class="min-w-0 text-right text-sm tabular-nums text-textSecondary">
            {{ formatNum(item.amount * (item.food?.calories ?? 0)) }}C /
            {{ formatNum(item.amount * (item.food?.protein ?? 0)) }}P /
            {{ formatNum(item.amount * (item.food?.fiber ?? 0)) }}F
        </span>
        <button
            class="flex h-9 w-9 shrink-0 items-center justify-center justify-self-end rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-cfRed"
            type="button"
            aria-label="Remove item"
            @click="emit('remove', rowIndex)"
        >
            <Trash2 :size="20" />
        </button>
    </div>
</template>
