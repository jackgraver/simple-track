<script setup lang="ts">
import { computed, ref, useId } from "vue";
import { Check, ChevronDown, ChevronRight, Circle, MoreVertical, SquarePen, Trash2 } from "lucide-vue-next";
import type { Meal } from "~/types/diet";
import { Badge, Button, IconButton, ListRow, Menu, type MenuItem } from "~/shared/ui";
import { macroLine, mealMacros } from "./dietDay";
import MealItemsList from "./MealItemsList.vue";

const props = defineProps<{
    meal: Meal;
    kind: "planned" | "logged";
    /** Planned only: a day log already points at this meal. */
    logged?: boolean;
    /** Planned only: one-tap Log is offered (the backend logs planned meals to today only). */
    canLog?: boolean;
    onLog?: (meal: Meal) => void | Promise<void>;
    onEdit: (meal: Meal) => void;
    onDelete?: (meal: Meal) => void | Promise<void>;
    /** Logged only: not saved yet (subtle tint; still deletable). */
    pending?: boolean;
}>();

const open = ref(false);
const panelId = useId();
const subtitle = computed(() => macroLine(mealMacros(props.meal)));

// The handlers toast their own failures; this only drives the pending state.
const busy = ref(false);
async function run(fn?: (meal: Meal) => void | Promise<void>) {
    if (!fn || busy.value) return;
    busy.value = true;
    try {
        await fn(props.meal);
    } finally {
        busy.value = false;
    }
}

const menuItems = computed<MenuItem[]>(() =>
    props.kind === "planned"
        ? [{ key: "edit", label: "Edit and log", icon: SquarePen }]
        : [
              { key: "edit", label: "Edit", icon: SquarePen },
              { key: "delete", label: "Delete", icon: Trash2, tone: "danger" },
          ],
);

function onMenu(key: string) {
    if (key === "edit") props.onEdit(props.meal);
    if (key === "delete") void run(props.onDelete);
}
</script>

<template>
    <ListRow :pending="busy || pending" :aria-expanded="open" :aria-controls="panelId" @click="open = !open">
        <template #leading>
            <Check v-if="kind === 'planned' && logged" class="size-5 text-success" aria-hidden="true" />
            <Circle v-else-if="kind === 'planned'" class="size-5 text-fg-subtle" aria-hidden="true" />
            <ChevronDown v-else-if="open" class="size-4" aria-hidden="true" />
            <ChevronRight v-else class="size-4" aria-hidden="true" />
        </template>
        <span :class="kind === 'planned' && logged ? 'text-fg-muted' : 'font-medium'">{{ meal.name }}</span>
        <template #subtitle>{{ subtitle }}</template>
        <template #end>
            <Badge v-if="kind === 'planned' && logged" tone="success">Logged</Badge>
            <template v-else>
                <Button
                    v-if="kind === 'planned' && canLog"
                    variant="secondary"
                    size="sm"
                    :loading="busy"
                    :aria-label="`Log ${meal.name}`"
                    @click="run(onLog)"
                >
                    Log
                </Button>
                <Menu :items="menuItems" @select="onMenu">
                    <IconButton :label="`More actions for ${meal.name}`" :disabled="busy">
                        <MoreVertical />
                    </IconButton>
                </Menu>
            </template>
        </template>
    </ListRow>
    <li v-if="open" :id="panelId" class="bg-surface-2">
        <MealItemsList :meal="meal" />
    </li>
</template>
