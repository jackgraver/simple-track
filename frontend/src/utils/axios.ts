import axios from 'axios';
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import { config } from '~/config/env';

let _apiClient: AxiosInstance | null = null;

function getTokenFromStorage(): string | null {
    if (typeof window === 'undefined') return null;
    return sessionStorage.getItem('auth_token');
}

export function createAxiosInstance() {
    const instance = axios.create({
        baseURL: config.apiBase,
        headers: {
            'Content-Type': 'application/json',
        },
    });

    instance.interceptors.request.use(
        (config: InternalAxiosRequestConfig) => {
            const token = getTokenFromStorage();
            if (token && config.headers) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        },
        (error) => {
            return Promise.reject(error);
        }
    );

    instance.interceptors.response.use(
        (response) => response,
        (error) => {
            if (error.response?.status === 401) {
                const token = getTokenFromStorage();
                if (token) {
                    sessionStorage.removeItem('auth_token');
                    sessionStorage.removeItem('auth_username');
                    if (typeof window !== 'undefined') {
                        window.location.href = '/signin';
                    }
                }
            }
            console.error('API Error:', error);
            return Promise.reject(error);
        }
    );

    return instance;
}

function getApiClient(): AxiosInstance {
    if (!_apiClient) {
        _apiClient = createAxiosInstance();
    }
    return _apiClient;
}

export const apiClient = new Proxy({} as AxiosInstance, {
    get(_target, prop) {
        const client = getApiClient();
        const value = client[prop as keyof AxiosInstance];
        if (typeof value === 'function') {
            return value.bind(client);
        }
        return value;
    }
});

