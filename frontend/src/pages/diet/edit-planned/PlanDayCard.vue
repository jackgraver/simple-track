<script setup lang="ts">
import { ArrowDown, ArrowUp, Copy, CopyPlus, Ellipsis, Eraser, Plus, Trash2 } from "lucide-vue-next";
import { computed, useId } from "vue";
import { Badge, Button, Card, IconButton, Menu, type MenuItem } from "~/shared/ui";
import { formatInt } from "~/pages/diet/home/dietDay";
import { kcalLine, shortDayLabel, statusText, type PlanDay } from "./domain/planWeek";

const props = defineProps<{
    day: PlanDay;
    /** The day before, for "Copy from Thu" on an empty day. */
    previous?: PlanDay;
    /** Desktop tray selection: the add button adds this meal straight away. */
    armedName?: string | null;
    busy: boolean;
    busyMealIds: Set<number>;
}>();

const emit = defineEmits<{
    add: [];
    copyTo: [];
    copyFromPrevious: [];
    clear: [];
    move: [index: number, dir: -1 | 1];
    remove: [mealId: number];
}>();

const headingId = useId();
const label = computed(() => shortDayLabel(props.day.date));
const prevShort = computed(() => props.previous?.date.toLocaleDateString(undefined, { weekday: "short" }));

const dayMenu = computed<MenuItem[]>(() => {
    const items: MenuItem[] = [];
    if (props.day.meals.length === 0 && (props.previous?.meals.length ?? 0) > 0) {
        items.push({ key: "copyPrev", label: `Copy from ${prevShort.value}`, icon: CopyPlus });
    }
    items.push(
        { key: "copy", label: "Copy to other days…", icon: Copy, disabled: props.day.meals.length === 0 },
        { key: "clear", label: "Clear day", icon: Eraser, tone: "danger", disabled: props.day.meals.length === 0 },
    );
    return items;
});
function onDayMenu(key: string) {
    if (key === "copyPrev") emit("copyFromPrevious");
    else if (key === "copy") emit("copyTo");
    else if (key === "clear") emit("clear");
}

function mealMenu(index: number): MenuItem[] {
    return [
        { key: "up", label: "Move earlier", icon: ArrowUp, disabled: index === 0 },
        { key: "down", label: "Move later", icon: ArrowDown, disabled: index === props.day.meals.length - 1 },
        { key: "remove", label: "Remove", icon: Trash2, tone: "danger" },
    ];
}
function onMealMenu(index: number, mealId: number, key: string) {
    if (key === "up") emit("move", index, -1);
    else if (key === "down") emit("move", index, 1);
    else if (key === "remove") emit("remove", mealId);
}

const MACROS = [
    ["P", "protein"],
    ["C", "carbs"],
    ["F", "fat"],
] as const;
const macroParts = computed(() =>
    MACROS.map(([short, key]) => ({
        short,
        planned: formatInt(props.day.planned[key]),
        target: props.day.target[key] > 0 ? formatInt(props.day.target[key]) : null,
    })),
);
</script>

<template>
    <Card as="section" padding="sm" :aria-labelledby="headingId">
        <!-- Mobile: day + actions on top, totals, then meals. Desktop: day and totals | meals | actions.
             The actions never move, so many quick adds hit the same spot. -->
        <div
            class="grid grid-cols-[1fr_auto] gap-x-3 gap-y-3 lg:grid-cols-[15rem_1fr_auto] lg:gap-x-6 lg:p-1"
            :aria-busy="busy || undefined"
        >
            <div class="col-start-1 row-start-1 flex items-baseline gap-2 self-center">
                <h3 :id="headingId" class="m-0 text-base font-semibold" :class="day.isPast ? 'text-fg-muted' : 'text-fg'">
                    {{ label }}
                </h3>
                <span v-if="day.isToday" class="text-xs font-medium text-accent">Today</span>
            </div>

            <div class="col-start-2 row-start-1 flex items-center gap-1 lg:col-start-3">
                <Button
                    size="sm"
                    variant="secondary"
                    :aria-label="armedName ? `Add ${armedName} to ${label}` : `Add meal to ${label}`"
                    @click="emit('add')"
                >
                    <Plus class="size-4 shrink-0" aria-hidden="true" />
                    <span class="max-w-40 truncate">{{ armedName ?? "Add" }}</span>
                </Button>
                <Menu :items="dayMenu" align="end" @select="onDayMenu">
                    <IconButton :label="`${label} actions`" size="sm">
                        <Ellipsis class="size-5" />
                    </IconButton>
                </Menu>
            </div>

            <!-- Totals -->
            <div class="col-span-2 flex flex-col gap-1.5 lg:col-span-1 lg:col-start-1 lg:row-start-2">
                <div class="flex items-baseline justify-between gap-2 tabular-nums">
                    <span class="text-sm font-medium text-fg">{{ kcalLine(day.planned.calories, day.target.calories) }}</span>
                    <span v-if="day.status === 'over'" class="text-xs text-warn">{{ statusText(day) }}</span>
                </div>
                <div
                    class="h-1.5 w-full overflow-hidden rounded-full bg-surface-3"
                    role="img"
                    :aria-label="`${kcalLine(day.planned.calories, day.target.calories)} planned. ${statusText(day)}`"
                >
                    <div
                        class="h-full rounded-full"
                        :class="day.status === 'over' ? 'bg-warn' : 'bg-macro-calories'"
                        :style="{ width: `${day.fill}%` }"
                    />
                </div>
                <p class="m-0 flex flex-wrap gap-x-3 text-xs text-fg-muted tabular-nums">
                    <span v-for="m in macroParts" :key="m.short">
                        <span class="text-fg">{{ m.short }} {{ m.planned }}</span
                        ><template v-if="m.target"> / {{ m.target }}</template>
                    </span>
                    <span v-if="day.isPast || (day.isToday && day.loggedKcal > 0)">
                        Logged {{ formatInt(day.loggedKcal) }}
                    </span>
                </p>
            </div>

            <!-- Meals: always visible; tap one for move / remove. -->
            <div class="col-span-2 flex min-w-0 lg:col-span-1 lg:col-start-2 lg:row-span-2 lg:row-start-1">
                <ul v-if="day.meals.length" class="m-0 flex list-none flex-wrap content-start gap-2 p-0">
                    <li
                        v-for="(meal, index) in day.meals"
                        :key="meal.id"
                        class="transition-opacity duration-150"
                        :class="busyMealIds.has(meal.id) ? 'opacity-60' : ''"
                    >
                        <Menu :items="mealMenu(index)" align="start" @select="(k) => onMealMenu(index, meal.id, k)">
                            <Button
                                size="sm"
                                variant="secondary"
                                :disabled="busyMealIds.has(meal.id)"
                                :aria-label="`${meal.name}, ${formatInt(meal.macros.calories)} kcal. Options`"
                            >
                                <span class="max-w-48 truncate">{{ meal.name }}</span>
                                <span class="text-xs text-fg-muted tabular-nums">{{
                                    formatInt(meal.macros.calories)
                                }}</span>
                            </Button>
                        </Menu>
                    </li>
                </ul>
                <Badge v-else-if="!day.isPast" tone="warn" class="self-start">Nothing planned</Badge>
                <span v-else class="self-center text-sm text-fg-subtle">No plan</span>
            </div>
        </div>
    </Card>
</template>
