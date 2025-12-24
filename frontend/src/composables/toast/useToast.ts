import { ref } from "vue";

type Toast = {
    id: number;
    message: string;
    type: "success" | "error" | "info";
};

const toasts = ref<Toast[]>([]);
let idCounter = 0;

export function useToast() {
    function push(message: string, type: Toast["type"] = "info") {
        const id = ++idCounter;
        toasts.value.push({ id, message, type });

        // Auto-remove after 3s
        setTimeout(() => {
            remove(id);
        }, 3000);
    }

    function remove(id: number) {
        toasts.value = toasts.value.filter((t) => t.id !== id);
    }

    return { toasts, push, remove };
}

// Singleton instance (so it's global)
export const toast = useToast();
