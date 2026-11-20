<script setup lang="ts">
import type { ExerciseGroup } from "~/api/workout/api";
import type { ComboboxOption, MenuItem } from "~/shared/ui";
import { ref, computed } from "vue";
import { ChevronRight, Dumbbell, MoreVertical, PartyPopper, Trash2 } from "lucide-vue-next";
import { Button, Card, Combobox, EmptyState, IconButton, List, ListRow, Menu } from "~/shared/ui";
import { useExercisesPaginated } from "~/api/workout/queries";
import StepMarker from "~/pages/gym/components/StepMarker.vue";
import StepProgress from "~/pages/gym/components/StepProgress.vue";
import { nextStepIndex, stepKindLabel, type WorkoutStep } from "../domain/workoutSteps";
import { buildSessionSummary } from "../domain/sessionSummary";
import SessionSummary from "./SessionSummary.vue";

const props = defineProps<{
    workoutName: string;
    dateLabel: string;
    steps: WorkoutStep[];
    exercises: ExerciseGroup[];
    removingId?: number | null;
}>();

const emit = defineEmits<{
    (e: "select-step", step: WorkoutStep): void;
    (e: "add-exercise", exerciseId: number): void;
    (e: "remove-exercise", index: number): void;
    (e: "finish"): void;
}>();

const next = computed(() => nextStepIndex(props.steps));
const nextStep = computed(() => (next.value >= 0 ? props.steps[next.value] : null));
const doneCount = computed(() => props.steps.filter((s) => s.done).length);
const allDone = computed(() => props.steps.length > 0 && next.value === -1);

const summary = computed(() => buildSessionSummary(props.exercises));
const hasGains = computed(() => summary.value.lifts.length > 0);

const stateOf = (i: number) => (props.steps[i]?.done ? "done" : i === next.value ? "next" : "todo");

const startLabel = (s: WorkoutStep) =>
    s.kind === "lift" ? "Start" : s.kind === "cardio" ? "Log cardio" : `Start ${stepKindLabel[s.kind].toLowerCase()}`;

// Add exercise (catalog search, skipping what's already in today's log)
const exerciseSearch = ref("");
const { data: exercisesPages, fetchNextPage, hasNextPage, isFetching } = useExercisesPaginated(exerciseSearch);
const exerciseOptions = computed<ComboboxOption[]>(() => {
    const all = exercisesPages.value?.pages.flatMap((p) => p.exercises) ?? [];
    const existing = new Set(
        props.exercises.map((g) => g.planned?.ID ?? g.logged?.exercise_id).filter((id): id is number => id != null),
    );
    return all.filter((ex) => !existing.has(ex.ID)).map((ex) => ({ label: ex.name, value: ex.ID }));
});

const exerciseIdAt = (index: number | undefined) => {
    if (index == null) return null;
    const g = props.exercises[index];
    return g?.logged?.exercise_id ?? g?.planned?.ID ?? null;
};
const isRemoving = (s: WorkoutStep) => props.removingId != null && exerciseIdAt(s.exerciseIndex) === props.removingId;

const rowMenu: MenuItem[] = [{ key: "remove", label: "Remove from this workout", icon: Trash2, tone: "danger" }];
const onRowMenu = (s: WorkoutStep, key: string) => {
    if (key === "remove" && s.exerciseIndex != null) emit("remove-exercise", s.exerciseIndex);
};
</script>

