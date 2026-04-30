import { Component, shallowRef } from "vue";

export type DialogOptions = {
    title: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
};

export type CustomDialogOptions = {
    title?: string;
    component: Component;
    componentProps?: Record<string, any>;
};

type DialogState =
    | { type: "confirm"; options: DialogOptions }
    | { type: "custom"; options: CustomDialogOptions };

const dialog = shallowRef<DialogState | null>(null);
let resolver: ((value: any) => void) | null = null;

export function useDialog() {
    /**
     * Open a confirm dialog.
     * @param options - The options for the confirm dialog
     * - title: The title of the dialog
     * - message: The message to display in the dialog
     * - confirmText: The text of the confirm button
     * - cancelText: The text of the cancel button
     * @returns Promise that resolves to the result of the confirm dialog
     */
    function confirm(options: DialogOptions): Promise<boolean | null> {
        dialog.value = { type: "confirm", options };
        return new Promise<boolean | null>((resolve) => {
            resolver = resolve;
        });
    }

    /**
     * Open dialog with custom component rendered inside it.
     * @param options - The options for the custom dialog
     * - title: The title of the dialog
     * - component: The component to render inside the dialog
     * - componentProps: The props to pass to the component
     * @returns Promise that resolves to the result of the custom dialog
     */
    function custom<T>(options: CustomDialogOptions): Promise<T | null> {
        dialog.value = { type: "custom", options };
        return new Promise<T | null>((resolve) => {
            resolver = resolve;
        });
    }

    /**
     * Resolve the dialog with a value.
     * @param value - The value to resolve the dialog with
     */
    function resolve<T>(value: T) {
        if (resolver) {
            resolver(value);
            resolver = null;
            dialog.value = null;
        }
    }

    /**
     * Cancel the dialog.
     */
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
