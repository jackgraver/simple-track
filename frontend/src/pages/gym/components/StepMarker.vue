<script setup lang="ts">
import { Check } from "lucide-vue-next";

// The leading marker of a workout step: a check when done, an accent ring on the next one, a plain
// number otherwise. Decorative: the row says "done"/"up next" in words for screen readers.
withDefaults(defineProps<{ n: number; state: "done" | "next" | "todo"; size?: "sm" | "md" }>(), { size: "md" });
</script>

<template>
    <span
        aria-hidden="true"
        class="flex shrink-0 items-center justify-center rounded-full font-semibold tabular-nums"
        :class="[
            size === 'sm' ? 'size-5 text-2xs' : 'size-7 text-xs',
            state === 'done'
                ? 'bg-success-soft text-success'
                : state === 'next'
                  ? 'border-2 border-accent text-accent'
                  : 'border border-line-strong text-fg-muted',
        ]"
    >
        <Check v-if="state === 'done'" :class="size === 'sm' ? 'size-3' : 'size-4'" stroke-width="3" />
        <template v-else>{{ n }}</template>
    </span>
</template>
