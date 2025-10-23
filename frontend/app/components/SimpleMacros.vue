<script setup lang="ts">
defineProps<{
    calories: number;
    protein: number;
    fiber: number;
}>();
</script>

<template>
    <div class="macros">
        <span class="macro calories">{{ calories }}C</span>
        <span class="macro protein">{{ protein }}P</span>
        <span class="macro fiber">{{ fiber }}F</span>
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
