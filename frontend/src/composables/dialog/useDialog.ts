import { ref } from "vue";

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

type DialogState<T = any> = 
    | { type: "confirm"; options: DialogOptions }
    | { type: "custom"; options: CustomDialogOptions<T> };

const dialog = ref<DialogState | null>(null);
let resolver: ((value: any) => void) | null = null;

export function useDialog() {
    function confirm(options: DialogOptions): Promise<boolean | null> {
        dialog.value = { type: "confirm", options };
        return new Promise<boolean | null>((resolve) => {
            resolver = resolve;
        });
    }

    function custom<T>(options: CustomDialogOptions<T>): Promise<T | null> {
        dialog.value = { type: "custom", options };
        return new Promise<T | null>((resolve) => {
            resolver = resolve;
        });
    }

    function resolve<T>(value: T) {
        if (resolver) {
            resolver(value);
            resolver = null;
            dialog.value = null;
        }
    }

    function cancel() {
        if (resolver) {
            resolver(null);
            resolver = null;
            dialog.value = null;
        }
    }

    return { dialog, confirm, custom, resolve, cancel };
}

export const dialogManager = useDialog();
