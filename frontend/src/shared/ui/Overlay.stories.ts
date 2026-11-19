import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Button from "./Button.vue";
import Field from "./Field.vue";
import Input from "./Input.vue";
import Overlay from "./Overlay.vue";

const meta = {
    title: "UI/Overlay",
    component: Overlay,
    tags: ["autodocs"],
    args: { title: "Edit meal", open: true },
    parameters: { layout: "fullscreen" },
    render: (args) => ({
        components: { Overlay, Button, Field, Input },
        setup: () => ({ args, open: ref(args.open), name: ref("Oats"), rows: Array.from({ length: 24 }, (_, i) => i + 1) }),
        template: `
        <div class="p-4">
            <Button variant="primary" @click="open = true">Open overlay</Button>
            <Overlay v-bind="args" v-model:open="open">
                <div class="flex flex-col gap-4">
                    <Field label="Name"><Input v-model="name" /></Field>
                    <p v-if="args.title === 'Long content'" v-for="r in rows" :key="r" class="m-0 text-sm text-fg-muted">Line {{ r }} of a long body that scrolls inside the overlay.</p>
                </div>
                <template #footer>
                    <Button variant="secondary" @click="open = false">Cancel</Button>
                    <Button variant="primary" @click="open = false">Save</Button>
                </template>
            </Overlay>
        </div>`,
    }),
} satisfies Meta<typeof Overlay>;
export default meta;
type Story = StoryObj<typeof meta>;

/** `auto`: a sheet below lg (swipe the handle or header down, ✕, Esc or the scrim close it), a centred dialog from lg. Use the 390 / 1280 viewports. */
export const Auto: Story = {};
export const Large: Story = { args: { size: "lg", title: "Long content" } };
export const ForcedSheet: Story = { args: { presentation: "sheet" } };
export const ForcedDialog: Story = { args: { presentation: "dialog", size: "sm", description: "Small dialog with a description." } };
export const NotDismissible: Story = { args: { dismissible: false, title: "Saving…" } };
