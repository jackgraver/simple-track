<script setup lang="ts">
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { computed } from "vue";

const props = defineProps<{
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
    amount: number;
}>();

const scaled = computed(() => {
    const a = Number(props.amount);
    return {
        calories: (props.calories ?? 0) * a,
        protein: (props.protein ?? 0) * a,
        carbs: (props.carbs ?? 0) * a,
        fat: (props.fat ?? 0) * a,
        fiber: (props.fiber ?? 0) * a,
    };
});
</script>

<template>
    <SimpleMacros
        :calories="scaled.calories"
        :protein="scaled.protein"
        :carbs="scaled.carbs"
        :fat="scaled.fat"
        :fiber="scaled.fiber"
    />
</template>
