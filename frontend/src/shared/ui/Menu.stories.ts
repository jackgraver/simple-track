import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { MoreVertical, Pencil, Trash2, Copy } from "lucide-vue-next";
import IconButton from "./IconButton.vue";
import Menu from "./Menu.vue";

const meta = {
    title: "UI/Menu",
    component: Menu,
    tags: ["autodocs"],
    args: {
        items: [
            { key: "edit", label: "Edit", icon: Pencil },
            { key: "copy", label: "Duplicate", icon: Copy },
            { key: "archive", label: "Archive (unavailable)", disabled: true },
            { key: "delete", label: "Delete", icon: Trash2, tone: "danger" },
        ],
    },
    render: (args) => ({
        components: { Menu, IconButton, MoreVertical },
        setup: () => ({ args }),
        template: `<div class="flex justify-end p-4"><Menu v-bind="args"><IconButton label="Actions"><MoreVertical /></IconButton></Menu></div>`,
    }),
} satisfies Meta<typeof Menu>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const AlignStart: Story = { args: { align: "start" } };
