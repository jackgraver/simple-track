<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

// No cards inside cards: inside one, use List dividers or a Section heading.
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
        class="w-full rounded-card border border-line bg-surface text-left text-fg"
        :class="[
            fill ? 'flex min-h-0 flex-col' : 'block',
            interactive || to
                ? 'cursor-pointer rounded-card border-line transition-colors duration-100 hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent'
                : '',
        ]"
    >
        <div v-if="$slots.header" class="flex items-center justify-between gap-3 border-b border-line px-4 py-3">
            <slot name="header" />
        </div>
        <div :class="[pad[padding], fill ? 'flex min-h-0 flex-1 flex-col' : '']"><slot /></div>
        <div v-if="$slots.footer" class="flex items-center justify-end gap-2 border-t border-line px-4 py-3">
            <slot name="footer" />
        </div>
    </component>
</template>
