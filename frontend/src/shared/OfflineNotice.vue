<script setup lang="ts">
import { computed } from "vue";
import { CloudOff } from "lucide-vue-next";
import { useOnline, usePausedMutationCount } from "~/api/online";

// Offline, writes are paused (shown as pending) and sent on reconnect. Say so once, app-wide, so a
// row that stays pending isn't read as "stuck". The queue lives in memory: closing the app drops it.
const online = useOnline();
const queued = usePausedMutationCount();
const text = computed(() => {
    const n = queued.value;
    if (n === 0) return "You're offline. Changes are sent when you reconnect.";
    return `You're offline. ${n} ${n === 1 ? "change" : "changes"} will be sent when you reconnect. Keep the app open.`;
});
</script>

<template>
    <p
        v-if="!online"
        role="status"
        class="m-0 mb-3 flex items-center gap-2 rounded-control bg-warn-soft px-3 py-2 text-xs text-fg"
    >
        <CloudOff aria-hidden="true" class="size-4 shrink-0 text-warn" />
        <span>{{ text }}</span>
    </p>
</template>
