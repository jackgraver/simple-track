<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { WorkoutPlan } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { computed, ref, watch } from "vue";
import { GripVertical, CalendarDays, Plus } from "lucide-vue-next";
import { Badge, Button, Card, EmptyState, ErrorState, PageHeader, Section, Select, Skeleton, useDelayedFlag } from "~/shared/ui";
import { useRouter } from "vue-router";
import { apiClient } from "~/api/client";
import {
    useActivateWorkoutProgram,
    useCreateWorkoutPlan,
    useCreateWorkoutProgram,
    useRenameWorkoutProgram,
    useWorkoutPrograms,
} from "~/api/workout/queries";

const router = useRouter();

const { data, isPending, isError, error, refetch } = useWorkoutPrograms();
const showSkeleton = useDelayedFlag(() => isPending.value);
const createProgram = useCreateWorkoutProgram();
const renameProgram = useRenameWorkoutProgram();
const activateProgram = useActivateWorkoutProgram();
const createPlan = useCreateWorkoutPlan();
const selectedProgramId = ref<number | null>(null);
const selectedProgram = computed(
    () =>
        data.value?.programs.find(
            (program) => program.ID === selectedProgramId.value,
        ) ??
        data.value?.programs.find((program) => program.is_active) ??
        data.value?.programs[0],
);
watch(
    () => selectedProgram.value?.ID,
    (id) => {
        if (id) selectedProgramId.value = id;
    },
    { immediate: true },
);
const plans = computed(() => selectedProgram.value?.plans || []);
const refresh = () => refetch();

const failToast = (what: string, err: unknown) =>
    toast.push(`${what}: ${getApiErrorMessage(err, "Something went wrong")}`, "error");

const promptCreateProgram = async () => {
    const name = window.prompt("Workout plan name");
    if (!name?.trim()) return;
    try {
        await createProgram.mutateAsync(name.trim());
    } catch (err: unknown) {
        failToast("Failed to create plan", err);
    }
};
const makeActive = async () => {
    if (!selectedProgram.value) return;
    try {
        await activateProgram.mutateAsync(selectedProgram.value.ID);
    } catch (err: unknown) {
        failToast("Failed to activate plan", err);
    }
};
const promptRenameProgram = async () => {
    if (!selectedProgram.value) return;
    const name = window.prompt(
        "Rename workout plan",
        selectedProgram.value.name,
    );
    if (!name?.trim()) return;
    try {
        await renameProgram.mutateAsync({
            id: selectedProgram.value.ID,
            name: name.trim(),
        });
    } catch (err: unknown) {
        failToast("Failed to rename plan", err);
    }
};
const programOptions = computed(() =>
    (data.value?.programs ?? []).map((program) => ({
        label: `${program.name}${program.is_active ? " (Active)" : ""}`,
        value: program.ID,
    })),
);
const programModel = computed({
    get: () => selectedProgramId.value ?? "",
    set: (v: number | "") => {
        selectedProgramId.value = v === "" ? null : v;
    },
});
const promptCreateDay = async (dayOfWeek: number | null = null) => {
    if (!selectedProgram.value) return;
    const name = window.prompt(
        dayOfWeek === null
            ? "Routine name"
            : `Routine name for ${dayNames[dayOfWeek]}`,
    );
    if (!name?.trim()) return;
    try {
        await createPlan.mutateAsync({
            programId: selectedProgram.value.ID,
            name: name.trim(),
            dayOfWeek,
        });
        await refresh();
        toast.push(`Created ${name.trim()}`, "success");
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to create day: " + message, "error");
    }
};

const dayNames = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];

const weekdaysSundayFirst = [
    { dow: 0, label: "Sunday" },
    { dow: 1, label: "Monday" },
    { dow: 2, label: "Tuesday" },
    { dow: 3, label: "Wednesday" },
    { dow: 4, label: "Thursday" },
    { dow: 5, label: "Friday" },
    { dow: 6, label: "Saturday" },
];

const getDayName = (dayOfWeek: number | null): string | undefined => {
    if (dayOfWeek === null) return undefined;
    return dayNames[dayOfWeek];
};

const assignedDays = (plan: WorkoutPlan) =>
    plan.assigned_days ?? (plan.day_of_week === null ? [] : [plan.day_of_week]);

const planByDay = computed(() => {
    const m: Partial<Record<number, WorkoutPlan>> = {};
    for (const p of plans.value) {
        const days =
            p.assigned_days ?? (p.day_of_week === null ? [] : [p.day_of_week]);
        for (const day of days) {
            m[day] = p;
        }
    }
    return m;
});

const planDays = computed(() => plans.value);

const draggingPlanId = ref<number | null>(null);
const draggingSourceDay = ref<number | null>(null);
const dropTargetKey = ref<number | "pool" | null>(null);
const dragHandleActive = ref(false);

const onHandleMouseDown = () => {
    dragHandleActive.value = true;
    const onUp = () => {
        dragHandleActive.value = false;
        window.removeEventListener("mouseup", onUp);
    };
    window.addEventListener("mouseup", onUp);
};

