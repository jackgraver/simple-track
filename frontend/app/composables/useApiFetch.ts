//TODO: grouped request

export function useAPIGet<T>(
    endpoint: string | Ref<string>,
    options: any = {},
) {
    const config = useRuntimeConfig();
    const baseUrl = config.public.apiBase || "http://localhost:8080";

    return useFetch<T>(endpoint, {
        baseURL: baseUrl,
        watch: [endpoint],
        ...options,
    });
}

export async function useAPIPost<T>(
    endpoint: string,
    method: "POST" | "DELETE" = "POST",
    body: Record<string, any>,
    options: any = {},
) {
    const config = useRuntimeConfig();
    const baseUrl = config.public.apiBase || "http://localhost:8080";

    const { data, error, status } = await useFetch<T>(
        `${baseUrl}/${endpoint}`,
        {
            method: method,
            body,
            ...options,
            onResponse({ response }) {
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`);
                }
            },
        },
    );

    if (error.value) {
        console.error("Network error:", error.value);
        return { data: null, error: error.value, status: status.value };
    }

    return { response: data.value, error: null, status: status.value };
}
