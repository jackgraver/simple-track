export function useApiFetch<T>(endpoint: string, options: any = {}) {
    const config = useRuntimeConfig();

    const baseUrl = config.public.apiBase || "http://localhost:8080";

    return useFetch<T>(`${baseUrl}/${endpoint}`, {
        ...options,
    });
}
