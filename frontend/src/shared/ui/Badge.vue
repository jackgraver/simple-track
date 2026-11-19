<script setup lang="ts">
export type Tone = "neutral" | "accent" | "success" | "warn" | "danger" | "info";

withDefaults(defineProps<{ tone?: Tone; variant?: "soft" | "outline"; size?: "sm" | "md" }>(), {
    tone: "neutral",
    variant: "soft",
    size: "sm",
});

// Full class strings so Tailwind can see them. Status is never colour alone: callers add a word or icon.
const soft: Record<Tone, string> = {
    neutral: "bg-surface-3 text-fg-muted",
    accent: "bg-accent-soft text-accent",
    success: "bg-success-soft text-success",
    warn: "bg-warn-soft text-warn",
    danger: "bg-danger-soft text-danger",
    info: "bg-info-soft text-info",
};
const outline: Record<Tone, string> = {
    neutral: "border-line-strong text-fg-muted",
    accent: "border-accent text-accent",
    success: "border-success text-success",
    warn: "border-warn text-warn",
    danger: "border-danger text-danger",
    info: "border-info text-info",
};
</script>

<template>
    <span
        class="inline-flex items-center gap-1 rounded-full font-medium whitespace-nowrap"
        :class="[
            variant === 'soft' ? soft[tone] : ['border', outline[tone]],
            size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm',
        ]"
    >
        <slot />
    </span>
</template>
