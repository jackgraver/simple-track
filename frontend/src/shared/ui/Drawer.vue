<script setup lang="ts">
// A panel that slides in from the left edge (mobile navigation). Reka Dialog supplies the focus trap
// and restore, Esc, scroll lock and ARIA; the look is ours. The swipe gesture lives with the caller
// (it needs the whole screen, not just the panel): it drives `offset` while a finger is down.
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from "reka-ui";
import { computed } from "vue";

const open = defineModel<boolean>("open", { required: true });
const props = withDefaults(
    defineProps<{
        title: string;
        /** Panel position in px from fully open: 0 = open, -width = off screen. */
        offset?: number;
        /** A finger is driving `offset`: follow it with no easing. */
        dragging?: boolean;
        /** Panel width in px, to fade the scrim with the offset. */
        width?: number;
        /** Play the slide-in. Off when a swipe already pulled the panel out. */
        animateEnter?: boolean;
    }>(),
    { offset: 0, dragging: false, width: 288, animateEnter: true },
);

// Scrim fades with the panel while it is being dragged or settling.
const scrimOpacity = computed(() => (props.offset === 0 ? undefined : Math.max(0, 1 + props.offset / props.width)));
</script>

<template>
    <DialogRoot v-model:open="open">
        <DialogPortal>
            <DialogOverlay
                class="fixed inset-0 z-overlay bg-overlay data-[state=closed]:animate-overlay-out"
                :class="animateEnter && 'data-[state=open]:animate-overlay-in'"
                :style="{ opacity: scrimOpacity, transition: dragging ? 'none' : 'opacity 150ms ease-out' }"
            />
            <DialogContent
                class="fixed inset-y-0 left-0 z-overlay flex w-72 max-w-[85vw] flex-col overflow-hidden border-r border-line bg-surface pt-safe pb-safe text-fg shadow-overlay outline-none data-[state=closed]:animate-drawer-out"
                :class="animateEnter && 'data-[state=open]:animate-drawer-in'"
                :style="{
                    '--drawer-drag': `${offset}px`,
                    transform: offset ? `translateX(${offset}px)` : undefined,
                    transition: dragging ? 'none' : 'transform 150ms ease-out',
                }"
            >
                <DialogTitle class="sr-only">{{ title }}</DialogTitle>
                <DialogDescription class="sr-only">{{ title }}</DialogDescription>
                <slot />
            </DialogContent>
        </DialogPortal>
    </DialogRoot>
</template>
