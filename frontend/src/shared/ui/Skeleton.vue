<script setup lang="ts">
// A placeholder the shape of the content. Decorative: the wrapper (SkeletonList, or the caller) announces loading.
withDefaults(defineProps<{ shape?: "line" | "block" | "circle"; width?: string; height?: string }>(), {
    shape: "line",
});
</script>

<template>
    <span
        aria-hidden="true"
        class="block animate-pulse bg-surface-3"
        :class="{ 'h-4 rounded-control': shape === 'line', 'h-24 rounded-card': shape === 'block', 'size-8 shrink-0 rounded-full': shape === 'circle' }"
        :style="{ width: width, height: height }"
    ></span>
</template>
