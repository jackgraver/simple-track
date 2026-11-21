<script setup lang="ts">
import type { Meal } from "~/types/diet";
import { EDIT_LOGGED_TYPE, EDIT_TYPE } from "~/pages/diet/logmeal/logmealMode";
import {
    Trash2,
    SquarePen,
    Check,
    ChevronRight,
    ChevronDown,
} from "lucide-vue-next";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { Button, Card, IconButton, List, ListRow } from "~/shared/ui";
import MealCardMacroDetails from "./MealCardMacroDetails.vue";
import { computed, ref } from "vue";
import {
    blockMacros,
    mealItemsToDisplayBlocks,
    type MealItemWithIndex,
} from "~/utils/mealItemGroups";

function formatNum(n: number) {
    const s = n.toFixed(2);
    return s.replace(/\.?0+$/, "");
}

function itemServingAmount(item: Meal["items"][number]): number {
    return (item.food?.serving_amount || 1) * Number(item.amount);
}

function macroTotalsForMeal(meal: Meal) {
    let calories = 0;
    let protein = 0;
    let carbs = 0;
    let fat = 0;
    let fiber = 0;
    for (const item of meal.items) {
        const a = Number(item.amount);
        calories += (item.food?.calories ?? 0) * a;
        protein += (item.food?.protein ?? 0) * a;
        carbs += (item.food?.carbs ?? 0) * a;
        fat += (item.food?.fat ?? 0) * a;
        fiber += (item.food?.fiber ?? 0) * a;
    }
    return { calories, protein, carbs, fat, fiber };
}

const props = defineProps<{
    meal: Meal;
    type: "planned" | "logged";
    onLogPlanned: (meal: Meal) => void | Promise<void>;
    onLogEdited: (
        meal: Meal,
        type: typeof EDIT_TYPE | typeof EDIT_LOGGED_TYPE,
    ) => void;
    onDelete: (meal: Meal) => void | Promise<void>;
}>();

const mealMacroTotals = computed(() => macroTotalsForMeal(props.meal));

const mealBlocks = computed(() => mealItemsToDisplayBlocks(props.meal.items));
const collapsedGroups = ref<Record<string, boolean>>({});
function isGroupExpanded(groupId: string): boolean {
    return !collapsedGroups.value[groupId];
}
function toggleGroupCollapse(groupId: string) {
    collapsedGroups.value = {
        ...collapsedGroups.value,
        [groupId]: !collapsedGroups.value[groupId],
    };
}

function pluralize(name: string | undefined, amount: number): string {
    if (!name) return "";
    return amount > 1 && name.charAt(name.length - 1) !== "s" ? "s" : "";
}

function groupMacroTotals(rows: MealItemWithIndex[]) {
    const m = blockMacros(rows);
    return {
        calories: m.calories,
        protein: m.protein,
        carbs: m.carbs,
        fat: m.fat,
        fiber: m.fiber,
    };
}

const isQuickLog = computed(() =>
    props.meal.items.some((item) => item.food?.quick_entry),
);

// The handlers toast their own failures; this only drives the pending state.
const busy = ref(false);
async function run(fn: () => void | Promise<void>) {
    if (busy.value) return;
    busy.value = true;
    try {
        await fn();
    } finally {
        busy.value = false;
    }
}

// "180g Salmon", "2 Bananas": a gram amount never pluralises the name.
function itemLabel(food: Meal["items"][number]): string {
    const name = food.food?.name ?? "";
    const amount = formatNum(itemServingAmount(food));
    if (food.food?.serving_type === "g") return `${amount}g ${name}`;
    return `${amount} ${name}${pluralize(name, Number(food.amount))}`;
}
</script>

<template>
    <Card padding="none" :aria-busy="busy || undefined">
        <div class="flex items-start gap-3 p-4">
            <div class="flex min-w-0 flex-1 flex-col gap-1">
                <h3 class="m-0 truncate text-base font-semibold text-fg">
                    {{ meal.name }}
                </h3>
            </div>
            <div class="flex shrink-0 items-center gap-1">
                <template v-if="type === 'logged'">
                    <IconButton
                        label="Edit meal"
                        :disabled="busy"
                        @click="onLogEdited(meal, EDIT_TYPE)"
                    >
                        <SquarePen class="size-5" />
                    </IconButton>
                    <IconButton
                        label="Delete meal"
                        variant="danger"
                        :disabled="busy"
                        @click="run(() => onDelete(meal))"
                    >
                        <Trash2 class="size-5" />
                    </IconButton>
                </template>
                <template v-else>
                    <IconButton
                        label="Edit and log meal"
                        :disabled="busy"
                        @click="onLogEdited(meal, EDIT_LOGGED_TYPE)"
                    >
                        <SquarePen class="size-5" />
                    </IconButton>
                    <Button
                        variant="primary"
                        :loading="busy"
                        @click="run(() => onLogPlanned(meal))"
                    >
                        <template #icon-left><Check class="size-4" /></template>
                        Log
                    </Button>
                </template>
            </div>
        </div>
        <SimpleMacros
            class="px-4 pb-3"
            :calories="mealMacroTotals.calories"
            :protein="mealMacroTotals.protein"
            :fat="mealMacroTotals.fat"
            :carbs="mealMacroTotals.carbs"
            :fiber="mealMacroTotals.fiber"
        />
        <List class="border-t border-line">
            <ListRow v-if="isQuickLog">Quick Log Entry</ListRow>
            <template v-else>
                <template v-for="(block, bi) in mealBlocks" :key="'mb-' + bi">
                    <template v-if="block.kind === 'ungrouped'">
                        <ListRow
                            v-for="{ item: food, index: i } in block.rows"
                            :key="'u-' + i"
                        >
                            {{ itemLabel(food) }}
                            <template #subtitle>
                                <MealCardMacroDetails
                                    :calories="food.food?.calories ?? 0"
                                    :protein="food.food?.protein ?? 0"
                                    :carbs="food.food?.carbs ?? 0"
                                    :fat="food.food?.fat ?? 0"
                                    :fiber="food.food?.fiber ?? 0"
                                    :amount="Number(food.amount)"
                                />
                            </template>
                        </ListRow>
                    </template>
                    <template v-else>
                        <ListRow
                            :aria-expanded="isGroupExpanded(block.groupId)"
                            @click="toggleGroupCollapse(block.groupId)"
                        >
                            <template #leading>
                                <ChevronDown
                                    v-if="isGroupExpanded(block.groupId)"
                                    class="size-4"
                                />
                                <ChevronRight v-else class="size-4" />
                            </template>
                            <span class="font-semibold">{{
                                block.label || "Group"
                            }}</span>
                            <template #subtitle>
                                <MealCardMacroDetails
                                    v-bind="groupMacroTotals(block.rows)"
                                    :amount="1"
                                />
                            </template>
                        </ListRow>
                        <template v-if="isGroupExpanded(block.groupId)">
                            <ListRow
                                v-for="{ item: food, index: i } in block.rows"
                                :key="'g-' + i"
                            >
                                <template #leading
                                    ><span class="block w-6"></span
                                ></template>
                                {{ itemLabel(food) }}
                                <template #subtitle>
                                    <MealCardMacroDetails
                                        :calories="food.food?.calories ?? 0"
                                        :protein="food.food?.protein ?? 0"
                                        :carbs="food.food?.carbs ?? 0"
                                        :fat="food.food?.fat ?? 0"
                                        :fiber="food.food?.fiber ?? 0"
                                        :amount="Number(food.amount)"
                                    />
                                </template>
                            </ListRow>
                        </template>
                    </template>
                </template>
            </template>
        </List>
    </Card>
</template>
