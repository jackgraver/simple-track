export function useAPIGet<T>(endpoint: string, options: any = {}) {
    const config = useRuntimeConfig();

    const baseUrl = config.public.apiBase || "http://localhost:8080";

    return useFetch<T>(`${baseUrl}/${endpoint}`, {
        ...options,
    });
}

export async function useAPIPost<T>(
    endpoint: string,
    body: Record<string, any>,
    options: any = {},
) {
    const config = useRuntimeConfig();
    const baseUrl = config.public.apiBase || "http://localhost:8080";

    const { data, error, status } = await useFetch<T>(
        `${baseUrl}/${endpoint}`,
        {
            method: "POST",
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
