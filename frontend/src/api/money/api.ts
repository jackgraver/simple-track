import { apiClient } from "~/api/client";

export type ContributionRule = {
    ID: number;
    investment_account_type_id: number;
    year: number;
    annual_limit: number;
};

export type ContributionStatus = {
    year: number;
    annual_limit: number;
    contributed: number;
    remaining: number;
};

export type ContributionRoom = {
    eligible_from_year: number;
    earned_room: number;
    contributed: number;
    remaining: number;
};

export type InvestmentAccountType = {
    ID: number;
    name: string;
    contribution_start_year?: number;
    rules?: ContributionRule[];
    contribution_status: ContributionStatus[];
    contribution_room?: ContributionRoom;
};

export type InvestmentAccount = {
    ID: number;
    name: string;
    investment_account_type_id?: number;
    investment_account_type?: InvestmentAccountType;
    current_balance: number;
    total_deposits: number;
    profit: number;
    contribution_status: ContributionStatus[];
    contribution_room?: ContributionRoom;
};

export type InvestmentDeposit = {
    ID: number;
    account_id: number;
    amount: number;
    date: string;
};

export async function listInvestmentAccounts() {
    const response = await apiClient.get<{ accounts: InvestmentAccount[] }>(
        "/money/investments",
    );
    return response.data.accounts;
}

export async function createInvestmentAccount(payload: {
    name: string;
    investment_account_type_id?: number;
    current_balance: number;
}) {
    await apiClient.post("/money/investments", payload);
}

export async function updateInvestmentAccount(
    accountId: number,
    payload: { name?: string; current_balance?: number },
) {
    await apiClient.patch(`/money/investments/${accountId}`, payload);
}

export async function deleteInvestmentAccount(accountId: number) {
    await apiClient.delete(`/money/investments/${accountId}`);
}

export async function listInvestmentAccountTypes() {
    const response = await apiClient.get<{
        account_types: InvestmentAccountType[];
    }>("/money/investments/account-types");
    return response.data.account_types;
}

export async function createInvestmentAccountType(payload: {
    name: string;
    contribution_start_year?: number;
}) {
    await apiClient.post("/money/investments/account-types", payload);
}

export async function updateInvestmentAccountType(
    accountTypeId: number,
    payload: { name: string; contribution_start_year?: number },
) {
    await apiClient.patch(
        `/money/investments/account-types/${accountTypeId}`,
        payload,
    );
}

export async function deleteInvestmentAccountType(accountTypeId: number) {
    await apiClient.delete(`/money/investments/account-types/${accountTypeId}`);
}

export async function listInvestmentDeposits(accountId: number) {
    const response = await apiClient.get<{ deposits: InvestmentDeposit[] }>(
        `/money/investments/${accountId}/deposits`,
    );
    return response.data.deposits;
}

export async function createInvestmentDeposit(
    accountId: number,
    payload: { amount: number; date: string },
) {
    await apiClient.post(`/money/investments/${accountId}/deposits`, payload);
}

export async function deleteInvestmentDeposit(
    accountId: number,
    depositId: number,
) {
    await apiClient.delete(
        `/money/investments/${accountId}/deposits/${depositId}`,
    );
}

export async function upsertContributionRule(
    accountTypeId: number,
    year: number,
    annualLimit: number,
) {
    await apiClient.put(
        `/money/investments/account-types/${accountTypeId}/contribution-rules/${year}`,
        { annual_limit: annualLimit },
    );
}

export async function deleteContributionRule(
    accountTypeId: number,
    year: number,
) {
    await apiClient.delete(
        `/money/investments/account-types/${accountTypeId}/contribution-rules/${year}`,
    );
}
