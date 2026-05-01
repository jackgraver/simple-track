<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import AddExerciseDialog from "~/pages/gym/plans/components/AddExerciseDialog.vue";
import CreateExerciseForPlanDialog from "~/pages/gym/plans/components/CreateExerciseForPlanDialog.vue";
import EditExerciseDialog from "~/pages/gym/plans/components/EditExerciseDialog.vue";
import { X, Plus, ChevronUp, ChevronDown, Pencil } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiPUT } from "~/api/client";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";

const route = useRoute();

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

const planIdParam = computed(() => {
    const raw = route.params.id;
    const n = typeof raw === "string" ? Number.parseInt(raw, 10) : NaN;
    return Number.isFinite(n) ? n : null;
});

const plan = computed(() => {
    const id = planIdParam.value;
    if (id === null) return null;
    return plans.value.find((p) => p.ID === id) ?? null;
});

const dayNames = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
];

const getDayName = (dayOfWeek: number | null): string => {
    if (dayOfWeek === null) return "Unassigned";
    return dayNames[dayOfWeek] ?? "Unknown";
};

const plannedCardioInput = ref("");
watch(
    plan,
    (p) => {
        plannedCardioInput.value = p?.planned_cardio_type?.trim() ?? "";
    },
    { immediate: true },
);

const savePlannedCardio = async () => {
    const p = plan.value;
    if (!p) return;
    try {
        await apiPUT(`workout/plans/${p.ID}/planned-cardio`, {
            type: plannedCardioInput.value.trim(),
        });
        toast.push("Planned cardio saved", "success");
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to save planned cardio: " + message, "error");
    }
};

const getAssignedDays = computed(() => {
    const assigned: number[] = [];
    plans.value.forEach((p) => {
        if (p.day_of_week !== null) {
            assigned.push(p.day_of_week);
        }
    });
    return assigned;
});

const isDayAssigned = (
    dayOfWeek: number,
    currentPlan: WorkoutPlan,
): boolean => {
    return (
        getAssignedDays.value.includes(dayOfWeek) &&
        plans.value.find((p) => p.day_of_week === dayOfWeek)?.ID !==
            currentPlan.ID
    );
};

const getAssignedPlanName = (
    dayOfWeek: number,
    currentPlan: WorkoutPlan,
): string => {
    const assignedPlan = plans.value.find(
        (p) => p.day_of_week === dayOfWeek && p.ID !== currentPlan.ID,
    );
    return assignedPlan?.name || "Assigned";
};

const removeExerciseFromPlan = async (p: WorkoutPlan, exercise: Exercise) => {
    try {
        await apiClient.delete(`workout/plans/${p.ID}/exercises/remove`, {
            data: { exercise_id: exercise.ID },
        });
        toast.push(`Removed ${exercise.name} from ${p.name}`, "success");
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to remove exercise: " + message, "error");
    }
};

const openAddExerciseToPlanDialog = () => {
    const p = plan.value;
    if (!p) return;
    dialogManager
        .custom<boolean>({
            title: `Add exercise to ${p.name}`,
            component: AddExerciseDialog,
            componentProps: {
                plan: p,
            },
        })
        .then((result) => {
            if (result !== null) {
                refresh();
            }
        });
};

const openCreateExerciseDialog = () => {
    dialogManager
        .custom<boolean>({
            title: "Create exercise",
            component: CreateExerciseForPlanDialog,
        })
        .then((result) => {
            if (result !== null) {
                refresh();
            }
        });
};

const openEditExerciseDialog = (exercise: Exercise) => {
    dialogManager
        .custom<boolean>({
            title: `Edit ${exercise.name}`,
            component: EditExerciseDialog,
            componentProps: { exercise },
        })
        .then((result) => {
            if (result !== null) {
                refresh();
            }
        });
};

const handleDayChange = async (p: WorkoutPlan, event: Event) => {
    const select = event.target as HTMLSelectElement;
    const day = parseInt(select.value);
    const previousValue =
        p.day_of_week !== null ? p.day_of_week.toString() : "";

    if (isNaN(day)) {
        select.value = previousValue;
        return;
    }

    const currentlyAssignedPlan = plans.value.find(
        (x) => x.day_of_week === day && x.ID !== p.ID,
    );

    if (currentlyAssignedPlan) {
        const confirmed = await dialogManager.confirm({
            title: "Day Already Assigned",
            message:
                dayNames[day] +
                ' is currently assigned to "' +
                currentlyAssignedPlan.name +
                '". Assigning "' +
                p.name +
                '" will unassign "' +
                currentlyAssignedPlan.name +
                '". Continue?',
            confirmText: "Yes, Reassign",
            cancelText: "Cancel",
        });

        if (!confirmed) {
            select.value = previousValue;
            return;
        }
    }

    await assignPlanToDay(p, day);
};

