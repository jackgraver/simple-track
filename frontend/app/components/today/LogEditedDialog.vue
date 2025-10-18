<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";

//TODO: better interaction between dialog and close dialog
const props = defineProps<{
    meal: Meal;
    onResolve: (meal: Meal) => void;
}>();

const meal = props.meal;
meal.ID = 0;

const logEditedMeal = async () => {
    props.onResolve(meal);
};
</script>

<template>
    <div class="container">
        <input type="text" id="name" v-model="meal.name" />

        <div class="items-grid">
            <div v-for="item in meal.items" :key="item.ID" class="item-row">
                <span>{{ item.food?.name }}</span>
                <input type="number" v-model="item.amount" />
            </div>
        </div>

        <button class="confirm-button" @click="logEditedMeal">
            Log Edited
        </button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 1rem;
    color: #fff;
}

.items-grid {
    display: grid;
    grid-template-columns: 1fr 100px; /* left = name, right = input */
    gap: 0.4rem 1rem; /* row gap, column gap */
    align-items: center;
}

.item-row {
    display: contents; /* keep grid alignment without extra nesting */
    text-align: left;
}

input {
    background-color: rgb(50, 50, 50);
    color: white;
    border: none;
    padding: 0.5rem;
}
</style>
