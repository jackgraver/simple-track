<script setup lang="ts">
// One action in a DropdownMenu. With `to` it's a link (RouterLink), otherwise it emits `select`.
// 44px rows below lg, compact from lg (PRINCIPLES §2).
import { computed, type Component } from "vue";
import { RouterLink, type RouteLocationRaw } from "vue-router";
import { DropdownMenuItem } from "reka-ui";

const props = defineProps<{
    icon?: Component;
    tone?: "neutral" | "danger";
    disabled?: boolean;
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
</script>

<template>
    <DropdownMenuItem
        :as-child="!!to"
        :disabled="disabled"
        :class="[base, toneClass]"
        @select="emit('select', $event)"
    >
        <RouterLink v-if="to" :to="to">
            <component :is="icon" v-if="icon" aria-hidden="true" class="size-4 shrink-0 text-fg-muted" />
            <span class="min-w-0 flex-1 truncate"><slot /></span>
            <span v-if="$slots.trailing" class="shrink-0 text-xs text-fg-subtle"><slot name="trailing" /></span>
        </RouterLink>
        <template v-else>
            <component
                :is="icon"
                v-if="icon"
                aria-hidden="true"
                class="size-4 shrink-0"
                :class="tone === 'danger' ? '' : 'text-fg-muted'"
            />
            <span class="min-w-0 flex-1 truncate"><slot /></span>
            <span v-if="$slots.trailing" class="shrink-0 text-xs text-fg-subtle"><slot name="trailing" /></span>
        </template>
    </DropdownMenuItem>
</template>
