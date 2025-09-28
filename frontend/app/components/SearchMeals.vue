<script lang="ts" setup>
import type { Meal, MealItem } from "~/types/models";

const props = defineProps<{
    ID: number;
    name: string;
    items: Partial<MealItem>[];
}>();

const emit = defineEmits<{
    (e: "update:ID", value: number): void;
    (e: "update:name", value: string): void;
    (e: "update:items", value: Partial<MealItem>[]): void;
    (e: "createMeal", value: string): void;
}>();

let currentSearch = ref("");
let showDropdown = ref(false);

const {
    data: meals,
    pending,
    error,
} = useApiFetch<Meal[]>("mealplan/meal/all");

const filteredMeals = computed(() => {
    if (!meals.value) return [];
    if (!currentSearch.value) return meals.value;
    const search = currentSearch.value.toLowerCase();
    return meals.value.filter((m) => m.name.toLowerCase().includes(search));
});

function createMeal(name: string) {
    emit("createMeal", name);
    showDropdown.value = false;
}

function submitMeal(meal: Meal) {
    emit("update:ID", meal.ID);
    emit("update:name", meal.name);
    emit("update:items", meal.items);
    showDropdown.value = false;
}
</script>
<template>
    <div>
        <input
            type="text"
            placeholder="Search meals..."
            :value="props.name"
            @input="
                (e) => {
                    emit('update:name', (e.target as HTMLInputElement).value);
                    currentSearch = (e.target as HTMLInputElement).value;
                    showDropdown = true;
                }
            "
        />
        <div v-if="pending">Loading Meals...</div>
        <div v-else-if="error">Error: {{ error.message }}</div>
        <div v-else-if="showDropdown" class="dropdown">
            <div
                v-if="showDropdown && filteredMeals.length === 0"
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
