import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { defineComponent, h } from "vue";
import { Button } from "~/shared/ui";
import DialogContainer from "./DialogContainer.vue";
import { dialogManager } from "./useDialog";

// A custom dialog body like the real ones: it gets onResolve/onCancel props.
const Body = defineComponent({
    props: { onResolve: Function, onCancel: Function },
    setup: (p) => () =>
        h("div", { class: "flex flex-col gap-3" }, [
            h("p", { class: "m-0 text-sm text-fg-muted" }, "Any component can render here and resolve the promise."),
            h(Button, { variant: "primary", onClick: () => p.onResolve?.(true) }, () => "Done"),
        ]),
});

const meta = {
    title: "UI/DialogContainer (dialogManager)",
    component: DialogContainer,
    parameters: { layout: "padded" },
    render: () => ({
        components: { DialogContainer, Button },
        setup: () => ({
            confirm: () => dialogManager.confirm({ title: "Delete balance?", message: "This can't be undone.", confirmText: "Delete" }),
            danger: () => dialogManager.confirm({ title: "Delete balance?", message: "This can't be undone.", confirmText: "Delete", tone: "danger" }),
            custom: () => dialogManager.custom({ title: "Custom body", component: Body }),
            wide: () => dialogManager.custom({ title: "Wide custom body", component: Body, wide: true }),
        }),
        template: `<div class="flex flex-wrap gap-2"><Button @click="confirm">confirm()</Button><Button variant="danger" @click="danger">confirm({ tone: "danger" })</Button><Button @click="custom">custom()</Button><Button @click="wide">custom({ wide })</Button><DialogContainer /></div>`,
    }),
} satisfies Meta<typeof DialogContainer>;
export default meta;
type Story = StoryObj<typeof meta>;
export const Playground: Story = {};
