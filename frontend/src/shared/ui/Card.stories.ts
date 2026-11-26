import type { Meta, StoryObj } from "@storybook/vue3-vite";
import Button from "./Button.vue";
import Card from "./Card.vue";

const meta = {
    title: "UI/Card",
    component: Card,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
    render: (args) => ({
        components: { Card },
        setup: () => ({ args }),
        template: `<div class="max-w-sm"><Card v-bind="args"><p class="m-0 text-sm text-fg-muted">Weekly average</p><p class="m-0 text-display tabular-nums">82.4 kg</p></Card></div>`,
    }),
} satisfies Meta<typeof Card>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const PaddingSmall: Story = { args: { padding: "sm" } };
export const Interactive: Story = { args: { interactive: true, as: "button" } };
export const HeaderAndFooter: Story = {
    render: (args) => ({
        components: { Card, Button },
        setup: () => ({ args }),
        template: `<div class="max-w-sm"><Card v-bind="args">
            <template #header><h3 class="m-0 text-sm font-semibold">Account</h3><Button size="sm" variant="ghost">Edit</Button></template>
            <p class="m-0 text-sm text-fg-muted">Everyday spending, updated today.</p>
            <template #footer><Button variant="ghost" size="sm">Cancel</Button><Button variant="primary" size="sm">Save</Button></template>
        </Card></div>`,
    }),
};

export const Tab: Story = {
    args: { padding: "none" },
    render: (args) => ({
        components: { Card },
        setup: () => ({ args }),
        template: `<div class="max-w-sm"><Card v-bind="args">
            <template #tab><span class="tabular-nums">Set 1 · 30 lb</span></template>
            <div class="grid grid-cols-2 divide-x divide-line">
                <p class="m-0 p-3 text-sm">Left · 10 reps</p>
                <p class="m-0 p-3 text-sm">Right · 9 reps</p>
            </div>
        </Card></div>`,
    }),
};
