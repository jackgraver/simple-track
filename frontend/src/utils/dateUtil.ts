// Reused formatters: building one per call (what toLocale*String does) is slow, and the activity
// heatmap formats ~700 labels per render.
const MONTH_SHORT = new Intl.DateTimeFormat("en-US", { month: "short" });
const MONTH_LONG = new Intl.DateTimeFormat("en-US", { month: "long" });
const WEEKDAY_LONG = new Intl.DateTimeFormat("en-US", { weekday: "long" });

export function formatDate(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getDate(); // local day
    const month = MONTH_SHORT.format(date);

    let suffix = "th";
    if (day % 10 === 1 && day !== 11) suffix = "st";
    else if (day % 10 === 2 && day !== 12) suffix = "nd";
    else if (day % 10 === 3 && day !== 13) suffix = "rd";

    return `${month} ${day}${suffix}`;
}

export function formatDateLong(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getDate(); // local day
    const month = MONTH_LONG.format(date);
    const year = date.getFullYear();
    const dayOfWeek = WEEKDAY_LONG.format(date);

    let suffix = "th";
    if (day % 10 === 1 && day !== 11) suffix = "st";
    else if (day % 10 === 2 && day !== 12) suffix = "nd";
    else if (day % 10 === 3 && day !== 13) suffix = "rd";

    return `${dayOfWeek} ${month} ${day}${suffix}, ${year}`;
}

export function formatDateShort(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getDate(); // local day

    let suffix = "th";
    if (day % 10 === 1 && day !== 11) suffix = "st";
    else if (day % 10 === 2 && day !== 12) suffix = "nd";
    else if (day % 10 === 3 && day !== 13) suffix = "rd";

    return `${day}${suffix}`;
}

export function monthName(isoDate: string): string {
    const date = new Date(isoDate);
    return MONTH_LONG.format(date);
}

export function isSameDay(a: string | Date, b: string | Date): boolean {
    const dateA = new Date(a);
    const dateB = new Date(b);

    return (
        dateA.getUTCFullYear() === dateB.getUTCFullYear() &&
        dateA.getUTCMonth() === dateB.getUTCMonth() &&
        dateA.getUTCDate() === dateB.getUTCDate()
    );
}

export function isSameMonth(month: number, b: string | Date): boolean {
    const dateB = new Date(b);

    return month - 1 === dateB.getUTCMonth();
}

export function dayOfWeek(dateStr: string): string {
    const date = new Date(dateStr);
    return WEEKDAY_LONG.format(date); // local
}

export function parseYmdDateQuery(raw: unknown): string | null {
    if (typeof raw !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(raw)) {
        return null;
    }
    return raw;
}

/** Local calendar date `YYYY-MM-DD` for today minus `offsetDays` (0 = today). */
export function ymdForDayOffset(offsetDays: number): string {
    const d = new Date();
    d.setDate(d.getDate() - offsetDays);
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
}

/** Matches backend `utils.ZerodTime(offset)`: local calendar day is today minus `offset` days (past → larger offset). */
export function dateAtDietDayOffset(offsetDays: number): Date {
    const now = new Date();
    return new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate() - offsetDays,
        12,
        0,
        0,
        0,
    );
}

/** Inverse of `dateAtDietDayOffset`: API offset from a local calendar date (`monthIndex` 0–11). */
export function dietDayOffsetFromLocalDate(
    year: number,
    monthIndex: number,
    day: number,
): number {
    const now = new Date();
    const todayMid = new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate(),
    );
    const pickedMid = new Date(year, monthIndex, day);
    return Math.round(
        (todayMid.getTime() - pickedMid.getTime()) / (24 * 60 * 60 * 1000),
    );
}

export function parseDietDayOffsetQuery(raw: unknown): number {
    const s = Array.isArray(raw) ? raw[0] : raw;
    if (typeof s !== "string") return 0;
    const v = Number.parseInt(s, 10);
    return Number.isNaN(v) ? 0 : v;
}

export function withDietDayOffsetQuery(
    offset: number,
    query: Record<string, string>,
): Record<string, string> {
    if (offset === 0) return query;
    return { ...query, offset: String(offset) };
}

/**
 * Where a week strip's arrows (or a swipe) land: the new week's Sunday, or today when that week
 * is the current one. `sundayOffset` is the shown week's Sunday as a day offset (0 = today,
 * positive = past); `step` is 1 for the next week, -1 for the previous one.
 */
export function weekLandingOffset(sundayOffset: number, step: 1 | -1): number {
    const sunday = sundayOffset - 7 * step;
    return sunday >= 0 && sunday <= 6 ? 0 : sunday;
}
