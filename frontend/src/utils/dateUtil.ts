export function formatDate(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getDate(); // local day
    const month = date.toLocaleString("en-US", { month: "short" });

    let suffix = "th";
    if (day % 10 === 1 && day !== 11) suffix = "st";
    else if (day % 10 === 2 && day !== 12) suffix = "nd";
    else if (day % 10 === 3 && day !== 13) suffix = "rd";

    return `${month} ${day}${suffix}`;
}

export function formatDateLong(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getDate(); // local day
    const month = date.toLocaleString("en-US", { month: "long" });
    const year = date.getFullYear();
    const dayOfWeek = date.toLocaleDateString("en-US", { weekday: "long" });

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
    return date.toLocaleString("en-US", { month: "long" });
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
    return date.toLocaleDateString("en-US", { weekday: "long" }); // local
}
