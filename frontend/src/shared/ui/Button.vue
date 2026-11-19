<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { Loader2 } from "lucide-vue-next";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "danger-solid";

const props = withDefaults(
    defineProps<{
        variant?: ButtonVariant;
        size?: "sm" | "md" | "lg";
        loading?: boolean;
        disabled?: boolean;
        block?: boolean;
        to?: string;
        type?: "button" | "submit" | "reset";
    }>(),
    { variant: "secondary", size: "md", type: "button" },
);

// `!` classes neutralise the global `button {}` rule in App.vue until phase 6 removes it, so
// callers never need overrides. Delete them with that rule.
const base =
    "m-0! inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 rounded-control! border! font-medium whitespace-nowrap shadow-none! transition-colors duration-100 select-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent";

const variants: Record<ButtonVariant, [rest: string, interactive: string]> = {
    primary: ["border-transparent! bg-fg text-fg-inverse", "hover:opacity-90 active:opacity-80"],
    secondary: ["border-line! bg-surface-2 text-fg", "hover:bg-surface-3 active:bg-line"],
    ghost: ["border-transparent! bg-transparent text-fg-muted", "hover:bg-surface-3 hover:text-fg active:bg-line"],
    danger: ["border-transparent! bg-danger-soft text-danger", "hover:bg-danger/25 active:bg-danger/35"],
    // text-fg-inverse, not on-solid: white on the dark-theme danger fill is under 4.5:1 (DECISIONS).
    "danger-solid": ["border-transparent! bg-danger text-fg-inverse", "hover:opacity-90 active:opacity-80"],
};

// 44px on mobile everywhere; compact from lg (PRINCIPLES §2).
const sizes = {
    sm: "min-h-touch px-3! text-sm lg:min-h-control-sm",
    md: "min-h-touch px-4! text-base lg:min-h-row lg:text-sm",
    lg: "min-h-14 px-6! text-base lg:min-h-touch",
} as const;

const inert = computed(() => props.disabled || props.loading);
const classes = computed(() => [
    base,
    variants[props.variant][0],
    inert.value ? "" : variants[props.variant][1],
    sizes[props.size],
    props.block ? "w-full" : "",
    props.disabled ? "cursor-not-allowed opacity-50" : "",
    props.loading ? "cursor-progress" : "",
]);
</script>

<template>
    <RouterLink
        v-if="to && !inert"
        :to="to"
        :class="[classes, 'no-underline']"
    >
        <slot name="icon-left" />
        <slot />
        <slot name="icon-right" />
    </RouterLink>
    <button
        v-else
        :type="type"
        :disabled="disabled"
        :aria-busy="loading || undefined"
        :aria-disabled="loading || undefined"
        :class="classes"
        @click="(e) => inert && e.preventDefault()"
    >
        <Loader2 v-if="loading" aria-hidden="true" class="size-4 animate-spin" />
        <slot v-else name="icon-left" />
        <slot />
        <slot name="icon-right" />
    </button>
</template>
