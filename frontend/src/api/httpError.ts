import axios from "axios";

export type ApiErrorBody = { error: string };

export function isApiErrorBody(value: unknown): value is ApiErrorBody {
    return (
        typeof value === "object" &&
        value !== null &&
        typeof (value as Record<string, unknown>).error === "string"
    );
}

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

const axiosGenericStatusRe = /^Request failed with status code \d+$/;

export function getApiErrorMessage(error: unknown, fallback: string): string {
    if (axios.isAxiosError(error)) {
        const data = error.response?.data;
        if (isApiErrorBody(data) && data.error.trim() !== "") return data.error;
        if (isRecord(data)) {
            const msg = data.message;
            if (typeof msg === "string" && msg.trim() !== "") return msg;
        }
        if (typeof data === "string" && data.trim() !== "") return data;
        const m = error.message ?? "";
        if (m && !axiosGenericStatusRe.test(m)) return m;
        return fallback;
    }
    if (error instanceof Error && error.message) return error.message;
    return fallback;
}
