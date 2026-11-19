import { ref } from "vue";

import type { ToastAction, ToastItem, ToastTone } from "~/shared/ui/toastTypes";

export type { ToastAction, ToastTone };
export type ToastOptions = { message: string; tone?: ToastTone; action?: ToastAction; duration?: number };

const MAX_VISIBLE = 3;
const EXIT_MS = 200; // let the exit animation finish before the toast leaves the list

const toasts = ref<ToastItem[]>([]);
let idCounter = 0;

export function useToast() {
    /**
     * `push("Saved", "success")` or `push({ message, tone, action: { label: "Undo", onClick }, duration })`.
     * Errors stay longer, and so do toasts with an action (time to hit Undo).
     */
    function push(input: string | ToastOptions, type: ToastTone = "info") {
        const o: ToastOptions = typeof input === "string" ? { message: input, tone: type } : input;
        const tone = o.tone ?? "info";
        const duration = o.duration ?? (o.action ? 6000 : tone === "error" ? 5000 : 3000);
        const id = ++idCounter;
        toasts.value.push({ id, message: o.message, tone, action: o.action, duration, open: true });
        // At most 3 on screen: dismiss the oldest.
        const live = toasts.value.filter((t) => t.open);
        for (const old of live.slice(0, Math.max(0, live.length - MAX_VISIBLE))) dismiss(old.id);
        return id;
    }

    /** Close with the exit animation, then drop it. */
    function dismiss(id: number) {
        const t = toasts.value.find((x) => x.id === id);
        if (!t || !t.open) return;
        t.open = false;
        setTimeout(() => remove(id), EXIT_MS);
    }

    function remove(id: number) {
        toasts.value = toasts.value.filter((t) => t.id !== id);
    }

    return { toasts, push, dismiss, remove };
}

// Singleton instance (so it's global)
export const toast = useToast();
