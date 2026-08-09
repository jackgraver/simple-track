<script setup lang="ts">
import type { WorkoutPlan } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { computed, ref, watch } from "vue";
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

const { data, isPending, refetch } = useWorkoutPrograms();
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

const promptCreateProgram = async () => {
    const name = window.prompt("Workout plan name");
    if (name?.trim()) await createProgram.mutateAsync(name.trim());
};
const promptRenameProgram = async () => {
    if (!selectedProgram.value) return;
    const name = window.prompt(
        "Rename workout plan",
        selectedProgram.value.name,
    );
    if (name?.trim())
        await renameProgram.mutateAsync({
            id: selectedProgram.value.ID,
            name: name.trim(),
        });
};
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
        const message = err instanceof Error ? err.message : String(err);
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
        const message = err instanceof Error ? err.message : String(err);
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
        const message = err instanceof Error ? err.message : String(err);
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
    "rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-left transition-opacity hover:bg-thirdBg/40";
const planDraggingClass = (planId: number) =>
    draggingPlanId.value === planId ? "opacity-60" : "";

const exerciseCountLabel = (n: number) => `${n} exercise${n === 1 ? "" : "s"}`;
</script>

<template>
    <div class="flex w-full flex-col gap-6 pb-8 pt-2">
        <div class="flex flex-col gap-1 border-b border-(--color-border) pb-3">
            <h1
                class="m-0 text-xl font-semibold tracking-tight text-textPrimary"
            >
                Workout schedule
            </h1>
            <p class="m-0 text-sm text-textSecondary">
                Choose an active plan, then arrange its routines through the
                week.
            </p>
        </div>
        <div
            v-if="data?.programs?.length"
            class="flex flex-wrap items-center gap-2"
        >
            <select
                v-model="selectedProgramId"
                class="rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm text-textPrimary"
            >
                <option
                    v-for="program in data.programs"
                    :key="program.ID"
                    :value="program.ID"
                >
                    {{ program.name }}{{ program.is_active ? " (Active)" : "" }}
                </option>
            </select>
            <button
                class="rounded-md border border-(--color-border) px-3 py-2 text-sm text-textPrimary"
                :disabled="selectedProgram?.is_active"
                @click="
                    selectedProgram &&
                    activateProgram.mutate(selectedProgram.ID)
                "
            >
                Make active
            </button>
            <button
                class="rounded-md border border-(--color-border) px-3 py-2 text-sm text-textPrimary"
                @click="promptRenameProgram"
            >
                Rename
            </button>
            <button
                class="rounded-md bg-(--color-cf-red) px-3 py-2 text-sm font-medium text-white"
                @click="promptCreateProgram"
            >
                New workout plan
            </button>
        </div>
        <button
            v-else-if="!isPending"
            class="self-start rounded-md bg-(--color-cf-red) px-3 py-2 text-sm font-medium text-white"
            @click="promptCreateProgram"
        >
            Create your first workout plan
        </button>
        <div v-if="isPending" class="text-center text-sm text-textSecondary">
            Loading…
        </div>
        <div
            v-else-if="!selectedProgram"
            class="rounded-lg border border-(--color-border) bg-firstBg p-8 text-center text-sm text-textSecondary"
        >
            No workout plans yet. Create one above to configure its week.
        </div>
        <template v-else>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-7">
                <div
                    v-for="slot in weekdaysSundayFirst"
                    :key="slot.dow"
                    class="flex min-h-32 flex-col gap-2 rounded-lg border border-(--color-border) bg-firstBg p-3 transition-shadow"
                    :class="
                        dropTargetKey === slot.dow
                            ? 'ring-2 ring-(--color-cf-red)/60'
                            : ''
                    "
                    @dragover="onDragOverDay($event, slot.dow)"
                    @drop.prevent="onDropDay(slot.dow)"
                >
                    <div class="flex flex-wrap items-center gap-1.5">
                        <span class="text-sm font-semibold text-textPrimary">{{
                            slot.label
                        }}</span>
                        <span
                            v-if="slot.dow === todayDow()"
                            class="rounded bg-(--color-cf-red)/20 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-(--color-cf-red)"
                            >Today</span
                        >
                    </div>
                    <template v-if="planByDay[slot.dow]">
                        <div
                            role="button"
                            tabindex="0"
                            draggable="true"
                            :class="[
                                planCardClasses,
                                planDraggingClass(planByDay[slot.dow]!.ID),
                            ]"
                            @dragstart="
                                onDragStart(
                                    $event,
                                    planByDay[slot.dow]!,
                                    slot.dow,
                                )
                            "
                            @dragend="onDragEnd"
                            @click="goDetail(planByDay[slot.dow]!.ID)"
                            @keydown.enter="goDetail(planByDay[slot.dow]!.ID)"
                        >
                            <div class="flex items-start gap-2">
                                <div
                                    class="mt-0.5 shrink-0 cursor-grab text-textSecondary/40 hover:text-textSecondary active:cursor-grabbing"
                                    @mousedown="onHandleMouseDown"
                                >
                                    <svg
                                        width="10"
                                        height="14"
                                        viewBox="0 0 10 14"
                                        fill="currentColor"
                                    >
                                        <circle cx="2" cy="2" r="1.5" />
                                        <circle cx="8" cy="2" r="1.5" />
                                        <circle cx="2" cy="7" r="1.5" />
                                        <circle cx="8" cy="7" r="1.5" />
                                        <circle cx="2" cy="12" r="1.5" />
                                        <circle cx="8" cy="12" r="1.5" />
                                    </svg>
                                </div>
                                <div class="min-w-0 flex-1">
                                    <div
                                        class="truncate text-sm font-medium text-textPrimary"
                                    >
                                        {{ planByDay[slot.dow]!.name }}
                                    </div>
                                    <div
                                        class="mt-1 text-xs text-textSecondary"
                                    >
                                        {{
                                            exerciseCountLabel(
                                                planByDay[slot.dow]!.exercises
                                                    .length,
                                            )
                                        }}
                                    </div>
                                    <div
                                        v-if="
                                            planByDay[
                                                slot.dow
                                            ]!.planned_cardio_type?.trim()
                                        "
                                        class="mt-1 truncate text-[11px] text-textSecondary"
                                    >
                                        Cardio:
                                        {{
                                            planByDay[slot.dow]!
                                                .planned_cardio_type
                                        }}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </template>
                    <div
                        v-else
                        class="flex flex-1 items-center justify-center rounded-md border border-dashed border-(--color-border) px-2 py-6 text-center text-xs text-textSecondary"
                        role="button"
                        tabindex="0"
                        @click="promptCreateDay(slot.dow)"
                        @keydown.enter="promptCreateDay(slot.dow)"
                    >
                        Add a routine
                    </div>
                </div>
            </div>
            <section class="flex flex-col gap-2">
                <div class="flex flex-wrap items-center justify-between gap-2">
                    <h2 class="m-0 text-sm font-semibold text-textPrimary">
                        Plan Days
                    </h2>
                    <button
                        type="button"
                        class="rounded-md bg-(--color-cf-red) px-3 py-1.5 text-xs font-medium text-white"
                        @click="promptCreateDay()"
                    >
                        Create day
                    </button>
                </div>
                <div
                    class="flex min-h-18 flex-wrap content-start gap-2 rounded-lg border border-(--color-border) bg-firstBg p-3 transition-shadow"
                    :class="
                        dropTargetKey === 'pool'
                            ? 'ring-2 ring-(--color-cf-red)/60'
                            : ''
                    "
                    @dragover="onDragOverPool($event)"
                    @drop.prevent="onDropPool()"
                >
                    <p
                        v-if="planDays.length === 0"
                        class="m-0 w-full text-center text-xs text-textSecondary"
                    >
                        No plan days yet — create one to add it to the week.
                    </p>
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
                            <div
                                class="mt-0.5 shrink-0 cursor-grab text-textSecondary/40 hover:text-textSecondary active:cursor-grabbing"
                                @mousedown="onHandleMouseDown"
                            >
                                <svg
                                    width="10"
                                    height="14"
                                    viewBox="0 0 10 14"
                                    fill="currentColor"
                                >
                                    <circle cx="2" cy="2" r="1.5" />
                                    <circle cx="8" cy="2" r="1.5" />
                                    <circle cx="2" cy="7" r="1.5" />
                                    <circle cx="8" cy="7" r="1.5" />
                                    <circle cx="2" cy="12" r="1.5" />
                                    <circle cx="8" cy="12" r="1.5" />
                                </svg>
                            </div>
                            <div class="min-w-0 flex-1">
                                <div
                                    class="truncate text-sm font-medium text-textPrimary"
                                >
                                    {{ p.name }}
                                </div>
                                <div class="mt-1 text-xs text-textSecondary">
                                    {{ exerciseCountLabel(p.exercises.length) }}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </template>
    </div>
</template>
