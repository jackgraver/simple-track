<script setup lang="ts">
import type { Meal } from "~/types/diet";

function formatNum(n: number) {
    const s = n.toFixed(2); // always 2 decimals
    return s.replace(/\.?0+$/, ""); // drop trailing zeros and optional dot
}

function mealMacros(meal: Meal) {
    let calories = 0,
        protein = 0,
        fiber = 0;
    for (const item of meal.items) {
        calories += (item.food?.calories ?? 0) * item.amount;
        protein += (item.food?.protein ?? 0) * item.amount;
        fiber += (item.food?.fiber ?? 0) * item.amount;
    }

    return h("span", { class: "details" }, [
        h("span", { class: "cal" }, `${formatNum(calories)}C`),
        " / ",
        h("span", { class: "pro" }, `${formatNum(protein)}P`),
        " / ",
        h("span", { class: "fib" }, `${formatNum(fiber)}F`),
    ]);
}

defineProps<{
    meal: Meal;
    type: "planned" | "logged";
    onLogPlanned: (meal: Meal) => void;
    onLogEdited: (meal: Meal) => void;
    onDelete: (meal: Meal) => void;
    onEdit: (meal: Meal) => void;
}>();
</script>

<template>
    <div class="card">
        <h3>{{ meal.name }} <component :is="mealMacros(meal)" /></h3>
        <div class="meal">
            <div class="left">
                <div class="foods">
                    <span
                        v-for="food in meal.items"
                        :key="food.ID"
                        class="food"
                    >
                        <span
                            >{{ food.amount }}{{ food.food?.unit }}
                            {{
                                `${food.food?.name}${food.amount > 1 ? "s" : ""}`
                            }}</span
                        >
                        <span class="details">
                            <span class="cal"
                                >{{
                                    food.food?.calories ?? 0 * food.amount
                                }}C</span
                            >
                            /
                            <span class="pro"
                                >{{
                                    food.food?.protein ?? 0 * food.amount
                                }}P</span
                            >
                            /
                            <span class="fib"
                                >{{
                                    food.food?.fiber ?? 0 * food.amount
                                }}F</span
                            >
                        </span>
                    </span>
                </div>
            </div>
            <div class="right">
                <div class="actions" v-if="type === 'logged'">
                    <button class="delete-button" @click="onDelete(meal)">
                        Delete
                    </button>
                    <button @click="onEdit(meal)">Edit</button>
                </div>

                <div class="actions" v-else-if="type === 'planned'">
                    <button @click="onLogEdited(meal)">Log Edited</button>
                    <button class="confirm-button" @click="onLogPlanned(meal)">
                        Log
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
    width: 500px;
    /* gap: 1rem; */
}

.meal {
    display: flex;
    flex-direction: row;
}

.card h3 {
    margin-top: 0;
    margin-bottom: 0.5rem;
    width: 100%;
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
    justify-content: flex-end; /* push actions to bottom */
    align-items: flex-end; /* align to right edge */
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
    color: #f87171;
}
.details .pro {
    color: #60a5fa;
}
.details .fib {
    color: #34d399;
}

.actions {
    display: flex;
}

button {
    white-space: nowrap;
}
</style>
