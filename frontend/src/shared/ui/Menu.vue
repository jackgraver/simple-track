<script setup lang="ts">
// Reka DropdownMenu: focus, typeahead, arrow keys, Esc, portal and collision handling come from
// the library. The trigger is whatever the caller puts in the slot (usually an IconButton).
import type { Component } from "vue";
import { DropdownMenuContent, DropdownMenuItem, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuTrigger } from "reka-ui";

export type MenuItem = {
    key: string;
    label: string;
    icon?: Component;
    tone?: "neutral" | "danger";
    disabled?: boolean;
};

defineProps<{ items: MenuItem[]; align?: "start" | "center" | "end" }>();
const emit = defineEmits<{ select: [key: string] }>();

const itemBase =
    "m-0 flex min-h-touch w-full cursor-pointer items-center gap-3 rounded-control px-3 text-base outline-none select-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50 lg:min-h-row lg:text-sm";
const tones = {
    neutral: "text-fg data-[highlighted]:bg-surface-3",
    danger: "text-danger data-[highlighted]:bg-danger-soft",
} as const;
const toneClass = (t?: MenuItem["tone"]) => tones[t === "danger" ? "danger" : "neutral"];
</script>

<template>
    <DropdownMenuRoot>
        <DropdownMenuTrigger as-child><slot /></DropdownMenuTrigger>
        <DropdownMenuPortal>
            <DropdownMenuContent
                :align="align ?? 'end'"
                :side-offset="6"
                :collision-padding="8"
                class="z-popover min-w-48 rounded-card border border-line bg-surface p-1 shadow-raised"
            >
                <DropdownMenuItem
                    v-for="item in items"
                    :key="item.key"
                    :disabled="item.disabled"
                    :class="[itemBase, toneClass(item.tone)]"
                    @select="emit('select', item.key)"
                >
                    <component :is="item.icon" v-if="item.icon" aria-hidden="true" class="size-4 shrink-0" />
                    <span class="min-w-0 flex-1 truncate">{{ item.label }}</span>
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenuPortal>
    </DropdownMenuRoot>
</template>
