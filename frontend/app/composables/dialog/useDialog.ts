import { ref, type VNode } from "vue";

export type DialogOptions = {
    title: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
};

type CustomDialogOptions = {
    title?: string;
    content: VNode; // arbitrary content
};

const dialog = ref<DialogOptions | CustomDialogOptions | null>(null);
let resolver: ((value: boolean) => void) | null = null;

export function useDialog() {
    // Confirm Dialog (predefined template)
    function confirm(options: DialogOptions): Promise<boolean> {
        dialog.value = options;
        return new Promise((resolve) => {
            resolver = resolve;
        });
    }

    // General-purpose custom dialog
    function custom(options: CustomDialogOptions): Promise<void> {
        dialog.value = options;
        return new Promise((resolve) => {
            resolver = () => resolve(); // resolve without boolean
        });
    }

    function resolve(result?: boolean) {
        if (resolver) resolver(result as any);
        resolver = null;
        dialog.value = null;
    }

    return { dialog, confirm, custom, resolve };
}

// Singleton
export const dialogManager = useDialog();
