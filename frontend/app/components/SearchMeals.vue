<script lang="ts" setup>
import type { Meal } from "~/types/models";

const {
    data: meals,
    pending,
    error,
} = useApiFetch<Meal[]>("mealplan/meal/all");

let currentSearch = ref("");

const filteredMeals = computed(() => {
    if (!meals.value) return [];
    if (!currentSearch.value) return meals.value;
    const search = currentSearch.value.toLowerCase();
    return meals.value.filter((m) => m.name.toLowerCase().includes(search));
});

const emit = defineEmits<{
    (e: "selectMeal", meal: Meal): void;
    (e: "createMeal", value: string): void;
}>();

function createMeal(name: string) {
    emit("createMeal", name);
    currentSearch.value = "";
}

function submitMeal(meal: Meal) {
    emit("selectMeal", meal);
    currentSearch.value = "";
}
</script>
<template>
    <div>
        <input
            type="text"
            placeholder="Search meals..."
            @input="
                (e) => (currentSearch = (e.target as HTMLInputElement).value)
            "
        />
        <div v-if="pending">Loading Meals...</div>
        <div v-else-if="error">Error: {{ error.message }}</div>
        <div v-else-if="currentSearch" class="dropdown">
            <div
                v-if="currentSearch && filteredMeals.length === 0"
                @click="createMeal(currentSearch)"
            >
                + Click to create new saved meal "{{ currentSearch }}"
            </div>
            <div
                v-for="meal in filteredMeals"
                :key="meal.ID"
                @click="submitMeal(meal)"
                class="dropdown-item"
            >
                {{ meal.name }}
            </div>
        </div>
    </div>
</template>

<style scoped>
.dropdown {
    position: absolute;
    z-index: 1000;
    background-color: #3b3b3b;
    border: 1px solid #ccc;
    border-radius: 4px;
    padding: 10px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.dropdown div {
    padding: 2px 4px;
    cursor: pointer;
}

.dropdown div:hover {
    background-color: #7e7e7e;
}
</style>
