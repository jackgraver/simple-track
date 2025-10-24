<script setup lang="ts">
defineProps<{
    calories: number;
    protein: number;
    fiber: number;
    fontSize?: string;
}>();
</script>

<template>
    <div class="macros">
        <span class="macro calories" :style="{ fontSize }"
            >{{ calories }}C</span
        >
        <span class="macro protein" :style="{ fontSize }">{{ protein }}P</span>
        <span class="macro fiber" :style="{ fontSize }">{{ fiber }}F</span>
    </div>
</template>

<style scoped>
.macros {
    display: flex;
    gap: 0.6rem;
    font-weight: 500;
    margin-top: 0.25rem;
}

.macro {
    font-size: 0.9rem;
}

.calories {
    color: #facc15;
}
.protein {
    color: #60a5fa;
}
.fiber {
    color: #34d399;
}
</style>
