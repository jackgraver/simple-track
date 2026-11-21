<script setup lang="ts">
import { computed, ref } from "vue";
import { ChevronDown, ChevronRight } from "lucide-vue-next";
import type { Meal } from "~/types/diet";
import { List, ListRow } from "~/shared/ui";
import { blockMacros, mealItemsToDisplayBlocks, type MealItemWithIndex } from "~/utils/mealItemGroups";
import { formatInt, macroLine, mealMacros } from "./dietDay";

const props = defineProps<{ meal: Meal }>();

const blocks = computed(() => mealItemsToDisplayBlocks(props.meal.items));
const isQuickLog = computed(() => props.meal.items.some((item) => item.food?.quick_entry));

const collapsed = ref<Record<string, boolean>>({});
const isOpen = (id: string) => !collapsed.value[id];
const toggle = (id: string) => {
    collapsed.value = { ...collapsed.value, [id]: !collapsed.value[id] };
};

type Item = Meal["items"][number];

function itemMacroLine(item: Item): string {
    const a = Number(item.amount) || 0;
    const f = item.food;
    return macroLine({
        calories: (f?.calories ?? 0) * a,
        protein: (f?.protein ?? 0) * a,
        carbs: (f?.carbs ?? 0) * a,
        fat: (f?.fat ?? 0) * a,
        fiber: (f?.fiber ?? 0) * a,
    });
}

function groupMacroLine(rows: MealItemWithIndex[]): string {
    return macroLine(blockMacros(rows));
}

function trimNum(n: number): string {
    return n.toFixed(2).replace(/\.?0+$/, "");
}

// "180g Salmon", "2 Bananas": a gram amount never pluralises the name.
function itemLabel(item: Item): string {
    const name = item.food?.name ?? "";
    const amount = trimNum((item.food?.serving_amount || 1) * Number(item.amount));
    if (item.food?.serving_type === "g") return `${amount}g ${name}`;
    const plural = Number(item.amount) > 1 && !name.endsWith("s") ? "s" : "";
    return `${amount} ${name}${plural}`;
}
</script>

<template>
    <List>
        <ListRow v-if="isQuickLog">
            Quick log entry
            <template #trailing>{{ formatInt(mealMacros(meal).calories) }} kcal</template>
        </ListRow>
        <template v-else>
            <template v-for="(block, bi) in blocks" :key="'b-' + bi">
                <template v-if="block.kind === 'ungrouped'">
                    <ListRow v-for="{ item, index } in block.rows" :key="'u-' + index">
                        {{ itemLabel(item) }}
                        <template #subtitle>{{ itemMacroLine(item) }}</template>
                    </ListRow>
                </template>
                <template v-else>
                    <ListRow :aria-expanded="isOpen(block.groupId)" @click="toggle(block.groupId)">
                        <template #leading>
                            <ChevronDown v-if="isOpen(block.groupId)" class="size-4" />
                            <ChevronRight v-else class="size-4" />
                        </template>
                        <span class="font-medium">{{ block.label || "Group" }}</span>
                        <template #subtitle>{{ groupMacroLine(block.rows) }}</template>
                    </ListRow>
                    <template v-if="isOpen(block.groupId)">
                        <ListRow v-for="{ item, index } in block.rows" :key="'g-' + index">
                            <template #leading><span class="block w-4" /></template>
                            {{ itemLabel(item) }}
                            <template #subtitle>{{ itemMacroLine(item) }}</template>
                        </ListRow>
                    </template>
                </template>
            </template>
        </template>
    </List>
</template>
