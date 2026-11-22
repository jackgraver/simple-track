<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { CalendarDays, ChevronRight, Ellipsis, GripVertical, Pencil, Plus } from "lucide-vue-next";
import { getApiErrorMessage } from "~/api/httpError";
import type { WorkoutPlan } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import {
    Badge,
    Button,
    Card,
    EmptyState,
    ErrorState,
    IconButton,
    List,
    ListRow,
    Menu,
    PageHeader,
    Section,
    Select,
    Skeleton,
    SkeletonList,
    useDelayedFlag,
} from "~/shared/ui";
import type { MenuItem } from "~/shared/ui";
import {
    useActivateWorkoutProgram,
    useCreateWorkoutPlan,
    useCreateWorkoutProgram,
    useRenameWorkoutProgram,
    useWorkoutPrograms,
} from "~/api/workout/queries";
import { useAssignPlanDay, useUnassignPlanDay } from "~/api/workout/routines";
import DayPicker from "./components/DayPicker.vue";
import NameDialog from "./components/NameDialog.vue";
import {
    type DaySlot,
    DAY_NAMES,
    buildWeek,
    formatWeekSummary,
    pickDefaultProgram,
    planMeta,
    weekSummary,
} from "./domain/week";
import { parseId } from "./domain/routine";

const router = useRouter();
const route = useRoute();

// --- data -------------------------------------------------------------------------------------
const { data, isPending, isError, error, refetch } = useWorkoutPrograms();
const showSkeleton = useDelayedFlag(() => isPending.value);
const programs = computed(() => data.value?.programs ?? []);

// `?program=` (set when leaving a routine page) picks which program shows first.
const selectedProgramId = ref<number | null>(parseId(route.query.program));
const program = computed(
    () => programs.value.find((p) => p.ID === selectedProgramId.value) ?? pickDefaultProgram(programs.value),
);
watch(
    () => program.value?.ID,
    (id) => {
        if (id !== undefined) selectedProgramId.value = id;
    },
    { immediate: true },
);
const programModel = computed({
    get: () => selectedProgramId.value ?? "",
    set: (v: number | "") => {
        selectedProgramId.value = v === "" ? null : v;
    },
});
const programOptions = computed(() =>
    programs.value.map((p) => ({
        // The badge beside the select says "Active" for the shown program; the suffix helps inside the open list.
        label: p.is_active && p.ID !== program.value?.ID ? `${p.name} (active)` : p.name,
        value: p.ID,
    })),
);

const plans = computed(() => program.value?.plans ?? []);
const week = computed(() => buildWeek(plans.value, new Date().getDay()));
const summary = computed(() => formatWeekSummary(weekSummary(week.value)));

const failToast = (what: string, err: unknown) =>
    toast.push(`${what}: ${getApiErrorMessage(err, "Something went wrong")}`, "error");

const detailRoute = (plan: WorkoutPlan) => ({
    name: "gym-program-day-detail",
    params: { programId: String(program.value?.ID ?? plan.workout_program_id ?? 0), id: String(plan.ID) },
});

// --- program actions ---------------------------------------------------------------------------
const createProgram = useCreateWorkoutProgram();
const renameProgram = useRenameWorkoutProgram();
const activateProgram = useActivateWorkoutProgram();
const createPlan = useCreateWorkoutPlan();

const makeActive = async () => {
    if (!program.value) return;
    try {
        await activateProgram.mutateAsync(program.value.ID);
    } catch (err) {
        failToast("Couldn't make the program active", err);
    }
};

type NameDialogMode = "new-program" | "rename-program" | "new-routine";
const nameDialog = ref<{ mode: NameDialogMode; dow: number | null } | null>(null);
const nameDialogOpen = computed({
    get: () => nameDialog.value !== null,
    set: (v: boolean) => {
        if (!v) nameDialog.value = null;
    },
});
const nameDialogCopy = computed(() => {
    const d = nameDialog.value;
    if (d?.mode === "rename-program")
        return { title: "Rename program", label: "Program name", initial: program.value?.name ?? "", confirm: "Rename", placeholder: "" };
    if (d?.mode === "new-routine")
        return {
            title: d.dow === null ? "New routine" : `New routine for ${DAY_NAMES[d.dow]}`,
            label: "Routine name",
            initial: "",
            confirm: "Create",
            placeholder: "e.g. Push, Upper A",
        };
    return { title: "New program", label: "Program name", initial: "", confirm: "Create", placeholder: "e.g. Push Pull Legs" };
});
const nameDialogLoading = computed(
    () => createProgram.isPending.value || renameProgram.isPending.value || createPlan.isPending.value,
);

