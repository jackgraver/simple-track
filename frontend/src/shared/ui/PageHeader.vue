<script setup lang="ts">
// The title of a page. Module pages keep `ModuleTitleBar` (the module's sub-navigation); sub-pages and
// detail pages use this (DECISIONS P7). Below lg the actions stack under the title, unless `inline`
// (for a single icon action, e.g. a ⋯ menu, which fits beside the title at every width).
defineProps<{ title: string; inline?: boolean }>();
</script>

<template>
    <header
        class="flex gap-3 border-b border-line pb-3 lg:flex-row lg:items-center lg:justify-between lg:gap-4"
        :class="inline ? 'flex-row items-start justify-between' : 'flex-col'"
    >
        <div class="flex min-w-0 flex-col gap-0.5">
            <h1 class="m-0 truncate text-xl font-semibold tracking-tight text-fg">{{ title }}</h1>
            <p v-if="$slots.meta" class="m-0 text-sm text-fg-muted"><slot name="meta" /></p>
        </div>
        <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2" :class="inline ? '-mr-2 lg:mr-0' : ''"><slot name="actions" /></div>
    </header>
</template>
