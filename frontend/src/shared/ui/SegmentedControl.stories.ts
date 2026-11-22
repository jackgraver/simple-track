import type { Meta, StoryObj } from "@storybook/vue3-vite";
// The component is generic, which Meta<typeof X> can't infer: type the story args by hand.
type Args = { label: string; modelValue: string; options: { label: string; value: string }[] };
import { ref } from "vue";
import SegmentedControl from "./SegmentedControl.vue";

const meta = {
    title: "UI/SegmentedControl",
    component: SegmentedControl as unknown as Meta<Args>["component"],
    tags: ["autodocs"],
    args: {
        label: "Theme",
        modelValue: "dark",
        options: [
            { label: "Dark", value: "dark" },
            { label: "Light", value: "light" },
            { label: "System", value: "system" },
        ],
    },
    render: (args: Args) => ({
        components: { SegmentedControl },
        setup: () => ({ args, v: ref(args.modelValue) }),
        template: `<div class="w-80"><SegmentedControl v-bind="args" v-model="v" /></div>`,
    }),
} satisfies Meta<Args>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const TwoOptions: Story = {
    args: { modelValue: "a", options: [{ label: "Metric", value: "a" }, { label: "Imperial", value: "b" }] },
};
