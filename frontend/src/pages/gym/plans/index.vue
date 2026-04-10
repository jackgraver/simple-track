<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import AddExerciseDialog from "~/shared/AddExerciseDialog.vue";
import CreateExerciseForPlanDialog from "~/shared/CreateExerciseForPlanDialog.vue";
import { X, Plus, ChevronUp, ChevronDown } from "lucide-vue-next";
import { computed, ref, watch } from "vue";
import { apiPUT } from "~/api/client";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "~/api/client";

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

const getDayName = (dayOfWeek: number | null): string | undefined => {
    if (dayOfWeek === null) return "Unassigned";
    return dayNames[dayOfWeek];
};

const selectedPlanId = ref<number | null>(null);

const defaultPlanIdForList = (list: WorkoutPlan[]): number | null => {
    if (list.length === 0) return null;
    const dow = new Date().getDay();
    const forToday = list.find((p) => p.day_of_week === dow);
    return forToday?.ID ?? list[0]!.ID;
};

watch(
    plans,
    (list) => {
        if (list.length === 0) {
            selectedPlanId.value = null;
            return;
        }
        if (
            selectedPlanId.value === null ||
            !list.some((p) => p.ID === selectedPlanId.value)
        ) {
            selectedPlanId.value = defaultPlanIdForList(list);
        }
    },
    { immediate: true },
);

const selectedPlan = computed(() => {
    const id = selectedPlanId.value;
    if (id === null) return null;
    return plans.value.find((p) => p.ID === id) ?? null;
});

const plannedCardioInput = ref("");
watch(
    selectedPlan,
    (p) => {
        plannedCardioInput.value = p?.planned_cardio_type?.trim() ?? "";
    },
    { immediate: true },
);

