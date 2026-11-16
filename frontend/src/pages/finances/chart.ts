import type { ChartData, ChartOptions } from "chart.js";
import { formatCompactMoney, formatDay, formatMoney } from "./format";

function themeColor(name: string, fallback: string) {
    if (typeof document === "undefined") return fallback;
    return (
        getComputedStyle(document.documentElement).getPropertyValue(name).trim() ||
        fallback
    );
}

export type MoneyPoint = { date: string; value: number };

/** Line chart of money values over time with a soft fill under the line. */
export function buildMoneyLineChart(
    points: MoneyPoint[],
    label: string,
    color = "rgb(52, 211, 153)",
): { data: ChartData<"line">; options: ChartOptions<"line"> } {
    const muted = themeColor("--color-text-secondary", "hsl(0, 0%, 70%)");
    const grid = themeColor("--color-border", "hsl(0, 0%, 30%)");
    const data: ChartData<"line"> = {
        labels: points.map((p) => formatDay(p.date)),
        datasets: [
            {
                label,
                data: points.map((p) => p.value),
                borderColor: color,
                backgroundColor: color.replace("rgb(", "rgba(").replace(")", ", 0.12)"),
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
