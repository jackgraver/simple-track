import type { Meta, StoryObj } from "@storybook/vue3-vite";
import Badge from "./Badge.vue";

const meta = {
    title: "UI/Badge",
    component: Badge,
    tags: ["autodocs"],
    render: (args) => ({ components: { Badge }, setup: () => ({ args }), template: `<Badge v-bind="args">Stale</Badge>` }),
} satisfies Meta<typeof Badge>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Neutral: Story = {};
export const Success: Story = { args: { tone: "success" } };
export const Danger: Story = { args: { tone: "danger" } };

/** Every tone × variant × size. */
export const Matrix: Story = {
    render: () => ({
        components: { Badge },
        setup: () => ({ tones: ["neutral", "accent", "success", "warn", "danger", "info"], variants: ["soft", "outline"], sizes: ["sm", "md"] }),
        template: `<div class="flex flex-col gap-3"><template v-for="v in variants" :key="v"><div v-for="s in sizes" :key="s" class="flex flex-wrap items-center gap-2">
            <Badge v-for="t in tones" :key="t" :tone="t" :variant="v" :size="s">{{ t }}</Badge></div></template></div>`,
    }),
};
