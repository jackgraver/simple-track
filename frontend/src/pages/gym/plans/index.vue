<script setup lang="ts">
import type { WorkoutPlan } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";

const router = useRouter();

const { data, isPending, refetch } = useQuery({
    queryKey: ["workout", "plans", "all"],
    queryFn: async () => {
        const res = await apiClient.get<{ plans: WorkoutPlan[] }>(
            "/workout/plans/all",
        );
        return res.data;
    },
});

const plans = computed(() => data.value?.plans || []);
const refresh = () => refetch();

const dayNames = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];

const weekdaysMondayFirst = [
    { dow: 1, label: "Monday" },
    { dow: 2, label: "Tuesday" },
    { dow: 3, label: "Wednesday" },
    { dow: 4, label: "Thursday" },
    { dow: 5, label: "Friday" },
    { dow: 6, label: "Saturday" },
    { dow: 0, label: "Sunday" },
];

const getDayName = (dayOfWeek: number | null): string | undefined => {
    if (dayOfWeek === null) return undefined;
    return dayNames[dayOfWeek];
};

const planByDay = computed(() => {
    const m: Partial<Record<number, WorkoutPlan>> = {};
    for (const p of plans.value) {
        if (p.day_of_week !== null) {
            m[p.day_of_week] = p;
        }
    }
    return m;
});

const unassignedPlans = computed(() =>
    plans.value.filter((p) => p.day_of_week === null),
);

const draggingPlanId = ref<number | null>(null);
const dropTargetKey = ref<number | "pool" | null>(null);

const todayDow = () => new Date().getDay();

const goDetail = (id: number) => {
    router.push({ name: "gym-plan-detail", params: { id: String(id) } });
};

const onDragStart = (e: DragEvent, plan: WorkoutPlan) => {
    draggingPlanId.value = plan.ID;
    e.dataTransfer?.setData("text/plain", String(plan.ID));
    if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = "move";
    }
};

const onDragEnd = () => {
    draggingPlanId.value = null;
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

const unassignPlanFromDay = async (plan: WorkoutPlan) => {
    try {
        await apiClient.delete<{ plan: WorkoutPlan }>(
            `workout/plans/${plan.ID}/assign-day`,
        );
        toast.push(
            `Unassigned ${plan.name} from ${getDayName(plan.day_of_week) ?? "day"}`,
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
    if (!draggedPlan || draggedPlan.day_of_week === dow) return;
    const occupant = plans.value.find(
        (p) => p.day_of_week === dow && p.ID !== draggedPlan.ID,
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
    dropTargetKey.value = null;
    draggingPlanId.value = null;
    if (rawId === null) return;
    const draggedPlan = plans.value.find((p) => p.ID === rawId);
    if (!draggedPlan || draggedPlan.day_of_week === null) return;
    await unassignPlanFromDay(draggedPlan);
};

const planCardClasses =
    "cursor-grab active:cursor-grabbing rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-left transition-opacity hover:bg-thirdBg/40";
const planDraggingClass = (planId: number) =>
    draggingPlanId.value === planId ? "opacity-60" : "";

const exerciseCountLabel = (n: number) =>
    `${n} exercise${n === 1 ? "" : "s"}`;
</script>

<template>
    <div class="flex w-full max-w-6xl flex-col gap-6 pb-8 pt-2">
        <div class="flex flex-col gap-1 border-b border-(--color-border) pb-3">
            <h1 class="m-0 text-xl font-semibold tracking-tight text-textPrimary">
                Workout schedule
            </h1>
            <p class="m-0 text-sm text-textSecondary">
                Drag plans onto a weekday or into Unassigned. Click a plan to
                edit exercises.
            </p>
        </div>
        <div v-if="isPending" class="text-center text-sm text-textSecondary">
            Loading…
        </div>
        <div
            v-else-if="!plans.length"
            class="rounded-lg border border-(--color-border) bg-firstBg p-8 text-center text-sm text-textSecondary"
        >
            No workout plans yet.
        </div>
        <template v-else>
            <div
                class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-7"
            >
                <div
                    v-for="slot in weekdaysMondayFirst"
                    :key="slot.dow"
                    class="flex min-h-[8rem] flex-col gap-2 rounded-lg border border-(--color-border) bg-firstBg p-3 transition-shadow"
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
                                onDragStart($event, planByDay[slot.dow]!)
                            "
                            @dragend="onDragEnd"
                            @click="goDetail(planByDay[slot.dow]!.ID)"
                            @keydown.enter="goDetail(planByDay[slot.dow]!.ID)"
                        >
                            <div
                                class="truncate text-sm font-medium text-textPrimary"
                            >
                                {{ planByDay[slot.dow]!.name }}
                            </div>
                            <div class="mt-1 text-xs text-textSecondary">
                                {{
                                    exerciseCountLabel(
                                        planByDay[slot.dow]!.exercises.length,
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
                                    planByDay[slot.dow]!.planned_cardio_type
                                }}
                            </div>
                        </div>
                    </template>
                    <div
                        v-else
                        class="flex flex-1 items-center justify-center rounded-md border border-dashed border-(--color-border) px-2 py-6 text-center text-xs text-textSecondary"
                    >
                        Drop a plan here
                    </div>
                </div>
            </div>
            <section class="flex flex-col gap-2">
                <h2 class="m-0 text-sm font-semibold text-textPrimary">
                    Unassigned
                </h2>
                <div
                    class="flex min-h-[4.5rem] flex-wrap content-start gap-2 rounded-lg border border-(--color-border) bg-firstBg p-3 transition-shadow"
                    :class="
                        dropTargetKey === 'pool'
                            ? 'ring-2 ring-(--color-cf-red)/60'
                            : ''
                    "
                    @dragover="onDragOverPool($event)"
                    @drop.prevent="onDropPool()"
                >
                    <p
                        v-if="unassignedPlans.length === 0"
                        class="m-0 w-full text-center text-xs text-textSecondary"
                    >
                        No unassigned plans — drag a scheduled plan here to
                        remove its weekday.
                    </p>
                    <div
                        v-for="p in unassignedPlans"
                        :key="p.ID"
                        role="button"
                        tabindex="0"
                        draggable="true"
                        :class="[planCardClasses, planDraggingClass(p.ID)]"
                        @dragstart="onDragStart($event, p)"
                        @dragend="onDragEnd"
                        @click="goDetail(p.ID)"
                        @keydown.enter="goDetail(p.ID)"
                    >
                        <div class="truncate text-sm font-medium text-textPrimary">
                            {{ p.name }}
                        </div>
                        <div class="mt-1 text-xs text-textSecondary">
                            {{ exerciseCountLabel(p.exercises.length) }}
                        </div>
                    </div>
                </div>
            </section>
        </template>
    </div>
</template>
