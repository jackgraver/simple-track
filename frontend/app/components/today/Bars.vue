<script setup lang="ts">
defineProps<{
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    plannedCalories: number;
    plannedProtein: number;
    plannedFiber: number;
}>();
</script>

<template>
    <div class="bars-container">
        <TodayBar
            :total="totalCalories"
            :planned="plannedCalories"
            type="calories"
        />
        <TodayBar
            :total="totalProtein"
            :planned="plannedProtein"
            type="protein"
        />
        <TodayBar :total="totalFiber" :planned="plannedFiber" type="fiber" />
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