const onNameSubmit = async (name: string) => {
    const d = nameDialog.value;
    if (!d) return;
    try {
        if (d.mode === "new-program") {
            const res = await createProgram.mutateAsync(name);
            selectedProgramId.value = res.program.ID;
        } else if (d.mode === "rename-program" && program.value) {
            await renameProgram.mutateAsync({ id: program.value.ID, name });
        } else if (d.mode === "new-routine" && program.value) {
            await createPlan.mutateAsync({ programId: program.value.ID, name, dayOfWeek: d.dow });
        }
        nameDialog.value = null;
    } catch (err) {
        failToast(d.mode === "rename-program" ? "Couldn't rename the program" : `Couldn't create ${name}`, err);
    }
};

const programMenu: MenuItem[] = [
    { key: "rename", label: "Rename program", icon: Pencil },
    { key: "new", label: "New program", icon: Plus },
];
const onProgramMenu = (key: string) => {
    nameDialog.value = { mode: key === "rename" ? "rename-program" : "new-program", dow: null };
};

// --- scheduling ---------------------------------------------------------------------------------
const assign = useAssignPlanDay();
const unassign = useUnassignPlanDay();
// Assign/unassign patch the programs cache at once and roll back with a Retry toast, so no pending state here.
const putOnDay = (plan: WorkoutPlan, dow: number) => assign.mutate({ planId: plan.ID, dayOfWeek: dow });
const makeRest = (plan: WorkoutPlan, dow: number) => unassign.mutate({ planId: plan.ID, dayOfWeek: dow });
/**
 * Drag from one day to another moves the routine: assign the new day, and clear the old one only once
 * that succeeded, so a failed assign (rolled back, toasted) leaves the routine where it was.
 */
const moveDay = (plan: WorkoutPlan, from: number, to: number) => {
    assign
        .mutateAsync({ planId: plan.ID, dayOfWeek: to })
        .then(
            () => unassign.mutate({ planId: plan.ID, dayOfWeek: from }),
            () => {}, // the helper already rolled back and toasted
        );
};

// Day picker (tap on mobile, click or Enter on desktop).
const pickerSlot = ref<DaySlot | null>(null);
const pickerOpen = computed({
    get: () => pickerSlot.value !== null,
    set: (v: boolean) => {
        if (!v) pickerSlot.value = null;
    },
});
const onPick = (plan: WorkoutPlan) => {
    const slot = pickerSlot.value;
    pickerSlot.value = null;
    if (!slot || slot.plan?.ID === plan.ID) return;
    void putOnDay(plan, slot.dow);
};
const onPickRest = () => {
    const slot = pickerSlot.value;
    pickerSlot.value = null;
    if (slot?.plan) void makeRest(slot.plan, slot.dow);
};
const onPickCreate = () => {
    const dow = pickerSlot.value?.dow ?? null;
    pickerSlot.value = null;
    nameDialog.value = { mode: "new-routine", dow };
};

// Drag and drop (desktop only: the columns only render at lg).
const dragging = ref<{ plan: WorkoutPlan; from: number | null } | null>(null);
const dropTarget = ref<number | "pool" | null>(null);

const onDragStart = (e: DragEvent, plan: WorkoutPlan, from: number | null) => {
    dragging.value = { plan, from };
    e.dataTransfer?.setData("text/plain", String(plan.ID));
    if (e.dataTransfer) e.dataTransfer.effectAllowed = "move";
};
const onDragEnd = () => {
    dragging.value = null;
    dropTarget.value = null;
};
const onDragOver = (e: DragEvent, target: number | "pool") => {
    if (!dragging.value) return;
    e.preventDefault();
    dropTarget.value = target;
};
const onDropDay = (dow: number) => {
    const d = dragging.value;
    onDragEnd();
    if (!d || d.from === dow) return;
    if (week.value[dow]?.plan?.ID === d.plan.ID) return;
    void (d.from === null ? putOnDay(d.plan, dow) : moveDay(d.plan, d.from, dow));
};
const onDropPool = () => {
    const d = dragging.value;
    onDragEnd();
    if (d && d.from !== null) void makeRest(d.plan, d.from);
};

