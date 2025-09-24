export function formatDateShort(isoDate: string): string {
    const date = new Date(isoDate);
    const day = date.getUTCDate(); // day of month
    const month = date.toLocaleString("en-US", { month: "short" });

    // Determine ordinal suffix
    let suffix = "th";
    if (day % 10 === 1 && day !== 11) suffix = "st";
    else if (day % 10 === 2 && day !== 12) suffix = "nd";
    else if (day % 10 === 3 && day !== 13) suffix = "rd";

    return `${month} ${day}${suffix}`;
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

export function dayOfWeek(year: number, month: number, day: number): string {
    // JS months are 0-based, so subtract 1
    const date = new Date(year, month - 1, day);
    return date.toLocaleDateString("en-US", { weekday: "long" });
}
