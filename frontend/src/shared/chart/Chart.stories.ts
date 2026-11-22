import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { computed } from "vue";
import { buildMoneyLineChart } from "~/pages/finances/chart";
import { buildTrackingBarChart, buildTrackingLineChart } from "~/pages/gym/trackingLogChart";
import Chart from "./Chart.vue";

// The builders call chartTheme(), so every story re-reads the tokens when the theme switches.
const days = Array.from({ length: 14 }, (_, i) => `Nov ${i + 1}`);
const meta = {
    title: "Chart/Chart",
    component: Chart,
    tags: ["autodocs"],
    decorators: [() => ({ template: `<div class="h-64 w-[min(90vw,40rem)]"><story /></div>` })],
} satisfies Meta;
export default meta;
type Story = StoryObj;

export const WeightLine: Story = {
    render: () => ({
        components: { Chart },
        setup() {
            const c = computed(() => buildTrackingLineChart(days, days.map((_, i) => 182 - i * 0.4 + (i % 3)), "Weight (lbs)"));
            return { c };
        },
        template: `<Chart class="h-full w-full" type="line" :data="c.data" :options="c.options" label="Body weight" />`,
    }),
};

export const StepsBar: Story = {
    render: () => ({
        components: { Chart },
        setup() {
            const c = computed(() => buildTrackingBarChart(days, days.map((_, i) => 4000 + ((i * 937) % 6000)), "Steps"));
            return { c };
        },
        template: `<Chart class="h-full w-full" type="bar" :data="c.data" :options="c.options" label="Daily steps" />`,
    }),
};

export const MoneyArea: Story = {
    render: () => ({
        components: { Chart },
        setup() {
            const c = computed(() =>
                buildMoneyLineChart(days.map((_, i) => ({ date: `2026-11-${String(i + 1).padStart(2, "0")}`, value: 12000 + i * 310 })), "Net worth"),
            );
            return { c };
        },
        template: `<Chart class="h-full w-full" type="line" :data="c.data" :options="c.options" label="Net worth" />`,
    }),
};
