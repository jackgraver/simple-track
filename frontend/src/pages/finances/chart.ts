import type { ChartData, ChartOptions } from "chart.js";
import { chartTheme, withAlpha } from "~/shared/chart/theme";
import { formatCompactMoney, formatDay, formatMoney } from "./format";

export type MoneyPoint = { date: string; value: number };

/** Line chart of money values over time with a soft fill under the line. */
export function buildMoneyLineChart(
    points: MoneyPoint[],
    label: string,
    color?: string,
): { data: ChartData<"line">; options: ChartOptions<"line"> } {
    const t = chartTheme();
    const muted = t.muted;
    const grid = t.grid;
    const line = color ?? t.success;
    const data: ChartData<"line"> = {
        labels: points.map((p) => formatDay(p.date)),
        datasets: [
            {
                label,
                data: points.map((p) => p.value),
                borderColor: line,
                backgroundColor: withAlpha(line, 0.12),
                fill: true,
                tension: 0.25,
                borderWidth: 2,
                pointRadius: points.length > 40 ? 0 : 3,
                pointHoverRadius: 5,
            },
        ],
    };
    const options: ChartOptions<"line"> = {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
            legend: { display: false },
            tooltip: {
                callbacks: {
                    label: (ctx) => `${label}: ${formatMoney(Number(ctx.parsed.y))}`,
                },
            },
        },
        scales: {
            x: {
                ticks: { color: muted, maxRotation: 0, autoSkip: true, maxTicksLimit: 6 },
                grid: { display: false },
            },
            y: {
                ticks: {
                    color: muted,
                    maxTicksLimit: 5,
                    callback: (value) => formatCompactMoney(Number(value)),
                },
                grid: { color: grid },
            },
        },
    };
    return { data, options };
}
