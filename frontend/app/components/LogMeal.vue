<script lang="ts" setup>
import { reactive } from "vue";
import type { Meal, MealItem, Food, MealPlanDay } from "~/types/models";

// Props: optional meal to prefill
const props = defineProps<{ meal?: Meal }>();

const { data: foodRes, pending, error } = useFetch<Food[]>("http://localhost:8080/mealplan/food/all");

const foodOptions = foodRes?.value || [];

// Reactive form state
const form = reactive({
    name: "",
    items: [] as MealItem[],
});

// Add a new food item to the meal
function addFoodItem() {
    const newItem: MealItem = {
        ID: Date.now(),
        meal_id: 0,
        food_id: -1,
        amount: -1,
        created_at: "",
        updated_at: "",
    };
    form.items.push(newItem);
}

// Remove food item
function removeFoodItem(index: number) {
    form.items.splice(index, 1);
}

// Submit the meal
function onSubmit(e: Event) {
    e.preventDefault();
    const newMeal: Meal = {
        ID: Date.now(), // temp ID
        name: form.name,
        items: form.items.map((i) => ({
            ID: i.ID!,
            meal_id: i.meal_id!,
            food_id: i.food_id!,
            amount: i.amount!,
            food: i.food, // can send optional nested Food if backend supports
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        })),
        created_at: "",
        updated_at: "",
    };
    console.log("Submitting meal:", newMeal);
    // send to backend here
}
</script>

<template>
    <h1>Log Food</h1>
    <ExpectedMeals @set-meal="(meal: MealPlanDay) => console.log(meal)"/>
    <form @submit="onSubmit">
        <SearchMeals />
        <div
            v-for="(item, index) in form.items"
            :key="item.ID"
            style="margin: 4px 0"
        >
            <span>{{ item.food ? item.food.name : "No food" }}</span>
            <!-- <select value={{ item.food.name }} selected>
                <option v-for="value in foodOptions">{{ value.name }}</option>
            </select> -->
             <input type="number" v-model.number="item.amount" min="1" />
            <!--<button type="button" class="delete-button" @click="removeFoodItem(index)">X</button> -->
        </div>
        <button type="button" @click="addFoodItem">Add Food Item</button>
        <button type="submit">Submit Meal</button>
    </form>
</template>

<style scoped>
form {
    width: 50%;
    display: flex;
    flex-direction: column;
}

.delete-button {
    background: rgb(255, 20, 20)
}
</style>
