import axios from "axios";

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

const axiosGenericStatusRe = /^Request failed with status code \d+$/;

export function getApiErrorMessage(error: unknown, fallback: string): string {
    if (axios.isAxiosError(error)) {
        const data = error.response?.data;
        if (isRecord(data)) {
            const apiErr = data.error;
            if (typeof apiErr === "string" && apiErr.trim() !== "") return apiErr;
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
