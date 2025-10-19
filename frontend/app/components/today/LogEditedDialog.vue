<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
//TODO: calculate amount based on inputting "xg" or "1.5 servings"

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

function addFoodItem() {
    //TODO: Need to think about how to handle foods that are not in the food list
}

function removeFoodItem(index: number) {
    meal.items.splice(index, 1);
}
</script>

<template>
    <div class="container">
        <input type="text" id="name" v-model="meal.name" />

        <div class="items-grid">
            <div
                v-for="(item, i) in meal.items"
                :key="item.ID"
                class="item-row"
            >
                <span>{{ item.food?.name }}</span>
                <div class="right">
                    <input type="number" v-model="item.amount" />
                    <button class="delete-button" @click="removeFoodItem(i)">
                        X
                    </button>
                </div>
            </div>
        </div>
        <button @click="addFoodItem">Add Food Item</button>
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

.right {
    display: flex;
    flex-direction: row;
}

input {
    background-color: rgb(50, 50, 50);
    color: white;
    border: none;
    padding: 0.5rem;
}
</style>
