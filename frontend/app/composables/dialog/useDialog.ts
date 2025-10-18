import { ref, type VNode } from "vue";

export type DialogOptions = {
    title: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
};

export type CustomDialogOptions<T = any> = {
    title?: string;
    component: Component;
    props?: Record<string, any>;
};

const dialog = ref<DialogOptions | CustomDialogOptions | null>(null);
let resolver: ((value: any) => void) | null = null;

export function useDialog() {
    // Confirm Dialog (predefined template)
    function confirm(options: DialogOptions): Promise<boolean> {
        dialog.value = options;
        return new Promise((resolve) => {
            resolver = resolve;
        });
    }

    // General-purpose custom dialog
    function custom<T>(options: CustomDialogOptions<T>): Promise<T> {
        dialog.value = options;
        return new Promise<T>((resolve) => {
            resolver = resolve as (value: T) => void;
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
