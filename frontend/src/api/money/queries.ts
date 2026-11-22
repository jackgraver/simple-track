import { queryOptions, useMutation, useQuery, useQueryClient } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";
import * as api from "./api";
import { moneyKeys } from "./keys";

export { moneyKeys };

/** Single source for hooks and route prefetch. */
export const moneyQueries = {
    accounts: (includeArchived = false) =>
        queryOptions({
            queryKey: moneyKeys.accounts(includeArchived),
            queryFn: () => api.fetchAccounts(includeArchived),
        }),
    netWorth: () =>
        queryOptions({
            queryKey: moneyKeys.netWorth,
            queryFn: api.fetchNetWorthHistory,
        }),
};

export function useAccounts(includeArchived: MaybeRefOrGetter<boolean> = false) {
    return useQuery({
        queryKey: computed(() => moneyKeys.accounts(toValue(includeArchived))),
        queryFn: () => api.fetchAccounts(toValue(includeArchived)),
        placeholderData: (previous) => previous,
    });
}

export function useAccount(id: MaybeRefOrGetter<number>) {
    return useQuery({
        queryKey: computed(() => moneyKeys.account(toValue(id))),
        queryFn: () => api.fetchAccount(toValue(id)),
        enabled: computed(() => Number.isFinite(toValue(id))),
    });
}

export function useBalances(id: MaybeRefOrGetter<number>) {
    return useQuery({
        queryKey: computed(() => moneyKeys.balances(toValue(id))),
        queryFn: () => api.fetchBalances(toValue(id)),
        enabled: computed(() => Number.isFinite(toValue(id))),
    });
}

export function useTransactions(id: MaybeRefOrGetter<number>) {
    return useQuery({
        queryKey: computed(() => moneyKeys.transactions(toValue(id))),
        queryFn: () => api.fetchTransactions(toValue(id)),
        enabled: computed(() => Number.isFinite(toValue(id))),
    });
}

export function useNetWorthHistory() {
    return useQuery(moneyQueries.netWorth());
}

export function useContributionTypes() {
    return useQuery({
        queryKey: moneyKeys.contributionTypes,
        queryFn: api.fetchContributionTypes,
    });
}

export function useContributionPresets() {
    return useQuery({
        queryKey: moneyKeys.presets,
        queryFn: api.fetchContributionPresets,
        staleTime: Infinity,
    });
}

/**
 * Any money write can change balances, net worth and contribution room at
 * once, so every mutation simply refreshes the whole money cache.
 */
export function useMoneyMutation<TArgs, TResult = unknown>(
    fn: (args: TArgs) => Promise<TResult>,
) {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: fn,
        onSuccess: () => {
            void queryClient.invalidateQueries({ queryKey: moneyKeys.all });
        },
    });
}
