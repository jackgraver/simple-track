export const moneyKeys = {
    all: ["money"] as const,
    accounts: (includeArchived: boolean) =>
        ["money", "accounts", { includeArchived }] as const,
    account: (id: number) => ["money", "account", id] as const,
    balances: (id: number) => ["money", "account", id, "balances"] as const,
    transactions: (id: number) =>
        ["money", "account", id, "transactions"] as const,
    netWorth: ["money", "net-worth"] as const,
    contributionTypes: ["money", "contribution-types"] as const,
    presets: ["money", "presets"] as const,
};
