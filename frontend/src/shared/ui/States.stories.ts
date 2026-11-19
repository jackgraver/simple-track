import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { Dumbbell } from "lucide-vue-next";
import Button from "./Button.vue";
import Card from "./Card.vue";
import EmptyState from "./EmptyState.vue";
import ErrorState from "./ErrorState.vue";
import Skeleton from "./Skeleton.vue";
import SkeletonList from "./SkeletonList.vue";

const meta = {
    title: "UI/States",
    component: EmptyState,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
} satisfies Meta<typeof EmptyState>;
export default meta;
type Story = StoryObj<typeof meta>;

const components = { EmptyState, ErrorState, Skeleton, SkeletonList, Card, Button, Dumbbell };

/** Loading, empty and error in the card they will live in. The skeleton appears only after ~150ms in the app (`useDelayedFlag`). */
export const Gallery: Story = {
    args: { title: "x" },
    render: () => ({
        components,
        setup: () => ({ Dumbbell, retry: () => {}, err: new Error("GET /api/workouts 503 Service Unavailable") }),
        template: `
        <div class="grid max-w-3xl gap-6 lg:grid-cols-2">
            <div><h3 class="mb-2 text-xs font-medium text-fg-muted">Loading</h3><Card padding="none"><SkeletonList :rows="4" /></Card></div>
            <div><h3 class="mb-2 text-xs font-medium text-fg-muted">Empty</h3><Card padding="none">
                <EmptyState :icon="Dumbbell" title="No workouts yet" description="Log a set and it shows up here.">
                    <template #action><Button variant="primary">Start workout</Button></template>
                </EmptyState></Card></div>
            <div><h3 class="mb-2 text-xs font-medium text-fg-muted">Error</h3><Card padding="none"><ErrorState :error="err" title="Couldn't load workouts" @retry="retry" /></Card></div>
            <div><h3 class="mb-2 text-xs font-medium text-fg-muted">Error (compact)</h3><Card padding="none"><ErrorState compact title="Couldn't load meals" @retry="retry" /></Card></div>
            <div><h3 class="mb-2 text-xs font-medium text-fg-muted">Skeleton shapes</h3><Card class="flex flex-col gap-3"><Skeleton width="40%" /><Skeleton shape="block" /><Skeleton shape="circle" /></Card></div>
        </div>`,
    }),
};

export const Empty: Story = { args: { title: "No saved meals", description: "Save a meal to reuse it.", icon: Dumbbell } };
