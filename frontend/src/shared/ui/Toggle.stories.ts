import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Toggle from "./Toggle.vue";

const meta = {
    title: "UI/Toggle",
    component: Toggle,
    tags: ["autodocs"],
    args: { label: "Auto-increase weight" },
    render: (args) => ({
        components: { Toggle },
        setup: () => ({ args, on: ref(!!args.modelValue) }),
        template: `<div class="w-80"><Toggle v-bind="args" v-model="on" /></div>`,
    }),
} satisfies Meta<typeof Toggle>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Off: Story = {};
export const On: Story = { args: { modelValue: true } };
export const Disabled: Story = { args: { disabled: true } };
