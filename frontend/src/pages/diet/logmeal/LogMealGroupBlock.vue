<script setup lang="ts">
import { ChevronDown, ChevronRight, MoreVertical } from "lucide-vue-next";
import { computed } from "vue";
import type { Food } from "~/types/diet";
import type { MealItemDisplayBlock } from "~/utils/mealItemGroups";
import { blockMacros } from "~/utils/mealItemGroups";
import { toast } from "~/composables/toast/useToast";
import LogMealItemRow from "./LogMealItemRow.vue";
import { formatKcal, shareOfMeal } from "./domain/mealCard";
import MacroValues from "./components/MacroValues.vue";
import { useCreateCompositeFood } from "./queries/useMealMutations";
import { IconButton, Input, List, Menu } from "~/shared/ui";

const props = defineProps<{
    block: Extract<MealItemDisplayBlock, { kind: "group" }>;
    expanded: boolean;
    selectedForGroup: Record<number, boolean>;
    selecting?: boolean;
    mealCalories: number;
}>();

const emit = defineEmits<{
    toggleCollapse: [groupId: string];
    setGroupLabel: [groupId: string, label: string];
    removeGroup: [indices: number[]];
    toggleSelect: [index: number];
    setItemAmount: [index: number, amount: number];
    removeItem: [index: number];
    swapVariant: [index: number, variant: Food];
}>();

const { mutateAsync: createCompositeAsync, isPending: savingComposite } =
    useCreateCompositeFood();

const label = computed({
    get: () => props.block.label,
    set: (v) => emit("setGroupLabel", props.block.groupId, String(v)),
});
const totals = computed(() => blockMacros(props.block.rows));
const share = computed(() => shareOfMeal(totals.value.calories, props.mealCalories));
const menuItems = computed(() => [
    {
        key: "composite",
        label: "Save as composite food",
        disabled: savingComposite.value,
    },
    { key: "remove", label: "Remove group", tone: "danger" as const },
]);

function onMenu(key: string) {
    if (key === "composite") void saveAsCompositeFood();
    else removeGroup();
}

function removeGroup() {
    emit(
        "removeGroup",
        props.block.rows.map((r) => r.index),
    );
}

async function saveAsCompositeFood() {
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
    <li :aria-busy="savingComposite || undefined">
        <div class="flex min-h-touch flex-wrap items-center gap-x-2 gap-y-1 bg-surface-2 px-4 py-2 lg:min-h-row">
            <IconButton
                :label="expanded ? 'Collapse group' : 'Expand group'"
                size="sm"
                @click="emit('toggleCollapse', block.groupId)"
            >
                <ChevronDown v-if="expanded" class="size-5" />
                <ChevronRight v-else class="size-5" />
            </IconButton>
            <div class="min-w-0 flex-1">
                <Input
                    v-model="label"
                    size="sm"
                    aria-label="Group name"
                    placeholder="Name this group"
                />
            </div>
            <span class="order-last flex w-full flex-wrap items-baseline gap-x-3 pl-12 text-sm text-fg-muted tabular-nums">
                <span><span class="font-medium text-fg">{{ formatKcal(totals.calories) }}</span> kcal</span>
                <MacroValues :macros="totals" />
                <span class="text-xs">{{ share }}% of meal kcal</span>
            </span>
            <Menu :items="menuItems" @select="onMenu">
                <IconButton label="Group options" size="sm">
                    <MoreVertical class="size-5" />
                </IconButton>
            </Menu>
        </div>
        <List v-if="expanded" class="border-l-2 border-line-strong">
            <LogMealItemRow
                v-for="{ item, index: i } in block.rows"
                :key="`g-${i}`"
                :item="item"
                :row-index="i"
                :selected="!!selectedForGroup[i]"
                :selecting="selecting"
                :meal-calories="mealCalories"
                @toggle-select="emit('toggleSelect', $event)"
                @set-item-amount="
                    (rowI, amt) => emit('setItemAmount', rowI, amt)
                "
                @swap-variant="(rowI, v) => emit('swapVariant', rowI, v)"
                @remove="emit('removeItem', $event)"
            />
        </List>
    </li>
</template>
