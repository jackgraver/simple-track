<script setup lang="ts">
// A vertical group of ListRows. `inset` pads the group so rows line up inside a Card with padding.
defineProps<{ inset?: boolean }>();
</script>

<template>
    <ul class="m-0 flex list-none flex-col divide-y divide-line p-0" :class="inset ? 'px-1' : ''">
        <slot />
    </ul>
</template>
