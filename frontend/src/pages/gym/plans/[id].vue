<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import { ArrowDown, ArrowUp, Dumbbell, Ellipsis, GripVertical, Pencil, Plus, SearchX, Trash2, X } from "lucide-vue-next";
import { getApiErrorMessage } from "~/api/httpError";
import { workoutKeys } from "~/api/workout/keys";
import { useAssignPlanDay, useUnassignPlanDay, useWorkoutPrograms } from "~/api/workout/queries";
import {
    useAddRoutineExercise,
    useDeleteRoutine,
    useRemoveRoutineExercise,
    useReorderRoutineExercises,
    useRenameRoutine,
    useRestoreRoutine,
    useSetRoutineCardio,
    useSetRoutineMobility,
} from "~/api/workout/routines";
import { dialogManager } from "~/composables/dialog/useDialog";
import { toast } from "~/composables/toast/useToast";
import type { Exercise } from "~/types/workout";
import {
    Button,
    Card,
    EmptyState,
    ErrorState,
    IconButton,
    Input,
    List,
    ListRow,
    Menu,
    NumberField,
    PageHeader,
    Skeleton,
    SkeletonList,
    useDelayedFlag,
    useIsDesktop,
} from "~/shared/ui";
import type { MenuItem } from "~/shared/ui";
import AddExerciseDialog from "./components/AddExerciseDialog.vue";
import EditExerciseDialog from "./components/EditExerciseDialog.vue";
import NameDialog from "./components/NameDialog.vue";
import RoutineItemList from "./components/RoutineItemList.vue";
import RoutineStep from "./components/RoutineStep.vue";
import {
    type DayToggle,
    cardioChanged,
    dayToggles,
    findRoutine,
    formatDays,
    moveItem,
    normaliseCardio,
    otherDaysLine,
    parseId,
    routineMeta,
} from "./domain/routine";
import { plural } from "./domain/week";

// One routine (a day of a program), laid out in session order: warmup, lifts, cardio, stretch.
// Everything saves as you go; deleting the routine can be undone from the toast.

const route = useRoute();
const router = useRouter();
const queryClient = useQueryClient();
const isDesktop = useIsDesktop();

// --- data -------------------------------------------------------------------------------------
const { data, isPending, isError, error, refetch } = useWorkoutPrograms();
const showSkeleton = useDelayedFlag(() => isPending.value);

const planId = computed(() => parseId(route.params.id));
const programId = computed(() => parseId(route.params.programId));
const found = computed(() => findRoutine(data.value?.programs ?? [], programId.value, planId.value));
const plan = computed(() => found.value?.plan ?? null);
const program = computed(() => found.value?.program ?? null);
const toggles = computed(() => (plan.value && program.value ? dayToggles(plan.value, program.value.plans) : []));
const otherDays = computed(() => otherDaysLine(toggles.value));

const failToast = (what: string, err: unknown) =>
    toast.push(`${what}: ${getApiErrorMessage(err, "Something went wrong")}`, "error");
const refreshPrograms = () => queryClient.invalidateQueries({ queryKey: workoutKeys.programs });

// Quiet per-step save state ("Saving…" then "Saved" for a moment), shown in the step heading.
type StepKey = "warmup" | "cardio" | "stretch";
const saveState = ref<Partial<Record<StepKey, "saving" | "saved">>>({});
const savedTimers: Partial<Record<StepKey, ReturnType<typeof setTimeout>>> = {};
async function saving(step: StepKey, what: string, run: () => Promise<unknown>) {
    clearTimeout(savedTimers[step]);
    saveState.value = { ...saveState.value, [step]: "saving" };
    try {
        await run();
        saveState.value = { ...saveState.value, [step]: "saved" };
        savedTimers[step] = setTimeout(() => {
            saveState.value = { ...saveState.value, [step]: undefined };
        }, 2000);
    } catch (err) {
        saveState.value = { ...saveState.value, [step]: undefined };
        failToast(what, err);
    }
}
const stepMeta = (step: StepKey, fallback: string) => {
    const s = saveState.value[step];
    return s === "saving" ? "Saving…" : s === "saved" ? "Saved" : fallback;
};

// --- header: rename / delete --------------------------------------------------------------------
const renameRoutine = useRenameRoutine();
const deleteRoutine = useDeleteRoutine();
const restoreRoutine = useRestoreRoutine();
const renameOpen = ref(false);

const routineMenu: MenuItem[] = [
    { key: "rename", label: "Rename", icon: Pencil },
    { key: "delete", label: "Delete routine", icon: Trash2, tone: "danger" },
];
const onRoutineMenu = (key: string) => {
    if (key === "rename") renameOpen.value = true;
    if (key === "delete") void removeRoutine();
};

