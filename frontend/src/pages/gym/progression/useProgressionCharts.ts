import { computed, type Ref } from "vue";
import type { ChartData, ChartOptions } from "chart.js";

export type ProgressionEntry = {
    date: string;
    weight: number;
    reps: number;
};

export type ProgressionRange = "3m" | "6m" | "1y" | "all";

export function progressionDayKeyFromApi(dateStr: string): string {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(dateStr);
    if (m) {
        return `${m[1]}-${m[2]}-${m[3]}`;
    }
    const d = new Date(dateStr);
    const y = d.getFullYear();
    const mo = String(d.getMonth() + 1).padStart(2, "0");
    const da = String(d.getDate()).padStart(2, "0");
    return `${y}-${mo}-${da}`;
}

function parseDayKeyToLocalDate(dayKey: string): Date {
    const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(dayKey);
    if (m) {
        return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
    }
    return new Date(dayKey);
}

export function formatShortDayLabel(dayKey: string): string {
    const d = parseDayKeyToLocalDate(dayKey);
    return d.toLocaleString("en-US", { month: "short", day: "numeric" });
}

export function roundWeightHalfLb(weight: number): number {
    return Math.round(weight * 2) / 2;
}

export function progressionRangeCutoffKey(
    range: ProgressionRange,
): string | null {
    if (range === "all") return null;
    const now = new Date();
    const end = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const months = range === "3m" ? 3 : range === "6m" ? 6 : 12;
    const start = new Date(end);
    start.setMonth(start.getMonth() - months);
    const y = start.getFullYear();
    const mo = String(start.getMonth() + 1).padStart(2, "0");
    const da = String(start.getDate()).padStart(2, "0");
    return `${y}-${mo}-${da}`;
}

type NormalizedSet = {
    dayKey: string;
    weightRounded: number;
    reps: number;
};

function normalizeEntries(entries: ProgressionEntry[]): NormalizedSet[] {
    return entries.map((e) => ({
        dayKey: progressionDayKeyFromApi(e.date),
        weightRounded: roundWeightHalfLb(Number(e.weight)),
        reps: Number(e.reps),
    }));
}

function filterByRange(rows: NormalizedSet[], range: ProgressionRange): NormalizedSet[] {
    const startKey = progressionRangeCutoffKey(range);
    if (startKey == null) return rows;
    return rows.filter((r) => r.dayKey >= startKey);
}

function groupByDay(rows: NormalizedSet[]): Map<string, NormalizedSet[]> {
    const m = new Map<string, NormalizedSet[]>();
    for (const r of rows) {
        const list = m.get(r.dayKey);
        if (list) list.push(r);
        else m.set(r.dayKey, [r]);
    }
    return m;
}

function readChartTheme(): {
    text: string;
    textMuted: string;
    grid: string;
    lineWeight: string;
    lineVolume: string;
} {
    if (typeof document === "undefined") {
        return {
            text: "hsl(0, 0%, 20%)",
            textMuted: "hsl(0, 0%, 45%)",
            grid: "hsl(0, 0%, 80%)",
            lineWeight: "hsl(142, 71%, 45%)",
            lineVolume: "hsl(215, 85%, 55%)",
        };
    }
    const root = document.documentElement;
    const text =
        getComputedStyle(root).getPropertyValue("--color-text-primary").trim() ||
        "hsl(0, 0%, 20%)";
    const textMuted =
        getComputedStyle(root).getPropertyValue("--color-text-secondary").trim() ||
        "hsl(0, 0%, 45%)";
    const grid =
        getComputedStyle(root).getPropertyValue("--color-border").trim() ||
        "hsl(0, 0%, 80%)";
    return {
        text,
        textMuted,
        grid,
        lineWeight: "hsl(142, 71%, 45%)",
        lineVolume: "hsl(215, 85%, 55%)",
    };
}

