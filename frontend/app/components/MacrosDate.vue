<script lang="ts" setup>
import type { Day } from "~/types/models";

const props = defineProps<{
    day: Day;
}>();

const day = props.day;
</script>

<template>
    <div class="macros">
        <div class="macro">
            <div class="macro-bar">
                <div
                    class="macro-fill calories"
                    :style="{
                        width: `${Math.min(100, (day.totalCalories ?? 0 / day.plan.calories) * 100)}%`,
                    }"
                >
                    {{ day.totalCalories ?? 0 }}
                </div>
            </div>
            <div class="macro-goal">{{ day.plan.calories }}</div>
        </div>
        <div class="macro">
            <div class="macro-bar">
                <div
                    class="macro-fill protein"
                    :style="{
                        width: `${Math.min(100, (day.totalProtein ?? 0 / day.plan.protein) * 100)}%`,
                    }"
                >
                    {{ day.totalProtein ?? 0 }}g
                </div>
            </div>
            <div class="macro-goal">{{ day.plan.protein }}g</div>
        </div>
        <div class="macro">
            <div class="macro-bar">
                <div
                    class="macro-fill fiber"
                    :style="{
                        width: `${Math.min(100, (day.totalFiber ?? 0 / day.plan.fiber) * 100)}%`,
                    }"
                >
                    {{ day.totalFiber ?? 0 }}g
                </div>
            </div>
            <div class="macro-goal">{{ day.plan.fiber }}g</div>
        </div>
    </div>
</template>

<style scoped>
.macros {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.macro {
    display: flex;
    align-items: center;
    gap: 4px;
}

.macro-bar {
    flex: 1;
    background: #444;
    border-radius: 4px;
    height: 20px;
    overflow: hidden;
    position: relative;
}

.macro-fill {
    height: 100%;
    color: #ffffff;
    font-weight: bold;
    text-align: right;
    padding: 0 8px;
    white-space: nowrap;
    line-height: 20px;
    border-radius: 4px 0 0 4px;
}

.macro-goal {
    width: 40px;
    text-align: right;
    font-weight: bold;
}

/* Colors */
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
