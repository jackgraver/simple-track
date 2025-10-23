<script setup lang="ts">
import type { Food } from "~/types/diet";

const props = defineProps<{
    foodName?: string;
    onResolve: (food: Food | null) => void;
}>();

const meal = ref<Food>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: props.foodName ?? "",
    calories: 0,
    protein: 0,
    fiber: 0,
    unit: "",
});

const createFood = async () => {
    console.log(meal.value);
    if (meal.value.unit === "Grams") meal.value.unit = "g";
    if (meal.value.unit === "Unit") meal.value.unit = "";

    const { response } = await useAPIPost<{ food: Food }>(
        `mealplan/food/add`,
        "POST",
        {
            food: meal.value,
        },
    );
    props.onResolve(response?.food ?? null);
};
</script>

<template>
    <div class="food-form">
        <div class="field">
            <label for="name">Food Name</label>
            <input type="text" id="name" v-model="meal.name" />
        </div>
        <div class="macros-row">
            <div class="field">
                <label for="calories">Calories</label>
                <input
                    type="number"
                    id="calories"
                    min="0"
                    v-model="meal.calories"
                />
            </div>
            <div class="field">
                <label for="protein">Protein</label>
                <input
                    type="number"
                    id="protein"
                    min="0"
                    v-model="meal.protein"
                />
            </div>
            <div class="field">
                <label for="fiber">Fiber</label>
                <input type="number" id="fiber" min="0" v-model="meal.fiber" />
            </div>
        </div>
        <div class="field">
            <label for="unit">Unit</label>
            <select id="unit" v-model="meal.unit">
                <option value="Grams">Grams</option>
                <option value="Unit">Unit</option>
            </select>
        </div>
        <button @click="createFood">Create</button>
    </div>
</template>

<style scoped>
.food-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    color: white;
}

.field {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.macros-row {
    display: flex;
    gap: 1rem;
}

input {
    background-color: rgb(50, 50, 50);
    color: white;
    border: none;
    border-radius: 4px;
    padding: 0.5rem;
    width: 100%;
    box-sizing: border-box;
}

label {
    font-size: 0.9rem;
    color: #ccc;
}
</style>
