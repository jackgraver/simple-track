import axios, { type AxiosRequestConfig } from 'axios';
import { config } from '~/config/env';

const BASE_URL = config.apiBase || "/api";

export const api = axios.create({
    baseURL: BASE_URL,
    headers: {
        "Content-Type": "application/json",
    },
    withCredentials: true,
    timeout: 10000,
});

export const apiClient = api;

export const apiGET = async <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    const response = await api.get<T>(url, config);
    return response.data as T;
};

export const apiPOST = async <T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig,
): Promise<T> => {
    const response = await api.post<T>(url, data, config);
    return response.data as T;
};

export const apiPUT = async <T = any>(url: string, data: any): Promise<T> => {
    const response = await api.put<T>(url, data);
    return response.data as T;
};

export const apiPATCH = async <T = any>(url: string, data: any): Promise<T> => {
    const response = await api.patch<T>(url, data);
    return response.data as T;
};

export const apiDELETE = async <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    const response = await api.delete<T>(url, config);
    return response.data as T;
};

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            //TODO: logout / refresh / redirect
        }
        return Promise.reject(error);
    },
);
