<script setup lang="ts">
import { computed } from "vue";
import { Check, Moon, Plus } from "lucide-vue-next";
import type { WorkoutPlan } from "~/types/workout";
import { Button, List, ListRow, Overlay } from "~/shared/ui";
import { type DaySlot, DAY_SHORT, assignedDays, planMeta } from "../domain/week";

// Pick what a weekday trains: one of the program's routines, or rest. The tap-friendly twin of drag-and-drop.
const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{ day: DaySlot | null; plans: WorkoutPlan[]; programId: number | null }>();
const emit = defineEmits<{ choose: [plan: WorkoutPlan]; rest: []; create: [] }>();

const current = computed(() => props.day?.plan ?? null);

/** "Also Mon, Wed" for routines already on other days, so picking one twice is a choice, not a surprise. */
const otherDays = (p: WorkoutPlan) => {
    const days = assignedDays(p).filter((d) => d !== props.day?.dow);
    return days.length ? `also ${days.map((d) => DAY_SHORT[d]).join(", ")}` : "";
};
const subtitle = (p: WorkoutPlan) => [...planMeta(p), otherDays(p)].filter(Boolean).join(" · ");
</script>

<template>
    <Overlay v-model:open="open" :title="day?.name ?? 'Day'" description="Choose what to train, or make it a rest day." size="sm">
        <List v-if="day" class="-mx-4 lg:-mx-5">
            <ListRow
                v-for="p in plans"
                :key="p.ID"
                :selected="current?.ID === p.ID"
                @click="emit('choose', p)"
            >
                {{ p.name }}
                <template #subtitle>{{ subtitle(p) }}</template>
                <template v-if="current?.ID === p.ID" #trailing>
                    <Check class="size-4 text-accent" aria-hidden="true" />
                    <span class="sr-only">Current</span>
                </template>
            </ListRow>
            <ListRow :selected="!current" @click="emit('rest')">
                <template #leading><Moon class="size-4" aria-hidden="true" /></template>
                Rest day
                <template v-if="!current" #trailing>
                    <Check class="size-4 text-accent" aria-hidden="true" />
                    <span class="sr-only">Current</span>
                </template>
            </ListRow>
            <ListRow @click="emit('create')">
                <template #leading><Plus class="size-4" aria-hidden="true" /></template>
                <span class="text-fg-muted">New routine for {{ day.name }}</span>
            </ListRow>
        </List>
        <template v-if="current && programId !== null" #footer>
            <Button :to="{ name: 'gym-program-day-detail', params: { programId: String(programId), id: String(current.ID) } }">
                Open {{ current.name }}
            </Button>
        </template>
    </Overlay>
</template>
