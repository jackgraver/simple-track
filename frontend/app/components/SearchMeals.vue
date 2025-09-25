<script lang="ts" setup>
import type { Meal } from '~/types/models';

const { data, pending, error } = useFetch<Meal[]>("http://localhost:8080/mealplan/meal/all");

let display = ref(false);
</script>

<template>  
    <input type="text" placeholder="Search meals..." />
    <div v-if="pending">Loading Meals...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else-if="display">
        <div v-for="meal in data" :key="meal.ID">
            {{ meal.name }}
        </div>
    </div>
</template>

<style scoped>

</style>