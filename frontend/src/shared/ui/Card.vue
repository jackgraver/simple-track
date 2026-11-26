<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

// No cards inside cards: inside one, use List dividers or a Section heading.
// `tab` slot: a folder tab on the top-left edge (a short label that belongs to the card, e.g. "Set 1").
// It sits in the card's top margin. A tabbed card has no border: tab and body are one filled shape,
// set off from the page by the surface fill alone.
const props = withDefaults(
    // `fill`: the card is a flex column whose body takes the remaining height (bento/grid cells).
    defineProps<{ padding?: "none" | "sm" | "md"; interactive?: boolean; as?: string; to?: string; fill?: boolean }>(),
    { padding: "md", as: "div" },
);

const tag = computed(() => (props.to ? RouterLink : props.as));
const pad = { none: "", sm: "p-3", md: "p-4 lg:p-5" } as const;
</script>

<template>
    <component
        :is="tag"
        :to="to"
        class="w-full rounded-card bg-surface text-left text-fg"
        :class="[
            $slots.tab ? 'relative mt-8 rounded-tl-none' : 'border border-line',
            fill ? 'flex min-h-0 flex-col' : 'block',
            interactive || to
                ? 'cursor-pointer rounded-card transition-colors duration-100 hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent'
                : '',
        ]"
    >
        <div
            v-if="$slots.tab"
            class="absolute bottom-full left-0 flex h-8 items-center gap-2 rounded-t-card bg-surface px-3 text-sm font-medium"
        >
            <slot name="tab" />
        </div>
        <div v-if="$slots.header" class="flex items-center justify-between gap-3 border-b border-line px-4 py-3">
            <slot name="header" />
        </div>
        <div :class="[pad[padding], fill ? 'flex min-h-0 flex-1 flex-col' : '']"><slot /></div>
        <div v-if="$slots.footer" class="flex items-center justify-end gap-2 border-t border-line px-4 py-3">
            <slot name="footer" />
        </div>
    </component>
</template>
