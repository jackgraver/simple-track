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
    /** Limit rule in effect this year (also set in the opening year). */
    annual_limit: number | null;
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

/** The shared rules of a type: editable by its creator and the admin. */
export type ContributionTypeRules = {
    name: string;
    eligibility_age: number | null;
    carry_forward_cap: number | null;
    lifetime_limit: number | null;
    withdrawals_restore_room: boolean;
};

/** The signed-in user's own figures for a type. */
export type ContributionRoomSettings = {
    start_year: number | null;
    opening_room_year: number | null;
    opening_room: number | null;
};

export type ContributionTypeSettings = ContributionTypeRules & ContributionRoomSettings;

export type ContributionType = ContributionTypeSettings & {
    ID: number;
    created_by: number;
    /** True for the type's creator and the admin (the rules and limits are shared). */
    can_edit: boolean;
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