export function useProgressionCharts(
    entries: Ref<ProgressionEntry[]>,
    range: Ref<ProgressionRange>,
) {
    const filteredRows = computed(() => {
        const normalized = normalizeEntries(entries.value);
        return filterByRange(normalized, range.value);
    });

    const byDay = computed(() => groupByDay(filteredRows.value));

    const combinedSeries = computed(() => {
        const dayMap = byDay.value;
        const sortedDayKeys = [...dayMap.keys()].sort();
        const labels: string[] = [];
        const weights: number[] = [];
        const topReps: number[] = [];
        const topSetVolumes: number[] = [];
        for (const day of sortedDayKeys) {
            const sets = dayMap.get(day)!;
            let maxW = -Infinity;
            for (const s of sets) {
                if (s.weightRounded > maxW) maxW = s.weightRounded;
            }
            let bestReps = 0;
            for (const s of sets) {
                if (s.weightRounded === maxW && s.reps > bestReps) bestReps = s.reps;
            }
            labels.push(formatShortDayLabel(day));
            weights.push(maxW);
            topReps.push(bestReps);
            topSetVolumes.push(maxW * bestReps);
        }
        return { labels, weights, topReps, topSetVolumes, sortedDayKeys };
    });

    const hasProgressionChartData = computed(
        () => combinedSeries.value.weights.length > 0,
    );

    const progressionChartData = computed<ChartData<"line">>(() => {
        const t = readChartTheme();
        const s = combinedSeries.value;
        return {
            labels: s.labels,
            datasets: [
                {
                    label: "Top set weight (lbs)",
                    data: s.weights,
                    yAxisID: "y",
                    borderColor: t.lineWeight,
                    backgroundColor: t.lineWeight,
                    tension: 0.2,
                    fill: false,
                    pointRadius: 3,
                    pointHoverRadius: 5,
                    borderWidth: 2,
                },
                {
                    label: "Top-set volume (lb×reps)",
                    data: s.topSetVolumes,
                    yAxisID: "y1",
                    borderColor: t.lineVolume,
                    backgroundColor: t.lineVolume,
                    tension: 0.2,
                    fill: false,
                    pointRadius: 3,
                    pointHoverRadius: 5,
                    borderWidth: 2,
                },
            ],
        };
    });

    const progressionChartOptions = computed<ChartOptions<"line">>(() => {
        const t = readChartTheme();
        const s = combinedSeries.value;
        return {
            responsive: true,
            maintainAspectRatio: false,
            interaction: { mode: "index", intersect: false },
            plugins: {
                legend: {
                    labels: { color: t.textMuted },
                },
                tooltip: {
                    callbacks: {
                        label: (ctx) => {
                            const i = ctx.dataIndex;
                            const w = s.weights[i];
                            const r = s.topReps[i];
                            const v = s.topSetVolumes[i];
                            if (ctx.datasetIndex === 0) {
                                if (w === undefined || r === undefined) return "";
                                return `Weight: ${w} lbs × ${r} reps`;
                            }
                            if (v === undefined) return "";
                            return `Top-set volume: ${v.toLocaleString()} (lb×reps)`;
                        },
                    },
                },
            },
            scales: {
                x: {
                    ticks: { color: t.textMuted, maxRotation: 45 },
                    grid: { color: t.grid },
                },
                y: {
                    position: "left",
                    ticks: { color: t.lineWeight },
                    grid: { color: t.grid },
                    title: {
                        display: true,
                        text: "Top set (lbs)",
                        color: t.textMuted,
                    },
                },
                y1: {
                    position: "right",
                    ticks: { color: t.lineVolume },
                    grid: { drawOnChartArea: false },
                    title: {
                        display: true,
                        text: "Top-set volume",
                        color: t.textMuted,
                    },
                },
            },
        };
    });

    return {
        filteredRows,
        progressionChartData,
        progressionChartOptions,
        hasProgressionChartData,
    };
}
