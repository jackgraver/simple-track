import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Combobox from "./Combobox.vue";

const options = ["Bench press", "Back squat", "Deadlift", "Overhead press", "Barbell row", "Pull-up", "Dip"].map((label, i) => ({
    label,
    value: i + 1,
}));

const meta = {
    title: "UI/Combobox",
    component: Combobox,
    tags: ["autodocs"],
    args: { options, placeholder: "Add exercise…" },
    render: (args) => ({
        components: { Combobox },
        setup: () => ({ args, picked: ref("nothing yet") }),
        template: `<div class="w-80"><Combobox v-bind="args" @select="(o) => (picked = o.label)" /><p class="mt-3 text-sm text-fg-muted">Picked: {{ picked }}</p></div>`,
    }),
} satisfies Meta<typeof Combobox>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Loading: Story = { args: { loading: true, hasMore: true } };
export const Empty: Story = { args: { options: [], emptyText: "No exercises yet" } };
export const Disabled: Story = { args: { disabled: true } };
