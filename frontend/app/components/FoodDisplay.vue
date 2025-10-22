<script setup lang="ts">
import type { Food } from "~/types/diet";

const props = defineProps<{
    item: Food;
}>();
</script>

<template>
    <div class="food-container">
        <h3 class="name">{{ item.name }}</h3>
        <div class="macros">
            <span class="macro calories">{{ item.calories }}C</span>
            <span class="macro protein">{{ item.protein }}P</span>
            <span class="macro fiber">{{ item.fiber }}F</span>
        </div>
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

.macros {
    display: flex;
    gap: 0.6rem;
    font-weight: 500;
    margin-top: 0.25rem;
}

.macro {
    font-size: 0.9rem;
}

.calories {
    color: #facc15;
}
.protein {
    color: #60a5fa;
}
.fiber {
    color: #34d399;
}
</style>
