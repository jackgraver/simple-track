<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
import { Check, Plus, Trash2 } from "lucide-vue-next";
//TODO: calculate amount based on inputting "xg" or "1.5 servings"

//TODO: better interaction between dialog and close dialog
const props = defineProps<{
    meal: Meal;
    onResolve: (meal: Meal) => void;
}>();

const addNew = ref(false);

const meal = props.meal;
meal.ID = 0;

const logEditedMeal = async () => {
    props.onResolve(meal);
};

function addFoodItem() {
    addNew.value = true;
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
                        <Trash2 :size="20" />
                    </button>
                </div>
            </div>
        </div>
        <input v-if="addNew" type="text" placeholder="Food Name" />
        <button @click="addFoodItem"><Plus :size="20" /></button>
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
/* 
input {
    background-color: rgb(50, 50, 50);
    color: white;
    border: none;
    padding: 0.5rem;
} */
</style>