<template>
    <div class="flex w-full flex-col gap-6">
        <!-- Where you are in the session -->
        <header class="flex flex-col gap-3">
            <div class="flex items-end justify-between gap-3">
                <div class="flex min-w-0 flex-col">
                    <p class="m-0 text-xs font-semibold uppercase tracking-wide text-fg-muted">{{ dateLabel }}</p>
                    <h1 class="m-0 truncate text-xl font-semibold tracking-tight text-fg">
                        {{ workoutName || "Workout" }}
                    </h1>
                </div>
                <p v-if="steps.length" class="m-0 shrink-0 text-sm text-fg-muted tabular-nums">
                    <span class="font-semibold text-fg">{{ doneCount }}</span> / {{ steps.length }} done
                </p>
            </div>
            <StepProgress v-if="steps.length" :steps="steps" :next="next" />
        </header>

        <!-- Up next: the one thing to do now -->
        <Card v-if="nextStep">
            <div class="flex flex-col gap-4">
                <div class="flex flex-col gap-1">
                    <p class="m-0 text-xs font-semibold uppercase tracking-wide text-accent">
                        Up next · {{ stepKindLabel[nextStep.kind] }}
                    </p>
                    <h2 class="m-0 text-xl font-semibold tracking-tight text-fg">{{ nextStep.name }}</h2>
                    <p v-if="nextStep.detail" class="m-0 text-sm text-fg-muted tabular-nums">{{ nextStep.detail }}</p>
                </div>
                <Button variant="primary" size="lg" block @click="emit('select-step', nextStep)">
                    {{ startLabel(nextStep) }}
                    <ChevronRight aria-hidden="true" class="size-5" />
                </Button>
            </div>
        </Card>
        <Card v-else-if="allDone">
            <div class="flex flex-col gap-4">
                <div class="flex items-center gap-3">
                    <span class="flex size-10 shrink-0 items-center justify-center rounded-full bg-success-soft text-success">
                        <PartyPopper aria-hidden="true" class="size-5" />
                    </span>
                    <div class="flex min-w-0 flex-col">
                        <h2 class="m-0 text-lg font-semibold text-fg">Workout complete</h2>
                        <p class="m-0 text-sm text-fg-muted tabular-nums">All {{ steps.length }} steps logged</p>
                    </div>
                </div>
                <Button variant="primary" size="lg" block @click="emit('finish')">Finish workout</Button>
            </div>
        </Card>

        <!-- After the workout, the payoff comes first; mid-session it sits below the list -->
        <SessionSummary v-if="allDone && hasGains" :summary="summary" complete />

        <!-- The whole session, in order -->
        <section aria-labelledby="session-plan" class="flex flex-col gap-3">
            <h2 id="session-plan" class="m-0 text-xs font-semibold uppercase tracking-wide text-fg-muted">Session</h2>
            <Card padding="none">
                <List v-if="steps.length">
                    <ListRow
                        v-for="(s, i) in steps"
                        :key="s.key"
                        :selected="stateOf(i) === 'next'"
                        :pending="isRemoving(s)"
                        @click="emit('select-step', s)"
                    >
                        <template #leading><StepMarker :n="i + 1" :state="stateOf(i)" /></template>
                        <span :class="s.done ? 'text-fg-muted' : 'text-fg'">{{ s.name }}</span>
                        <span class="sr-only">{{ s.done ? ", done" : stateOf(i) === "next" ? ", up next" : "" }}</span>
                        <template v-if="s.detail" #subtitle>
                            <span class="tabular-nums" :class="s.done ? 'text-fg' : ''">{{ s.detail }}</span>
                        </template>
                        <template #end>
                            <Menu v-if="s.kind === 'lift'" :items="rowMenu" @select="onRowMenu(s, $event)">
                                <IconButton :label="`More actions for ${s.name}`" :disabled="isRemoving(s)">
                                    <MoreVertical />
                                </IconButton>
                            </Menu>
                            <span v-else aria-hidden="true" class="size-touch lg:size-row" />
                        </template>
                    </ListRow>
                </List>
                <EmptyState
                    v-else
                    :icon="Dumbbell"
                    title="Nothing planned"
                    description="Add an exercise below, or pick a plan for this day on the Gym page."
                />
            </Card>
            <Combobox
                :options="exerciseOptions"
                :has-more="hasNextPage ?? false"
                :loading="isFetching"
                placeholder="Add an exercise…"
                @select="(o: ComboboxOption) => emit('add-exercise', o.value as number)"
                @load-more="() => hasNextPage && fetchNextPage()"
                @search="(q: string) => (exerciseSearch = q)"
            />
        </section>

        <SessionSummary v-if="!allDone && hasGains" :summary="summary" :complete="false" />

        <Button v-if="!allDone" variant="secondary" size="lg" block @click="emit('finish')">
            {{ doneCount > 0 ? "Finish workout" : "Back to Gym" }}
        </Button>
    </div>
</template>