const onRename = async (name: string) => {
    const p = plan.value;
    if (!p) return;
    try {
        await renameRoutine.mutateAsync({ planId: p.ID, name });
        renameOpen.value = false;
    } catch (err) {
        failToast("Couldn't rename the routine", err);
    }
};

/** Soft delete, leave for the schedule, and offer Undo (restore, then come back here). */
const removeRoutine = async () => {
    const p = plan.value;
    const prog = program.value;
    if (!p || !prog) return;
    const back = { name: "gym-program-day-detail", params: { programId: String(prog.ID), id: String(p.ID) } };
    try {
        await deleteRoutine.mutateAsync(p.ID);
    } catch (err) {
        failToast("Couldn't delete the routine", err);
        return;
    }
    await router.push({ name: "gym-plans", query: { program: String(prog.ID) } });
    toast.push({
        message: `Deleted ${p.name}`,
        action: {
            label: "Undo",
            onClick: () => {
                restoreRoutine
                    .mutateAsync(p.ID)
                    .then(() => router.push(back))
                    .catch((err: unknown) => failToast(`Couldn't restore ${p.name}`, err));
            },
        },
    });
};

// --- days ---------------------------------------------------------------------------------------
const assign = useAssignPlanDay();
const unassign = useUnassignPlanDay();
const pendingDays = ref<Set<number>>(new Set());

const toggleDay = async (t: DayToggle) => {
    const p = plan.value;
    if (!p || pendingDays.value.has(t.dow)) return;
    if (!t.on && t.takenBy) {
        const ok = await dialogManager.confirm({
            title: `Move ${t.name} to ${p.name}?`,
            message: `${t.name} is ${t.takenBy.name} now. ${t.takenBy.name} keeps its other days.`,
            confirmText: `Move to ${p.name}`,
            cancelText: "Cancel",
        });
        if (!ok) return;
    }
    pendingDays.value = new Set(pendingDays.value).add(t.dow);
    try {
        if (t.on) await unassign.mutateAsync({ planId: p.ID, dayOfWeek: t.dow });
        else await assign.mutateAsync({ planId: p.ID, dayOfWeek: t.dow });
    } catch (err) {
        failToast(t.on ? `Couldn't take ${t.name} off` : `Couldn't add ${t.name}`, err);
    } finally {
        const next = new Set(pendingDays.value);
        next.delete(t.dow);
        pendingDays.value = next;
    }
};

// --- warmup / stretch ---------------------------------------------------------------------------
const setMobility = useSetRoutineMobility();
const saveMobility = (step: "warmup" | "stretch", items: string[]) => {
    const p = plan.value;
    if (!p) return;
    const pre = step === "warmup" ? items : (p.pre_mobility_items ?? []);
    const post = step === "stretch" ? items : (p.post_mobility_items ?? []);
    void saving(step, `Couldn't save the ${step}`, () => setMobility.mutateAsync({ planId: p.ID, pre, post }));
};

const stretchCount = (n: number) => `${n} ${n === 1 ? "stretch" : "stretches"}`;

// --- cardio (saved when focus leaves the fields, or on Enter) ----------------------------------
const setCardio = useSetRoutineCardio();
const cardioType = ref("");
const cardioMinutes = ref<number | null>(null);
watch(
    () => [plan.value?.ID, plan.value?.planned_cardio_type, plan.value?.planned_cardio_minutes] as const,
    ([, type, minutes]) => {
        cardioType.value = type?.trim() ?? "";
        cardioMinutes.value = minutes || null;
    },
    { immediate: true },
);
const cardioBox = ref<HTMLElement | null>(null);
const saveCardio = () => {
    const p = plan.value;
    if (!p) return;
    const next = normaliseCardio(cardioType.value, cardioMinutes.value);
    if (!cardioChanged(p, next)) return;
    void saving("cardio", "Couldn't save cardio", () => setCardio.mutateAsync({ planId: p.ID, ...next }));
};
const onCardioFocusOut = (e: FocusEvent) => {
    if (cardioBox.value?.contains(e.relatedTarget as Node | null)) return;
    saveCardio();
};

// --- exercises ------------------------------------------------------------------------------------
const reorder = useReorderRoutineExercises();
const removeExercise = useRemoveRoutineExercise();
const addExercise = useAddRoutineExercise();
const exercises = computed(() => plan.value?.exercises ?? []);
const reordering = computed(() => reorder.isPending.value);

