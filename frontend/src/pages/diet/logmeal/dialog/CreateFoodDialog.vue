<script setup lang="ts">
import { ref } from "vue";
import type { Food } from "~/types/diet";
import { useCreateFood } from "../queries/useFoodMutations";

const props = defineProps<{
    foodName?: string;
    onResolve: (food: Food) => void;
    onCancel?: () => void;
}>();

const meal = ref<Food>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: props.foodName ?? "",
    calories: 0,
    protein: 0,
    fiber: 0,
    carbs: 0,
    serving_type: "",
    serving_amount: 0,
});

const createFoodMutation = useCreateFood();

const createFood = async () => {
    if (meal.value.serving_type === "Grams") meal.value.serving_type = "g";
    if (meal.value.serving_type === "Unit") meal.value.serving_type = "";

    try {
        const response = await createFoodMutation.mutateAsync(meal.value);
        if (response?.food) {
            props.onResolve(response.food);
        }
    } catch (error) {
        console.error("Failed to create food:", error);
    }
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
            <div class="field">
                <label for="carbs">Carbs</label>
                <input type="number" id="carbs" min="0" v-model="meal.carbs" />
            </div>
        </div>
        <div class="field">
            <label for="unit">Serving Type</label>
            <select id="unit" v-model="meal.serving_type">
                <option value="none" selected disabled hidden>
                    Select Serving Type
                </option>
                <option value="Unit">Unit</option>
                <option value="Grams">Grams</option>
            </select>
            <label
                v-if="
                    meal.serving_type === 'Unit' ||
                    meal.serving_type === 'Grams'
                "
                >Serving Amount</label
            >
            <input
                type="number"
                v-if="meal.serving_type === 'Unit'"
                v-model="meal.serving_amount"
            />
            <input
                type="number"
                v-if="meal.serving_type === 'Grams'"
                v-model="meal.serving_amount"
            />
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
    padding: 1.5rem 0.5rem;
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
