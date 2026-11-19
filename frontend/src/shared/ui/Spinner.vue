<script setup lang="ts">
import { Loader2 } from "lucide-vue-next";

withDefaults(defineProps<{ size?: "sm" | "md"; label?: string }>(), { size: "md", label: "Loading" });
</script>

<template>
    <span role="status" class="inline-flex items-center">
        <Loader2 aria-hidden="true" class="animate-spin text-fg-muted" :class="size === 'sm' ? 'size-4' : 'size-5'" />
        <span class="sr-only">{{ label }}</span>
    </span>
</template>
