import { useQuery } from "@tanstack/vue-query";
import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { getApiErrorMessage } from "~/api/httpError";
import { useOptimisticMutation, useServerMutation } from "~/api/optimistic";
import * as api from "./api";
import { compareBalances, compareTransactions, removeById, restoreById } from "./cache";
import { moneyKeys } from "./keys";
import type { AccountBalance, AccountTransaction } from "./types";

export { moneyKeys };

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
    return useQuery({
        queryKey: moneyKeys.netWorth,
        queryFn: api.fetchNetWorthHistory,
    });
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
 * once, so every mutation refreshes the whole money cache on settle (skipped
 * while an optimistic sibling on an overlapping key is still pending).
 * Wait-for-server: `useServerMutation` toasts the API error once (the global
 * handler is suppressed), so call sites only catch to stop their own flow.
 * `failureMessage` is the toast fallback when the API sends no message.
 */
export function useMoneyMutation<TArgs, TResult = unknown>(
    fn: (args: TArgs) => Promise<TResult>,
    failureMessage = "Couldn’t save",
) {
    return useServerMutation<TArgs, TResult>({
        mutationKey: moneyKeys.mutations.write,
        mutationFn: fn,
        errorMessage: (err) => getApiErrorMessage(err, failureMessage),
        invalidate: () => [moneyKeys.all],
    });
}

export type DeleteRowVars = { accountId: number; id: number };

/**
 * Delete a transaction: the row leaves the list now; account totals, net worth
 * and room stay as they are until the refetch on settle (partial optimistic).
 */
export function useDeleteTransaction() {
    return useOptimisticMutation<DeleteRowVars, void>({
        mutationKey: moneyKeys.mutations.deleteTransaction,
        mutationFn: ({ accountId, id }) => api.deleteTransaction(accountId, id),
        optimistic: ({ accountId }) => [
            {
                queryKey: moneyKeys.transactions(accountId),
                apply: (old: AccountTransaction[], v: DeleteRowVars) => removeById(old, v.id),
                revert: (cur: AccountTransaction[], v: DeleteRowVars, snap: AccountTransaction[]) =>
                    restoreById(cur, snap, v.id, compareTransactions),
            },
        ],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to delete transaction"),
        invalidate: () => [moneyKeys.all],
    });
}

/** Delete a balance: the history row leaves now; the current balance and charts refresh on settle. */
export function useDeleteBalance() {
    return useOptimisticMutation<DeleteRowVars, void>({
        mutationKey: moneyKeys.mutations.deleteBalance,
        mutationFn: ({ accountId, id }) => api.deleteBalance(accountId, id),
        optimistic: ({ accountId }) => [
            {
                queryKey: moneyKeys.balances(accountId),
                apply: (old: AccountBalance[], v: DeleteRowVars) => removeById(old, v.id),
                revert: (cur: AccountBalance[], v: DeleteRowVars, snap: AccountBalance[]) =>
                    restoreById(cur, snap, v.id, compareBalances),
            },
        ],
        errorMessage: (err) => getApiErrorMessage(err, "Failed to delete balance"),
        invalidate: () => [moneyKeys.all],
    });
}
