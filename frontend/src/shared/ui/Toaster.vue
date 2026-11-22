<script setup lang="ts">
// Reka Toast gives the live region, pause on hover/focus, swipe to dismiss and the F8 hotkey;
// the look is ours. `toast.push` (composables/toast/useToast.ts) is the API; ToastContainer feeds this.
import { ToastAction, ToastClose, ToastDescription, ToastProvider, ToastRoot, ToastViewport } from "reka-ui";
import { AlertCircle, CheckCircle2, Info, X } from "lucide-vue-next";
import type { ToastItem } from "./toastTypes";

// `tabBar`: the mobile tab bar is showing, so sit above it. `top`: focus pages (gym logging) keep their
// primary button at the bottom, so toasts go to the top on mobile instead of covering it.
defineProps<{ toasts: ToastItem[]; tabBar?: boolean; top?: boolean }>();
const emit = defineEmits<{ (e: "dismiss", id: number): void; (e: "action", id: number): void }>();

const icons = { success: CheckCircle2, error: AlertCircle, info: Info };
const iconTone = { success: "text-success", error: "text-danger", info: "text-info" } as const;

function onOpen(t: ToastItem, open: boolean) {
    if (!open) emit("dismiss", t.id);
}
</script>

<template>
    <ToastProvider swipe-direction="right" label="Notifications">
        <ToastRoot
            v-for="t in toasts"
            :key="t.id"
            :open="t.open"
            :duration="t.duration"
            :type="t.tone === 'error' ? 'foreground' : 'background'"
            class="pointer-events-auto flex items-center gap-3 rounded-card border border-line bg-surface p-3 text-sm text-fg shadow-overlay data-[state=closed]:animate-toast-out data-[state=open]:animate-toast-in data-[swipe=cancel]:translate-x-0 data-[swipe=cancel]:transition-transform data-[swipe=end]:translate-x-(--reka-toast-swipe-end-x) data-[swipe=end]:transition-transform data-[swipe=move]:translate-x-(--reka-toast-swipe-move-x)"
            @update:open="(o: boolean) => onOpen(t, o)"
        >
            <component :is="icons[t.tone]" aria-hidden="true" class="size-5 shrink-0" :class="iconTone[t.tone]" />
            <ToastDescription class="m-0 min-w-0 flex-1">{{ t.message }}</ToastDescription>
            <ToastAction v-if="t.action" as-child :alt-text="`${t.action.label}: ${t.message}`">
                <button
                    type="button"
                    class="min-h-touch shrink-0 cursor-pointer rounded-control px-3 py-0 text-sm font-medium text-accent hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-accent lg:min-h-control-sm"
                    @click="emit('action', t.id)"
                >
                    {{ t.action.label }}
                </button>
            </ToastAction>
            <ToastClose as-child>
                <button
                    type="button"
                    aria-label="Dismiss"
                    class="flex size-8 shrink-0 cursor-pointer items-center justify-center rounded-control text-fg-subtle hover:bg-surface-3 hover:text-fg focus-visible:outline-2 focus-visible:outline-accent"
                >
                    <X aria-hidden="true" class="size-4" />
                </button>
            </ToastClose>
        </ToastRoot>
        <ToastViewport
            class="pointer-events-none fixed z-toast m-0 flex list-none flex-col gap-2 p-0 outline-none inset-x-4 lg:inset-x-auto lg:right-4 lg:bottom-4 lg:w-96"
            :class="[
                top ? 'max-lg:top-3 max-lg:pt-safe' : tabBar ? 'max-lg:bottom-above-tabbar' : 'max-lg:bottom-4 max-lg:pb-safe',
                top ? 'lg:bottom-4' : '',
            ]"
        />
    </ToastProvider>
</template>