const goDetail = (plan: WorkoutPlan) => router.push(detailRoute(plan));
</script>

<template>
    <div class="flex w-full flex-col gap-6 pb-8 pt-2">
        <PageHeader title="Schedule">
            <template v-if="program" #meta>{{ summary }}</template>
            <template v-if="program" #actions>
                <Button variant="primary" class="max-lg:hidden" @click="nameDialog = { mode: 'new-routine', dow: null }">
                    <template #icon-left><Plus class="size-4" /></template>
                    New routine
                </Button>
            </template>
        </PageHeader>

        <ErrorState v-if="isError" :error="error" title="Couldn't load your programs" @retry="refetch()" />

        <template v-else-if="isPending">
            <div v-if="showSkeleton" class="flex flex-col gap-6" role="status" aria-busy="true">
                <Skeleton height="2.75rem" width="20rem" />
                <SkeletonList :rows="7" class="lg:hidden" />
                <div class="hidden grid-cols-7 gap-3 lg:grid">
                    <Skeleton v-for="n in 7" :key="n" shape="block" height="9rem" />
                </div>
            </div>
        </template>

        <EmptyState
            v-else-if="!program"
            :icon="CalendarDays"
            title="No programs yet"
            description="A program is a set of routines laid out over the week, like Push Pull Legs."
        >
            <template #action>
                <Button variant="primary" @click="nameDialog = { mode: 'new-program', dow: null }">Create a program</Button>
            </template>
        </EmptyState>

        <template v-else>
            <!-- Program bar -->
            <div class="flex items-center gap-2">
                <div class="min-w-0 flex-1 lg:max-w-xs">
                    <Select v-model="programModel" :options="programOptions" aria-label="Program" />
                </div>
                <Badge v-if="program.is_active" tone="success" size="md">Active</Badge>
                <Button v-else :loading="activateProgram.isPending.value" @click="makeActive">Make active</Button>
                <Menu :items="programMenu" align="end" @select="onProgramMenu">
                    <IconButton label="Program actions"><Ellipsis /></IconButton>
                </Menu>
            </div>

            <Section title="This week">
                <!-- Mobile: one row per day; tap to choose a routine or rest. -->
                <Card padding="none" class="lg:hidden">
                    <List>
                        <ListRow
                            v-for="slot in week"
                            :key="slot.dow"
                            @click="pickerSlot = slot"
                        >
                            <template #leading>
                                <span
                                    class="w-9 text-sm font-medium"
                                    :class="slot.isToday ? 'text-accent' : 'text-fg-muted'"
                                >{{ slot.short }}</span>
                            </template>
                            <span v-if="slot.plan" class="font-medium">{{ slot.plan.name }}</span>
                            <span v-else class="text-fg-subtle">Rest</span>
                            <template v-if="slot.plan" #subtitle>{{ planMeta(slot.plan).join(" · ") }}</template>
                            <template #trailing>
                                <Badge v-if="slot.isToday" tone="accent">Today</Badge>
                                <ChevronRight v-if="slot.plan" class="size-4 text-fg-subtle" aria-hidden="true" />
                                <Plus v-else class="size-4 text-fg-subtle" aria-hidden="true" />
                            </template>
                        </ListRow>
                    </List>
                </Card>

                <!-- Desktop: seven columns. Drag a routine between days or to "Not scheduled"; click to choose. -->
                <div class="hidden grid-cols-7 gap-2 lg:grid">
                    <Card
                        v-for="slot in week"
                        :key="slot.dow"
                        padding="sm"
                        interactive
                        fill
                        role="button"
                        tabindex="0"
                        :aria-label="`${slot.name}: ${slot.plan ? slot.plan.name : 'rest day'}. Choose routine`"
                        :draggable="slot.plan ? 'true' : 'false'"
                        class="min-h-36"
                        :class="[
                            slot.plan ? '' : 'border-dashed',
                            dropTarget === slot.dow ? 'outline-2 outline-accent' : '',
                            dragging && dragging.from === slot.dow ? 'opacity-60' : '',
                        ]"
                        @click="pickerSlot = slot"
                        @keydown.enter.prevent="pickerSlot = slot"
                        @keydown.space.prevent="pickerSlot = slot"
                        @dragstart="slot.plan && onDragStart($event, slot.plan, slot.dow)"
                        @dragend="onDragEnd"
                        @dragover="onDragOver($event, slot.dow)"
                        @dragleave="dropTarget === slot.dow && (dropTarget = null)"
                        @drop.prevent="onDropDay(slot.dow)"
                    >
                        <div class="flex flex-1 flex-col gap-2">
                            <div class="flex items-center justify-between gap-1">
                                <span class="text-xs font-medium" :class="slot.isToday ? 'text-accent' : 'text-fg-muted'">
                                    {{ slot.short }}
                                </span>
                                <Badge v-if="slot.isToday" tone="accent">Today</Badge>
                            </div>
                            <div v-if="slot.plan" class="flex items-start gap-1.5">
                                <GripVertical class="mt-0.5 size-4 shrink-0 cursor-grab text-fg-subtle" aria-hidden="true" />
                                <div class="flex min-w-0 flex-col gap-1">
                                    <span class="text-sm font-semibold break-words text-fg">{{ slot.plan.name }}</span>
                                    <span class="flex flex-col gap-0.5 text-xs text-fg-muted">
                                        <span v-for="m in planMeta(slot.plan)" :key="m">{{ m }}</span>
                                    </span>
                                </div>
                            </div>
                            <span v-else class="flex flex-1 items-center justify-center text-xs text-fg-subtle">Rest</span>
                        </div>
                    </Card>
                </div>
            </Section>

            <!-- The program's routine library. A routine can be on several days: drag one onto a day to add
                 that day (desktop), or pick it from a day's sheet. Drop a day here to clear that day. -->
            <Section title="Routines">
                <template v-if="plans.length" #actions>
                    <span class="text-xs text-fg-subtle max-lg:hidden">Drag onto a day to add it</span>
                </template>
                <Card
                    padding="none"
                    :class="dropTarget === 'pool' ? 'outline-2 outline-accent' : ''"
                    @dragover="onDragOver($event, 'pool')"
                    @dragleave="dropTarget === 'pool' && (dropTarget = null)"
                    @drop.prevent="onDropPool"
                >
                    <div v-if="plans.length" class="flex flex-wrap content-start gap-2 p-3">
                        <div
                            v-for="p in plans"
                            :key="p.ID"
                            role="button"
                            tabindex="0"
                            draggable="true"
                            :aria-label="`Open ${p.name}`"
                            class="flex min-h-touch min-w-36 cursor-pointer items-start gap-2 rounded-control border border-line bg-surface-2 px-3 py-2 text-left transition-colors duration-100 hover:bg-surface-3 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent lg:cursor-grab lg:active:cursor-grabbing"
                            :class="dragging?.plan.ID === p.ID && dragging.from === null ? 'opacity-60' : ''"
                            @click="goDetail(p)"
                            @keydown.enter.prevent="goDetail(p)"
                            @dragstart="onDragStart($event, p, null)"
                            @dragend="onDragEnd"
                        >
                            <GripVertical class="mt-0.5 hidden size-4 shrink-0 text-fg-subtle lg:block" aria-hidden="true" />
                            <span class="flex min-w-0 flex-col">
                                <span class="truncate text-sm font-medium text-fg">{{ p.name }}</span>
                                <span class="text-xs text-fg-muted">{{ planMeta(p).join(" · ") }}</span>
                            </span>
                        </div>
                    </div>
                    <EmptyState
                        v-else
                        title="No routines in this program"
                        description="Create a routine, then put it on a day."
                    />
                </Card>
            </Section>

            <Button variant="primary" size="lg" block class="lg:hidden" @click="nameDialog = { mode: 'new-routine', dow: null }">
                <template #icon-left><Plus class="size-4" /></template>
                New routine
            </Button>
        </template>

        <DayPicker
            v-model:open="pickerOpen"
            :day="pickerSlot"
            :plans="plans"
            :program-id="program?.ID ?? null"
            @choose="onPick"
            @rest="onPickRest"
            @create="onPickCreate"
        />
        <NameDialog
            v-model:open="nameDialogOpen"
            :title="nameDialogCopy.title"
            :label="nameDialogCopy.label"
            :initial="nameDialogCopy.initial"
            :placeholder="nameDialogCopy.placeholder"
            :confirm-text="nameDialogCopy.confirm"
            :loading="nameDialogLoading"
            @submit="onNameSubmit"
        />
    </div>
</template>
