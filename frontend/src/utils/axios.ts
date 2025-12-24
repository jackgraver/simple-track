import axios from 'axios';
import type { AxiosInstance } from 'axios';
import { config } from '~/config/env';

let _apiClient: AxiosInstance | null = null;

export function createAxiosInstance() {
    const instance = axios.create({
        baseURL: config.apiBase,
        headers: {
            'Content-Type': 'application/json',
        },
    });

    instance.interceptors.response.use(
        (response) => response,
        (error) => {
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

