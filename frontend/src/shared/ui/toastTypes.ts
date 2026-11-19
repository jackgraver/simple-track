export type ToastTone = "success" | "error" | "info";
export type ToastAction = { label: string; onClick: () => void };
export type ToastItem = {
    id: number;
    message: string;
    tone: ToastTone;
    action?: ToastAction;
    duration: number;
    open: boolean;
};