const move = async (from: number, to: number) => {
    const p = plan.value;
    if (!p) return;
    const next = moveItem(p.exercises, from, to);
    if (next === p.exercises) return;
    try {
        await reorder.mutateAsync({ planId: p.ID, exerciseIds: next.map((e) => e.ID) });
    } catch (err) {
        failToast("Couldn't reorder", err);
    }
};

/** Remove now; Undo adds it back and restores its place. */
const remove = async (exercise: Exercise) => {
    const p = plan.value;
    if (!p) return;
    const order = p.exercises.map((e) => e.ID);
    try {
        await removeExercise.mutateAsync({ planId: p.ID, exerciseId: exercise.ID });
    } catch (err) {
        failToast(`Couldn't remove ${exercise.name}`, err);
        return;
    }
    toast.push({
        message: `Removed ${exercise.name}`,
        action: {
            label: "Undo",
            onClick: () => {
                addExercise
                    .mutateAsync({ planId: p.ID, exerciseId: exercise.ID })
                    .then(() => reorder.mutateAsync({ planId: p.ID, exerciseIds: order }))
                    .catch((err: unknown) => failToast(`Couldn't put ${exercise.name} back`, err));
            },
        },
    });
};

const openAdd = () => {
    const p = plan.value;
    if (!p) return;
    void dialogManager
        .custom<boolean>({ title: `Add to ${p.name}`, component: AddExerciseDialog, componentProps: { plan: p } })
        .then((ok) => {
            if (ok) void refreshPrograms();
        });
};
const openEdit = (exercise: Exercise) => {
    void dialogManager
        .custom<boolean>({ title: `Edit ${exercise.name}`, component: EditExerciseDialog, componentProps: { exercise } })
        .then((ok) => {
            if (ok) void refreshPrograms();
        });
};

const rowMenu = (index: number): MenuItem[] => [
    { key: "edit", label: "Edit exercise", icon: Pencil },
    { key: "up", label: "Move up", icon: ArrowUp, disabled: index === 0 || reordering.value },
    { key: "down", label: "Move down", icon: ArrowDown, disabled: index === exercises.value.length - 1 || reordering.value },
    { key: "remove", label: "Remove from routine", icon: X, tone: "danger" },
];
const onRowMenu = (key: string, exercise: Exercise, index: number) => {
    if (key === "edit") openEdit(exercise);
    if (key === "up") void move(index, index - 1);
    if (key === "down") void move(index, index + 1);
    if (key === "remove") void remove(exercise);
};

// Drag to reorder (desktop; the row menu covers keyboard and touch).
const dragFrom = ref<number | null>(null);
const dragOver = ref<number | null>(null);
const onDragStart = (e: DragEvent, index: number) => {
    dragFrom.value = index;
    if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", String(index));
    }
};
const onDragOver = (e: DragEvent, index: number) => {
    if (dragFrom.value === null) return;
    e.preventDefault();
    dragOver.value = index;
};
const onDragEnd = () => {
    dragFrom.value = null;
    dragOver.value = null;
};
const onDrop = (index: number) => {
    const from = dragFrom.value;
    onDragEnd();
    if (from !== null) void move(from, index);
};
</script>

