<script setup lang="ts">
// One overlay: a bottom sheet below `lg`, a centred dialog from `lg`. Reka Dialog supplies the focus trap
// and restore, Esc, scroll lock and ARIA; the look, the sheet and swipe-to-close are ours.
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from "reka-ui";
import { X } from "lucide-vue-next";
import { computed, ref } from "vue";
import IconButton from "./IconButton.vue";
import { useIsDesktop } from "./useIsDesktop";

const open = defineModel<boolean>("open", { required: true });
const props = withDefaults(
    defineProps<{
        title: string;
        description?: string;
        presentation?: "auto" | "dialog" | "sheet";
        size?: "sm" | "md" | "lg";
        dismissible?: boolean;
    }>(),
    { presentation: "auto", size: "md", dismissible: true },
);

const isDesktop = useIsDesktop();
const sheet = computed(() => (props.presentation === "auto" ? !isDesktop.value : props.presentation === "sheet"));

const width = { sm: "w-96", md: "w-lg", lg: "w-2xl" } as const;

function guard(e: Event) {
    if (!props.dismissible) e.preventDefault();
}

// Swipe down on the handle/header to close. A pointer handler: Reka has no sheet gesture.
const dragY = ref(0);
const dragging = ref(false);
let pressed = false;
let startY = 0;
let startT = 0;
// Capture the pointer only once it moves, so a plain tap on the ✕ still clicks.
function onDown(e: PointerEvent) {
    if (!sheet.value || !props.dismissible) return;
    pressed = true;
    startY = e.clientY;
    startT = e.timeStamp;
}
function onMove(e: PointerEvent) {
    if (!pressed) return;
    const dy = e.clientY - startY;
    if (!dragging.value && Math.abs(dy) > 4) {
        dragging.value = true;
        (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    }
    if (dragging.value) dragY.value = Math.max(0, dy);
}
function onUp(e: PointerEvent) {
    pressed = false;
    if (!dragging.value) return;
    dragging.value = false;
    const velocity = dragY.value / Math.max(1, e.timeStamp - startT); // px per ms
    if (dragY.value > 96 || (dragY.value > 40 && velocity > 0.6)) open.value = false;
    else dragY.value = 0;
}
function afterClose(isOpen: boolean) {
    open.value = isOpen;
    if (!isOpen) dragY.value = 0;
}
</script>

<template>
    <DialogRoot :open="open" @update:open="afterClose">
        <DialogPortal>
            <DialogOverlay
                class="fixed inset-0 z-overlay bg-overlay data-[state=closed]:animate-overlay-out data-[state=open]:animate-overlay-in"
            />
            <DialogContent
                class="fixed z-overlay flex flex-col overflow-hidden border border-line bg-surface text-fg shadow-overlay outline-none"
                :class="
                    sheet
                        ? 'inset-x-0 bottom-0 max-h-[90dvh] rounded-t-dialog border-b-0 pb-safe data-[state=closed]:animate-sheet-out data-[state=open]:animate-sheet-in'
                        : [
                              'top-[10vh] left-1/2 max-h-[80vh] max-w-[90vw] -translate-x-1/2 rounded-dialog data-[state=closed]:animate-dialog-out data-[state=open]:animate-dialog-in',
                              width[size],
                          ]
                "
                :style="sheet ? { '--sheet-drag': `${dragY}px`, transform: dragY ? `translateY(${dragY}px)` : undefined, transition: dragging ? 'none' : 'transform 150ms ease-out' } : undefined"
                @escape-key-down="guard"
                @pointer-down-outside="guard"
                @interact-outside="guard"
            >
                <div
                    v-if="sheet && dismissible"
                    class="flex shrink-0 touch-none justify-center pt-2"
                    @pointerdown="onDown"
                    @pointermove="onMove"
                    @pointerup="onUp"
                    @pointercancel="onUp"
                >
                    <span aria-hidden="true" class="h-1 w-10 rounded-full bg-line-strong"></span>
                </div>
                <header
                    class="flex shrink-0 items-center justify-between gap-3 border-b border-line px-4 py-2 lg:px-5"
                    :class="sheet && dismissible ? 'touch-none' : ''"
                    @pointerdown="onDown"
                    @pointermove="onMove"
                    @pointerup="onUp"
                    @pointercancel="onUp"
                >
                    <DialogTitle class="m-0 min-w-0 flex-1 truncate text-base font-semibold">{{ title }}</DialogTitle>
                    <IconButton v-if="dismissible" label="Close" size="sm" @click="open = false"><X /></IconButton>
                </header>
                <DialogDescription v-if="!description" class="sr-only">{{ title }}</DialogDescription>
                <DialogDescription
                    v-else
                    class="m-0 px-4 pt-4 text-sm text-fg-muted lg:px-5"
                    :class="$slots.default ? '' : 'pb-4'"
                >
                    {{ description }}
                </DialogDescription>
                <div v-if="$slots.default" class="min-h-0 flex-1 overflow-y-auto p-4 text-left lg:p-5"><slot /></div>
                <footer
                    v-if="$slots.footer"
                    class="flex shrink-0 flex-col-reverse gap-2 border-t border-line px-4 py-3 lg:flex-row lg:justify-end lg:px-5"
                >
                    <slot name="footer" />
                </footer>
            </DialogContent>
        </DialogPortal>
    </DialogRoot>
</template>
