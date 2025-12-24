<script setup lang="ts">
import MacroBar from "./MacroBar.vue";

defineProps<{
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
    plannedCalories: number;
    plannedProtein: number;
    plannedFiber: number;
    plannedCarbs: number;
}>();
</script>

<template>
    <div class="bars-container">
        <MacroBar
            :total="totalCalories"
            :planned="plannedCalories"
            type="calories"
            :indicate-overflow="true"
        />
        <MacroBar
            :total="totalProtein"
            :planned="plannedProtein"
            type="protein"
            :indicate-overflow="true"
        />
        <MacroBar :total="totalFiber" :planned="plannedFiber" type="fiber" />
        <MacroBar :total="totalCarbs" :planned="plannedCarbs" type="carbs" />
    </div>
</template>

<style scoped>
.bars-container {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    width: 100%;
}
</style>
