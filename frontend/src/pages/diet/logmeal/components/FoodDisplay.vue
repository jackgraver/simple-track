<script setup lang="ts">
import { computed } from "vue";
import type { Food } from "~/types/diet";
import SimpleMacros from "~/shared/SimpleMacros.vue";

const props = defineProps<{
    item: Food & { entry_kind?: string };
}>();

const isComposite = computed(() => props.item.entry_kind === "composite");
</script>

<template>
    <div class="food-container">
        <h3 class="name">
            {{ item.name }}
            <span v-if="isComposite" class="recipe-tag">Recipe</span>
        </h3>
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

.recipe-tag {
    margin-left: 0.35rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: #93c5fd;
    vertical-align: middle;
}
</style>
