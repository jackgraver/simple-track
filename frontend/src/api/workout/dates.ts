/**
 * Day offsets (0 = today, 1 = yesterday, -1 = tomorrow, as the gym pages and the backend use them)
 * versus absolute `YYYY-MM-DD` dates. Cache keys and mutation variables use dates; the offset only
 * exists at the HTTP edge. It is recomputed from the date when a request is sent, so a request that
 * starts after midnight still targets the day it was made for.
 */

const MS_PER_DAY = 86_400_000;

function pad(n: number): string {
    return String(n).padStart(2, "0");
}

function localDateString(d: Date): string {
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** Local calendar date for a day offset, e.g. `offsetToDate(1)` is yesterday. */
export function offsetToDate(offset: number, now: Date = new Date()): string {
    const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() - offset);
    return localDateString(d);
}

/** Inverse of `offsetToDate` for the current local date. */
export function dateToOffset(date: string, now: Date = new Date()): number {
    const [y, m, d] = date.split("-").map(Number);
    const target = Date.UTC(y, m - 1, d);
    const today = Date.UTC(now.getFullYear(), now.getMonth(), now.getDate());
    return Math.round((today - target) / MS_PER_DAY);
}
