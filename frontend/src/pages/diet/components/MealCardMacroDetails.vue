<script setup lang="ts">
import SimpleMacros from "~/shared/SimpleMacros.vue";
import { computed } from "vue";

const props = defineProps<{
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    amount: number;
}>();

const scaled = computed(() => {
    const a = Number(props.amount);
    return {
        calories: (props.calories ?? 0) * a,
        protein: (props.protein ?? 0) * a,
        carbs: (props.carbs ?? 0) * a,
        fat: (props.fat ?? 0) * a,
    };
});
</script>

<template>
    <div class="meal-card-macro-details">
        <SimpleMacros
            :calories="scaled.calories"
            :protein="scaled.protein"
            :carbs="scaled.carbs"
            :fat="scaled.fat"
        />
    </div>
</template>

<style scoped>
.meal-card-macro-details {
    opacity: 0;
    visibility: hidden;
    transition: visibility 0.3s ease;
    transition-delay: 0.5s;
}
.meal-card-macro-details :deep(.macros) {
    margin-top: 0;
    gap: 0.45rem;
}
.meal-card-macro-details :deep(.macro) {
    font-size: clamp(0.62rem, 2.5vw, 0.85rem);
}
</style>
