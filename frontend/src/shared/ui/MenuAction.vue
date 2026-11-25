<script setup lang="ts">
import { computed, type Component } from "vue";
import { RouterLink, type RouteLocationRaw } from "vue-router";
import { DropdownMenuItem } from "reka-ui";
import { Loader2 } from "lucide-vue-next";

const props = defineProps<{
    icon?: Component;
    tone?: "neutral" | "danger";
    disabled?: boolean;
    /** Keeps the menu open with a spinner until the caller clears it. */
    loading?: boolean;
    to?: RouteLocationRaw;
}>();
const emit = defineEmits<{ select: [event: Event] }>();

const base =
    "m-0 flex min-h-touch w-full cursor-pointer items-center gap-3 rounded-control px-3 text-base no-underline outline-none select-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50 lg:min-h-row lg:text-sm";
const tones = {
    neutral: "text-fg data-[highlighted]:bg-surface-3",
    danger: "text-danger data-[highlighted]:bg-danger-soft",
} as const;
const toneClass = computed(() => tones[props.tone === "danger" ? "danger" : "neutral"]);

function onSelect(e: Event) {
    if (props.loading) {
        e.preventDefault();
        return;
    }
    emit("select", e);
}
</script>

<template>
    <DropdownMenuItem
        :as="to ? RouterLink : 'div'"
        v-bind="to ? { to } : {}"
        :disabled="disabled"
        :aria-busy="loading || undefined"
        :class="[base, toneClass]"
        @select="onSelect"
    >
        <Loader2 v-if="loading" aria-hidden="true" class="size-4 shrink-0 animate-spin" />
        <component :is="icon" v-else-if="icon" aria-hidden="true" class="size-4 shrink-0" />
        <span class="min-w-0 flex-1 truncate"><slot /></span>
        <span v-if="$slots.trailing" class="shrink-0 text-xs text-fg-subtle"><slot name="trailing" /></span>
    </DropdownMenuItem>
</template>
