import axios, { type AxiosRequestConfig } from 'axios';
import { config } from '~/config/env';
import { markUnauthorized, notifyUnauthorized } from '~/composables/auth/session';

export const api = axios.create({
    baseURL: config.apiBase,
    headers: {
        "Content-Type": "application/json",
    },
    withCredentials: true,
    timeout: 10000,
});

export const apiClient = api;

export const apiGET = async <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    const response = await api.get<T>(url, config);
    return response.data as T;
};

export const apiPOST = async <T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
): Promise<T> => {
    const response = await api.post<T>(url, data, config);
    return response.data as T;
};

export const apiPUT = async <T = unknown>(
    url: string,
    data: unknown,
    config?: AxiosRequestConfig,
): Promise<T> => {
    const response = await api.put<T>(url, data, config);
    return response.data as T;
};

export const apiPATCH = async <T = unknown>(
    url: string,
    data: unknown,
    config?: AxiosRequestConfig,
): Promise<T> => {
    const response = await api.patch<T>(url, data, config);
    return response.data as T;
};

export const apiDELETE = async <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    const response = await api.delete<T>(url, config);
    return response.data as T;
};

function isAuthCredentialRequest(url: string): boolean {
    return (
        url.includes("/auth/me") ||
        url.includes("/auth/login") ||
        url.includes("/auth/register")
    );
}

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            markUnauthorized();
            const reqUrl = typeof error.config?.url === "string" ? error.config.url : "";
            if (!isAuthCredentialRequest(reqUrl)) {
                notifyUnauthorized();
            }
        }
        return Promise.reject(error);
    },
);
