<script setup lang="ts">
import type { Meal } from "~/types/diet";
import { EDIT_LOGGED_TYPE, EDIT_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { Trash2, SquarePen, Check } from "lucide-vue-next";
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { computed } from "vue";

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
    let fiber = 0;
    for (const item of meal.items) {
        const a = Number(item.amount);
        calories += (item.food?.calories ?? 0) * a;
        protein += (item.food?.protein ?? 0) * a;
        fiber += (item.food?.fiber ?? 0) * a;
    }
    return { calories, protein, fiber };
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
</script>

<template>
    <div class="card">
        <h3 class="meal-title">
            {{ meal.name }}
            <SimpleMacros
                class="title-macros"
                :calories="mealMacroTotals.calories"
                :protein="mealMacroTotals.protein"
                :fiber="mealMacroTotals.fiber"
                font-size="0.9rem"
            />
        </h3>
        <div class="meal">
            <div class="left">
                <div class="foods">
                    <span
                        v-for="(food, i) in meal.items"
                        :key="`${food.food_id}-${i}`"
                        class="food"
                    >
                        <span
                            >({{ formatNum(itemServingAmount(food))
                            }}{{
                                food.food?.serving_type === "g" ? "g" : ""
                            }}) {{ food.food?.name
                            }}{{
                                Number(food.amount) > 1 ? "s" : ""
                            }}</span
                        >
                        <span class="details">
                            <span class="cal"
                                >{{
                                    formatNum(
                                        (food.food?.calories ?? 0) *
                                            Number(food.amount),
                                    )
                                }}C</span
                            >
                            /
                            <span class="pro"
                                >{{
                                    formatNum(
                                        (food.food?.protein ?? 0) *
                                            Number(food.amount),
                                    )
                                }}P</span
                            >
                            /
                            <span class="fib"
                                >{{
                                    formatNum(
                                        (food.food?.fiber ?? 0) *
                                            Number(food.amount),
                                    )
                                }}F</span
                            >
                        </span>
                    </span>
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
}

.foods {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.right {
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: flex-end;
}

.food .details {
    opacity: 0;
    visibility: hidden;
    margin-left: 0.25rem;
    transition: visibility 0.3s ease;
    transition-delay: 0.5s;
}

.meal h3:hover ~ .food .details {
    opacity: 1;
    visibility: visible;
    transition-delay: 0s;
}

.food:hover .details {
    opacity: 1;
    visibility: visible;
    transition-delay: 0s;
}

.meal span {
    color: gray;
}

.meal span:hover {
    transition-delay: 0s;
}

.details .cal {
    color: orange;
}
.details .pro {
    color: #60a5fa;
}
.details .fib {
    color: green;
}

.actions {
    display: flex;
}

button {
    white-space: nowrap;
}
</style>
