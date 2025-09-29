import { ref } from "vue";

type DialogOptions = {
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
};

const dialog = ref<DialogOptions | null>(null);
let resolver: ((value: boolean) => void) | null = null;

export function useDialog() {
    function confirm(options: DialogOptions): Promise<boolean> {
        dialog.value = options;

        return new Promise((resolve) => {
            resolver = resolve;
        });
    }

    function resolve(result: boolean) {
        if (resolver) {
            resolver(result);
            resolver = null;
        }
        dialog.value = null;
    }

    return { dialog, confirm, resolve };
}

// Singleton
export const dialogManager = useDialog();
