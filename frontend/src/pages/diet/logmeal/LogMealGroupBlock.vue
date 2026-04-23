<script setup lang="ts">
import { ChevronDown, ChevronRight, Trash2 } from "lucide-vue-next";
import type { MealItemDisplayBlock } from "~/utils/mealItemGroups";
import { blockMacros } from "~/utils/mealItemGroups";
import { formatNum } from "./logmealItemFormat";
import { mealItemsListGridClass } from "./mealItemsListGrid";
import LogMealItemRow from "./LogMealItemRow.vue";

defineProps<{
    block: Extract<MealItemDisplayBlock, { kind: "group" }>;
    expanded: boolean;
    selectedForGroup: Record<number, boolean>;
}>();

const emit = defineEmits<{
    toggleCollapse: [groupId: string];
    setGroupLabel: [groupId: string, label: string];
    removeGroup: [indices: number[]];
    toggleSelect: [index: number];
    amountPlusMinus: [index: number, direction: "plus" | "minus"];
    removeItem: [index: number];
}>();
</script>
<template>
    <div class="rounded-md bg-secondBg/40 py-2 px-4">
        <div
            :class="[
                mealItemsListGridClass,
                'gap-y-1',
                expanded && 'border-b border-secondBg/60',
            ]"
        >
            <div
                class="col-span-2 flex min-w-0 w-full max-w-full items-center justify-start gap-2 self-stretch"
            >
                <button
                    type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded border border-secondBg bg-secondBg text-textPrimary transition-colors hover:border-thirdBg hover:bg-thirdBg"
                    @click="emit('toggleCollapse', block.groupId)"
                >
                    <ChevronRight v-if="!expanded" :size="18" />
                    <ChevronDown v-else :size="18" />
                </button>
                <input
                    type="text"
                    class="min-w-0 flex-1 basis-0 rounded border border-secondBg bg-secondBg px-2 py-1 text-left text-base font-medium text-textPrimary placeholder:text-textSecondary focus:border-thirdBg focus:outline-none"
                    :value="block.label"
                    placeholder="Name this group"
                    @input="
                        emit(
                            'setGroupLabel',
                            block.groupId,
                            ($event.target as HTMLInputElement).value,
                        )
                    "
                    @click.stop
                />
            </div>
            <span class="text-center text-xs text-textSecondary">—</span>
            <span
                class="min-w-0 text-right text-sm tabular-nums text-textSecondary"
            >
                {{ formatNum(blockMacros(block.rows).calories) }}C /
                {{ formatNum(blockMacros(block.rows).protein) }}P /
                {{ formatNum(blockMacros(block.rows).fiber) }}F
            </span>
            <button
                class="flex h-9 w-9 shrink-0 items-center justify-center justify-self-end rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-cfRed"
                type="button"
                aria-label="Remove group"
                @click="
                    emit(
                        'removeGroup',
                        block.rows.map((r) => r.index),
                    )
                "
            >
                <Trash2 :size="20" />
            </button>
        </div>
        <div v-if="expanded">
            <LogMealItemRow
                v-for="{ item, index: i } in block.rows"
                :key="`g-${i}`"
                :item="item"
                :row-index="i"
                :selected="!!selectedForGroup[i]"
                compact-name
                @toggle-select="emit('toggleSelect', $event)"
                @amount-plus-minus="
                    (rowI, direction) =>
                        emit('amountPlusMinus', rowI, direction)
                "
                @remove="emit('removeItem', $event)"
            />
        </div>
    </div>
</template>