const assignPlanToDay = async (p: WorkoutPlan, dayOfWeek: number) => {
    try {
        await apiClient.post<{ plan: WorkoutPlan }>(
            `workout/plans/${p.ID}/assign-day`,
            {
                day_of_week: dayOfWeek,
            },
        );
        toast.push(`Assigned ${p.name} to ${dayNames[dayOfWeek]}`, "success");
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to assign day: " + message, "error");
    }
};

const unassignPlanFromDay = async (p: WorkoutPlan) => {
    try {
        await apiClient.delete<{ plan: WorkoutPlan }>(
            `workout/plans/${p.ID}/assign-day`,
        );
        toast.push(
            `Unassigned ${p.name} from ${getDayName(p.day_of_week)}`,
            "success",
        );
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to unassign day: " + message, "error");
    }
};

const moveExerciseInPlan = async (
    p: WorkoutPlan,
    index: number,
    delta: number,
) => {
    const list = p.exercises;
    const next = index + delta;
    if (next < 0 || next >= list.length) return;
    const reordered = [...list];
    const a = reordered[index];
    const b = reordered[next];
    if (a === undefined || b === undefined) return;
    reordered[index] = b;
    reordered[next] = a;
    try {
        await apiPUT(`workout/plans/${p.ID}/exercises/reorder`, {
            exercise_ids: reordered.map((e) => e.ID),
        });
        toast.push("Exercise order updated", "success");
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to reorder: " + message, "error");
    }
};
</script>

