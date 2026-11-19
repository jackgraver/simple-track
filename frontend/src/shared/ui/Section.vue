<script setup lang="ts">
import { useId } from "vue";

// A titled group on a page. `gap-6` between sections is the page's job (its flex column).
defineProps<{ title: string }>();
const id = useId();
</script>

<template>
    <section :aria-labelledby="id" class="flex flex-col gap-3">
        <div class="flex items-center justify-between gap-3">
            <h2 :id="id" class="m-0 text-base font-semibold text-fg">{{ title }}</h2>
            <div v-if="$slots.actions" class="flex items-center gap-2"><slot name="actions" /></div>
        </div>
        <slot />
    </section>
</template>
