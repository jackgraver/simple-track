import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { ChevronRight, Dumbbell, MoreVertical, Pencil, Trash2 } from "lucide-vue-next";
import Badge from "./Badge.vue";
import Card from "./Card.vue";
import IconButton from "./IconButton.vue";
import List from "./List.vue";
import ListRow from "./ListRow.vue";

const meta = {
    title: "UI/List",
    component: ListRow,
    tags: ["autodocs"],
    parameters: { layout: "padded" },
} satisfies Meta<typeof ListRow>;
export default meta;
type Story = StoryObj<typeof meta>;

const components = { List, ListRow, Card, Badge, IconButton, ChevronRight, Dumbbell, MoreVertical, Pencil, Trash2 };

/**
 * One list, every state. Rows with a click handler are buttons. Desktop (lg+): Edit/Delete show on
 * hover or focus-within. Mobile: the trailing "more" IconButton is the way in (never hover-only).
 */
export const Showcase: Story = {
    render: () => ({
        components,
        setup: () => ({ ping: () => {} }),
        template: `
        <div class="max-w-xl"><Card padding="none"><List>
            <ListRow @click="ping">
                <template #leading><Dumbbell class="size-5" /></template>
                Bench press
                <template #subtitle>3 × 8 · 80 kg</template>
                <template #trailing><ChevronRight class="size-4" /></template>
            </ListRow>
            <ListRow selected @click="ping">
                <template #leading><Dumbbell class="size-5" /></template>
                Overhead press (selected)
                <template #subtitle>3 × 10 · 40 kg</template>
                <template #trailing><Badge tone="success">+2.5</Badge></template>
            </ListRow>
            <ListRow pending>
                <template #leading><Dumbbell class="size-5" /></template>
                Squat (pending)
                <template #subtitle>Saving…</template>
            </ListRow>
            <ListRow disabled>
                <template #leading><Dumbbell class="size-5" /></template>
                Deadlift (disabled)
            </ListRow>
            <ListRow @click="ping">
                Row with actions
                <template #subtitle>Hover or focus on desktop</template>
                <template #trailing>
                    <IconButton label="More actions" size="sm" class="lg:hidden"><MoreVertical /></IconButton>
                </template>
                <template #actions>
                    <IconButton label="Edit" size="sm"><Pencil /></IconButton>
                    <IconButton label="Delete" size="sm" variant="danger"><Trash2 /></IconButton>
                </template>
            </ListRow>
            <ListRow>A static row with a very long title that should truncate instead of wrapping onto a second line at any width
                <template #trailing>12:30</template>
            </ListRow>
        </List></Card></div>`,
    }),
};