const todayDow = () => new Date().getDay();

const goDetail = (id: number) => {
    if (selectedProgram.value) {
        router.push({
            name: "gym-program-day-detail",
            params: {
                programId: String(selectedProgram.value.ID),
                id: String(id),
            },
        });
        return;
    }
    router.push({ name: "gym-plan-detail", params: { id: String(id) } });
};

const onDragStart = (
    e: DragEvent,
    plan: WorkoutPlan,
    sourceDay: number | null,
) => {
    if (!dragHandleActive.value) {
        e.preventDefault();
        return;
    }
    draggingPlanId.value = plan.ID;
    draggingSourceDay.value = sourceDay;
    e.dataTransfer?.setData("text/plain", String(plan.ID));
    if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = "move";
    }
};

const onDragEnd = () => {
    draggingPlanId.value = null;
    draggingSourceDay.value = null;
    dropTargetKey.value = null;
};

const onDragOverDay = (e: DragEvent, dow: number) => {
    e.preventDefault();
    dropTargetKey.value = dow;
};

const onDragOverPool = (e: DragEvent) => {
    e.preventDefault();
    dropTargetKey.value = "pool";
};

const assignPlanToDay = async (plan: WorkoutPlan, dayOfWeek: number) => {
    try {
        await apiClient.post<{ plan: WorkoutPlan }>(
            `workout/plans/${plan.ID}/assign-day`,
            {
                day_of_week: dayOfWeek,
            },
        );
        toast.push(
            `Assigned ${plan.name} to ${dayNames[dayOfWeek]}`,
            "success",
        );
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to assign day: " + message, "error");
    }
};

const unassignPlanFromDay = async (plan: WorkoutPlan, dayOfWeek: number) => {
    try {
        await apiClient.delete<{ plan: WorkoutPlan }>(
            `workout/plans/${plan.ID}/assign-day`,
            { params: { day_of_week: dayOfWeek } },
        );
        toast.push(
            `Unassigned ${plan.name} from ${getDayName(dayOfWeek) ?? "day"}`,
            "success",
        );
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to unassign day: " + message, "error");
    }
};

const onDropDay = async (dow: number) => {
    const rawId = draggingPlanId.value;
    dropTargetKey.value = null;
    draggingPlanId.value = null;
    if (rawId === null) return;
    const draggedPlan = plans.value.find((p) => p.ID === rawId);
    if (!draggedPlan || assignedDays(draggedPlan).includes(dow)) return;
    const occupant = plans.value.find(
        (p) => assignedDays(p).includes(dow) && p.ID !== draggedPlan.ID,
    );
    if (occupant) {
        const confirmed = await dialogManager.confirm({
            title: "Day Already Assigned",
            message:
                dayNames[dow] +
                ' is currently assigned to "' +
                occupant.name +
                '". Assigning "' +
                draggedPlan.name +
                '" will unassign "' +
                occupant.name +
                '". Continue?',
            confirmText: "Yes, Reassign",
            cancelText: "Cancel",
        });
        if (!confirmed) return;
    }
    await assignPlanToDay(draggedPlan, dow);
};

const onDropPool = async () => {
    const rawId = draggingPlanId.value;
    const sourceDay = draggingSourceDay.value;
    dropTargetKey.value = null;
    draggingPlanId.value = null;
    draggingSourceDay.value = null;
    if (rawId === null || sourceDay === null) return;
    const draggedPlan = plans.value.find((p) => p.ID === rawId);
    if (!draggedPlan || !assignedDays(draggedPlan).includes(sourceDay)) return;
    await unassignPlanFromDay(draggedPlan, sourceDay);
};

const planCardClasses =
    "rounded-control border border-line bg-surface-2 px-3 py-2 text-left transition-colors duration-100 hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent";
const planDraggingClass = (planId: number) =>
    draggingPlanId.value === planId ? "opacity-60" : "";

const exerciseCountLabel = (n: number) => `${n} exercise${n === 1 ? "" : "s"}`;
</script>

