import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { Trash2, X } from "lucide-vue-next";
import IconButton from "./IconButton.vue";

const meta = {
    title: "UI/IconButton",
    component: IconButton,
    tags: ["autodocs"],
    args: { label: "Close", variant: "ghost", size: "md" },
    render: (args) => ({
        components: { IconButton, X },
        setup: () => ({ args }),
        template: `<IconButton v-bind="args"><X /></IconButton>`,
    }),
} satisfies Meta<typeof IconButton>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Ghost: Story = {};
export const Secondary: Story = { args: { variant: "secondary" } };
export const Danger: Story = { args: { variant: "danger", label: "Delete" } };
export const Disabled: Story = { args: { disabled: true } };

export const Matrix: Story = {
    render: () => ({
        components: { IconButton, Trash2 },
        setup: () => ({ variants: ["ghost", "secondary", "danger"], sizes: ["md", "sm"] }),
        template: `
        <div class="flex flex-col gap-3">
            <div v-for="s in sizes" :key="s" class="flex items-center gap-2">
                <span class="w-8 text-2xs text-fg-subtle">{{ s }}</span>
                <template v-for="v in variants" :key="v">
                    <IconButton :variant="v" :size="s" label="Delete"><Trash2 /></IconButton>
                    <IconButton :variant="v" :size="s" label="Delete" disabled><Trash2 /></IconButton>
                </template>
            </div>
        </div>`,
    }),
};
