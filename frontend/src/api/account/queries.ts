import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import { computed } from "vue";
import { isAdmin } from "~/composables/auth/session";
import { createSetupLink, createUser, fetchAuthEvents, fetchUsers, setUserDisabled } from "./api";

export const accountKeys = {
    users: ["admin", "users"] as const,
    events: ["admin", "auth-events"] as const,
};

export function useAdminUsers() {
    return useQuery({ queryKey: accountKeys.users, queryFn: fetchUsers, enabled: computed(() => isAdmin.value) });
}

export function useAuthEvents() {
    return useInfiniteQuery({
        queryKey: accountKeys.events,
        queryFn: ({ pageParam }) => fetchAuthEvents(pageParam || undefined),
        initialPageParam: 0,
        getNextPageParam: (last) => (last.length < 50 ? undefined : last[last.length - 1]?.id),
        enabled: computed(() => isAdmin.value),
    });
}

/** Account changes also write audit events, so both lists refresh. */
function useInvalidateAccounts() {
    const qc = useQueryClient();
    return () => {
        void qc.invalidateQueries({ queryKey: accountKeys.users });
        void qc.invalidateQueries({ queryKey: accountKeys.events });
    };
}

export function useCreateUser() {
    const invalidate = useInvalidateAccounts();
    return useMutation({ mutationFn: createUser, onSettled: invalidate });
}

export function useCreateSetupLink() {
    const invalidate = useInvalidateAccounts();
    return useMutation({ mutationFn: createSetupLink, onSettled: invalidate });
}

export function useSetUserDisabled() {
    const invalidate = useInvalidateAccounts();
    return useMutation({
        mutationFn: ({ id, disabled }: { id: number; disabled: boolean }) => setUserDisabled(id, disabled),
        onSettled: invalidate,
    });
}