<template>
    <div class="flex w-full flex-col gap-6 pb-8 pt-2">
        <PageHeader title="Workout schedule">
            <template #meta>Choose an active plan, then arrange its routines through the week.</template>
        </PageHeader>
        <ErrorState
            v-if="isError"
            :error="error"
            title="Couldn't load your workout plans"
            @retry="refetch()"
        />
        <template v-else-if="isPending">
            <div v-if="showSkeleton" class="flex flex-col gap-4" role="status" aria-busy="true">
                <Skeleton height="2.75rem" width="24rem" />
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-7">
                    <Skeleton v-for="n in 7" :key="n" shape="block" height="8rem" />
                </div>
            </div>
        </template>
        <template v-else>
            <div v-if="data?.programs?.length" class="flex flex-col gap-2 lg:flex-row lg:flex-wrap lg:items-center">
                <div class="lg:w-64">
                    <Select v-model="programModel" :options="programOptions" aria-label="Workout plan" />
                </div>
                <Button
                    :disabled="selectedProgram?.is_active"
                    :loading="activateProgram.isPending.value"
                    @click="makeActive"
                >
                    Make active
                </Button>
                <Button :loading="renameProgram.isPending.value" @click="promptRenameProgram">Rename</Button>
                <Button variant="primary" :loading="createProgram.isPending.value" @click="promptCreateProgram">
                    New workout plan
                </Button>
            </div>
            <EmptyState
                v-if="!selectedProgram"
                :icon="CalendarDays"
                title="No workout plans yet"
                description="Create one to configure its week."
            >
                <template #action>
                    <Button variant="primary" :loading="createProgram.isPending.value" @click="promptCreateProgram">
                        Create your first workout plan
                    </Button>
                </template>
            </EmptyState>
            <template v-else>
                <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-7">
                    <Card
                        v-for="slot in weekdaysSundayFirst"
                        :key="slot.dow"
                        padding="sm"
                        class="min-h-32 transition-shadow"
                        :class="dropTargetKey === slot.dow ? 'ring-2 ring-accent' : ''"
                        @dragover="onDragOverDay($event, slot.dow)"
                        @drop.prevent="onDropDay(slot.dow)"
                    >
                        <div class="flex flex-col gap-2">
                        <div class="flex flex-wrap items-center gap-1.5">
                            <span class="text-sm font-semibold text-fg">{{ slot.label }}</span>
                            <Badge v-if="slot.dow === todayDow()" tone="accent">Today</Badge>
                        </div>
                        <div
                            v-if="planByDay[slot.dow]"
                            role="button"
                            tabindex="0"
                            draggable="true"
                            :class="[planCardClasses, planDraggingClass(planByDay[slot.dow]!.ID)]"
                            @dragstart="onDragStart($event, planByDay[slot.dow]!, slot.dow)"
                            @dragend="onDragEnd"
                            @click="goDetail(planByDay[slot.dow]!.ID)"
                            @keydown.enter="goDetail(planByDay[slot.dow]!.ID)"
                        >
                            <div class="flex items-start gap-2">
                                <span
                                    class="mt-0.5 hidden shrink-0 cursor-grab text-fg-subtle hover:text-fg-muted active:cursor-grabbing lg:block"
                                    aria-hidden="true"
                                    @mousedown="onHandleMouseDown"
                                >
                                    <GripVertical class="size-4" />
                                </span>
                                <div class="min-w-0 flex-1">
                                    <div class="truncate text-sm font-medium text-fg">{{ planByDay[slot.dow]!.name }}</div>
                                    <div class="mt-1 text-xs text-fg-muted">
                                        {{ exerciseCountLabel(planByDay[slot.dow]!.exercises.length) }}
                                    </div>
                                    <div
                                        v-if="planByDay[slot.dow]!.planned_cardio_type?.trim()"
                                        class="mt-1 truncate text-2xs text-fg-muted"
                                    >
                                        Cardio: {{ planByDay[slot.dow]!.planned_cardio_type }}
                                    </div>
                                </div>
                            </div>
                        </div>
                        <Button v-else variant="secondary" block @click="promptCreateDay(slot.dow)">
                            <template #icon-left><Plus class="size-4" /></template>
                            Add
                        </Button>
                        </div>
                    </Card>
                </div>
                <Section title="Plan Days">
                    <template #actions>
                        <Button size="sm" :loading="createPlan.isPending.value" @click="promptCreateDay()">
                            Create day
                        </Button>
                    </template>
                    <Card
                        padding="sm"
                        class="min-h-18 transition-shadow"
                        :class="dropTargetKey === 'pool' ? 'ring-2 ring-accent' : ''"
                        @dragover="onDragOverPool($event)"
                        @drop.prevent="onDropPool()"
                    >
                        <div class="flex flex-wrap content-start gap-2">
                        <EmptyState
                            v-if="planDays.length === 0"
                            class="w-full"
                            title="No plan days yet"
                            description="Create one to add it to the week."
                        />
                        <div
                            v-for="p in planDays"
                            :key="p.ID"
                            role="button"
                            tabindex="0"
                            draggable="true"
                            :class="[planCardClasses, planDraggingClass(p.ID)]"
                            @dragstart="onDragStart($event, p, null)"
                            @dragend="onDragEnd"
                            @click="goDetail(p.ID)"
                            @keydown.enter="goDetail(p.ID)"
                        >
                            <div class="flex items-start gap-2">
                                <span
                                    class="mt-0.5 hidden shrink-0 cursor-grab text-fg-subtle hover:text-fg-muted active:cursor-grabbing lg:block"
                                    aria-hidden="true"
                                    @mousedown="onHandleMouseDown"
                                >
                                    <GripVertical class="size-4" />
                                </span>
                                <div class="min-w-0 flex-1">
                                    <div class="truncate text-sm font-medium text-fg">{{ p.name }}</div>
                                    <div class="mt-1 text-xs text-fg-muted">{{ exerciseCountLabel(p.exercises.length) }}</div>
                                </div>
                            </div>
                        </div>
                        </div>
                    </Card>
                </Section>
            </template>
        </template>
    </div>
</template>
