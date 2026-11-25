<script setup lang="ts">
// `#trigger` is as-child, for an existing primitive (IconButton). `#trigger-content` is for triggers
// no primitive covers (account row, avatar): the menu renders the button, styled by `shape`.
import { DropdownMenuContent, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuTrigger } from "reka-ui";

withDefaults(
    defineProps<{
        align?: "start" | "center" | "end";
        side?: "top" | "right" | "bottom" | "left";
        width?: "sm" | "md";
        shape?: "row" | "circle";
        triggerLabel?: string;
        active?: boolean;
    }>(),
    { align: "end", side: "bottom", width: "sm", shape: "row" },
);
const open = defineModel<boolean>("open", { default: false });

const widths = { sm: "min-w-48", md: "min-w-64" } as const;
const shapes = {
    row: "w-full gap-2.5 rounded-control px-2 py-1.5",
    circle: "size-touch justify-center rounded-full",
} as const;
</script>

<template>
    <DropdownMenuRoot v-model:open="open">
        <DropdownMenuTrigger v-if="$slots.trigger" as-child><slot name="trigger" /></DropdownMenuTrigger>
        <DropdownMenuTrigger
            v-else
            :aria-label="triggerLabel"
            class="m-0 flex cursor-pointer items-center border-0 text-left text-fg transition-colors duration-100 outline-none hover:bg-surface-3 focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-canvas data-[state=open]:bg-surface-3"
            :class="[shapes[shape], active ? 'bg-surface-3' : 'bg-transparent']"
        >
            <slot name="trigger-content" />
        </DropdownMenuTrigger>
        <DropdownMenuPortal>
            <DropdownMenuContent
                :align="align"
                :side="side"
                :side-offset="6"
                :collision-padding="8"
                class="z-popover max-w-(--reka-dropdown-menu-content-available-width) rounded-card border border-line bg-surface p-1 shadow-raised"
                :class="widths[width]"
            >
                <slot />
            </DropdownMenuContent>
        </DropdownMenuPortal>
    </DropdownMenuRoot>
</template>
