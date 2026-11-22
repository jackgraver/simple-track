<script setup lang="ts">
import type { Component } from "vue";

// One line of what this is plus the action that fills it. Inside a Card or List it has no border of its own.
defineProps<{ title: string; description?: string; icon?: Component }>();
</script>

<template>
    <div class="flex flex-col items-center gap-3 px-4 py-10 text-center">
        <component :is="icon" v-if="icon" aria-hidden="true" class="size-8 text-fg-subtle" />
        <div class="flex flex-col items-center gap-1">
            <p class="m-0 text-sm font-medium text-fg">{{ title }}</p>
            <p v-if="description" class="m-0 max-w-xs text-sm text-fg-muted">{{ description }}</p>
        </div>
        <div v-if="$slots.action" class="mt-1"><slot name="action" /></div>
    </div>
</template>
