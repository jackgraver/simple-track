<script setup lang="ts">
import { ref } from "vue";
import Meal from "../home/components/Meal.vue";

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
    <div>
        <h1>Diet</h1>
        <Meal
            @date-change="datechange"
            :date-offset="dateOffset"
            :key="dateOffset"
        />
    </div>
</template>

<style scoped></style>
