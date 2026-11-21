<script setup lang="ts">
import { useId } from "vue";

// One step of the session timeline on the routine page: a numbered dot on a rail, a heading,
// and the step's editor. The rail stops at the last step.
defineProps<{ step: number; title: string; meta?: string; last?: boolean }>();
const id = useId();
</script>

<template>
    <section :aria-labelledby="id" class="grid grid-cols-[1.5rem_1fr] gap-x-3 lg:gap-x-4">
        <div class="flex flex-col items-center" aria-hidden="true">
            <span
                class="flex size-6 shrink-0 items-center justify-center rounded-full border border-line bg-surface-2 text-xs font-semibold text-fg-muted tabular-nums"
            >
                {{ step }}
            </span>
            <span v-if="!last" class="mt-1 w-px flex-1 bg-line" />
        </div>
        <div class="flex min-w-0 flex-col gap-3" :class="last ? '' : 'pb-8'">
            <div class="flex min-h-6 items-center justify-between gap-3">
                <h2 :id="id" class="m-0 flex min-w-0 items-baseline gap-2 text-base font-semibold text-fg">
                    {{ title }}
                    <span v-if="meta" class="truncate text-sm font-normal text-fg-muted">{{ meta }}</span>
                </h2>
                <div v-if="$slots.actions" class="flex shrink-0 items-center gap-2"><slot name="actions" /></div>
            </div>
            <slot />
        </div>
    </section>
</template>
