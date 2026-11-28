import { apiClient } from "~/api/client";
import type {
    Account,
    AccountBalance,
    AccountSettings,
    AccountTransaction,
    ContributionPreset,
    ContributionType,
    ContributionRoomSettings,
    ContributionTypeRules,
    ContributionTypeSettings,
    NetWorthPoint,
    TransactionKind,
} from "./types";

// ---- Accounts ----

export async function fetchAccounts(includeArchived = false) {
    const { data } = await apiClient.get<{ accounts: Account[] }>(
        "/money/accounts",
        { params: includeArchived ? { archived: "true" } : undefined },
    );
    return data.accounts;
}

export async function fetchAccount(accountId: number) {
    const { data } = await apiClient.get<{ account: Account }>(
        `/money/accounts/${accountId}`,
    );
    return data.account;
}

export async function createAccount(
    payload: AccountSettings & {
        opening_balance?: number | null;
        balance_date?: string;
    },
) {
    const { data } = await apiClient.post<{ account: Account }>(
        "/money/accounts",
        payload,
    );
    return data.account;
}

export async function updateAccount(accountId: number, payload: AccountSettings) {
    await apiClient.put(`/money/accounts/${accountId}`, payload);
}

export async function deleteAccount(accountId: number) {
    await apiClient.delete(`/money/accounts/${accountId}`);
}

// ---- Balances ----

export async function fetchBalances(accountId: number) {
    const { data } = await apiClient.get<{ balances: AccountBalance[] }>(
        `/money/accounts/${accountId}/balances`,
    );
    return data.balances;
}

export async function setBalance(
    accountId: number,
    payload: { balance: number; date: string },
) {
    await apiClient.put(`/money/accounts/${accountId}/balances`, payload);
}

export async function setBalances(payload: {
    date: string;
    balances: { account_id: number; balance: number }[];
}) {
    await apiClient.put("/money/balances", payload);
}

export async function deleteBalance(accountId: number, balanceId: number) {
    await apiClient.delete(`/money/accounts/${accountId}/balances/${balanceId}`);
}

export async function fetchNetWorthHistory() {
    const { data } = await apiClient.get<{ points: NetWorthPoint[] }>(
        "/money/net-worth/history",
    );
    return data.points;
}

// ---- Transactions ----

export async function fetchTransactions(accountId: number) {
    const { data } = await apiClient.get<{ transactions: AccountTransaction[] }>(
        `/money/accounts/${accountId}/transactions`,
    );
    return data.transactions;
}

export async function createTransaction(
    accountId: number,
    payload: {
        kind: TransactionKind;
        amount: number;
        date: string;
        note: string;
        /** Also move the balance by the amount, atomically on the server. */
        update_balance?: boolean;
    },
) {
    await apiClient.post(`/money/accounts/${accountId}/transactions`, payload);
}

export async function deleteTransaction(accountId: number, transactionId: number) {
    await apiClient.delete(
        `/money/accounts/${accountId}/transactions/${transactionId}`,
    );
}

// ---- Contribution types ----

export async function fetchContributionTypes() {
    const { data } = await apiClient.get<{
        contribution_types: ContributionType[];
        birth_year: number | null;
    }>("/money/contribution-types");
    return data;
}

export async function fetchContributionPresets() {
    const { data } = await apiClient.get<{ presets: ContributionPreset[] }>(
        "/money/contribution-types/presets",
    );
    return data.presets;
}

export async function createContributionTypeFromPreset(payload: {
    key: string;
    name?: string;
    start_year?: number | null;
    opening_room_year?: number | null;
    opening_room?: number | null;
}) {
    await apiClient.post("/money/contribution-types/presets", payload);
}

export async function createContributionType(payload: ContributionTypeSettings) {
    await apiClient.post("/money/contribution-types", payload);
}

/** Edits the shared rules (creator or admin only). */
export async function updateContributionType(
    id: number,
    payload: ContributionTypeRules,
) {
    await apiClient.put(`/money/contribution-types/${id}`, payload);
}

/** Saves the signed-in user's own start year and opening room for a type. */
export async function setContributionRoom(
    id: number,
    payload: ContributionRoomSettings,
) {
    await apiClient.put(`/money/contribution-types/${id}/room`, payload);
}

export async function deleteContributionType(id: number) {
    await apiClient.delete(`/money/contribution-types/${id}`);
}

export async function upsertContributionLimit(
    id: number,
    year: number,
    annualLimit: number,
) {
    await apiClient.put(`/money/contribution-types/${id}/limits/${year}`, {
        annual_limit: annualLimit,
    });
}

export async function deleteContributionLimit(id: number, year: number) {
    await apiClient.delete(`/money/contribution-types/${id}/limits/${year}`);
}

export async function saveBirthYear(birthYear: number) {
    await apiClient.patch("/auth/me", { birth_year: birthYear });
}
