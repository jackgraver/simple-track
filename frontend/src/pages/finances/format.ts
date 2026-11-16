import type { AccountKind } from "~/api/money/types";

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

export const ui = {
    card: "rounded-lg border border-(--color-border) bg-firstBg p-4",
    input: "w-full rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary outline-none focus:border-textSecondary",
    label: "flex flex-col gap-1 text-xs font-medium text-textSecondary",
    btnPrimary:
        "rounded-md bg-green-700 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-green-600 disabled:cursor-not-allowed disabled:opacity-50",
    btnSecondary:
        "rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm text-textPrimary transition-colors hover:bg-thirdBg disabled:cursor-not-allowed disabled:opacity-50",
    btnDanger:
        "rounded-md px-3 py-2 text-sm text-(--color-cf-red) transition-colors hover:bg-secondBg disabled:opacity-50",
    positive: "text-emerald-400",
    negative: "text-(--color-cf-red)",
};
