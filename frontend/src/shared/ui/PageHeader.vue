<script setup lang="ts">
// The title of a page. Module pages keep `ModuleTitleBar` (the module's sub-navigation); sub-pages and
// detail pages use this (DECISIONS P7). Below lg the actions stack under the title.
defineProps<{ title: string }>();
</script>

<template>
    <header class="flex flex-col gap-3 border-b border-line pb-3 lg:flex-row lg:items-center lg:justify-between lg:gap-4">
        <div class="flex min-w-0 flex-col gap-0.5">
            <h1 class="m-0 truncate text-xl font-semibold tracking-tight text-fg">{{ title }}</h1>
            <p v-if="$slots.meta" class="m-0 text-sm text-fg-muted"><slot name="meta" /></p>
        </div>
        <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2"><slot name="actions" /></div>
    </header>
</template>
