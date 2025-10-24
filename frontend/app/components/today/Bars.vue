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
            :indicate-overflow="true"
        />
        <TodayBar
            :total="totalProtein"
            :planned="plannedProtein"
            type="protein"
            :indicate-overflow="true"
        />
        <TodayBar :total="totalFiber" :planned="plannedFiber" type="fiber" />
        <TodayBar :total="0" :planned="0" type="carbs" />
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
