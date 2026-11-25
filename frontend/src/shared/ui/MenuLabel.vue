<script setup lang="ts">
import { DropdownMenuLabel } from "reka-ui";

defineProps<{ variant?: "section" | "header" }>();
</script>

<template>
    <DropdownMenuLabel
        v-if="variant === 'header'"
        class="m-0 flex items-center gap-3 px-3 py-2.5 text-sm text-fg"
    >
        <slot />
    </DropdownMenuLabel>
    <DropdownMenuLabel
        v-else
        class="m-0 px-3 pt-2 pb-1 text-xs font-semibold tracking-wide text-fg-muted uppercase"
    >
        <slot />
    </DropdownMenuLabel>
</template>
