<script setup lang="ts">
import type { MealItem } from "~/types/diet";
import { Minus, Plus, Trash2 } from "lucide-vue-next";
import { nextTick, ref, watch } from "vue";
import { formatNum, itemServingAmount } from "./logmealItemFormat";
import { mealItemsListGridClass } from "./mealItemsListGrid";

const props = withDefaults(
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
    setItemAmount: [index: number, amount: number];
}>();

const qtyEditing = ref(false);
const qtyDraft = ref("");

watch(
    () => props.item.amount,
    () => {
        qtyEditing.value = false;
    },
);

function enterQtyEdit() {
    if (!props.item.food) return;
    qtyDraft.value = String(itemServingAmount(props.item));
    qtyEditing.value = true;
    nextTick(() => {
        const el = document.getElementById(
            `log-meal-qty-${props.rowIndex}`,
        ) as HTMLInputElement | null;
        el?.focus();
        el?.select();
    });
}

function onQtyDraftInput(e: Event) {
    qtyDraft.value = (e.target as HTMLInputElement).value;
}

function cancelQtyEdit() {
    qtyEditing.value = false;
    qtyDraft.value = "";
}

function commitQtyEdit() {
    if (!qtyEditing.value) return;
    const trimmed = qtyDraft.value.trim();
    const item = props.item;
    if (!item.food) {
        cancelQtyEdit();
        return;
    }
    if (trimmed === "") {
        cancelQtyEdit();
        return;
    }
    const n = Number(trimmed);
    if (!Number.isFinite(n) || n < 0) {
        cancelQtyEdit();
        return;
    }
    const serving = item.food.serving_amount || 1;
    const idx = props.rowIndex;
    cancelQtyEdit();
    if (n <= 0) {
        emit("remove", idx);
        return;
    }
    emit("setItemAmount", idx, n / serving);
}
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
                v-if="!qtyEditing"
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                type="button"
                @click="emit('amountPlusMinus', rowIndex, 'minus')"
            >
                <Minus :size="18" />
            </button>
            <div
                v-if="!qtyEditing"
                class="flex min-w-11 shrink-0 cursor-pointer select-none items-center justify-center gap-0 text-center text-sm text-textPrimary hover:opacity-90"
                role="button"
                tabindex="0"
                @click="enterQtyEdit"
                @keydown.enter.prevent="enterQtyEdit"
                @keydown.space.prevent="enterQtyEdit"
            >
                {{ formatNum(itemServingAmount(item))
                }}<span
                    v-if="item.food?.serving_type === 'g'"
                    class="text-textSecondary"
                    >g</span
                >
            </div>
            <div
                v-else
                class="flex shrink-0 items-center justify-center gap-0.5"
            >
                <input
                    :id="`log-meal-qty-${rowIndex}`"
                    type="number"
                    class="h-9 w-30 min-w-11 shrink-0 rounded border border-secondBg bg-secondBg px-1 text-center text-sm tabular-nums text-textPrimary focus:border-thirdBg focus:outline-none"
                    :value="qtyDraft"
                    min="0"
                    step="any"
                    inputmode="decimal"
                    @input="onQtyDraftInput"
                    @blur="commitQtyEdit"
                    @keydown.enter.prevent="commitQtyEdit"
                    @keydown.escape.prevent="cancelQtyEdit"
                />
                <span
                    v-if="item.food?.serving_type === 'g'"
                    class="text-sm text-textSecondary"
                    >g</span
                >
            </div>
            <button
                v-if="!qtyEditing"
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                type="button"
                @click="emit('amountPlusMinus', rowIndex, 'plus')"
            >
                <Plus :size="18" />
            </button>
        </div>
        <span
            class="min-w-0 text-right text-sm tabular-nums text-textSecondary"
        >
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
