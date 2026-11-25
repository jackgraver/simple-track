import type { Meta, StoryObj } from "@storybook/vue3-vite";
import Avatar from "./Avatar.vue";

const meta = {
    title: "UI/Avatar",
    component: Avatar,
    tags: ["autodocs"],
    args: { name: "dev", size: "md" },
} satisfies Meta<typeof Avatar>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Sizes: Story = {
    render: (args) => ({
        components: { Avatar },
        setup: () => ({ args }),
        template: `<div class="flex items-center gap-3 p-4"><Avatar v-bind="args" size="sm" /><Avatar v-bind="args" size="md" /><Avatar v-bind="args" size="lg" /></div>`,
    }),
};
export const BrokenImage: Story = { args: { src: "/does-not-exist.png", size: "lg" } };
