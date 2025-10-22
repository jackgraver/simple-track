<script setup lang="ts">
import { useRoute } from "vue-router";
import type { Food, Meal, MealItem } from "~/types/diet";
import SearchList from "~/components/SearchList.vue";
import { Check, Plus, Trash2, ChevronUp, ChevronDown } from "lucide-vue-next";
import FoodDisplay from "~/components/FoodDisplay.vue";

const route = useRoute();
const id = route.params.id as string | undefined;

// start with a default empty meal
const meal = ref<Meal>({
    ID: 0,
    created_at: "",
    updated_at: "",
    name: "",
    items: [],
});

if (id) {
    const { data, pending, error } = await useAPIGet<{ meal: Meal }>(
        `mealplan/meal/${id}`,
    );
    if (data.value) {
        meal.value = data.value.meal; // replace default with fetched
    } else {
        console.warn("No meal found for ID:", id);
    }
}

const addFood = async (food: Food) => {
    meal.value.items.push({
        meal_id: meal.value.ID,
        food_id: food.ID,
        food: food,
        amount: 1,
    } as MealItem);
};

const setMeal = async (newMeal: Meal) => {
    meal.value = newMeal;
};
</script>

<template>
    <div class="page-wrapper">
        <div v-if="meal" class="container">
            <div class="cell left">
                <h1>Log Meal</h1>
                <input type="text" id="name" v-model="meal.name" />
                <div class="meal-items">
                    <div
                        v-for="(item, i) in meal.items"
                        :key="item.ID"
                        class="items-rows"
                    >
                        <button @click="item.amount--"><ChevronDown /></button>
                        <input
                            class="small-input"
                            type="text"
                            v-model="item.amount"
                        />
                        <span v-if="item.food!.unit === 'g'">g</span>
                        <span>{{ item.food?.name }}</span>
                        <button @click="item.amount++"><ChevronUp /></button>
                        <span>
                            {{ item.amount * item.food!.calories }}C /
                            {{ item.amount * item.food!.protein }}P /
                            {{ item.amount * item.food!.fiber }}F
                        </span>
                        <button class="delete-button">
                            <Trash2 :size="20" />
                        </button>
                    </div>
                </div>
                <button>Log Meal</button>
            </div>
            <div class="cell right-top">
                <h2>Add Foods</h2>
                <SearchList
                    :route="'mealplan/food/all'"
                    :onSelect="addFood"
                    :displayComponent="FoodDisplay"
                    :prefilter="meal?.items.map((i) => i.ID)"
                />
            </div>
            <div class="cell right-bottom">
                <h2>Select Saved Meal</h2>
                <SearchList
                    :key="meal.ID"
                    :route="'mealplan/meal/all'"
                    :on-select="setMeal"
                    :prefilter="[meal.ID]"
                />
            </div>
        </div>
    </div>
</template>

<style scoped>
.page-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
    padding-block: 2rem;
    box-sizing: border-box;
    background: rgb(26, 26, 26);
    overflow: hidden;
}

.container {
    display: grid;
    grid-template-columns: 2fr 1fr;
    grid-template-rows: 1fr 1fr;
    gap: 1rem;
    width: 80%;
    height: 100%;
    max-height: 100%;
}

.cell {
    border: 1px solid rgb(82, 82, 82);
    border-radius: 5px;
    overflow: auto;
    min-height: 0;
    background: rgb(40, 40, 40);
}

.cell > h1,
h2 {
    margin-top: 0;
}

.left {
    grid-row: 1 / span 2;
    padding: 1rem;
}

.right-top,
.right-bottom {
    padding: 1rem;
}

.meal-items {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
}

.items-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.small-input {
    width: 16px;
    text-align: center;
}
</style>
