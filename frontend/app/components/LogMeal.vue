<script lang="ts" setup>
import { reactive, ref, watch } from "vue";
import type { Meal, MealItem, Food } from "~/types/models";

const {
    data: foodRes,
    pending,
    error,
} = useApiFetch<Food[]>("mealplan/food/all");

const foodOptions = foodRes && foodRes.value ? foodRes.value : [];

const form = reactive({
    ID: 0,
    name: "",
    items: [] as Partial<MealItem>[],
});

function newMeal(name: string) {
    form.name = name;
    form.items = [
        {
            ID: 0,
            meal_id: 0,
            food_id: 0,
            amount: 0,
            food: undefined,
        },
    ];
}

// Add a new empty food item
function addFoodItem() {
    form.items.push({
        ID: 0,
        meal_id: form.ID ?? 0,
        food_id: 0,
        amount: 0,
        food: undefined,
    });
    form.ID = 0;
    if (!form.name.endsWith(" Modified")) {
        form.name = form.name + " Modified";
    }
}

// Remove a food item
function removeFoodItem(index: number) {
    form.items.splice(index, 1);
}

function onFoodSelect(item: MealItem) {
    const food = foodOptions.find((f) => f.ID === item.food_id);
    item.food = food ? { ...food } : undefined;
}

// Submit the meal
function onSubmit(e: Event) {
    e.preventDefault();
    const payload = {
        meal_id: form.ID ?? 0,
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
    };
    const { data, error } = useApiFetch<Meal>("mealplan/meal/log", {
        method: "POST",
        body: payload,
    });
    if (error) {
        console.log("error", error);
    }
    if (data) {
        console.log("data", data);
    }
}

function checkEmpty() {
    console.log("check", form);
    if (form.name) {
        if (form.items.length > 0 && form.items[-1]?.food_id !== 0) {
            return true;
        }
    }
    return false;
}
</script>

<template>
    <h1>Log Food</h1>
    <ExpectedMeals
        v-model:ID="form.ID"
        v-model:name="form.name"
        v-model:items="form.items"
        :confirmMeal="
            (meal) => {
                console.log('confirming meal', meal);
            }
        "
    />
    <form @submit="onSubmit">
        <SearchMeals
            v-model:ID="form.ID"
            v-model:name="form.name"
            v-model:items="form.items"
            @update:model-value="form.name = $event"
            @select-meal="(meal: Meal) => console.log('huh?', meal)"
            @create-meal="newMeal"
        />
        <div v-for="(item, index) in form.items" :key="item.ID">
            <select
                v-model="item.food_id"
                @change="onFoodSelect(item as MealItem)"
            >
                <option value="0">Select food</option>
                <option v-if="pending" value="0">Loading...</option>
                <option v-else-if="error" value="0">
                    Error: {{ error.message }}
                </option>
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
        <button type="button" @click="addFoodItem" v-if="checkEmpty()">
            Add Food Item
        </button>
        <button type="submit" v-if="checkEmpty()">Submit Meal</button>
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
