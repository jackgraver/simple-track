import type { ChartData, ChartOptions } from "chart.js";
import { chartTheme, type ChartTheme } from "~/shared/chart/theme";

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
    t: ChartTheme,
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
            ticks: { color: t.muted, maxRotation: 45 },
            grid: { color: t.grid },
        },
        y: {
            ticks: { color: t.muted },
            grid: { color: t.grid },
            ...(yMin != null ? { min: yMin } : {}),
            ...(yMax != null ? { max: yMax } : {}),
            title: {
                display: true,
                text: datasetLabel,
                color: t.muted,
            },
        },
    };
}

export function buildTrackingLineChart(
    labels: string[],
    values: number[],
    datasetLabel: string,
    lineColor?: string,
): { data: ChartData<"line">; options: ChartOptions<"line"> } {
    const t = chartTheme();
    const lineColor_ = lineColor ?? t.series[0];
    const data: ChartData<"line"> = {
        labels,
        datasets: [
            {
                label: datasetLabel,
                data: values,
                borderColor: lineColor_,
                backgroundColor: lineColor_,
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
            legend: { labels: { color: t.muted } },
        },
        scales: buildTrackingScaleOptions(datasetLabel, t, values),
    };
    return { data, options };
}

export function buildTrackingBarChart(
    labels: string[],
    values: number[],
    datasetLabel: string,
    barColor?: string,
): { data: ChartData<"bar">; options: ChartOptions<"bar"> } {
    const t = chartTheme();
    const barColor_ = barColor ?? t.series[1];
    const data: ChartData<"bar"> = {
        labels,
        datasets: [
            {
                label: datasetLabel,
                data: values,
                backgroundColor: barColor_,
                borderColor: barColor_,
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
            legend: { labels: { color: t.muted } },
        },
        scales: buildTrackingScaleOptions(datasetLabel, t, values, 0),
    };
    return { data, options };
}
