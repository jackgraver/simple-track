import { ref, watch, type Ref } from 'vue';
import { apiClient } from '~/utils/axios';
import { config } from '~/config/env';

export function useAPIGet<T>(
    endpoint: string | Ref<string>,
    options: any = {},
) {
    const data = ref<T | null>(null);
    const pending = ref(true);
    const error = ref<Error | null>(null);

    const fetchData = async () => {
        const url = typeof endpoint === 'string' ? endpoint : endpoint.value;
        pending.value = true;
        error.value = null;
        
        try {
            const response = await apiClient.get<T>(url);
            data.value = response.data;
        } catch (err: any) {
            error.value = err;
            data.value = null;
        } finally {
            pending.value = false;
        }
    };

    if (typeof endpoint === 'string') {
        fetchData();
    } else {
        watch(endpoint, fetchData, { immediate: true });
    }

    return { data, pending, error, refresh: fetchData };
}

export async function useAPIPost<T>(
    endpoint: string,
    method: "POST" | "DELETE" = "POST",
    body: Record<string, any>,
    options: any = {},
    watch: boolean = true,
) {
    try {
        const response = await apiClient.request<T>({
            url: endpoint,
            method,
            data: body,
            ...options,
        });

        return { response: response.data, error: null, status: response.status };
    } catch (err: any) {
        console.error("Network error:", err);
        return { 
            response: null, 
            error: err, 
            status: err.response?.status || 500 
        };
    }
}

