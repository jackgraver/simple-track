<script setup lang="ts">
import type { Food, MealItem } from "~/types/diet";
import { ChevronDown, Trash2 } from "lucide-vue-next";
import { computed } from "vue";
import { itemServingAmount } from "./logmealItemFormat";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { Checkbox, IconButton, Menu, NumberField } from "~/shared/ui";

const props = defineProps<{
    item: MealItem;
    rowIndex: number;
    selected: boolean;
    compactName?: boolean;
}>();

const emit = defineEmits<{
    toggleSelect: [index: number];
    remove: [index: number];
    setItemAmount: [index: number, amount: number];
    swapVariant: [index: number, variant: Food];
}>();

const name = computed(() => props.item.food?.name ?? "");
const hasVariants = computed(
    () =>
        !!props.item.food?.variant_group_id &&
        (props.item.food.variants?.length ?? 0) > 0,
);
const variantItems = computed(() =>
    (props.item.food?.variants ?? []).map((v) => ({
        key: String(v.ID),
        label: `${v.name} · ${Math.round(v.calories)}C ${Math.round(v.protein)}P`,
    })),
);
function pickVariant(key: string) {
    const v = props.item.food?.variants?.find((x) => String(x.ID) === key);
    if (v) emit("swapVariant", props.rowIndex, v);
}

// Quantity is shown in serving units (g, pieces); the stored amount is in servings.
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
        class="flex flex-col gap-1 px-3 py-2 transition-colors duration-100 lg:px-4"
        :class="selected ? 'bg-accent-soft' : ''"
    >
        <div class="flex items-center gap-2">
            <Checkbox
                :model-value="selected"
                :label="`Select ${name}`"
                @update:model-value="emit('toggleSelect', rowIndex)"
            >
                <span class="sr-only">Select {{ name }}</span>
            </Checkbox>
            <span
                class="min-w-0 flex-1 truncate font-medium text-fg"
                :class="compactName ? 'text-sm' : 'text-base lg:text-sm'"
                :title="name"
                >{{ name }}</span
            >
            <Menu v-if="hasVariants" :items="variantItems" @select="pickVariant">
                <IconButton label="Swap variant" size="sm">
                    <ChevronDown class="size-5" />
                </IconButton>
            </Menu>
            <IconButton
                label="Remove item"
                variant="danger"
                size="sm"
                @click="emit('remove', rowIndex)"
            >
                <Trash2 class="size-5" />
            </IconButton>
        </div>
        <div class="flex flex-wrap items-center gap-x-4 gap-y-1 pl-8">
            <NumberField
                v-model="qty"
                class="w-48"
                :label="`${name} quantity`"
                :min="0"
                :step="1"
                :decimals="2"
                :unit="item.food?.serving_type === 'g' ? 'g' : undefined"
            />
            <SimpleMacros
                :calories="item.amount * (item.food?.calories ?? 0)"
                :protein="item.amount * (item.food?.protein ?? 0)"
                :fat="item.amount * (item.food?.fat ?? 0)"
                :carbs="item.amount * (item.food?.carbs ?? 0)"
                :fiber="item.amount * (item.food?.fiber ?? 0)"
                decimals="none"
            />
        </div>
    </li>
</template>
