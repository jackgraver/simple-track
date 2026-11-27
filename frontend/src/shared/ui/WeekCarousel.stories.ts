import type { Meta, StoryObj } from "@storybook/vue3-vite";
import { computed, ref, type Component } from "vue";
import IconButton from "./IconButton.vue";
import WeekCarousel from "./WeekCarousel.vue";

// The component is generic, which Meta<typeof X> can't infer: type the story args by hand.
type Args = { label: string };
type Day = { key: string; isSelected: boolean; date: Date };

const pad = (n: number) => String(n).padStart(2, "0");
const keyOf = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;

/** Arrows, swipes and tile taps all work; arrows land on Sunday like the gym and diet strips. */
const meta = {
    title: "UI/WeekCarousel",
    component: WeekCarousel as unknown as Meta<Args>["component"],
    tags: ["autodocs"],
    args: { label: "Days this week" },
    render: (args: Args) => ({
        components: { WeekCarousel: WeekCarousel as unknown as Component, IconButton },
        setup() {
            const start = new Date(2026, 10, 22); // a Sunday
            const sunday = ref(start);
            const selected = ref(keyOf(new Date(2026, 10, 26)));
            const days = computed<Day[]>(() =>
                Array.from({ length: 7 }, (_, i) => {
                    const date = new Date(sunday.value.getFullYear(), sunday.value.getMonth(), sunday.value.getDate() + i);
                    return { key: keyOf(date), date, isSelected: keyOf(date) === selected.value };
                }),
            );
            const move = (step: 1 | -1) => {
                const s = sunday.value;
                sunday.value = new Date(s.getFullYear(), s.getMonth(), s.getDate() + 7 * step);
                selected.value = keyOf(sunday.value);
            };
            const pick = (d: Day) => (selected.value = d.key);
            return { args, days, move, pick };
        },
        template: `
            <div class="flex w-96 flex-col gap-3 rounded-card bg-surface p-3">
                <div class="flex justify-end gap-2">
                    <IconButton label="Previous week" size="sm" @click="move(-1)">‹</IconButton>
                    <IconButton label="Next week" size="sm" @click="move(1)">›</IconButton>
                </div>
                <WeekCarousel v-bind="args" :days="days" @prev="move(-1)" @next="move(1)">
                    <template #day="{ day }">
                        <button
                            type="button"
                            class="flex min-h-touch w-full flex-col items-center rounded-control border border-transparent py-2 text-fg hover:bg-surface-3"
                            @click="pick(day)"
                        >
                            <span class="text-2xs text-fg-muted uppercase">{{ day.date.toLocaleDateString(undefined, { weekday: "short" }) }}</span>
                            <span class="text-base font-semibold tabular-nums">{{ day.date.getDate() }}</span>
                        </button>
                    </template>
                </WeekCarousel>
            </div>`,
    }),
} satisfies Meta<Args>;
export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
