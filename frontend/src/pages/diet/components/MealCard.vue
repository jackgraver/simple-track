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
    for (const item of meal.items) {
        const a = Number(item.amount);
        calories += (item.food?.calories ?? 0) * a;
        protein += (item.food?.protein ?? 0) * a;
        carbs += (item.food?.carbs ?? 0) * a;
        fat += (item.food?.fat ?? 0) * a;
    }
    return { calories, protein, carbs, fat };
}

const props = defineProps<{
    meal: Meal;
    type: "planned" | "logged";
    onLogPlanned: (meal: Meal) => void;
    onLogEdited: (
        meal: Meal,
        type: typeof EDIT_TYPE | typeof EDIT_LOGGED_TYPE,
    ) => void;
    onDelete: (meal: Meal) => void;
    onEdit: (meal: Meal) => void;
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
    };
}
</script>

<template>
    <div class="card">
        <h3 class="meal-title">
            {{ meal.name }}
            <SimpleMacros
                class="title-macros"
                :calories="mealMacroTotals.calories"
                :protein="mealMacroTotals.protein"
                :fat="mealMacroTotals.fat"
                :carbs="mealMacroTotals.carbs"
                font-size="0.9rem"
            />
        </h3>
        <div class="meal">
            <div class="left">
                <div class="foods">
                    <template
                        v-for="(block, bi) in mealBlocks"
                        :key="'mb-' + bi"
                    >
                        <template v-if="block.kind === 'ungrouped'">
                            <span
                                v-for="{ item: food, index: i } in block.rows"
                                :key="'u-' + i"
                                class="food"
                            >
                                <span class="food-line-label"
                                    >{{ formatNum(itemServingAmount(food))
                                    }}{{
                                        food.food?.serving_type === "g"
                                            ? "g"
                                            : ""
                                    }}
                                    {{ food.food?.name
                                    }}{{
                                        pluralize(
                                            food.food?.name,
                                            Number(food.amount),
                                        )
                                    }}</span
                                >
                                <MealCardMacroDetails
                                    class="macro-keep-size"
                                    :calories="food.food?.calories ?? 0"
                                    :protein="food.food?.protein ?? 0"
                                    :carbs="food.food?.carbs ?? 0"
                                    :fat="food.food?.fat ?? 0"
                                    :amount="Number(food.amount)"
                                />
                            </span>
                        </template>
                        <template v-else>
                            <div class="food-group">
                                <button
                                    type="button"
                                    class="group-header food"
                                    @click="toggleGroupCollapse(block.groupId)"
                                >
                                    <ChevronRight
                                        v-if="!isGroupExpanded(block.groupId)"
                                        :size="16"
                                        class="chev"
                                    />
                                    <ChevronDown
                                        v-else
                                        :size="16"
                                        class="chev"
                                    />
                                    <span class="group-title food-line-label">{{
                                        block.label || "Group"
                                    }}</span>
                                    <MealCardMacroDetails
                                        class="group-header-macros macro-keep-size"
                                        v-bind="groupMacroTotals(block.rows)"
                                        :amount="1"
                                    />
                                </button>
                                <div
                                    v-if="isGroupExpanded(block.groupId)"
                                    class="group-children"
                                >
                                    <span
                                        v-for="{
                                            item: food,
                                            index: i,
                                        } in block.rows"
                                        :key="'g-' + i"
                                        class="food food-child"
                                    >
                                        <span class="food-line-label"
                                            >{{
                                                formatNum(
                                                    itemServingAmount(food),
                                                )
                                            }}{{
                                                food.food?.serving_type === "g"
                                                    ? "g"
                                                    : ""
                                            }}
                                            {{ food.food?.name
                                            }}{{
                                                Number(food.amount) > 1
                                                    ? "s"
                                                    : ""
                                            }}</span
                                        >
                                        <MealCardMacroDetails
                                            class="macro-keep-size"
                                            :calories="food.food?.calories ?? 0"
                                            :protein="food.food?.protein ?? 0"
                                            :carbs="food.food?.carbs ?? 0"
                                            :fat="food.food?.fat ?? 0"
                                            :amount="Number(food.amount)"
                                        />
                                    </span>
                                </div>
                            </div>
                        </template>
                    </template>
                </div>
            </div>
            <div class="right">
                <div class="actions" v-if="type === 'logged'">
                    <button @click="onLogEdited(meal, EDIT_TYPE)">
                        <SquarePen :size="20" /></button
                    ><button class="delete-button" @click="onDelete(meal)">
                        <Trash2 :size="20" />
                    </button>
                </div>

                <div class="actions" v-else-if="type === 'planned'">
                    <button @click="onLogEdited(meal, EDIT_LOGGED_TYPE)">
                        <SquarePen :size="20" />
                    </button>
                    <button class="confirm-button" @click="onLogPlanned(meal)">
                        <Check :size="20" />
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.card {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 1rem;
    border: 1px solid #333;
    border-radius: 0.5rem;
    background: #1a1a1a;
    color: #fff;
}

.meal {
    display: flex;
    flex-direction: row;
}

.card h3,
.meal-title {
    margin-top: 0;
    margin-bottom: 0.5rem;
    width: 100%;
}

.meal-title {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.35rem 0.75rem;
}

.meal-title :deep(.macros) {
    margin-top: 0;
}

.left {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.foods {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.food {
    display: flex;
    flex-wrap: nowrap;
    align-items: baseline;
    gap: 0.25rem 0.5rem;
    min-width: 0;
}
.food-line-label {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.macro-keep-size {
    flex-shrink: 0;
}
.macro-keep-size :deep(.macros) {
    gap: clamp(0.25rem, 1.5vw, 0.45rem);
}
.macro-keep-size :deep(.macro) {
    font-size: clamp(0.62rem, 2.5vw, 0.85rem);
}

.food-group {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
}

.group-header {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    gap: 0.25rem 0.5rem;
    width: 100%;
    min-width: 0;
    margin: 0;
    padding: 0;
    cursor: pointer;
    text-align: left;
    color: inherit;
    font: inherit;
    box-shadow: none;
}

.group-header .chev {
    flex-shrink: 0;
    color: #888;
}

.group-title {
    font-weight: 600;
    color: #ccc;
}

.group-header-macros :deep(.macros) {
    margin-top: 0;
    gap: 0.45rem;
}

.group-children {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding-left: 1.25rem;
    border-left: 2px solid #333;
    margin-left: 0.35rem;
}

.food-child {
    font-size: 0.92em;
}

.right {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: flex-end;
}

.meal span {
    color: gray;
}

.meal span:hover {
    transition-delay: 0s;
}

.meal-title:hover ~ .meal :deep(.meal-card-macro-details),
.food:hover :deep(.meal-card-macro-details) {
    opacity: 1;
    visibility: visible;
    transition-delay: 0s;
}

.actions {
    display: flex;
}

button {
    white-space: nowrap;
}
</style>
