<script setup lang="ts">
import GymDate from "~/shared/GymDate.vue";
</script>

<template>
    <div>
        <h1>Workouts</h1>
        <Calendar :fetchURL="`workout/month`" :display-component="GymDate" />
    </div>
</template>

<style scoped></style>
