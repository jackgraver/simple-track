export type AccountKind =
    | "chequing"
    | "savings"
    | "cash"
    | "investment"
    | "property"
    | "credit_card"
    | "loan"
    | "other";

export type ContributionLimit = {
    ID: number;
    contribution_type_id: number;
    year: number;
    annual_limit: number;
};

export type RoomYear = {
    year: number;
    new_room: number;
    limit_year: number | null;
    is_opening: boolean;
    carried_in: number;
    restored: number;
    lifetime_capped: number;
    available: number;
    contributed: number;
    withdrawn: number;
    remaining: number;
    carried_out: number;
    over_contributed: boolean;
};

export type RoomSummary = {
    start_year: number;
    available: number;
    contributed_this_year: number;
    withdrawn_this_year: number;
    lifetime_contributed: number;
    lifetime_remaining: number | null;
    over_contributed: boolean;
    years: RoomYear[];
};

export type ContributionTypeSettings = {
    name: string;
    start_year: number | null;
    eligibility_age: number | null;
    carry_forward_cap: number | null;
    lifetime_limit: number | null;
    withdrawals_restore_room: boolean;
    opening_room_year: number | null;
    opening_room: number | null;
};

export type ContributionType = ContributionTypeSettings & {
    ID: number;
    limits: ContributionLimit[];
    accounts: { id: number; name: string }[];
    room: RoomSummary | null;
    room_issue?: string;
};

export type ContributionPreset = {
    key: string;
    name: string;
    description: string;
    eligibility_age: number | null;
    carry_forward_cap: number | null;
    lifetime_limit: number | null;
    withdrawals_restore_room: boolean;
    needs_start_year: boolean;
    suggest_opening_room: boolean;
    limits: { year: number; annual_limit: number }[];
};

export type AccountSettings = {
    name: string;
    kind: AccountKind;
    institution: string;
    notes: string;
    contribution_type_id: number | null;
    archived: boolean;
};

export type Account = {
    ID: number;
    name: string;
    kind: AccountKind;
    institution: string;
    notes: string;
    contribution_type_id: number | null;
    contribution_type?: { ID: number; name: string } | null;
    current_balance: number;
    archived_at: string | null;
    is_liability: boolean;
    balance_date: string | null;
    previous_balance: number | null;
    previous_balance_date: string | null;
    total_deposits: number;
    total_withdrawals: number;
    net_contributions: number;
    transaction_count: number;
    gain: number;
};

export type AccountBalance = {
    ID: number;
    account_id: number;
    date: string;
    balance: number;
};

export type TransactionKind = "deposit" | "withdrawal";

export type AccountTransaction = {
    ID: number;
    account_id: number;
    kind: TransactionKind;
    amount: number;
    date: string;
    note: string;
};

export type NetWorthPoint = {
    date: string;
    assets: number;
    liabilities: number;
    net_worth: number;
};
