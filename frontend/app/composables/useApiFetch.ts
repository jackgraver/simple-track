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
    watch: boolean = true,
) {
    const config = useRuntimeConfig();
    const baseUrl = config.public.apiBase || "http://localhost:8080";

    try {
        const data = await $fetch<T>(
            `${baseUrl}/${endpoint}`,
            {
                method: method,
                body,
                ...options,
            },
        );

        return { response: data, error: null, status: 200 };
    } catch (error: any) {
        console.error("Network error:", error);
        return { 
            response: null, 
            error: error, 
            status: error.status || 500 
        };
    }
}
