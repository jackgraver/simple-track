<script setup lang="ts">
import { ref, computed } from "vue";
import type { Food } from "~/types/diet";
import { Plus } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import CreateFoodDialog from "~/components/CreateFoodDialog.vue";

const props = defineProps<{ onSelect(food: Food): void }>();

const emit = defineEmits<{ (e: "loaded"): void }>();

const { data, pending, error } = await useAPIGet<{ foods: Food[] }>(
    "mealplan/food/all",
);

if (data.value) emit("loaded");

const search = ref("");
const dropdown = ref(false);

const filteredFoods = computed(() => {
    const foods = data?.value?.foods ?? [];
    const term = search.value.trim().toLowerCase();
    return term
        ? foods.filter((f) => f.name.toLowerCase().includes(term))
        : foods;
});

function itemSelected(food: Food) {
    dropdown.value = false;
    props.onSelect(food);
}

function createFoodOption() {}
</script>

<template>
    <div class="food-select">
        <input
            class="input"
            type="text"
            placeholder="Search foods..."
            v-model="search"
            @focus="dropdown = true"
            @blur="dropdown = false"
        />

        <div v-if="dropdown" class="dropdown-container">
            <div v-if="pending" class="status">Loading...</div>
            <div v-else-if="error" class="status error">
                Error loading foods
            </div>
            <div v-else class="dropdown">
                <template v-if="filteredFoods?.length">
                    <div
                        v-for="food in filteredFoods"
                        :key="food.ID"
                        class="option"
                        @mousedown.prevent="itemSelected(food)"
                    >
                        <span>{{ food.name }}</span>
                        <span class="macros">
                            {{ food.calories }} / {{ food.protein }} /
                            {{ food.fiber }}
                        </span>
                    </div>
                </template>

                <div
                    v-else
                    class="empty-option"
                    @mousedown.prevent="createFoodOption"
                >
                    <Plus :size="18" class="icon" />
                    <span>Create {{ search }}"</span>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.food-select {
    position: relative;
    width: 100%;
    display: flex;
    flex-direction: column;
}

.input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.5rem 0.75rem;
    border: none;
    border-radius: 6px;
    background-color: rgb(50, 50, 50);
    color: white;
    font-size: 0.95rem;
}

.input:focus {
    outline: 1px solid #777;
}

.dropdown-container {
    position: absolute;
    top: 100%;
    left: 0;
    width: 100%;
    margin-top: 4px;
    z-index: 10;
}

.dropdown {
    background-color: rgb(43, 43, 43);
    border: 1px solid #3d3d3d;
    border-radius: 8px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
    overflow: hidden;
}

.option {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.6rem 0.8rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    cursor: pointer;
}

.option:last-child {
    border-bottom: none;
}

.option:hover {
    background-color: rgb(60, 60, 60);
}

.macros {
    color: gray;
    font-size: 0.8rem;
}

.empty-option {
    display: flex;
    align-items: center;
    gap: 6px;
    justify-content: center;
    padding: 0.6rem 0.8rem;
    color: #bbb;
    cursor: pointer;
}

.empty-option:hover {
    background-color: rgb(60, 60, 60);
    color: white;
}

.status {
    text-align: center;
    padding: 0.8rem;
    color: gray;
}

.status.error {
    color: rgb(255, 100, 100);
}
</style>
