<script setup lang="ts">
import type { Food } from "~/types/diet";
import SimpleMacros from "./SimpleMacros.vue";

const props = defineProps<{
    item: Food;
}>();
</script>

<template>
    <div class="food-container">
        <h3 class="name">{{ item.name }}</h3>
        <SimpleMacros
            :calories="item.calories"
            :protein="item.protein"
            :fiber="item.fiber"
        />
    </div>
</template>

<style scoped>
.food-container {
    display: flex;
    flex-direction: column; /* stack vertically */
    align-items: flex-start; /* left-align everything */
    border-radius: 4px;
    color: white;
    line-height: 1.3;
}

.name {
    font-size: 1rem;
    margin: 0;
    font-weight: 600;
}
</style>
