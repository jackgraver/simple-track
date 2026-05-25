import type { ChartData, ChartOptions } from "chart.js";

function readChartTheme(): {
    textMuted: string;
    grid: string;
} {
    if (typeof document === "undefined") {
        return {
            textMuted: "hsl(0, 0%, 45%)",
            grid: "hsl(0, 0%, 80%)",
        };
    }
    const root = document.documentElement;
    const textMuted =
        getComputedStyle(root).getPropertyValue("--color-text-secondary").trim() ||
        "hsl(0, 0%, 45%)";
    const grid =
        getComputedStyle(root).getPropertyValue("--color-border").trim() ||
        "hsl(0, 0%, 80%)";
    return { textMuted, grid };
}

export function trackingDayKey(dateStr: string): string {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(dateStr);
    if (m) return `${m[1]}-${m[2]}-${m[3]}`;
    const d = new Date(dateStr);
    const y = d.getFullYear();
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${y}-${mo}-${da}`;
}

function formatShortLabel(dayKey: string): string {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(dayKey);
    if (!m) return dayKey;
    const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
    return d.toLocaleString("en-US", { month: "short", day: "numeric" });
}

function ymdAddDays(ymd: string, deltaDays: number): string {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(ymd);
    if (!m) return ymd;
    const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
    d.setDate(d.getDate() + deltaDays);
    const y = d.getFullYear();
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${y}-${mo}-${da}`;
}

/** One point per calendar day through `selectedYmd` (inclusive); latest row wins if duplicates. */
export function seriesThroughSelectedDay(
    rows: { date: string; value: number }[],
    selectedYmd: string,
): { labels: string[]; values: number[] } {
    const byDay = new Map<string, number>();
    for (const r of rows) {
        const k = trackingDayKey(r.date);
        if (k > selectedYmd) continue;
        byDay.set(k, r.value);
    }
    const keys = [...byDay.keys()].sort();
    return {
        labels: keys.map(formatShortLabel),
        values: keys.map((k) => byDay.get(k)!),
    };
}

/** Fixed window of `dayCount` calendar days ending on `selectedYmd`; missing days are 0. */
export function seriesLastDaysThroughSelectedDay(
    rows: { date: string; value: number }[],
    selectedYmd: string,
    dayCount = 14,
): { labels: string[]; values: number[] } {
    const byDay = new Map<string, number>();
    for (const r of rows) {
        byDay.set(trackingDayKey(r.date), r.value);
    }
    const keys: string[] = [];
    for (let i = dayCount - 1; i >= 0; i--) {
        keys.push(ymdAddDays(selectedYmd, -i));
    }
    return {
        labels: keys.map(formatShortLabel),
        values: keys.map((k) => byDay.get(k) ?? 0),
    };
}

const Y_AXIS_PADDING = 5;

function buildTrackingScaleOptions(
    datasetLabel: string,
    t: ReturnType<typeof readChartTheme>,
    values: number[],
    yMinFixed?: number,
) {
    let yMin: number | undefined = yMinFixed;
    let yMax: number | undefined;
    if (values.length > 0) {
        const dataMin = Math.min(...values);
        const dataMax = Math.max(...values);
        yMin = yMinFixed ?? dataMin - Y_AXIS_PADDING;
        yMax = dataMax + Y_AXIS_PADDING;
    }
    return {
        x: {
            ticks: { color: t.textMuted, maxRotation: 45 },
            grid: { color: t.grid },
        },
        y: {
            ticks: { color: t.textMuted },
            grid: { color: t.grid },
            ...(yMin != null ? { min: yMin } : {}),
            ...(yMax != null ? { max: yMax } : {}),
            title: {
                display: true,
                text: datasetLabel,
                color: t.textMuted,
            },
        },
    };
}

export function buildTrackingLineChart(
    labels: string[],
    values: number[],
    datasetLabel: string,
    lineColor: string,
): { data: ChartData<"line">; options: ChartOptions<"line"> } {
    const t = readChartTheme();
    const data: ChartData<"line"> = {
        labels,
        datasets: [
            {
                label: datasetLabel,
                data: values,
                borderColor: lineColor,
                backgroundColor: lineColor,
                tension: 0.2,
                fill: false,
                pointRadius: 3,
                pointHoverRadius: 5,
                borderWidth: 2,
            },
        ],
    };
    const options: ChartOptions<"line"> = {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
            legend: { labels: { color: t.textMuted } },
        },
        scales: buildTrackingScaleOptions(datasetLabel, t, values),
    };
    return { data, options };
}

export function buildTrackingBarChart(
    labels: string[],
    values: number[],
    datasetLabel: string,
    barColor: string,
): { data: ChartData<"bar">; options: ChartOptions<"bar"> } {
    const t = readChartTheme();
    const data: ChartData<"bar"> = {
        labels,
        datasets: [
            {
                label: datasetLabel,
                data: values,
                backgroundColor: barColor,
                borderColor: barColor,
                borderWidth: 0,
                borderRadius: 4,
            },
        ],
    };
    const options: ChartOptions<"bar"> = {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
            legend: { labels: { color: t.textMuted } },
        },
        scales: buildTrackingScaleOptions(datasetLabel, t, values, 0),
    };
    return { data, options };
}