const savePlannedCardio = async () => {
    const plan = selectedPlan.value;
    if (!plan) return;
    try {
        await apiPUT(`workout/plans/${plan.ID}/planned-cardio`, {
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
    plans.value.forEach((plan) => {
        if (plan.day_of_week !== null) {
            assigned.push(plan.day_of_week);
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

const planOptionLabel = (plan: WorkoutPlan) => {
    if (plan.day_of_week === null) return plan.name;
    return `${plan.name} (${dayNames[plan.day_of_week]})`;
};

const removeExerciseFromPlan = async (
    plan: WorkoutPlan,
    exercise: Exercise,
) => {
    try {
        await apiClient.delete(`workout/plans/${plan.ID}/exercises/remove`, {
            data: { exercise_id: exercise.ID },
        });
        toast.push(`Removed ${exercise.name} from ${plan.name}`, "success");
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to remove exercise: " + message, "error");
    }
};

const openAddExerciseToPlanDialog = () => {
    const plan = selectedPlan.value;
    if (!plan) return;
    dialogManager
        .custom<boolean>({
            title: `Add exercise to ${plan.name}`,
            component: AddExerciseDialog,
            props: {
                plan,
            },
        })
        .then((result) => {
            if (result !== null) {
                refresh();
            }
        });
};

const openCreateExerciseDialog = () => {
    const plan = selectedPlan.value;
    if (!plan) return;
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

const handleDayChange = async (plan: WorkoutPlan, event: Event) => {
    const select = event.target as HTMLSelectElement;
    const day = parseInt(select.value);
    const previousValue =
        plan.day_of_week !== null ? plan.day_of_week.toString() : "";

    if (isNaN(day)) {
        select.value = previousValue;
        return;
    }

    const currentlyAssignedPlan = plans.value.find(
        (p) => p.day_of_week === day && p.ID !== plan.ID,
    );

    if (currentlyAssignedPlan) {
        const confirmed = await dialogManager.confirm({
            title: "Day Already Assigned",
            message:
                dayNames[day] +
                ' is currently assigned to "' +
                currentlyAssignedPlan.name +
                '". Assigning "' +
                plan.name +
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

    await assignPlanToDay(plan, day);
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
            `Unassigned ${plan.name} from ${getDayName(plan.day_of_week)}`,
            "success",
        );
        await refresh();
    } catch (err: unknown) {
        const message = err instanceof Error ? err.message : String(err);
        toast.push("Failed to unassign day: " + message, "error");
    }
};

const moveExerciseInPlan = async (
    plan: WorkoutPlan,
    index: number,
    delta: number,
) => {
    const list = plan.exercises;
    const next = index + delta;
    if (next < 0 || next >= list.length) return;
    const reordered = [...list];
    const a = reordered[index];
    const b = reordered[next];
    if (a === undefined || b === undefined) return;
    reordered[index] = b;
    reordered[next] = a;
    try {
        await apiPUT(`workout/plans/${plan.ID}/exercises/reorder`, {
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
    <div class="plans-page">
        <div class="top">
            <h1 class="title">Workout plans</h1>
            <div class="toolbar">
                <label class="plan-pick">
                    <span class="sr-only">Plan</span>
                    <select
                        v-model.number="selectedPlanId"
                        class="plan-select"
                        :disabled="!plans.length"
                    >
                        <option v-for="p in plans" :key="p.ID" :value="p.ID">
                            {{ planOptionLabel(p) }}
                        </option>
                    </select>
                </label>
                <button
                    type="button"
                    class="add-button"
                    :disabled="!selectedPlan"
                    @click="openAddExerciseToPlanDialog"
                >
                    <Plus :size="18" />
                    Add To Plan
                </button>
                <button
                    type="button"
                    class="add-button secondary"
                    :disabled="!selectedPlan"
                    @click="openCreateExerciseDialog"
                >
                    <Plus :size="18" />
                    Create Exercise
                </button>
            </div>
        </div>
        <div v-if="isPending" class="loading">Loading...</div>
        <div v-else-if="!plans.length" class="empty-global">
            No workout plans yet.
        </div>
        <template v-else-if="selectedPlan">
            <div class="plan-card">
                <div class="plan-header">
                    <div class="plan-title-section">
                        <h2 class="plan-name">{{ selectedPlan.name }}</h2>
                        <div class="day-assignment">
                            <span
                                v-if="selectedPlan.day_of_week !== null"
                                class="day-badge"
                            >
                                {{ getDayName(selectedPlan.day_of_week) }}
                            </span>
                            <span v-else class="day-badge unassigned"
                                >Unassigned</span
                            >
                        </div>
                    </div>
                </div>
                <div class="day-selector-section">
                    <label>Assign to day</label>
                    <div class="day-selector">
                        <select
                            :value="
                                selectedPlan.day_of_week !== null
                                    ? selectedPlan.day_of_week
                                    : ''
                            "
                            class="day-select"
                            @change="handleDayChange(selectedPlan, $event)"
                        >
                            <option value="">— Select day —</option>
                            <option
                                v-for="(dayName, index) in dayNames"
                                :key="index"
                                :value="index"
                            >
                                {{ dayName
                                }}{{
                                    isDayAssigned(index, selectedPlan)
                                        ? ` (${getAssignedPlanName(index, selectedPlan)})`
                                        : ""
                                }}
                            </option>
                        </select>
                        <button
                            v-if="selectedPlan.day_of_week !== null"
                            type="button"
                            class="unassign-button"
                            @click="unassignPlanFromDay(selectedPlan)"
                        >
                            Clear
                        </button>
                    </div>
                </div>
                <div class="planned-cardio-section">
                    <label class="planned-cardio-label">Planned cardio (type)</label>
                    <div class="planned-cardio-row">
                        <input
                            v-model="plannedCardioInput"
                            type="text"
                            class="planned-cardio-input"
                            placeholder="e.g. Bike, Run"
                        />
                        <button
                            type="button"
                            class="planned-cardio-save"
                            @click="savePlannedCardio"
                        >
                            Save
                        </button>
                    </div>
                </div>
                <div class="exercises-list">
                    <div
                        v-for="(exercise, exerciseIndex) in selectedPlan.exercises"
                        :key="exercise.ID"
                        class="exercise-item"
                    >
                        <span>{{ exercise.name }}</span>
                        <div class="exercise-item-actions">
                            <button
                                type="button"
                                class="reorder-button"
                                :disabled="exerciseIndex === 0"
                                title="Move up"
                                @click="
                                    moveExerciseInPlan(
                                        selectedPlan,
                                        exerciseIndex,
                                        -1,
                                    )
                                "
                            >
                                <ChevronUp :size="16" />
                            </button>
                            <button
                                type="button"
                                class="reorder-button"
                                :disabled="
                                    exerciseIndex ===
                                    selectedPlan.exercises.length - 1
                                "
                                title="Move down"
                                @click="
                                    moveExerciseInPlan(
                                        selectedPlan,
                                        exerciseIndex,
                                        1,
                                    )
                                "
                            >
                                <ChevronDown :size="16" />
                            </button>
                            <button
                                type="button"
                                class="remove-button"
                                @click="
                                    removeExerciseFromPlan(selectedPlan, exercise)
                                "
                            >
                                <X :size="16" />
                            </button>
                        </div>
                    </div>
                    <div
                        v-if="selectedPlan.exercises.length === 0"
                        class="empty-state"
                    >
                        No exercises in this plan
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<style scoped>
.plans-page {
    max-width: 28rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}
.top {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}
.title {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
}
.toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
}
.plan-pick {
    flex: 1;
    min-width: 10rem;
}
.plan-select {
    width: 100%;
    padding: 0.5rem 0.65rem;
    background: #2a2a2a;
    color: #fff;
    border: 1px solid #444;
    border-radius: 0.25rem;
    font-size: 0.9rem;
    cursor: pointer;
}
.plan-select:focus {
    outline: none;
    border-color: #4a9eff;
}
.sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    border: 0;
}
.loading,
.empty-global {
    text-align: center;
    padding: 1.5rem;
    color: #aaa;
}
.plan-card {
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 0.5rem;
    padding: 1rem;
}
.plan-header {
    margin-bottom: 0.75rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid #333;
}
.plan-title-section {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
}
.plan-name {
    margin: 0;
    font-size: 1.25rem;
}
.day-assignment {
    display: flex;
    align-items: center;
}
.day-badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    background: #4a9eff;
    color: #fff;
    border-radius: 0.25rem;
    font-size: 0.8rem;
    font-weight: 500;
}
.day-badge.unassigned {
    background: #666;
}
.day-selector-section {
    margin-bottom: 0.75rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid #333;
}
.day-selector-section label {
    display: block;
    margin-bottom: 0.35rem;
    font-size: 0.85rem;
    color: #aaa;
}
.day-selector {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}
.day-select {
    flex: 1;
    padding: 0.5rem;
    background: #2a2a2a;
    color: #fff;
    border: 1px solid #444;
    border-radius: 0.25rem;
    font-size: 0.9rem;
    cursor: pointer;
}
.day-select:focus {
    outline: none;
    border-color: #4a9eff;
}
.unassign-button {
    padding: 0.5rem 0.75rem;
    background: #666;
    color: #fff;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.85rem;
}
.unassign-button:hover {
    background: #777;
}
.add-button {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.5rem 0.85rem;
    background: #4a9eff;
    color: #fff;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.9rem;
}
.add-button:hover:not(:disabled) {
    background: #3a8eef;
}
.add-button.secondary {
    background: #2f2f2f;
    border: 1px solid #444;
}
.add-button.secondary:hover:not(:disabled) {
    background: #3a3a3a;
}
.add-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}
.planned-cardio-section {
    margin-bottom: 0.75rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid #333;
}
.planned-cardio-label {
    display: block;
    margin-bottom: 0.35rem;
    font-size: 0.85rem;
    color: #aaa;
}
.planned-cardio-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}
.planned-cardio-input {
    flex: 1;
    padding: 0.5rem 0.65rem;
    background: #2a2a2a;
    color: #fff;
    border: 1px solid #444;
    border-radius: 0.25rem;
    font-size: 0.9rem;
}
.planned-cardio-save {
    padding: 0.5rem 0.75rem;
    background: #4a9eff;
    color: #fff;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.85rem;
}
.planned-cardio-save:hover {
    background: #3a8eef;
}
.exercises-list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
}
.exercise-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.65rem 0.75rem;
    background: #2a2a2a;
    border-radius: 0.25rem;
    border: 1px solid #444;
}
.exercise-item-actions {
    display: flex;
    align-items: center;
    gap: 0.15rem;
}
.reorder-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    background: transparent;
    border: none;
    color: #aaa;
    cursor: pointer;
    border-radius: 0.25rem;
}
.reorder-button:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    color: #fff;
}
.reorder-button:disabled {
    opacity: 0.25;
    cursor: not-allowed;
}
.remove-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    background: transparent;
    border: none;
    color: #ff6b6b;
    cursor: pointer;
    border-radius: 0.25rem;
}
.remove-button:hover {
    background: rgba(255, 107, 107, 0.1);
}
.empty-state {
    padding: 1rem;
    text-align: center;
    color: #888;
    font-style: italic;
}
</style>
