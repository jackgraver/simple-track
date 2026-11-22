<script setup lang="ts">
import { offsetToDate } from "~/api/workout/dates";
import { computed } from "vue";
import { ArrowLeftRight, Check, Moon, Plus } from "lucide-vue-next";
import { useSwitchWorkoutPlan } from "~/api/workout/queries";
import { getApiErrorMessage } from "~/api/httpError";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import SwitchWorkoutPlanDialog from "~/pages/gym/components/SwitchWorkoutPlanDialog.vue";
import StepMarker from "~/pages/gym/components/StepMarker.vue";
import StepProgress from "~/pages/gym/components/StepProgress.vue";
import { useWorkoutStore } from "~/pages/gym/logging/store/useWorkoutStore";
import { buildWorkoutSteps, nextStepIndex } from "~/pages/gym/logging/domain/workoutSteps";
import { Button, Card, ErrorState, Menu, Skeleton, useDelayedFlag, type MenuItem } from "~/shared/ui";
import { weekdayLong, type useGymWeek } from "./useGymWeek";

const props = defineProps<{ offset: number; week: ReturnType<typeof useGymWeek> }>();

// The same query and store the logging page uses, so opening it after this is instant.
const store = useWorkoutStore(() => props.offset);
const switchPlan = useSwitchWorkoutPlan();
const showSkeleton = useDelayedFlag(() => store.pending.value);

const plan = computed(() => store.data.value?.day?.workout_plan ?? null);
const steps = computed(() =>
    buildWorkoutSteps({
        log: store.log.value,
        plannedCardio: store.plannedCardio.value,
        loggedCardio: store.loggedCardio.value,
        plannedPreMobility: store.plannedPreMobility.value,
        loggedPreMobility: store.loggedPreMobility.value,
        plannedPostMobility: store.plannedPostMobility.value,
        loggedPostMobility: store.loggedPostMobility.value,
    }),
);
const next = computed(() => nextStepIndex(steps.value));
const doneCount = computed(() => steps.value.filter((s) => s.done).length);

type Phase = "rest" | "fresh" | "progress" | "complete";
const phase = computed<Phase>(() => {
    if (!steps.value.length) return "rest";
    if (next.value === -1) return "complete";
    return doneCount.value > 0 ? "progress" : "fresh";
});

const selected = computed(() => props.week.selected.value);
const dayName = computed(() => {
    if (props.offset === 0) return "Today";
    if (props.offset === 1) return "Yesterday";
    if (props.offset === -1) return "Tomorrow";
    return weekdayLong(selected.value);
});
const dateLine = computed(() =>
    selected.value.toLocaleDateString(undefined, {
        weekday: props.offset === 0 || Math.abs(props.offset) === 1 ? "long" : undefined,
        month: "short",
        day: "numeric",
    }),
);

const scheduled = computed(() => props.week.scheduledFor(selected.value));
const swappedFrom = computed(() =>
    scheduled.value && plan.value && scheduled.value.ID !== plan.value.ID ? scheduled.value.name : null,
);

const counts = computed(() => {
    const lifts = steps.value.filter((s) => s.kind === "lift").length;
    const parts: string[] = [];
    if (lifts) parts.push(`${lifts} ${lifts === 1 ? "lift" : "lifts"}`);
    for (const s of steps.value) if (s.kind !== "lift") parts.push(s.name);
    return parts.join(" · ");
});

const nextUp = computed(() => {
    const n = props.week.nextScheduled.value;
    if (!n) return null;
    const when = n.inDays === 1 ? "tomorrow" : weekdayLong(n.date);
    return `Next up: ${n.plan.name}, ${when}`;
});

const isFuture = computed(() => props.offset < 0);
const primaryLabel = computed(() => {
    if (phase.value === "progress") return "Continue workout";
    if (phase.value === "complete") return "Review workout";
    if (isFuture.value) return "View plan";
    return props.offset === 0 ? "Start workout" : "Log this workout";
});

const loggingRoute = computed(() => ({
    name: "logging",
    query: props.offset === 0 ? {} : { offset: String(props.offset) },
}));

// Plan switching: the active program's days first, then "rest", then the full picker for other programs.
const menuItems = computed<MenuItem[]>(() => {
    const items: MenuItem[] = (props.week.activeProgram.value?.plans ?? []).map((p) => ({
        key: `plan:${p.ID}`,
        label: p.ID === scheduled.value?.ID && p.ID !== plan.value?.ID ? `${p.name} (scheduled)` : p.name,
        icon: p.ID === plan.value?.ID ? Check : undefined,
        disabled: p.ID === plan.value?.ID,
    }));
    if (plan.value) items.push({ key: "none", label: "Make it a rest day", icon: Moon });
    items.push({ key: "more", label: "Other plans…", icon: Plus });
    return items;
});

