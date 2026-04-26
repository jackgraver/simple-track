<script setup lang="ts">
import { ChevronDown, ChevronRight, MoreVertical } from "lucide-vue-next";
import { onMounted, onUnmounted, ref } from "vue";
import type { Food } from "~/types/diet";
import type { MealItemDisplayBlock } from "~/utils/mealItemGroups";
import { blockMacros } from "~/utils/mealItemGroups";
import { toast } from "~/composables/toast/useToast";
import { formatNum } from "./logmealItemFormat";
import { mealItemsListGridClass } from "./mealItemsListGrid";
import LogMealItemRow from "./LogMealItemRow.vue";
import { useCreateCompositeFood } from "./queries/useMealMutations";

const props = defineProps<{
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
    setItemAmount: [index: number, amount: number];
    removeItem: [index: number];
    swapVariant: [index: number, variant: Food];
}>();

const menuOpen = ref(false);
const menuRoot = ref<HTMLElement | null>(null);
const { mutateAsync: createCompositeAsync, isPending: savingComposite } =
    useCreateCompositeFood();

function onDocumentClick(ev: MouseEvent) {
    const el = menuRoot.value;
    if (!el || el.contains(ev.target as Node)) return;
    menuOpen.value = false;
}

onMounted(() => document.addEventListener("click", onDocumentClick));
onUnmounted(() => document.removeEventListener("click", onDocumentClick));

function removeGroup() {
    menuOpen.value = false;
    emit(
        "removeGroup",
        props.block.rows.map((r) => r.index),
    );
}

async function saveAsCompositeFood() {
    menuOpen.value = false;
    const name = props.block.label.trim();
    if (!name) {
        toast.push(
            "Name the group before saving as a composite food.",
            "error",
        );
        return;
    }
    const items = props.block.rows
        .filter((r) => r.item.food_id)
        .map((r) => ({
            food_id: r.item.food_id,
            amount: Number(r.item.amount),
        }));
    if (items.length === 0) {
        toast.push("This group has no foods to save.", "error");
        return;
    }
    try {
        await createCompositeAsync({ name, items });
        toast.push(
            "Composite food saved. You can find it in Add Foods.",
            "success",
        );
    } catch {
        toast.push("Could not save composite food.", "error");
    }
}
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
            <div ref="menuRoot" class="relative justify-self-end">
                <button
                    type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded text-textSecondary transition-colors hover:bg-secondBg hover:text-textPrimary"
                    :aria-expanded="menuOpen"
                    aria-haspopup="true"
                    aria-label="Group options"
                    @click.stop="menuOpen = !menuOpen"
                >
                    <MoreVertical :size="20" />
                </button>
                <div
                    v-if="menuOpen"
                    class="absolute right-0 top-full z-20 mt-0.5 min-w-52 rounded-md border border-secondBg bg-firstBg py-1 shadow-lg"
                    role="menu"
                    @click.stop
                >
                    <button
                        type="button"
                        class="flex w-full items-center px-3 py-2 text-left text-sm text-textPrimary hover:bg-secondBg disabled:cursor-wait disabled:opacity-50"
                        role="menuitem"
                        :disabled="savingComposite"
                        @click="saveAsCompositeFood"
                    >
                        Save as composite food
                    </button>
                    <button
                        type="button"
                        class="flex w-full items-center px-3 py-2 text-left text-sm text-cfRed hover:bg-secondBg"
                        role="menuitem"
                        @click="removeGroup"
                    >
                        Remove
                    </button>
                </div>
            </div>
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
                @set-item-amount="
                    (rowI, amt) => emit('setItemAmount', rowI, amt)
                "
                @swap-variant="
                    (rowI, v) => emit('swapVariant', rowI, v)
                "
                @remove="emit('removeItem', $event)"
            />
        </div>
    </div>
</template>
