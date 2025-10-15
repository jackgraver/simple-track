<script setup lang="ts"></script>

<template>
    <body>
        <div class="section">
            <h1>Food</h1>
            <TodayMeal />
        </div>
        <div class="section">
            <h1>Gym</h1>
        </div>
    </body>
</template>

<style scoped>
body {
    height: 100vh;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    justify-content: center;
    align-items: center;
}

section {
    display: flex;
    flex: 1;
}
</style>

<!-- <LogMeal />
    <ManageMealPlan />
    <LogWorkout /> -->
