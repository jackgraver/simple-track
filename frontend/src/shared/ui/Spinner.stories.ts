import type { Meta, StoryObj } from "@storybook/vue3-vite";
import Spinner from "./Spinner.vue";

const meta = { title: "UI/Spinner", component: Spinner, tags: ["autodocs"] } satisfies Meta<typeof Spinner>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Medium: Story = {};
export const Small: Story = { args: { size: "sm" } };
