<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useDietDays } from "~/api/diet/queries";
import { formatInt } from "~/pages/diet/home/dietDay";
import { dateAtDietDayOffset } from "~/utils/dateUtil";
import { Button, Checkbox, ErrorState, Overlay, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { kcalLine, plannedKcal, shortDayLabel, type PlanDay } from "./domain/planWeek";

const open = defineModel<boolean>("open", { required: true });
const props = defineProps<{ source: PlanDay | null; pending: boolean }>();
const emit = defineEmits<{ copy: [offsets: number[]] }>();

/** The next 14 days after the source, never before today. */
const CANDIDATES = 14;
const offsets = computed(() => {
    if (!props.source) return [];
    const first = Math.min(props.source.offset - 1, 0);
    return Array.from({ length: CANDIDATES }, (_, i) => first - i).filter((o) => o !== props.source?.offset);
});
const queries = useDietDays(offsets);
const loading = computed(() => queries.value.some((q) => q.isPending));
const failed = computed(() => queries.value.find((q) => q.isError));
const showSkeleton = useDelayedFlag(loading);

const days = computed(() =>
    offsets.value.map((offset, i) => {
        const day = queries.value[i]?.data?.day;
        return {
            offset,
            date: dateAtDietDayOffset(offset),
            meals: day?.plannedMeals.length ?? 0,
            kcal: plannedKcal(day?.plannedMeals ?? []),
            target: day?.plan?.calories ?? 0,
        };
    }),
);

const selected = ref<Set<number>>(new Set());
watch(open, (o) => {
    if (o) selected.value = new Set();
});

function toggle(offset: number, on: boolean | "indeterminate") {
    const next = new Set(selected.value);
    if (on === true) next.add(offset);
    else next.delete(offset);
    selected.value = next;
}
function selectEmpty() {
    selected.value = new Set(days.value.filter((d) => d.meals === 0).map((d) => d.offset));
}

/** Accessible name: the same facts the row shows, e.g. "Mon 16 Nov, 3 planned, 1,800 / 2,200 kcal". */
function dayCheckboxLabel(d: { date: Date; meals: number; kcal: number; target: number }): string {
    const month = d.date.toLocaleDateString(undefined, { month: "short" });
    const plan = d.meals === 0 ? "nothing planned" : `${d.meals} planned, ${kcalLine(d.kcal, d.target)}`;
    return `${shortDayLabel(d.date)} ${month}, ${plan}`;
}

const title = computed(() => (props.source ? `Copy ${shortDayLabel(props.source.date)} to…` : "Copy day"));
const description = computed(() => {
    const s = props.source;
    if (!s) return undefined;
    const n = s.meals.length;
    return `Adds its ${n} ${n === 1 ? "meal" : "meals"} (${formatInt(s.planned.calories)} kcal) after anything already planned.`;
});
const count = computed(() => selected.value.size);
</script>

<template>
    <Overlay v-model:open="open" :title="title" :description="description">
        <div class="flex flex-col gap-3">
            <div class="flex flex-wrap gap-2">
                <Button size="sm" variant="ghost" :disabled="loading" @click="selectEmpty">Select empty days</Button>
                <Button size="sm" variant="ghost" :disabled="count === 0" @click="selected = new Set()">Clear</Button>
            </div>
            <ErrorState
                v-if="failed"
                compact
                title="Couldn't load the next days"
                :error="failed.error"
                @retry="queries.forEach((q) => q.isError && q.refetch())"
            />
            <SkeletonList v-else-if="showSkeleton" :rows="7" :leading="false" />
            <ul v-else-if="!loading" class="m-0 flex list-none flex-col divide-y divide-line p-0">
                <li v-for="d in days" :key="d.offset">
                    <Checkbox
                        :model-value="selected.has(d.offset)"
                        :label="dayCheckboxLabel(d)"
                        @update:model-value="(v) => toggle(d.offset, v)"
                    >
                        <span class="flex flex-col py-1">
                            <span class="font-medium">
                                {{ shortDayLabel(d.date) }}
                                <span class="font-normal text-fg-muted">{{
                                    d.date.toLocaleDateString(undefined, { month: "short" })
                                }}</span>
                            </span>
                            <span class="text-xs text-fg-muted tabular-nums">
                                <template v-if="d.meals === 0">Nothing planned</template>
                                <template v-else
                                    >{{ d.meals }} planned · {{ kcalLine(d.kcal, d.target) }}</template
                                >
                            </span>
                        </span>
                    </Checkbox>
                </li>
            </ul>
        </div>
        <template #footer>
            <Button variant="ghost" @click="open = false">Cancel</Button>
            <Button variant="primary" :disabled="count === 0" :loading="pending" @click="emit('copy', [...selected])">
                {{ count ? `Copy to ${count} ${count === 1 ? "day" : "days"}` : "Copy" }}
            </Button>
        </template>
    </Overlay>
</template>
