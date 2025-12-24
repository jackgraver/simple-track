<script setup lang="ts">
import type { Day, Food, Meal, MealItem } from "~/types/diet";
import { Check, Plus, Trash2, ChevronUp, ChevronDown } from "lucide-vue-next";
import { ref } from "vue";
//TODO: calculate amount based on inputting "xg" or "1.5 servings"

//TODO: better interaction between dialog and close dialog
const props = defineProps<{
    meal: Meal;
    onResolve: (meal: Meal) => void;
}>();

const addNew = ref(false);
const foodsReady = ref(false);

const meal = props.meal;
meal.ID = 0;

const logEditedMeal = async () => {
    props.onResolve(meal);
};

function addFoodItem(food: Food) {
    addNew.value = false;
    props.meal.items.push({
        meal_id: meal.ID,
        food_id: food.ID,
        food: food,
        amount: 1,
    } as MealItem);
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
                <button @click="item.amount--"><ChevronDown /></button>
                <input class="small-input" type="text" v-model="item.amount" />
                <span v-if="item.food!.serving_type === `g`">g</span>
                <span>{{ item.food?.name }}</span>
                <button @click="item.amount++"><ChevronUp /></button>
                <span
                    >{{ item.amount * item.food!.calories }}C /
                    {{ item.amount * item.food!.protein }}P /
                    {{ item.amount * item.food!.fiber }}F</span
                >

                <button class="delete-button" @click="removeFoodItem(i)">
                    <Trash2 :size="20" />
                </button>
            </div>
        </div>
        <SearchFoods
            v-show="addNew"
            :on-select="addFoodItem"
            @loaded="foodsReady = true"
        />
        <button :disabled="!foodsReady" @click="addNew = true">
            <Plus :size="20" />
        </button>
        <button class="confirm-button" @click="logEditedMeal">
            <Check :size="20" />
        </button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    color: #fff;
}

.items-grid {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
}

.item-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.right {
    display: flex;
    flex-direction: row;
    padding-left: 12px;
}

.small-input {
    width: 15px;
}
</style>
