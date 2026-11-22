import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ref } from "vue";
import Button from "./Button.vue";
import Toaster from "./Toaster.vue";
import type { ToastItem } from "./toastTypes";

const meta = {
    title: "UI/Toast",
    component: Toaster,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
    args: { toasts: [] },
    render: (args) => ({
        components: { Toaster, Button },
        setup() {
            const toasts = ref<ToastItem[]>([]);
            let id = 0;
            const add = (t: Partial<ToastItem> & { message: string }) => {
                const item: ToastItem = { id: ++id, tone: "info", duration: 4000, open: true, ...t };
                toasts.value.push(item);
            };
            const dismiss = (i: number) => {
                const t = toasts.value.find((x) => x.id === i);
                if (t) t.open = false;
                setTimeout(() => (toasts.value = toasts.value.filter((x) => x.id !== i)), 200);
            };
            return { args, toasts, add, dismiss };
        },
        template: `
        <div class="flex flex-wrap gap-2">
            <Button @click="add({ message: 'Meal logged', tone: 'success' })">Success</Button>
            <Button @click="add({ message: 'Couldn\\'t save the set. Try again.', tone: 'error', duration: 8000 })">Error</Button>
            <Button @click="add({ message: 'Set deleted', tone: 'info', duration: 8000, action: { label: 'Undo', onClick: () => add({ message: 'Restored', tone: 'success' }) } })">With Undo</Button>
            <Button variant="ghost" @click="add({ message: 'Notice' })">Info</Button>
            <Toaster :toasts="toasts" :tab-bar="args.tabBar" :top="args.top" @dismiss="dismiss" @action="(i) => { toasts.find((t) => t.id === i)?.action?.onClick(); dismiss(i); }" />
        </div>`,
    }),
} satisfies Meta<typeof Toaster>;
export default meta;
type Story = StoryObj<typeof meta>;

/** Desktop: bottom-right. Mobile: full-width above the tab bar (`tabBar`). Hover or focus pauses the timer; swipe right dismisses. */
export const Playground: Story = { args: { toasts: [], tabBar: true } };
/** Focus pages (gym logging) put toasts at the top on mobile so the bottom "Log set" button stays clear. */
export const TopOnMobile: Story = { args: { toasts: [], top: true } };
