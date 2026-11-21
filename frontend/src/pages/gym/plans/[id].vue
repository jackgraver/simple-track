<script setup lang="ts">
import { getApiErrorMessage } from "~/api/httpError";
import type { WorkoutPlan, Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import AddExerciseDialog from "~/pages/gym/plans/components/AddExerciseDialog.vue";
import CreateExerciseForPlanDialog from "~/pages/gym/plans/components/CreateExerciseForPlanDialog.vue";
import EditExerciseDialog from "~/pages/gym/plans/components/EditExerciseDialog.vue";
import MobilityItemEditor from "~/pages/gym/plans/components/MobilityItemEditor.vue";
import { X, Plus, ChevronUp, ChevronDown, Pencil, Dumbbell, SearchX } from "lucide-vue-next";
import { Badge, Button, Card, EmptyState, ErrorState, IconButton, Input, List, ListRow, NumberField, PageHeader, Section, Select, SkeletonList, useDelayedFlag } from "~/shared/ui";
import { computed, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { apiClient, apiPUT } from "~/api/client";
import { useWorkoutPlansAll } from "~/api/workout/queries";

const route = useRoute();

const { data, isPending, isError, error, refetch } = useWorkoutPlansAll();
const showSkeleton = useDelayedFlag(() => isPending.value);
const cardioSaving = ref(false);
const busyExerciseId = ref<number | null>(null);
const daySelectKey = ref(0);

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

const assignedDays = (p: WorkoutPlan): number[] =>
    p.assigned_days ?? (p.day_of_week === null ? [] : [p.day_of_week]);

const plannedCardioInput = ref("");
const plannedCardioMinutesInput = ref<number | null>(null);
const dynamicWarmupItems = ref<string[]>([]);
const staticStretchingItems = ref<string[]>([]);
const mobilitySaving = ref(false);
watch(
    plan,
    (p) => {
        plannedCardioInput.value = p?.planned_cardio_type?.trim() ?? "";
        plannedCardioMinutesInput.value = p?.planned_cardio_minutes ?? null;
        dynamicWarmupItems.value = [...(p?.pre_mobility_items ?? [])];
        staticStretchingItems.value = [...(p?.post_mobility_items ?? [])];
    },
    { immediate: true },
);

const savePlannedCardio = async () => {
    const p = plan.value;
    if (!p || cardioSaving.value) return;
    cardioSaving.value = true;
    try {
        await apiPUT(`workout/plans/${p.ID}/planned-cardio`, {
            type: plannedCardioInput.value.trim(),
            minutes: Math.max(
                0,
                Math.round(Number(plannedCardioMinutesInput.value) || 0),
            ),
        });
        toast.push("Planned cardio saved", "success");
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to save planned cardio: " + message, "error");
    } finally {
        cardioSaving.value = false;
    }
};

const savePlannedMobility = async () => {
    const p = plan.value;
    if (!p || mobilitySaving.value) return;
    mobilitySaving.value = true;
    try {
        await apiPUT(`workout/plans/${p.ID}/planned-mobility`, {
            pre_mobility_items: dynamicWarmupItems.value,
            post_mobility_items: staticStretchingItems.value,
        });
        toast.push("Mobility plan saved", "success");
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to save mobility plan: " + message, "error");
    } finally {
        mobilitySaving.value = false;
    }
};

const isDayAssigned = (
    dayOfWeek: number,
    currentPlan: WorkoutPlan,
): boolean => {
    return plans.value.some(
        (p) => p.ID !== currentPlan.ID && assignedDays(p).includes(dayOfWeek),
    );
};

const getAssignedPlanName = (
    dayOfWeek: number,
    currentPlan: WorkoutPlan,
): string => {
    const assignedPlan = plans.value.find(
        (p) => p.ID !== currentPlan.ID && assignedDays(p).includes(dayOfWeek),
    );
    return assignedPlan?.name || "Assigned";
};

const removeExerciseFromPlan = async (p: WorkoutPlan, exercise: Exercise) => {
    busyExerciseId.value = exercise.ID;
    try {
        await apiClient.delete(`workout/plans/${p.ID}/exercises/remove`, {
            data: { exercise_id: exercise.ID },
        });
        toast.push(`Removed ${exercise.name} from ${p.name}`, "success");
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to remove exercise: " + message, "error");
    } finally {
        busyExerciseId.value = null;
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

const handleDayChange = async (p: WorkoutPlan, value: number | "") => {
    // Re-keying the Select puts it back on "— Select day —" after every pick.
    const reset = () => {
        daySelectKey.value++;
    };
    const day = typeof value === "number" ? value : NaN;
    const currentDays = assignedDays(p);

    if (isNaN(day) || currentDays.includes(day)) {
        reset();
        return;
    }

    const currentlyAssignedPlan = plans.value.find(
        (x) => x.ID !== p.ID && assignedDays(x).includes(day),
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
            reset();
            return;
        }
    }

    await assignPlanToDay(p, day);
    reset();
};
const dayOptions = computed(() =>
    dayNames.map((dayName, index) => ({
        label: plan.value && isDayAssigned(index, plan.value) ? `${dayName} (${getAssignedPlanName(index, plan.value)})` : dayName,
        value: index,
    })),
);

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
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to assign day: " + message, "error");
    }
};

const unassignPlanFromDay = async (p: WorkoutPlan, dayOfWeek: number) => {
    try {
        await apiClient.delete<{ plan: WorkoutPlan }>(
            `workout/plans/${p.ID}/assign-day`,
            { params: { day_of_week: dayOfWeek } },
        );
        toast.push(
            `Unassigned ${p.name} from ${getDayName(dayOfWeek)}`,
            "success",
        );
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
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
    busyExerciseId.value = a.ID;
    try {
        await apiPUT(`workout/plans/${p.ID}/exercises/reorder`, {
            exercise_ids: reordered.map((e) => e.ID),
        });
        toast.push("Exercise order updated", "success");
        await refresh();
    } catch (err: unknown) {
        const message = getApiErrorMessage(err, "Something went wrong");
        toast.push("Failed to reorder: " + message, "error");
    } finally {
        busyExerciseId.value = null;
    }
};
</script>

<template>
    <div class="flex w-full max-w-3xl flex-col gap-4 pb-8 pt-2">
        <ErrorState v-if="isError" :error="error" title="Couldn't load this plan" @retry="refetch()" />
        <SkeletonList v-else-if="isPending && showSkeleton" :rows="4" :leading="false" />
        <span v-else-if="isPending" />
        <EmptyState
            v-else-if="planIdParam === null || !plan"
            :icon="SearchX"
            :title="planIdParam === null ? 'Invalid plan link' : 'Plan not found'"
        >
            <template #action>
                <Button :to="{ name: 'gym-plans' }">Back to schedule</Button>
            </template>
        </EmptyState>
        <template v-else>
            <PageHeader :title="plan.name">
                <template #meta>
                    <Badge v-if="assignedDays(plan).length">{{ assignedDays(plan).map(getDayName).join(", ") }}</Badge>
                    <Badge v-else>Not assigned to a day</Badge>
                </template>
            </PageHeader>
            <div class="flex flex-col gap-2 lg:flex-row">
                <Button @click="openAddExerciseToPlanDialog">
                    <template #icon-left><Plus class="size-4" /></template>
                    Add exercise to plan
                </Button>
                <Button variant="ghost" @click="openCreateExerciseDialog">
                    <template #icon-left><Plus class="size-4" /></template>
                    Create exercise
                </Button>
            </div>
            <Card>
                <Section title="Assign to day">
                    <div class="flex flex-wrap gap-2">
                        <div class="min-w-48 flex-1">
                            <Select
                                :key="daySelectKey"
                                aria-label="Assign to day"
                                model-value=""
                                :options="dayOptions"
                                placeholder="— Select day —"
                                @update:model-value="handleDayChange(plan, $event)"
                            />
                        </div>
                        <Button v-for="day in assignedDays(plan)" :key="day" @click="unassignPlanFromDay(plan, day)">
                            {{ getDayName(day) }}
                            <template #icon-right><X class="size-3.5" /></template>
                        </Button>
                    </div>
                </Section>
            </Card>
            <Card>
                <Section title="Planned cardio">
                    <div class="flex flex-col gap-2 lg:flex-row">
                        <div class="min-w-0 flex-1">
                            <Input v-model="plannedCardioInput" aria-label="Planned cardio type" placeholder="e.g. Bike, Run" />
                        </div>
                        <div class="lg:w-44">
                            <NumberField
                                v-model="plannedCardioMinutesInput"
                                label="planned cardio minutes"
                                unit="min"
                                :min="0"
                                :step="1"
                                :decimals="0"
                            />
                        </div>
                        <Button :loading="cardioSaving" @click="savePlannedCardio">Save</Button>
                    </div>
                </Section>
            </Card>
            <div class="flex flex-col gap-3">
                <MobilityItemEditor
                    v-model="dynamicWarmupItems"
                    title="Dynamic warmup"
                    description="Movements to complete before the workout."
                    placeholder="Add a warmup movement…"
                />
                <MobilityItemEditor
                    v-model="staticStretchingItems"
                    title="Static stretching"
                    description="Stretches to complete after the workout."
                    placeholder="Add a stretch…"
                />
                <Button class="lg:self-end" :loading="mobilitySaving" @click="savePlannedMobility">
                    {{ mobilitySaving ? "Saving…" : "Save mobility" }}
                </Button>
            </div>
            <Section title="Exercises">
                <Card padding="none">
                    <List v-if="plan.exercises.length">
                        <ListRow
                            v-for="(exercise, exerciseIndex) in plan.exercises"
                            :key="exercise.ID"
                            :pending="busyExerciseId === exercise.ID"
                        >
                            {{ exercise.name }}
                            <template #trailing>
                                <IconButton
                                    label="Edit exercise"
                                    size="sm"
                                    :disabled="busyExerciseId !== null"
                                    @click="openEditExerciseDialog(exercise)"
                                >
                                    <Pencil />
                                </IconButton>
                                <IconButton
                                    label="Move up"
                                    size="sm"
                                    :disabled="exerciseIndex === 0 || busyExerciseId !== null"
                                    @click="moveExerciseInPlan(plan, exerciseIndex, -1)"
                                >
                                    <ChevronUp />
                                </IconButton>
                                <IconButton
                                    label="Move down"
                                    size="sm"
                                    :disabled="exerciseIndex === plan.exercises.length - 1 || busyExerciseId !== null"
                                    @click="moveExerciseInPlan(plan, exerciseIndex, 1)"
                                >
                                    <ChevronDown />
                                </IconButton>
                                <IconButton
                                    label="Remove from plan"
                                    size="sm"
                                    variant="danger"
                                    :disabled="busyExerciseId !== null"
                                    @click="removeExerciseFromPlan(plan, exercise)"
                                >
                                    <X />
                                </IconButton>
                            </template>
                        </ListRow>
                    </List>
                    <EmptyState v-else :icon="Dumbbell" title="No exercises in this plan" description="Add one above." />
                </Card>
            </Section>
        </template>
    </div>
</template>
