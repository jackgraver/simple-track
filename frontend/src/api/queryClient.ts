import { MutationCache, QueryCache, QueryClient } from "@tanstack/vue-query";
import axios from "axios";

import { getApiErrorMessage } from "~/api/httpError";
import { toast } from "~/composables/toast/useToast";

/** staleTime tiers (DESIGN.md §3). Hooks pick one explicitly when the default is wrong. */
export const STALE = { hot: 30_000, default: 60_000, catalog: 10 * 60_000, static: Infinity } as const;

const GC_TIME = 24 * 60 * 60_000;
const MAX_QUERY_RETRIES = 2;

/** HTTP status of an axios error, or undefined for network errors and non-axios errors. */
function statusOf(err: unknown): number | undefined {
    return axios.isAxiosError(err) ? err.response?.status : undefined;
}

/** A 4xx response: the request itself is wrong, so repeating it cannot help. */
export function isClientError(err: unknown): boolean {
    const status = statusOf(err);
    return status !== undefined && status >= 400 && status < 500;
}

/** Query retry policy: up to 2 retries, never for 4xx. */
export function shouldRetryQuery(failureCount: number, err: unknown): boolean {
    return failureCount < MAX_QUERY_RETRIES && !isClientError(err);
}

type Notify = (message: string) => void;

/** Factory so tests can inject a notifier. The app uses the `queryClient` singleton below. */
export function createQueryClient(notifyError: Notify = (m) => toast.push(m, "error")): QueryClient {
    return new QueryClient({
        queryCache: new QueryCache({
            onError: (err, query) => {
                // A failed background refetch with cached data on screen gets a quiet toast. A cold-load
                // error is rendered by the page's own error state. 401 redirects to sign-in instead.
                if (statusOf(err) === 401) return;
                if (query.state.data !== undefined && !query.meta?.silent) notifyError("Couldn’t refresh");
            },
        }),
        mutationCache: new MutationCache({
            onError: (err, _vars, _ctx, mutation) => {
                if (statusOf(err) === 401 || mutation.meta?.handled) return;
                notifyError(getApiErrorMessage(err, "Couldn’t save"));
            },
        }),
        defaultOptions: {
            queries: {
                staleTime: STALE.default,
                gcTime: GC_TIME,
                retry: shouldRetryQuery,
                refetchOnWindowFocus: true,
                refetchOnReconnect: true,
            },
            // Idempotent hooks opt in to retries explicitly. 'always': a wait-for-server write (a form)
            // fails at once offline with "You're offline" and keeps its inputs, instead of pausing
            // with a spinner until reconnect. Optimistic writes opt back into 'online' (queued).
            mutations: { retry: 0, networkMode: "always" },
        },
    });
}

export const queryClient = createQueryClient();
