<script setup lang="ts">
function formatNum(n: number): string {
    const s = n.toFixed(2); // always 2 decimals
    return s.replace(/\.?0+$/, ""); // drop trailing zeros and optional dot
}

function calcWidth(total: number, planned: number): number {
    return Math.min(100, (total / planned) * 100);
}

defineProps<{
    total: number;
    planned: number;
    type: "calories" | "protein" | "fiber";
}>();
</script>

<template>
    <div class="fill-container">
        <div
            class="fill"
            :class="type"
            :style="{
                width: `${calcWidth(total, planned)}%`,
            }"
        >
            <span>{{
                formatNum(total ?? 0) + " / " + formatNum(planned ?? 0)
            }}</span>
        </div>
    </div>
</template>

<style scoped>
.fill-container {
    flex: 1;
    height: 20px;
    width: 250px;
    border-radius: 4px;
    background-color: #252525;
}

.fill {
    height: 100%;
    color: #ffffff;
    font-weight: bold;
    text-align: right;
    white-space: nowrap;
    line-height: 20px;
    border-radius: 4px;
    transition: width 0.6s ease-in-out;
}

.fill span {
    padding: 0 8px;
}

.calories {
    background-color: orange;
}
.protein {
    background-color: blue;
}
.fiber {
    background-color: green;
}
</style>
