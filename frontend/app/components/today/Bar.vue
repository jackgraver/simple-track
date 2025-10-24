<script setup lang="ts">
function formatNum(n: number): string {
    const s = n.toFixed(2); // always 2 decimals
    return s.replace(/\.?0+$/, ""); // drop trailing zeros and optional dot
}

function calcWidth(total: number, planned: number): number {
    return Math.min(100, (total / planned) * 100);
}

function determineOverflow(total: number, planned: number): string {
    if (planned <= 0 || !props.indicateOverflow) return "";
    const overflow = (total / planned) * 100 - 100;

    if (overflow > 20) return "num30";
    if (overflow > 10) return "num20";
    if (overflow > 0) return "num10";
    return "";
}

const props = defineProps<{
    total: number;
    planned: number;
    type: "calories" | "protein" | "fiber" | "carbs";
    indicateOverflow?: boolean;
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
            <span :class="determineOverflow(total, planned)">{{
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
    background-color: #60a5fa;
}
.fiber {
    background-color: green;
}

.num10 {
    color: yellow;
}
.num20 {
    color: orange;
}
.num30 {
    color: red;
}
</style>