<template>
    <div class="flex w-full max-w-3xl flex-col gap-4 pb-8 pt-2">
        <div class="flex flex-wrap items-center gap-3">
            <router-link
                :to="{ name: 'gym-plans' }"
                class="text-sm text-textSecondary underline-offset-2 transition-colors hover:text-textPrimary hover:underline"
                >← Schedule</router-link
            >
            <span
                v-if="plan"
                class="text-xs font-medium uppercase tracking-wide text-textSecondary"
                >Configure plan</span
            >
        </div>
        <div v-if="isPending" class="text-center text-sm text-textSecondary">
            Loading…
        </div>
        <div
            v-else-if="planIdParam === null"
            class="rounded-lg border border-(--color-border) bg-firstBg p-6 text-center text-sm text-textSecondary"
        >
            Invalid plan link.
            <router-link
                :to="{ name: 'gym-plans' }"
                class="text-textPrimary underline underline-offset-2 hover:no-underline"
                >Back to schedule</router-link
            >
        </div>
        <div
            v-else-if="!plan"
            class="rounded-lg border border-(--color-border) bg-firstBg p-6 text-center text-sm text-textSecondary"
        >
            Plan not found.
            <router-link
                :to="{ name: 'gym-plans' }"
                class="text-textPrimary underline underline-offset-2 hover:no-underline"
                >Back to schedule</router-link
            >
        </div>
        <template v-else>
            <div
                class="flex flex-col gap-2 border-b border-(--color-border) pb-3"
            >
                <h1
                    class="m-0 text-xl font-semibold tracking-tight text-textPrimary"
                >
                    {{ plan.name }}
                </h1>
                <div class="flex flex-wrap items-center gap-2">
                    <span
                        v-if="plan.day_of_week !== null"
                        class="rounded-md bg-secondBg px-2 py-0.5 text-xs font-medium text-textPrimary"
                        >{{ getDayName(plan.day_of_week) }}</span
                    >
                    <span
                        v-else
                        class="rounded-md bg-thirdBg/60 px-2 py-0.5 text-xs font-medium text-textSecondary"
                        >Not Assigned to a Day</span
                    >
                </div>
            </div>
            <div class="flex flex-wrap gap-2">
                <button
                    type="button"
                    class="inline-flex items-center gap-1.5 rounded-md bg-secondBg px-3 py-2 text-sm font-medium text-textPrimary transition-colors hover:bg-thirdBg disabled:opacity-50"
                    @click="openAddExerciseToPlanDialog"
                >
                    <Plus class="size-4 shrink-0" />
                    Add Exercise to plan
                </button>
                <button
                    type="button"
                    class="inline-flex items-center gap-1.5 rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm font-medium text-textPrimary transition-colors hover:bg-secondBg disabled:opacity-50"
                    @click="openCreateExerciseDialog"
                >
                    <Plus class="size-4 shrink-0" />
                    Create exercise
                </button>
            </div>
            <section
                class="rounded-lg border border-(--color-border) bg-firstBg p-4"
            >
                <label
                    class="mb-2 block text-xs font-medium uppercase tracking-wide text-textSecondary"
                    >Assign to day</label
                >
                <div class="flex flex-wrap gap-2">
                    <select
                        :value="
                            plan.day_of_week !== null ? plan.day_of_week : ''
                        "
                        class="min-w-48 flex-1 rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm text-textPrimary transition-colors hover:bg-secondBg focus:outline-none focus:ring-2 focus:ring-(--color-cf-red)/40"
                        @change="handleDayChange(plan, $event)"
                    >
                        <option value="">— Select day —</option>
                        <option
                            v-for="(dayName, index) in dayNames"
                            :key="index"
                            :value="index"
                        >
                            {{ dayName
                            }}{{
                                isDayAssigned(index, plan)
                                    ? ` (${getAssignedPlanName(index, plan)})`
                                    : ""
                            }}
                        </option>
                    </select>
                    <button
                        v-if="plan.day_of_week !== null"
                        type="button"
                        class="rounded-md border border-(--color-border) bg-secondBg px-3 py-2 text-sm font-medium text-textPrimary transition-colors hover:bg-thirdBg"
                        @click="unassignPlanFromDay(plan)"
                    >
                        Clear
                    </button>
                </div>
            </section>
            <section
                class="rounded-lg border border-(--color-border) bg-firstBg p-4"
            >
                <label
                    class="mb-2 block text-xs font-medium uppercase tracking-wide text-textSecondary"
                    >Planned cardio (type)</label
                >
                <div class="flex flex-wrap gap-2">
                    <input
                        v-model="plannedCardioInput"
                        type="text"
                        class="min-w-0 flex-1 rounded-md border border-(--color-border) bg-firstBg px-3 py-2 text-sm text-textPrimary placeholder:text-textSecondary/60 hover:bg-secondBg focus:outline-none focus:ring-2 focus:ring-(--color-cf-red)/40"
                        placeholder="e.g. Bike, Run"
                    />
                    <button
                        type="button"
                        class="rounded-md bg-secondBg px-3 py-2 text-sm font-medium text-textPrimary transition-colors hover:bg-thirdBg"
                        @click="savePlannedCardio"
                    >
                        Save
                    </button>
                </div>
            </section>
            <section class="flex flex-col gap-2">
                <h2 class="m-0 text-sm font-semibold text-textPrimary">
                    Exercises
                </h2>
                <div class="flex flex-col gap-2">
                    <div
                        v-for="(exercise, exerciseIndex) in plan.exercises"
                        :key="exercise.ID"
                        class="flex items-center justify-between gap-2 rounded-md border border-(--color-border) bg-secondBg px-3 py-2"
                    >
                        <div class="flex min-w-0 flex-1 items-center gap-2">
                            <button
                                type="button"
                                class="shrink-0 rounded p-1 text-textSecondary transition-colors hover:bg-firstBg hover:text-textPrimary"
                                title="Edit exercise"
                                @click="openEditExerciseDialog(exercise)"
                            >
                                <Pencil class="size-3.5" />
                            </button>
                            <span
                                class="min-w-0 truncate text-sm text-textPrimary"
                                >{{ exercise.name }}</span
                            >
                        </div>
                        <div class="flex shrink-0 items-center gap-0.5">
                            <button
                                type="button"
                                class="rounded p-1 text-textSecondary transition-colors hover:bg-firstBg hover:text-textPrimary disabled:opacity-25"
                                :disabled="exerciseIndex === 0"
                                title="Move up"
                                @click="
                                    moveExerciseInPlan(plan, exerciseIndex, -1)
                                "
                            >
                                <ChevronUp class="size-4" />
                            </button>
                            <button
                                type="button"
                                class="rounded p-1 text-textSecondary transition-colors hover:bg-firstBg hover:text-textPrimary disabled:opacity-25"
                                :disabled="
                                    exerciseIndex === plan.exercises.length - 1
                                "
                                title="Move down"
                                @click="
                                    moveExerciseInPlan(plan, exerciseIndex, 1)
                                "
                            >
                                <ChevronDown class="size-4" />
                            </button>
                            <button
                                type="button"
                                class="rounded p-1 text-(--color-cf-red) transition-colors hover:bg-(--color-cf-red)/10"
                                title="Remove"
                                @click="removeExerciseFromPlan(plan, exercise)"
                            >
                                <X class="size-4" />
                            </button>
                        </div>
                    </div>
                    <div
                        v-if="plan.exercises.length === 0"
                        class="rounded-md border border-dashed border-(--color-border) py-6 text-center text-sm italic text-textSecondary"
                    >
                        No exercises in this plan
                    </div>
                </div>
            </section>
        </template>
    </div>
</template>
