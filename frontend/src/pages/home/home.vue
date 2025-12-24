<script setup lang="ts">
import Meal from "./components/Meal.vue";
import Gym from "./components/Gym.vue";

import { ref } from 'vue';

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
    <main>
        <section>
            <Meal
                @date-change="datechange"
                :date-offset="dateOffset"
                :key="dateOffset"
            />
        </section>
        <section>
            <Gym :date-offset="dateOffset" :key="dateOffset" />
        </section>
    </main>
</template>

<style scoped>
main {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    align-items: center;
    width: 100%;
    height: 100%;
    padding: 1rem;
    padding-top: 0;
    text-align: center;
}

main > section {
    justify-self: start;
    text-align: left;
}
</style>
