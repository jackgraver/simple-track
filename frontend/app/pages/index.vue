<script setup lang="ts"></script>

<template>
    <main class="page">
        <section class="home-section">
            <TodayMeal />
        </section>
        <section class="home-section">
            <TodayGym />
        </section>
    </main>
</template>

<style scoped>
.page {
    display: grid;
    row-gap: 1.25rem;
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

.home-section {
    display: flex;
    justify-content: center;
}

.home-section > * {
    width: 75%;
    max-width: 1200px;
}
</style>
