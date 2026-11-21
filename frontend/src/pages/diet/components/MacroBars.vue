<script setup lang="ts">
import MacroBar from "./macrobar/MacroBar.vue";

defineProps<{
    totalCalories: number;
    totalProtein: number;
    totalFat: number;
    totalCarbs: number;
    plannedCalories: number;
    plannedProtein: number;
    plannedFat: number;
    plannedCarbs: number;
    /** Stay 2×2 at lg too (for narrow side-by-side spots). */
    twoByTwo?: boolean;
}>();
</script>

<template>
    <div class="grid w-full grid-cols-2 gap-x-4 gap-y-3" :class="twoByTwo ? '' : 'lg:grid-cols-4'">
        <MacroBar
            :total="totalCalories"
            :planned="plannedCalories"
            type="calories"
            :indicate-overflow="true"
        />
        <MacroBar
            :total="totalProtein"
            :planned="plannedProtein"
            type="protein"
            :indicate-overflow="true"
        />
        <MacroBar :total="totalCarbs" :planned="plannedCarbs" type="carbs" />
        <MacroBar :total="totalFat" :planned="plannedFat" type="fat" />
    </div>
</template>
