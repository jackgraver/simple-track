<script setup lang="ts">
const dateOffset = ref(0);

const datechange = (direction: "next" | "prev"): void => {
    if (direction === "next") {
        dateOffset.value++;
    } else if (direction === "prev") {
        dateOffset.value--;
    }
};
</script>

<template>
    <main class="page">
        <section class="home-section">
            <TodayMeal
                @date-change="datechange"
                :date-offset="dateOffset"
                :key="dateOffset"
            />
        </section>
        <section class="home-section">
            <TodayGym :date-offset="dateOffset" :key="dateOffset" />
        </section>
    </main>
</template>

<style scoped>
.page {
    display: grid;
    row-gap: 1.25rem;
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
