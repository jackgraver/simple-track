import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { Plus } from "lucide-vue-next";
import Button from "./Button.vue";

const variants = ["primary", "secondary", "ghost", "danger", "danger-solid"] as const;
const sizes = ["sm", "md", "lg"] as const;
const states = [
    { name: "default", props: {} },
    { name: "hover", props: {} },
    { name: "focus", props: {} },
    { name: "disabled", props: { disabled: true } },
    { name: "loading", props: { loading: true } },
];

const meta = {
    title: "UI/Button",
    component: Button,
    tags: ["autodocs"],
    args: { variant: "primary", size: "md" },
    render: (args) => ({
        components: { Button, Plus },
        setup: () => ({ args }),
        template: `<Button v-bind="args">Log set</Button>`,
    }),
} satisfies Meta<typeof Button>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {};
export const Secondary: Story = { args: { variant: "secondary" } };
export const Ghost: Story = { args: { variant: "ghost" } };
export const Danger: Story = { args: { variant: "danger" } };
export const DangerSolid: Story = { args: { variant: "danger-solid" } };
export const Loading: Story = { args: { loading: true } };
export const Disabled: Story = { args: { disabled: true } };

export const WithIcons: Story = {
    render: (args) => ({
        components: { Button, Plus },
        setup: () => ({ args }),
        template: `<Button v-bind="args"><template #icon-left><Plus class="size-4" /></template>Add food</Button>`,
    }),
};

/** Mobile primary CTA: `size="lg" block` at the bottom of the content. */
export const MobileCta: Story = {
    args: { size: "lg", block: true },
    parameters: { layout: "padded", viewport: { defaultViewport: "phone" } },
};

/** Every variant × size × state. Focus: Tab through the grid. Hover: use the pointer. */
export const Matrix: Story = {
    render: () => ({
        components: { Button, Plus },
        setup: () => ({ variants, sizes, states }),
        template: `
        <div class="flex flex-col gap-6">
            <section v-for="v in variants" :key="v" class="flex flex-col gap-2">
                <h3 class="text-xs font-medium text-fg-muted">{{ v }}</h3>
                <div v-for="s in sizes" :key="s" class="flex flex-wrap items-center gap-2">
                    <span class="w-8 text-2xs text-fg-subtle">{{ s }}</span>
                    <Button v-for="st in states" :key="st.name" :variant="v" :size="s" v-bind="st.props">
                        <template #icon-left><Plus class="size-4" /></template>{{ st.name }}
                    </Button>
                </div>
            </section>
        </div>`,
    }),
};
