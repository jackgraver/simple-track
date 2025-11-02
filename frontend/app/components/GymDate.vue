<script setup lang="ts">
import { formatDate } from "~/utils/dateUtil";

const props = defineProps<{
    day: any;
}>();
</script>

<template>
    <span>{{ day.workout_plan?.name }}</span>
</template>

<style scoped></style>