<template>
    <div class="flex w-full max-w-3xl flex-col gap-6 pb-8 pt-2">
        <ErrorState v-if="isError" :error="error" title="Couldn't load this routine" @retry="refetch()" />

        <template v-else-if="isPending">
            <div v-if="showSkeleton" class="flex flex-col gap-6" role="status" aria-busy="true">
                <Skeleton height="2.5rem" width="12rem" />
                <Skeleton height="2.75rem" />
                <SkeletonList :rows="4" :leading="false" />
            </div>
        </template>

        <EmptyState
            v-else-if="!plan || !program"
            :icon="SearchX"
            :title="planId === null ? 'Invalid routine link' : 'Routine not found'"
            description="It may have been deleted."
        >
            <template #action>
                <Button :to="{ name: 'gym-plans' }">Back to schedule</Button>
            </template>
        </EmptyState>

        <template v-else>
            <PageHeader :title="plan.name" inline>
                <template #meta>{{ routineMeta(program, plan) }}</template>
                <template #actions>
                    <Menu :items="routineMenu" align="end" @select="onRoutineMenu">
                        <IconButton label="Routine actions"><Ellipsis /></IconButton>
                    </Menu>
                </template>
            </PageHeader>

            <!-- Days: which weekdays this routine is on. -->
            <section aria-labelledby="routine-days" class="flex flex-col gap-2">
                <div class="flex items-baseline justify-between gap-3">
                    <h2 id="routine-days" class="m-0 text-sm font-medium text-fg-muted">Days</h2>
                    <span class="text-sm text-fg-muted">{{ formatDays(toggles) }}</span>
                </div>
                <div class="grid grid-cols-7 gap-1 lg:flex lg:gap-2">
                    <Button
                        v-for="t in toggles"
                        :key="t.dow"
                        size="sm"
                        :variant="t.on ? 'secondary' : 'ghost'"
                        :aria-pressed="t.on"
                        :aria-label="t.takenBy && !t.on ? `${t.name} (${t.takenBy.name})` : t.name"
                        :title="t.takenBy && !t.on ? `${t.name}: ${t.takenBy.name}` : t.name"
                        :loading="pendingDays.has(t.dow)"
                        class="px-0 aria-pressed:border-accent aria-pressed:text-accent lg:w-14"
                        @click="toggleDay(t)"
                    >
                        {{ t.short }}
                    </Button>
                </div>
                <p v-if="otherDays" class="m-0 text-xs text-fg-subtle">Other routines: {{ otherDays }}</p>
            </section>

            <!-- The session, in order. -->
            <div class="flex flex-col">
                <RoutineStep
                    :step="1"
                    title="Warmup"
                    :meta="stepMeta('warmup', plural((plan.pre_mobility_items ?? []).length, 'move'))"
                >
                    <RoutineItemList
                        :items="plan.pre_mobility_items ?? []"
                        label="warmup"
                        placeholder="Add a warmup move…"
                        @change="saveMobility('warmup', $event)"
                    />
                </RoutineStep>

                <RoutineStep :step="2" title="Lifts" :meta="plural(exercises.length, 'exercise')">
                    <template #actions>
                        <Button variant="primary" size="sm" class="max-lg:hidden" @click="openAdd">
                            <template #icon-left><Plus class="size-4" /></template>
                            Add exercise
                        </Button>
                    </template>
                    <Card padding="none">
                        <List v-if="exercises.length">
                            <ListRow
                                v-for="(exercise, index) in exercises"
                                :key="exercise.ID"
                                :draggable="isDesktop ? 'true' : 'false'"
                                :class="[
                                    dragFrom === index ? 'opacity-60' : '',
                                    dragOver === index && dragFrom !== index ? 'bg-surface-2' : '',
                                ]"
                                @click="openEdit(exercise)"
                                @dragstart="onDragStart($event, index)"
                                @dragover="onDragOver($event, index)"
                                @drop.prevent="onDrop(index)"
                                @dragend="onDragEnd"
                            >
                                <template #leading>
                                    <GripVertical class="hidden size-4 cursor-grab text-fg-subtle lg:block" aria-hidden="true" />
                                    <span class="w-5 text-right text-sm text-fg-muted tabular-nums">{{ index + 1 }}</span>
                                </template>
                                {{ exercise.name }}
                                <template #end>
                                    <Menu :items="rowMenu(index)" align="end" @select="onRowMenu($event, exercise, index)">
                                        <IconButton :label="`${exercise.name} actions`" size="sm"><Ellipsis /></IconButton>
                                    </Menu>
                                </template>
                            </ListRow>
                        </List>
                        <EmptyState v-else :icon="Dumbbell" title="No lifts yet" description="Add the exercises for this routine, in the order you do them." />
                    </Card>
                    <Button variant="primary" block class="lg:hidden" @click="openAdd">
                        <template #icon-left><Plus class="size-4" /></template>
                        Add exercise
                    </Button>
                </RoutineStep>

                <RoutineStep :step="3" title="Cardio" :meta="stepMeta('cardio', 'Optional')">
                    <form ref="cardioBox" class="flex gap-2" @submit.prevent="saveCardio" @focusout="onCardioFocusOut">
                        <div class="min-w-0 flex-1">
                            <Input v-model="cardioType" aria-label="Cardio type" placeholder="e.g. Bike, Run" autocomplete="off" />
                        </div>
                        <div class="w-36">
                            <NumberField v-model="cardioMinutes" label="cardio minutes" unit="min" :min="0" :step="5" :decimals="0" />
                        </div>
                    </form>
                </RoutineStep>

                <RoutineStep
                    :step="4"
                    title="Stretch"
                    :meta="stepMeta('stretch', stretchCount(plan.post_mobility_items?.length ?? 0))"
                    last
                >
                    <RoutineItemList
                        :items="plan.post_mobility_items ?? []"
                        label="stretch"
                        placeholder="Add a stretch…"
                        @change="saveMobility('stretch', $event)"
                    />
                </RoutineStep>
            </div>

            <NameDialog
                v-model:open="renameOpen"
                title="Rename routine"
                label="Routine name"
                :initial="plan.name"
                confirm-text="Rename"
                :loading="renameRoutine.isPending.value"
                @submit="onRename"
            />
        </template>
    </div>
</template>
