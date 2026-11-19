import type { AccountKind, RoomYear } from "~/api/money/types";

const moneyFormatter = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
});

const compactFormatter = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    notation: "compact",
    maximumFractionDigits: 1,
});

export function formatMoney(value: number) {
    return moneyFormatter.format(value);
}

export function formatCompactMoney(value: number) {
    return compactFormatter.format(value);
}

export function formatSignedMoney(value: number) {
    const formatted = moneyFormatter.format(Math.abs(value));
    if (value > 0) return `+${formatted}`;
    if (value < 0) return `−${formatted}`;
    return formatted;
}

/** Backend dates are local midnight timestamps; keep only the day part. */
export function dayKey(date: string) {
    return date.slice(0, 10);
}

export function formatDay(date: string) {
    const [y, m, d] = dayKey(date).split("-").map(Number);
    return new Date(y, m - 1, d).toLocaleDateString("en-CA", {
        month: "short",
        day: "numeric",
        year: "numeric",
    });
}

export function daysSince(date: string) {
    const [y, m, d] = dayKey(date).split("-").map(Number);
    const then = new Date(y, m - 1, d).getTime();
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
    return Math.round((today - then) / 86_400_000);
}

export function formatAge(date: string) {
    const days = daysSince(date);
    if (days <= 0) return "today";
    if (days === 1) return "yesterday";
    if (days < 30) return `${days} days ago`;
    const months = Math.round(days / 30);
    if (months < 12) return `${months} mo ago`;
    return `${Math.round(days / 365)} yr ago`;
}

export function todayKey() {
    const now = new Date();
    const m = String(now.getMonth() + 1).padStart(2, "0");
    const d = String(now.getDate()).padStart(2, "0");
    return `${now.getFullYear()}-${m}-${d}`;
}

/** Parse a user-typed amount, tolerating "$", commas and spaces. */
export function parseAmount(raw: string | number | null | undefined): number | null {
    if (raw === null || raw === undefined) return null;
    const cleaned = String(raw).replace(/[$,\s]/g, "");
    if (cleaned === "" || cleaned === "-") return null;
    const value = Number(cleaned);
    return Number.isFinite(value) ? value : null;
}

export const accountKindLabels: Record<AccountKind, string> = {
    chequing: "Chequing",
    savings: "Savings",
    cash: "Cash",
    investment: "Investment",
    property: "Property",
    credit_card: "Credit card",
    loan: "Loan",
    other: "Other",
};

export const accountKindOrder: AccountKind[] = [
    "chequing",
    "savings",
    "cash",
    "investment",
    "property",
    "other",
    "credit_card",
    "loan",
];

export function isLiabilityKind(kind: AccountKind) {
    return kind === "credit_card" || kind === "loan";
}

/** Balances older than this are flagged as stale. */
export const STALE_BALANCE_DAYS = 35;

/** Text colour for a money change: good news green, bad news red (never the only signal: the sign is shown too). */
export function changeClass(good: boolean) {
    return good ? "text-success" : "text-danger";
}

export type YearAllowance = {
    /** The annual limit for the year. */
    limit: number;
    /** Withdrawals from the previous year added back this year. */
    restored: number;
    /** limit + restored: what can go in this year without dipping into older room. */
    allowance: number;
    contributed: number;
    /** Allowance left this year, never more than the total room left. */
    left: number;
    /** Contributed more than this year's allowance (uses carried-forward room). */
    over: boolean;
};

/**
 * The year's own allowance, separate from room carried forward from earlier
 * years. Null when the type has no annual limit (e.g. RRSP opening room only).
 */
export function yearAllowance(year: RoomYear): YearAllowance | null {
    if (year.annual_limit === null) return null;
    const allowance = year.annual_limit + year.restored;
    const left = Math.round(Math.min(allowance - year.contributed, year.remaining) * 100) / 100;
    return {
        limit: year.annual_limit,
        restored: year.restored,
        allowance,
        contributed: year.contributed,
        left,
        over: year.contributed > allowance + 0.005,
    };
}
