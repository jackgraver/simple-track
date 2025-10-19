<script setup lang="ts"></script>

<template>
    <body>
        <div>
            <TodayMeal />
        </div>
        <div>
            <TodayGym />
        </div>
    </body>
</template>

<style scoped>
body {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    justify-content: center;
    align-items: center;
    font-family:
        system-ui,
        -apple-system,
        BlinkMacSystemFont,
        "Segoe UI",
        Roboto,
        "Helvetica Neue",
        Arial,
        sans-serif;
}
</style>
