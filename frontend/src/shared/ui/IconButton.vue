<script setup lang="ts">
withDefaults(
    defineProps<{
        label: string;
        variant?: "ghost" | "secondary" | "danger";
        size?: "md" | "sm";
        disabled?: boolean;
        type?: "button" | "submit" | "reset";
    }>(),
    { variant: "ghost", size: "md", type: "button" },
);

// See Button.vue for why the `!` classes are here.
const base =
    "m-0! inline-flex shrink-0 cursor-pointer items-center justify-center rounded-control! border! p-0! shadow-none! transition-colors duration-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent disabled:cursor-not-allowed disabled:opacity-50";
const variants = {
    ghost: "border-transparent! bg-transparent text-fg-muted enabled:hover:bg-surface-3 enabled:hover:text-fg enabled:active:bg-line",
    secondary: "border-line! bg-surface-2 text-fg enabled:hover:bg-surface-3 enabled:active:bg-line",
    danger: "border-transparent! bg-transparent text-fg-muted enabled:hover:bg-danger-soft enabled:hover:text-danger enabled:active:bg-danger/25",
} as const;
// Hit area: 44px below lg, compact from lg. The glyph stays size-5 (slot content).
const sizes = { md: "size-touch lg:size-row", sm: "size-touch lg:size-control-sm" } as const;
</script>

<template>
    <button
        :type="type"
        :aria-label="label"
        :title="label"
        :disabled="disabled"
        :class="[base, variants[variant], sizes[size]]"
    >
        <span aria-hidden="true" class="inline-flex size-5 items-center justify-center [&>svg]:size-5"><slot /></span>
    </button>
</template>
