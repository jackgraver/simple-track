<script setup lang="ts">
import type { Food, MealItem } from "~/types/diet";
import { ChevronDown, Minus, Plus, Trash2 } from "lucide-vue-next";
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { formatNum, itemServingAmount } from "./logmealItemFormat";
import SimpleMacros from "~/shared/SimpleMacros.vue";

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
    swapVariant: [index: number, variant: Food];
}>();

const hasVariants = computed(
    () =>
        !!props.item.food?.variant_group_id &&
        (props.item.food.variants?.length ?? 0) > 0,
);

const variantOpen = ref(false);
const variantAnchor = ref<HTMLElement | null>(null);
const variantMenu = ref<HTMLElement | null>(null);
const variantMenuStyle = ref<Record<string, string>>({});

function updateVariantMenuPosition() {
    const el = variantAnchor.value;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const minW = 224;
    let left = rect.left;
    const maxLeft = window.innerWidth - minW - 8;
    if (left > maxLeft) left = Math.max(8, maxLeft);
    variantMenuStyle.value = {
        top: `${rect.bottom + 4}px`,
        left: `${left}px`,
        minWidth: `${Math.max(rect.width, minW)}px`,
    };
}

function toggleVariantOpen() {
    variantOpen.value = !variantOpen.value;
    if (variantOpen.value) nextTick(updateVariantMenuPosition);
}

watch(variantOpen, (open) => {
    if (open) nextTick(updateVariantMenuPosition);
});

function onDocClick(ev: MouseEvent) {
    const target = ev.target as Node;
    if (variantAnchor.value?.contains(target)) return;
    if (variantMenu.value?.contains(target)) return;
    variantOpen.value = false;
}

function closeVariantOnScroll() {
    if (variantOpen.value) variantOpen.value = false;
}

onMounted(() => {
    document.addEventListener("click", onDocClick);
    window.addEventListener("scroll", closeVariantOnScroll, true);
    window.addEventListener("resize", closeVariantOnScroll);
});
onUnmounted(() => {
    document.removeEventListener("click", onDocClick);
    window.removeEventListener("scroll", closeVariantOnScroll, true);
    window.removeEventListener("resize", closeVariantOnScroll);
});

function pickVariant(v: Food) {
    variantOpen.value = false;
    emit("swapVariant", props.rowIndex, v);
}

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
            'grid w-full min-w-0 grid-cols-[1.5rem_minmax(0,1fr)_minmax(2.5rem,auto)_2.25rem] items-center gap-x-2 gap-y-1 border-b border-secondBg py-2.5 transition-colors last:border-b-0 sm:grid-cols-[2rem_minmax(0,1fr)_9rem_11rem_2.5rem] sm:gap-x-3',
            selected && 'bg-thirdBg/25',
        ]"
    >
        <input
            type="checkbox"
            class="col-start-1 row-span-2 h-4 w-4 place-self-center accent-thirdBg sm:row-span-1"
            :checked="selected"
            @change="emit('toggleSelect', rowIndex)"
        />
        <div
            class="col-start-2 row-start-1 flex min-w-0 items-center gap-1 font-medium text-textPrimary"
            :class="compactName ? 'text-sm' : 'text-base'"
        >
            <div
                v-if="hasVariants"
                ref="variantAnchor"
                class="flex min-w-0 items-center gap-0.5"
            >
                <span
                    class="min-w-0 truncate"
                    :title="item.food?.name"
                    >{{ item.food?.name ?? "" }}</span
                >
                <button
                    type="button"
                    class="m-2! flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-textSecondary transition-colors hover:bg-secondBg hover:text-textPrimary shadow-none!"
                    :aria-expanded="variantOpen"
                    aria-haspopup="listbox"
                    aria-label="Swap variant"
                    @click.stop="toggleVariantOpen"
                >
                    <ChevronDown :size="28" :stroke-width="2.25" />
                </button>
            </div>
            <span
                v-else
                class="min-w-0 truncate"
                :title="item.food?.name"
                >{{ item.food?.name ?? "" }}</span
            >
            <Teleport to="body">
                <div
                    v-if="variantOpen && hasVariants"
                    ref="variantMenu"
                    class="fixed z-50 max-h-56 overflow-y-auto rounded-md border border-thirdBg bg-secondBg shadow-lg"
                    :style="variantMenuStyle"
                    role="listbox"
                    @click.stop
                >
                    <button
                        v-for="v in item.food!.variants"
                        :key="v.ID"
                        type="button"
                        class="flex w-full flex-col px-3 py-2 m-0! text-left text-sm text-textPrimary shadow-none! rounded-none! hover:bg-thirdBg"
                        role="option"
                        @click="pickVariant(v)"
                    >
                        <span class="font-medium">{{ v.name }}</span>
                        <SimpleMacros
                            :calories="v.calories"
                            :protein="v.protein"
                            :fat="v.fat"
                            :carbs="v.carbs"
                            :fiber="v.fiber"
                            font-size="0.75rem"
                            decimals="none"
                        />
                    </button>
                </div>
            </Teleport>
        </div>
        <div
            class="col-start-3 row-span-2 row-start-1 flex items-center justify-center gap-1 tabular-nums sm:row-span-1"
        >
            <button
                v-if="!qtyEditing"
                class="hidden md:flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
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
                    class="h-9 w-20 min-w-11 shrink-0 rounded border border-secondBg bg-secondBg px-1 text-center text-sm tabular-nums text-textPrimary focus:border-thirdBg focus:outline-none sm:w-30"
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
                class="hidden md:flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                type="button"
                @click="emit('amountPlusMinus', rowIndex, 'plus')"
            >
                <Plus :size="18" />
            </button>
        </div>
        <span
            class="col-span-2 col-start-2 row-start-2 flex min-w-0 justify-start text-sm tabular-nums text-textSecondary sm:col-span-1 sm:col-start-4 sm:row-start-1 sm:justify-end"
        >
            <SimpleMacros
                class="mt-0! min-w-0 flex-wrap gap-x-2! gap-y-0!"
                :calories="item.amount * (item.food?.calories ?? 0)"
                :protein="item.amount * (item.food?.protein ?? 0)"
                :fat="item.amount * (item.food?.fat ?? 0)"
                :carbs="item.amount * (item.food?.carbs ?? 0)"
                :fiber="item.amount * (item.food?.fiber ?? 0)"
                font-size="0.75rem"
                decimals="none"
            />
        </span>
        <button
            class="col-start-4 row-span-2 row-start-1 flex h-9 w-9 shrink-0 items-center justify-center justify-self-end rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-cfRed sm:col-start-5 sm:row-span-1 sm:row-start-1"
            type="button"
            aria-label="Remove item"
            @click="emit('remove', rowIndex)"
        >
            <Trash2 :size="20" />
        </button>
    </div>
</template>
