<script setup lang="ts">
// A panel that slides in from the left edge (mobile navigation). Reka Dialog supplies the focus trap
// and restore, Esc, scroll lock and ARIA; the look and motion are ours. State and motion live in a
// SlideOver (slideOver.ts) the caller creates, so a gesture elsewhere on the screen can drive it.
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from "reka-ui";
import { motion, useTransform } from "motion-v";
import type { SlideOver } from "./slideOver";

const props = defineProps<{ title: string; panel: SlideOver }>();

// A spring carrying a fast fling can overshoot past open: hold the panel at the edge instead.
const panelX = useTransform(props.panel.x, (v) => Math.min(0, v));
const scrimOpacity = useTransform(props.panel.x, (v) => Math.min(1, Math.max(0, 1 + v / props.panel.width())));
</script>

<template>
    <!-- Reka sees the dialog as open until the panel has slid out, so it stays mounted meanwhile. -->
    <DialogRoot :open="panel.present.value" @update:open="(v) => !v && panel.settle(false)">
        <DialogPortal>
            <DialogOverlay as-child>
                <motion.div
                    class="fixed inset-0 z-overlay bg-overlay"
                    :class="!panel.open.value && 'pointer-events-none'"
                    :style="{ opacity: scrimOpacity }"
                />
            </DialogOverlay>
            <DialogContent as-child>
                <motion.div
                    class="fixed inset-y-0 left-0 z-overlay flex w-72 max-w-[85vw] flex-col overflow-hidden border-r border-line bg-surface pt-safe pb-safe text-fg shadow-overlay outline-none"
                    :class="!panel.open.value && 'pointer-events-none'"
                    :style="{ x: panelX }"
                >
                    <DialogTitle class="sr-only">{{ title }}</DialogTitle>
                    <DialogDescription class="sr-only">{{ title }}</DialogDescription>
                    <slot />
                </motion.div>
            </DialogContent>
        </DialogPortal>
    </DialogRoot>
</template>
