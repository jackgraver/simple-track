import {
    useQuery,
    type UseQueryOptions,
    type UseQueryResult,
} from "@tanstack/react-query";
// api.ts
import axios from "axios";

export const api = axios.create({
    baseURL: "http://localhost:8080/",
    withCredentials: true, // optional
});

async function fetcher<T>(url: string): Promise<T> {
    const res = await api.get<T>(url);
    return res.data;
}

export function useFetchQuery<T>(
    key: string | unknown[],
    url: string,
    options?: Omit<UseQueryOptions<T>, "queryKey" | "queryFn">,
): UseQueryResult<T> {
    return useQuery<T>({
        queryKey: Array.isArray(key) ? key : [key],
        queryFn: () => fetcher<T>(url),
        ...options,
    });
}
