<script setup lang="ts">
import LoggingHeader from "./LoggingHeader.vue";

defineProps<{
    pending?: boolean;
    error?: Error | null;
    empty?: boolean;
    emptyMessage?: string;
}>();

const emit = defineEmits<{
    (e: "back"): void;
}>();
</script>

<template>
    <div
        v-if="pending"
        class="logging-page-shell logging-page-shell--pad logging-page-shell--loading flex w-full flex-col gap-6 self-stretch"
    >
        <LoggingHeader @back="emit('back')">
            <template #center>
                <h2
                    class="logging-title m-0 min-w-0 truncate text-center text-lg font-medium text-zinc-400"
                >
                    Loading…
                </h2>
            </template>
            <template #right>
                <span class="set-indicator text-zinc-500">Set —</span>
            </template>
        </LoggingHeader>
    </div>
    <div
        v-else-if="error"
        class="logging-page-shell logging-page-shell--pad flex w-full flex-col gap-4 self-stretch"
    >
        <div>Error: {{ error.message }}</div>
    </div>
    <div
        v-else-if="empty"
        class="logging-page-shell logging-page-shell--pad flex w-full flex-col gap-4 self-stretch"
    >
        <slot name="empty">
            <div>{{ emptyMessage ?? "Nothing to show." }}</div>
            <button
                class="rounded border border-zinc-600 px-4 py-2 hover:bg-zinc-800"
                type="button"
                @click="emit('back')"
            >
                Back
            </button>
        </slot>
    </div>
    <div
        v-else
        class="logging-page-shell logging-page-shell--pad flex w-full flex-col gap-4 self-stretch"
    >
        <slot />
    </div>
</template>

<style scoped>
.logging-page-shell--pad {
    width: 100%;
    max-width: 100%;
}
@media (max-width: 767px) {
    .logging-page-shell--pad {
        padding: 0.5rem 0;
    }
}
</style>
