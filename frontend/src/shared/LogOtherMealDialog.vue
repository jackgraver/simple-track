<script setup lang="ts">
import type { Meal, MealItem } from "~/types/diet";

const form = reactive({
    ID: 0,
    name: "",
    items: [] as Partial<MealItem>[],
});
</script>

<template>
    <div>
        <SearchMeals
            v-model:ID="form.ID"
            v-model:name="form.name"
            v-model:items="form.items"
            @update:model-value="form.name = $event"
            @select-meal="(meal: Meal) => console.log('huh?', meal)"
            @create-meal=""
        />
    </div>
</template>

<style scoped></style>
