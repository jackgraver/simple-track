import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Field from "./Field.vue";
import NumberField from "./NumberField.vue";

const meta = {
    title: "UI/NumberField",
    component: NumberField,
    tags: ["autodocs"],
    args: { label: "weight", unit: "kg", step: 2.5, min: 0 },
    render: (args) => ({
        components: { NumberField, Field },
        setup: () => ({ args, value: ref<number | null>((args.modelValue as number | null | undefined) ?? 60) }),
        template: `<div class="w-80"><Field label="Weight"><NumberField v-bind="args" v-model="value" /></Field></div>`,
    }),
} satisfies Meta<typeof NumberField>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Weight: Story = {};
export const Reps: Story = { args: { label: "reps", unit: "reps", step: 1, decimals: 0, modelValue: 8 } };
export const Empty: Story = { args: { modelValue: null, placeholder: "0" } };
export const AtMinimum: Story = { args: { modelValue: 0 } };
export const Large: Story = { args: { size: "lg" } };
export const Invalid: Story = { args: { invalid: true } };
export const Disabled: Story = { args: { disabled: true } };
