<script setup lang="ts">
import { computed, defineAsyncComponent, h, nextTick, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { dialogManager } from "~/composables/dialog/useDialog";
import { DEFAULT_EXERCISE_LOAD_TYPE, type ExerciseLoadType } from "~/types/workout";
import { ChartLine } from "lucide-vue-next";
import { Button, IconButton, SegmentedControl, Skeleton } from "~/shared/ui";
import type { SetSide } from "~/types/workout";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import { buildWorkoutSteps, formatSets } from "../../domain/workoutSteps";
import { findExerciseGroupByExerciseId, parseExerciseIdParam } from "../domain/exerciseRouteGroup";
import { previousRepsForSetAtWeight } from "../domain/loggedSetDefaults";
import {
    formatExerciseWeightSetup,
    isPlateLoadedExercise,
    parseWeightSetup,
    type WeightSetupDialogResult,
} from "../domain/weightSetup";
import { overloadStatus, plannedSetCount, progressVerdict, visibleMachineSettings } from "../domain/liveLogging";
import { buildUpNext, type UpNextItem } from "../domain/upNext";
import { completePairCount, hasSides, isOpen, pairSets, previousSideSet, setsForProgression, sideLabel } from "../domain/unilateral";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import WeightSetupDialog from "../dialog/WeightSetupDialog.vue";
import WeightSheet, { type WeightSheetResult } from "../dialog/WeightSheet.vue";
import NotesSheet from "../dialog/NotesSheet.vue";
import FirstSideDialog from "../dialog/FirstSideDialog.vue";
import LoggingHeader from "./LoggingHeader.vue";
import NumberSlider from "./NumberSlider.vue";
import RestClock from "./RestClock.vue";
import SetupPanel from "./SetupPanel.vue";
import TodayVsLastList from "./TodayVsLastList.vue";
import UpNextView from "./UpNextView.vue";

// The chart (and chart.js, ~60KB compressed) only matters once the sheet opens: load it then.
// Same skeleton the panel shows while its data loads, after the same short delay.
const ExerciseProgressionChartPanel = defineAsyncComponent({
    loader: () => import("~/pages/gym/progression/ExerciseProgressionChartPanel.vue"),
    loadingComponent: () => h(Skeleton, { shape: "block", height: "16rem" }),
    delay: 150,
});

// One screen for the whole lift → log → rest loop (docs/projects/ui-overhaul/DECISIONS.md, 2026-11-21).
// Top to bottom: rest clock, weight + setup, the set just logged, reps slider, one primary action.
const route = useRoute();
const router = useRouter();
const { offset, goBackToLogging } = useLoggingRouteContext();
const store = useWorkoutStore(offset);
const { log, pending } = store;

const exerciseId = computed(() => parseExerciseIdParam(route));
const exerciseGroup = computed(() => {
    if (exerciseId.value == null || pending.value) return null;
    return findExerciseGroupByExerciseId(log.value, exerciseId.value);
});
const session = useExerciseLoggingSession({ exerciseGroup, pending, offset, router });

/* ---------- exercise facts ---------- */
const exercise = computed(() => session.exerciseGroup?.planned ?? session.exerciseGroup?.logged?.exercise ?? null);
const exerciseName = computed(() => exercise.value?.name ?? "");
const cues = computed(() => exercise.value?.cues ?? "");
const machineSettings = computed(() => visibleMachineSettings(exercise.value?.machine_settings));
const previousSets = computed(() =>
    (session.exerciseGroup?.previous?.sets ?? []).map((set) => ({
        weight: set.weight,
        reps: set.reps,
        side: set.side ?? null,
        pair_id: set.pair_id ?? null,
        client_id: set.client_id ?? null,
    })),
);
// Unilateral sets count one per pair (its weaker side) for targets, overload and the verdict.
const previousForProgression = computed(() => setsForProgression(previousSets.value));
const repRollover = computed(() => {
    const g = session.exerciseGroup;
    const rollover = g?.planned?.rep_rollover ?? g?.logged?.exercise?.rep_rollover ?? g?.previous?.exercise?.rep_rollover;
    return typeof rollover === "number" && Number.isFinite(rollover) && rollover > 0 ? rollover : undefined;
});
const loadType = computed<ExerciseLoadType>(
    () => (exercise.value?.load_type ?? DEFAULT_EXERCISE_LOAD_TYPE) as ExerciseLoadType,
);
const plates = computed(() => {
    if (!isPlateLoadedExercise(loadType.value)) return null;
    const setup = session.currentWeightSetup.trim();
    if (!setup) return "";
    const parsed = parseWeightSetup(setup);
    if (!parsed.parsed) return setup;
    return formatExerciseWeightSetup(loadType.value, parsed.platesPerSide) || "No plates";
});

/* ---------- sets ---------- */
// Unilateral: complete left + right pairs. A pair with one side logged isn't done yet.
const loggedCount = computed(() =>
    session.unilateral ? completePairCount(session.loggedSets) : session.loggedSets.length,
);
/** One side of the latest set is logged and the other isn't yet. */
const pairOpen = computed(() => {
    if (!session.unilateral) return false;
    const pairs = pairSets(session.loggedSets);
    const last = pairs[pairs.length - 1];
    return !!last && isOpen(last);
});
const targetSets = computed(() => plannedSetCount(previousForProgression.value));
const plannedDone = computed(() => loggedCount.value >= targetSets.value && !pairOpen.value);
const editing = computed(() => session.editingSetIndex != null);
const overload = computed(() => overloadStatus(previousForProgression.value, repRollover.value, session.currentWeight));
const verdict = computed(() => progressVerdict(setsForProgression(session.loggedSets), repRollover.value));

/* ---------- reps slider ---------- */
const extraSetOpen = ref(false);
const showSlider = computed(() => editing.value || !plannedDone.value || extraSetOpen.value);
const previousRep = computed(() => {
    const side = session.currentSide;
    if (!side) return previousRepsForSetAtWeight(previousSets.value, session.currentSetNumber - 1, session.currentWeight);
    const lastTime = previousSideSet(previousSets.value, session.currentSetNumber - 1, side);
    return previousRepsForSetAtWeight(lastTime ? [lastTime] : [], 0, session.currentWeight);
});

/* ---------- side (unilateral) ---------- */
const sideOptions: { label: string; value: SetSide }[] = [
    { label: "Left", value: "left" },
    { label: "Right", value: "right" },
];
// The first side is asked for when Set 1 is logged (FirstSideDialog), so until then no side is shown.
const askFirstSide = computed(() => session.unilateral && !editing.value && session.loggedSets.length === 0);
const sideSuffix = computed(() =>
    session.currentSide && !askFirstSide.value ? ` · ${sideLabel(session.currentSide)}` : "",
);
// Only to fix a logged set's side: new sides just alternate.
const showSidePicker = computed(() => editing.value && session.currentSide != null);
const sliderMarkers = computed(() => {
    const markers: { value: number; label: string; color?: "orange"; labelPosition?: "bottom" }[] = [];
    if (previousRep.value != null) {
        markers.push({ value: previousRep.value, label: "Last time", color: "orange", labelPosition: "bottom" });
    }
    if (repRollover.value != null) markers.push({ value: repRollover.value, label: "Rollover" });
    return markers;
});

/* ---------- view: logging, or "Up next" after finishing ---------- */
const view = ref<"log" | "next">("log");
const setupOpen = ref(false);
const setupExpanded = computed(() => session.loggedSets.length === 0 || setupOpen.value);
watch(exerciseId, () => {
    view.value = "log";
    setupOpen.value = false;
    extraSetOpen.value = false;
});

const scrollToTop = () => window.scrollTo({ top: 0, behavior: "instant" });

/* ---------- actions ---------- */
// No scrolling after a log: the clock is already at the top, and a jump on top of the setup card
// collapsing read as a flicker on the phone. No busy state either: the set appears on tap
// (optimistic, saved in the background; a failed save is marked in the list below).
const logSet = async () => {
    if (askFirstSide.value && session.currentSide) {
        const side = await dialogManager.custom<SetSide>({
            title: "Which side was that?",
            component: FirstSideDialog,
            componentProps: { side: session.currentSide },
        });
        if (side == null) return;
        session.setSide(side);
    }
    if (session.addNextSet()) extraSetOpen.value = false;
};

const editingRow = computed(() =>
    session.editingSetIndex != null ? session.loggedSets[session.editingSetIndex] : undefined,
);
/** Today's list shows paired cards (any set has a side), same rule as TodayVsLastList. */
const pairedView = computed(() => hasSides(session.loggedSets));
const deleteEditedSide = () => {
    if (session.editingSetIndex != null) session.deleteSet(session.editingSetIndex);
};

// Tap a set in the list below to edit it with the slider; tap it again (or "Back to set N") to stop.
// No scrolling: starting or stopping can show/hide the slider above the list, so the page is shifted
// by the same amount and the tapped row stays under the finger (iOS Safari has no scroll anchoring).
const selectSet = (index: number, row?: HTMLElement) => {
    const before = row?.getBoundingClientRect().top;
    if (session.editingSetIndex === index) session.cancelEditingSet();
    else session.editSet(index);
    if (row && before != null) {
        void nextTick(() => {
            const shift = row.getBoundingClientRect().top - before;
            if (shift) window.scrollBy({ top: shift, behavior: "instant" });
        });
    }
};

const finish = () => {
    if (session.finishLogging()) {
        view.value = "next";
        scrollToTop();
    }
};

const openPlatesDialog = async () => {
    const result = await dialogManager.custom<WeightSetupDialogResult>({
        title: "Plates",
        component: WeightSetupDialog,
        componentProps: {
            currentSetup: session.currentWeightSetup,
            targetWeightLbs: session.currentWeight,
            loadType: loadType.value,
        },
        wide: true,
    });
    if (result == null) return;
    session.updateWeightSetup(result.weightSetup);
    if (result.syncWeight) session.commitWeightFromInput(result.totalLbs);
};

// Today's weight is the automatic +5: the setup card offers "stay" vs "go up" until the first set.
// A weight typed in the sheet that's neither of the two ends the choice.
const progressionChoice = computed(() => {
    const p = session.weightProgression;
    if (!p || session.loggedSets.length > 0) return null;
    return session.currentWeight === p.nextWeight || session.currentWeight === p.previousWeight ? p : null;
});

const openWeightSheet = async () => {
    const result = await dialogManager.custom<WeightSheetResult>({
        title: `Weight & setup · ${exerciseName.value}`,
        component: WeightSheet,
        componentProps: {
            weight: session.currentWeight,
            step: session.weightIncrement,
            overload: overload.value,
            plates: plates.value,
            machineSettings: machineSettings.value,
            cues: cues.value,
        },
    });
    if (result == null) return;
    session.commitWeightFromInput(result.weight);
    if (result.action === "plates") await openPlatesDialog();
};

const openNotes = async () => {
    const result = await dialogManager.custom<string>({
        title: `Notes · ${exerciseName.value}`,
        component: NotesSheet,
        componentProps: { notes: session.notes },
    });
    if (result != null) session.updateNotes(result);
};

const openProgression = () => {
    const id = exerciseId.value;
    if (id == null) return;
    void dialogManager.custom<void>({
        title: `${exerciseName.value} Progression`,
        component: ExerciseProgressionChartPanel,
        componentProps: { exerciseId: id, lastTimeSets: previousForProgression.value, initialRange: "6m" },
        wide: true,
    });
};

/* ---------- up next ---------- */

const steps = computed(() =>
    buildWorkoutSteps({
        log: log.value,
        plannedCardio: store.plannedCardio.value,
        loggedCardio: store.loggedCardio.value,
        plannedPreMobility: store.plannedPreMobility.value,
        loggedPreMobility: store.loggedPreMobility.value,
        plannedPostMobility: store.plannedPostMobility.value,
        loggedPostMobility: store.loggedPostMobility.value,
    }),
);
const upNext = computed(() => buildUpNext(steps.value, log.value, exerciseId.value));
const doneCount = computed(() => steps.value.filter((s) => s.done).length);

const goTo = (item: UpNextItem) => {
    const query = route.query;
    if (item.kind === "lift" && item.exerciseId != null) {
        void router.push({ name: "logging-exercise", params: { id: String(item.exerciseId) }, query });
    } else if (item.kind === "cardio") {
        void router.push({ name: "logging-cardio", query });
    } else if (item.kind === "warmup" || item.kind === "cooldown") {
        void router.push({ name: "logging-mobility", params: { slot: item.kind === "warmup" ? "pre" : "post" }, query });
    }
};
</script>

<template>
    <div class="flex w-full flex-col gap-4 pb-6">
        <LoggingHeader
            :title="view === 'next' ? 'Up next' : exerciseName"
            :subtitle="
                view === 'next'
                    ? `${doneCount} of ${steps.length} done`
                    : editing
                      ? `Editing set ${session.currentSetNumber}${sideSuffix}`
                      : plannedDone
                        ? loggedCount > targetSets
                          ? `${loggedCount} sets · done`
                          : `${loggedCount} of ${targetSets} sets · done`
                        : `Set ${session.currentSetNumber} of ${targetSets}${sideSuffix}`
            "
            @back="goBackToLogging"
        >
            <template #right>
                <IconButton label="Progression chart" @click="openProgression"><ChartLine /></IconButton>
            </template>
        </LoggingHeader>

        <RestClock />

        <UpNextView
            v-if="view === 'next'"
            :finished-name="exerciseName"
            :finished-sets="formatSets(session.loggedSets)"
            :verdict="verdict"
            :items="upNext"
            @go="goTo"
            @list="goBackToLogging"
        />

        <template v-else>
            <SetupPanel
                :expanded="setupExpanded"
                :weight="session.currentWeight"
                :overload="overload"
                :last-time="formatSets(previousSets)"
                :machine-settings="machineSettings"
                :plates="plates"
                :cues="cues"
                :progression="progressionChoice"
                @edit="openWeightSheet"
                @choose="session.chooseWeightProgression"
            />
            <!-- Unilateral, editing a logged side: fix which side it was. -->
            <section v-if="showSidePicker && session.currentSide" class="flex flex-col gap-1" aria-label="Side">
                <p class="m-0 text-xs font-semibold tracking-wide text-fg-muted uppercase">Side</p>
                <SegmentedControl
                    :model-value="session.currentSide"
                    :options="sideOptions"
                    label="Side"
                    @update:model-value="session.setSide"
                />
            </section>
            <section v-if="showSlider" class="flex flex-col gap-1" aria-label="Reps">
                <p class="m-0 text-xs font-semibold tracking-wide uppercase" :class="editing ? 'text-accent' : 'text-fg-muted'">
                    {{ editing ? `Editing set ${session.currentSetNumber}${sideSuffix}` : "Reps" }}
                </p>
                <NumberSlider
                    label=""
                    :model-value="session.currentReps"
                    :markers="sliderMarkers"
                    integer-only
                    :max="50"
                    @update:model-value="session.commitRepsFromInput"
                />
            </section>

            <!-- Pinned above the tab bar so Log/Finish stays on screen on short phones. -->
            <div class="sticky bottom-at-tabbar z-raised -mx-4 flex flex-col gap-2 bg-canvas px-4 py-2 lg:static lg:mx-0 lg:p-0">
                <Button v-if="showSlider" variant="primary" size="lg" block @click="logSet">
                    {{
                        editing
                            ? `Save Set ${session.currentSetNumber}${sideSuffix}`
                            : `Log Set ${session.currentSetNumber}${sideSuffix}`
                    }}
                </Button>
                <Button v-if="plannedDone && !editing && !extraSetOpen" variant="primary" size="lg" block @click="finish">
                    Finish {{ exerciseName }}
                </Button>
                <Button
                    v-if="plannedDone && !editing && !extraSetOpen"
                    variant="ghost"
                    class="self-center"
                    @click="extraSetOpen = true"
                >
                    + Add set {{ loggedCount + 1 }}
                </Button>
                <Button v-if="extraSetOpen && !editing" variant="ghost" class="self-center" @click="extraSetOpen = false">
                    Cancel extra set
                </Button>
                <!-- The paired cards have no ✕: a side (or a bilateral set among them) is deleted from its edit mode. -->
                <Button v-if="editing && editingRow && pairedView" variant="danger" size="lg" block @click="deleteEditedSide">
                    {{
                        editingRow.side
                            ? `Delete ${sideLabel(editingRow.side).toLowerCase()} side`
                            : `Delete set ${session.currentSetNumber}`
                    }}
                </Button>
            </div>

            <div class="mt-4 flex flex-col gap-4 border-t border-dashed border-line pt-6">
                <TodayVsLastList
                    :sets="session.loggedSets"
                    :previous="previousSets"
                    :done="loggedCount"
                    :verdict="verdict"
                    :editing-index="session.editingSetIndex"
                    :target="targetSets"
                    @select="selectSet"
                    @delete="session.deleteSet"
                    @retry="session.retrySet"
                />
                <div class="flex flex-wrap gap-x-2 gap-y-1">
                    <Button
                        v-if="session.loggedSets.length > 0 && !plannedDone && !editing"
                        variant="ghost"
                        size="sm"
                        @click="finish"
                    >
                        Finish early
                    </Button>
                    <Button variant="ghost" size="sm" @click="openNotes">{{ session.notes ? "Edit notes" : "Add notes" }}</Button>
                    <Button v-if="session.loggedSets.length > 0" variant="ghost" size="sm" @click="setupOpen = !setupOpen">
                        {{ setupOpen ? "Hide setup" : "Show setup" }}
                    </Button>
                </div>
            </div>
        </template>
    </div>
</template>
