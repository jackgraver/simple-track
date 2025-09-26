<script lang="ts" setup>
import { reactive, ref, watch } from "vue";
import type { Meal, MealItem, Food } from "~/types/models";

const {
    data: foodRes,
    pending,
    error,
} = useFetch<Food[]>("mealplan/food/all");

const foodOptions = foodRes && foodRes.value ? foodRes.value : [];
console.log("Food options:", foodOptions);

const currentMeal = ref<Meal | null>(null);

// Reactive form state
const form = reactive({
    ID: -1,
    name: "",
    items: [] as Partial<MealItem>[],
});

// When a preset meal is selected, fill the form
watch(currentMeal, (meal) => {
    if (meal) {
        form.ID = meal.ID;
        form.name = meal.name;
        form.items = meal.items.map((i) => ({
            ...i,
            food: i.food ? { ...i.food } : undefined,
        }));
    }
});

function newMealName(name: string) {
    form.name = name;
    form.items = [
        {
            ID: -1,
            meal_id: currentMeal.value?.ID ?? -1,
            food_id: -1,
            amount: 0,
            food: undefined,
        },
    ];
}

// Add a new empty food item
function addFoodItem() {
    form.items.push({
        ID: -1,
        meal_id: currentMeal.value?.ID ?? -1,
        food_id: -1,
        amount: 0,
        food: undefined,
    });
}

// Remove a food item
function removeFoodItem(index: number) {
    form.items.splice(index, 1);
}

function onFoodSelect(item: MealItem) {
    item.food = foodOptions.find((f) => f.ID === item.food_id);
    if (currentMeal && item.food_id !== -1) {
        const mealItem = currentMeal.value?.items.find(
            (mi) => mi.ID === item.ID,
        );
        if (mealItem && mealItem.food_id === -1) {
            mealItem.food_id = item.food_id!;
            mealItem.food = item.food;
        }
    }
}

// Submit the meal
function onSubmit(e: Event) {
    e.preventDefault();
    const newMeal: Meal = {
        ID: currentMeal.value?.ID ?? -1,
        name: form.name,
        items: form.items.map((i) => ({
            ID: i.ID!,
            meal_id: i.meal_id!,
            food_id: i.food_id!,
            amount: i.amount!,
            food: i.food,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        })),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
    };
    console.log("Submitting meal:", newMeal);
    // send to backend here
}
</script>

<template>
    <h1>Log Food</h1>
    <ExpectedMeals @select-meal="(meal: Meal) => (currentMeal = meal)" />
    <form @submit="onSubmit">
        <SearchMeals
            @select-meal="(meal: Meal) => (currentMeal = meal)"
            @create-meal="newMealName"
        />
        <!-- <input
            type="text"
            placeholder="Meal name..."
            v-model="form.name"
            style="margin-bottom: 8px"
        /> -->
        <div
            v-for="(item, index) in form.items"
            :key="item.ID"
            style="margin-bottom: 8px"
        >
            <select
                v-model="item.food_id"
                @change="onFoodSelect(item as MealItem)"
            >
                <option value="-1">Select food</option>
                <option v-for="f in foodOptions" :key="f.ID" :value="f.ID">
                    {{ f.name }}
                </option>
            </select>
            <input type="number" v-model.number="item.amount" min="1" />
            <button
                type="button"
                @click="removeFoodItem(index)"
                class="delete-button"
            >
                X
            </button>
        </div>
        <button type="button" @click="addFoodItem">Add Food Item</button>
        <button type="submit" :disabled="!currentMeal?.name || currentMeal?.items.length === 0">Submit Meal</button>
    </form>
</template>

<style scoped>
form {
    width: 50%;
    display: flex;
    flex-direction: column;
}

input,
select {
    margin-right: 4px;
    margin-bottom: 4px;
}

button {
    margin-top: 4px;
}

button.delete-button {
    background: rgb(255, 20, 20);
    color: white;
}
button.delete-button:hover {
    transform: scale(1.05);
}
</style>
