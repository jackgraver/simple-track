<script setup lang="ts">
import { AlertCircle, RotateCw } from "lucide-vue-next";
import { computed } from "vue";
import Button from "./Button.vue";

// Inline error for a failed load (a failed *action* is a toast). The headline is ours, never the raw
// error: `error.message` goes in a <details> for the curious. Retry shows only when `@retry` is bound.
const props = withDefaults(
    defineProps<{ error?: unknown; title?: string; compact?: boolean; onRetry?: () => void }>(),
    { title: "Couldn't load this" },
);

const detail = computed(() => {
    const e = props.error;
    if (e instanceof Error) return e.message;
    return typeof e === "string" ? e : "";
});
</script>

<template>
    <div
        v-if="compact"
        role="alert"
        class="flex items-center gap-2 px-3 py-2 text-sm text-danger"
    >
        <AlertCircle aria-hidden="true" class="size-4 shrink-0" />
        <span class="min-w-0 flex-1">{{ title }}</span>
        <Button v-if="onRetry" size="sm" variant="ghost" @click="onRetry">Retry</Button>
    </div>
    <div v-else role="alert" class="flex flex-col items-center gap-3 px-4 py-10 text-center">
        <AlertCircle aria-hidden="true" class="size-8 text-danger" />
        <div class="flex flex-col items-center gap-1">
            <p class="m-0 text-sm font-medium text-fg">{{ title }}</p>
            <p class="m-0 max-w-xs text-sm text-fg-muted">Check your connection and try again.</p>
            <details v-if="detail" class="mt-1 max-w-xs text-xs text-fg-subtle">
                <summary class="cursor-pointer">Details</summary>
                <p class="m-0 mt-1 break-words">{{ detail }}</p>
            </details>
        </div>
        <Button v-if="onRetry" variant="secondary" @click="onRetry">
            <template #icon-left><RotateCw class="size-4" /></template>
            Retry
        </Button>
    </div>
</template>
