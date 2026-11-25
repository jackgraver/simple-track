<script setup lang="ts">
// Reka DropdownMenu shell: focus, typeahead, arrow keys, Esc, portal and collision handling come
// from the library. Build the content from MenuAction, MenuLabel, MenuSeparator and MenuRadioGroup.
//
// Trigger: either `#trigger` (as-child: pass one focusable element, e.g. an IconButton) or
// `#trigger-content`, which renders inside a built-in unstyled-ish <button> sized by `trigger-class`
// (for triggers that aren't a primitive, like the side nav account row).
import { DropdownMenuContent, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuTrigger } from "reka-ui";

withDefaults(
    defineProps<{
        align?: "start" | "center" | "end";
        side?: "top" | "right" | "bottom" | "left";
        /** Accessible name for the built-in trigger button when its content has no text. */
        triggerLabel?: string;
        triggerClass?: string;
        /** Tailwind min-width class for the panel. */
        width?: "sm" | "md";
    }>(),
    { align: "end", side: "bottom", width: "sm" },
);
const open = defineModel<boolean>("open", { default: false });

const widths = { sm: "min-w-48", md: "min-w-64" } as const;
</script>

<template>
    <DropdownMenuRoot v-model:open="open">
        <DropdownMenuTrigger v-if="$slots.trigger" as-child><slot name="trigger" /></DropdownMenuTrigger>
        <DropdownMenuTrigger
            v-else
            :aria-label="triggerLabel"
            class="m-0 flex cursor-pointer items-center rounded-control border-0 bg-transparent p-0 text-left text-fg transition-colors duration-100 outline-none hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent data-[state=open]:bg-surface-3"
            :class="triggerClass"
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
