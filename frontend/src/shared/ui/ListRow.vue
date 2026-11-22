<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { RouterLink } from "vue-router";

defineOptions({ inheritAttrs: false });
const props = defineProps<{ to?: string; selected?: boolean; pending?: boolean; disabled?: boolean }>();
const attrs = useAttrs();

// RouterLink for `to`, <button> when a click handler is attached, otherwise a plain <div>.
// Actions (desktop hover) sit beside the main element, never inside it, so buttons never nest.
const clickable = computed(() => !!attrs.onClick);
const main = computed(() => (props.to && !props.disabled ? RouterLink : clickable.value ? "button" : "div"));
const interactive = computed(() => !props.disabled && (!!props.to || clickable.value));
</script>

<template>
    <li
        class="group relative flex items-center"
        :class="[
            pending ? 'opacity-60' : '',
            disabled ? 'opacity-50' : '',
            selected ? 'bg-surface-2' : '',
            interactive ? 'transition-colors duration-100 hover:bg-surface-3' : '',
        ]"
        :aria-busy="pending || undefined"
    >
        <component
            :is="main"
            v-bind="{ ...attrs, ...(main === RouterLink ? { to } : {}), ...(main === 'button' ? { type: 'button', disabled } : {}) }"
            :aria-current="selected ? 'true' : undefined"
            class="flex min-h-touch min-w-0 flex-1 items-center gap-3 px-3 py-2 text-left text-base text-fg no-underline lg:min-h-row lg:px-4 lg:text-sm"
            :class="[
                interactive
                    ? 'cursor-pointer focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent'
                    : '',
                disabled ? 'cursor-not-allowed' : '',
            ]"
        >
            <span v-if="$slots.leading" class="flex shrink-0 items-center text-fg-muted"><slot name="leading" /></span>
            <span class="flex min-w-0 flex-1 flex-col">
                <span class="truncate"><slot /></span>
                <span v-if="$slots.subtitle" class="truncate text-xs text-fg-muted"><slot name="subtitle" /></span>
            </span>
            <span v-if="$slots.trailing" class="flex shrink-0 items-center gap-2 text-sm text-fg-muted tabular-nums">
                <slot name="trailing" />
            </span>
        </component>
        <!-- Always visible, at every width: a sibling of the main element so a clickable row can hold a
             44px IconButton or a Menu trigger without nesting buttons. -->
        <div v-if="$slots.end" class="flex shrink-0 items-center gap-1 pr-1 lg:pr-2"><slot name="end" /></div>
        <!-- Desktop only; on mobile the trailing slot holds an IconButton that opens a sheet or menu. -->
        <div
            v-if="$slots.actions"
            class="hidden shrink-0 items-center gap-1 pr-3 opacity-0 transition-opacity duration-100 group-focus-within:opacity-100 group-hover:opacity-100 lg:flex"
        >
            <slot name="actions" />
        </div>
    </li>
</template>
