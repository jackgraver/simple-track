<script lang="ts" setup>
import { reactive } from "vue";
import type { Meal, MealItem, Food } from "~/types/models";

// Props: optional meal to prefill
const props = defineProps<{ meal?: Meal }>();

// Sample food list
const foodOptions: Food[] = [
    {
        ID: 1,
        name: "Apple",
        unit: "piece",
        calories: 95,
        protein: 0.5,
        fiber: 4,
        created_at: "",
        updated_at: "",
    },
    {
        ID: 2,
        name: "Banana",
        unit: "piece",
        calories: 105,
        protein: 1.3,
        fiber: 3,
        created_at: "",
        updated_at: "",
    },
];

// Reactive form state
const form = reactive({
    name: props.meal?.name || "",
    items:
        props.meal?.items.map((i) => ({
            ...i,
            food: i.food ? { ...i.food } : undefined,
        })) || ([] as MealItem[]),
    selectedFoodId: 0,
    amount: 1,
});

// Add a new food item to the meal
function addFoodItem() {
    if (!form.selectedFoodId || form.amount <= 0) return;
    const food = foodOptions.find((f) => f.ID === form.selectedFoodId);
    if (!food) return;

    const newItem: MealItem = {
        ID: Date.now(), // temp unique ID for client side
        meal_id: 0, // will be assigned by DB
        food_id: food.ID,
        food: food,
        amount: form.amount,
        created_at: "",
        updated_at: "",
    };
    form.items.push(newItem);
    form.selectedFoodId = 0;
    form.amount = 1;
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
    <form @submit="onSubmit">
        <label for="mealName">Meal Name</label>
        <input
            id="mealName"
            type="text"
            v-model="form.name"
            placeholder="Enter meal name"
        />

        <div>
            <select v-model="form.selectedFoodId">
                <option value="0">Select food</option>
                <option
                    v-for="food in foodOptions"
                    :key="food.ID"
                    :value="food.ID"
                >
                    {{ food.name }}
                </option>
            </select>
            <input type="number" v-model.number="form.amount" min="1" />
            <button type="button" @click="addFoodItem">Add Food Item</button>
        </div>

        <div
            v-for="(item, index) in form.items"
            :key="item.ID"
            style="margin: 4px 0"
        >
            <span>{{ item.food?.name }} ({{ item.amount }})</span>
            <button type="button" @click="removeFoodItem(index)">X</button>
        </div>

        <button type="submit">Submit Meal</button>
    </form>
</template>

<style scoped>
form {
    width: 50%;
    display: flex;
    flex-direction: column;
}

input,
select,
button {
    background: rgb(71, 71, 71);
    color: white;
    border: none;
    padding: 4px;
    margin: 4px 0;
}
</style>