const onMenu = async (key: string) => {
    if (key === "more") {
        void dialogManager.custom<boolean>({
            title: "Switch workout plan",
            component: SwitchWorkoutPlanDialog,
            componentProps: { dayOffset: props.offset, currentPlanId: plan.value?.ID ?? null },
        });
        return;
    }
    const planId = key === "none" ? null : Number(key.slice("plan:".length));
    try {
        await switchPlan.mutateAsync({ date: offsetToDate(props.offset), planId });
        toast.push(planId == null ? "Marked as a rest day" : "Workout plan updated", "success");
    } catch (err: unknown) {
        toast.push(getApiErrorMessage(err, "Couldn't switch the plan"), "error");
    }
};
</script>

<template>
    <Card>
        <ErrorState
            v-if="store.error.value"
            :error="store.error.value"
            title="Couldn't load this day's workout"
            @retry="store.refetch()"
        />
        <div v-else-if="store.pending.value" class="flex flex-col gap-3" aria-busy="true">
            <template v-if="showSkeleton">
                <Skeleton width="8rem" height="0.75rem" />
                <Skeleton width="10rem" height="2.25rem" />
                <Skeleton shape="block" height="7rem" />
            </template>
        </div>
        <div
            v-else
            class="flex flex-col gap-5 lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] lg:grid-rows-[auto_1fr] lg:gap-x-8"
        >
            <!-- Heading: what day, what workout -->
            <header class="flex flex-col gap-1">
                <p class="m-0 text-xs font-semibold uppercase tracking-wide text-fg-muted">
                    {{ dayName }}<template v-if="dayName !== dateLine"> · {{ dateLine }}</template>
                </p>
                <h2 class="m-0 text-display font-semibold tracking-tight text-fg">
                    {{ phase === "rest" && !plan ? "Rest day" : (plan?.name ?? "Workout") }}
                </h2>
                <p v-if="phase === 'rest'" class="m-0 text-sm text-fg-muted">
                    <template v-if="plan">Nothing planned in {{ plan.name }}.</template>
                    <template v-else-if="nextUp">{{ nextUp }}</template>
                    <template v-else>Nothing scheduled.</template>
                </p>
                <p v-else class="m-0 text-sm text-fg-muted">
                    {{ counts }}<template v-if="swappedFrom"> · swapped from {{ swappedFrom }}</template>
                </p>
            </header>

            <!-- The session at a glance (desktop: right column) -->
            <div v-if="steps.length" class="flex flex-col gap-3 lg:col-start-2 lg:row-span-2 lg:row-start-1">
                <div v-if="phase !== 'fresh'" class="flex flex-col gap-1.5">
                    <div class="flex items-baseline justify-between text-xs">
                        <span class="font-medium" :class="phase === 'complete' ? 'text-success' : 'text-fg'">
                            {{ phase === "complete" ? "All done" : "In progress" }}
                        </span>
                        <span class="text-fg-muted tabular-nums">{{ doneCount }} of {{ steps.length }}</span>
                    </div>
                    <StepProgress :steps="steps" :next="next" />
                </div>
                <ol class="m-0 flex list-none flex-col gap-2 p-0">
                    <li v-for="(s, i) in steps" :key="s.key" class="flex min-w-0 items-center gap-3">
                        <StepMarker :n="i + 1" :state="s.done ? 'done' : i === next ? 'next' : 'todo'" size="sm" />
                        <span
                            class="min-w-0 flex-1 truncate text-sm"
                            :class="s.done ? 'text-fg-muted' : 'text-fg'"
                        >
                            {{ s.name }}
                            <span class="sr-only">{{ s.done ? "(done)" : i === next ? "(up next)" : "" }}</span>
                        </span>
                        <span v-if="s.detail" class="max-w-[45%] shrink-0 truncate text-xs text-fg-muted tabular-nums">
                            {{ s.detail }}
                        </span>
                    </li>
                </ol>
            </div>

            <!-- Actions: thumb reach on mobile (last), under the heading on desktop -->
            <div class="flex gap-2 lg:col-start-1 lg:row-start-2 lg:self-end">
                <template v-if="phase === 'rest'">
                    <Menu :items="menuItems" align="start" @select="onMenu">
                        <Button variant="secondary" size="lg" class="flex-1 lg:flex-none" :loading="switchPlan.isPending.value">
                            <ArrowLeftRight aria-hidden="true" class="size-4" />
                            Train anyway
                        </Button>
                    </Menu>
                    <Button variant="ghost" size="lg" :to="loggingRoute">Log freestyle</Button>
                </template>
                <template v-else>
                    <Button
                        :variant="phase === 'complete' ? 'secondary' : 'primary'"
                        size="lg"
                        class="flex-1"
                        :to="loggingRoute"
                    >
                        {{ primaryLabel }}
                    </Button>
                    <Menu :items="menuItems" align="end" @select="onMenu">
                        <Button variant="secondary" size="lg" :loading="switchPlan.isPending.value" aria-label="Change workout">
                            <ArrowLeftRight aria-hidden="true" class="size-4" />
                            <span class="hidden sm:inline">Change</span>
                        </Button>
                    </Menu>
                </template>
            </div>
        </div>
    </Card>
</template>
