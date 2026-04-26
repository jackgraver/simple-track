<script setup lang="ts">
import { ref } from "vue";
import type { Food } from "~/types/diet";
import SearchList from "~/shared/SearchList.vue";
import FoodDisplay from "~/pages/diet/logmeal/components/FoodDisplay.vue";
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

const relatedFood = ref<Food | null>(null);

const createFoodMutation = useCreateFood();

function clearRelated() {
    relatedFood.value = null;
}

async function onPickRelated(
    row: Food & { entry_kind?: string },
): Promise<boolean> {
    if (row.entry_kind === "composite") return false;
    relatedFood.value = row as Food;
    return true;
}

const createFood = async () => {
    if (meal.value.serving_type === "Grams") meal.value.serving_type = "g";
    if (meal.value.serving_type === "Unit") meal.value.serving_type = "";

    try {
        const response = await createFoodMutation.mutateAsync({
            food: meal.value,
            relatedFoodId: relatedFood.value?.ID,
        });
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
        <div class="field related-field">
            <label>Related to existing food (optional)</label>
            <div
                v-if="relatedFood"
                class="related-chip"
            >
                <span class="related-name">{{ relatedFood.name }}</span>
                <button type="button" class="related-clear" @click="clearRelated">
                    Clear
                </button>
            </div>
            <div class="related-search">
                <SearchList
                    route="diet/meals/food/all"
                    :on-select="onPickRelated"
                    :display-component="FoodDisplay"
                />
            </div>
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

.related-field {
    gap: 0.5rem;
}

.related-chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    padding: 0.45rem 0.6rem;
    border-radius: 4px;
    background: rgb(55, 55, 55);
}

.related-name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.9rem;
}

.related-clear {
    flex-shrink: 0;
    cursor: pointer;
    border: none;
    border-radius: 4px;
    padding: 0.25rem 0.5rem;
    background: rgb(70, 70, 70);
    color: white;
    font-size: 0.8rem;
}

.related-search {
    max-height: 14rem;
    min-height: 6rem;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: 4px;
    border: 1px solid rgb(60, 60, 60);
}
</style>
