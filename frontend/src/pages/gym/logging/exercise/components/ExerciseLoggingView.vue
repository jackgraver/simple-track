<script setup lang="ts">
import type { ExerciseGroup } from "../../store/useWorkoutStore";
import {
    DEFAULT_EXERCISE_LOAD_TYPE,
    type ExerciseLoadType,
} from "~/types/workout";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "../domain/exerciseRouteGroup";
import { pickBestLoggedSet } from "../domain/loggedSetDefaults";
import LoggingHeader from "./LoggingHeader.vue";
import LoggedSetsList from "./LoggedSetsList.vue";
import NumberSlider from "./NumberSlider.vue";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Pencil } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import EditCuesDialog from "../dialog/EditCuesDialog.vue";
import WeightSetupDialog from "../dialog/WeightSetupDialog.vue";
import {
    computeExerciseLoadLbs,
    formatExerciseWeightSetup,
    parseWeightSetup,
    type WeightSetupDialogResult,
} from "../domain/weightSetup";
import ExerciseProgressionChartPanel from "~/pages/gym/progression/ExerciseProgressionChartPanel.vue";
import { toast } from "~/composables/toast/useToast";
import { formatDate, isSameDay } from "~/utils/dateUtil";

const route = useRoute();
const router = useRouter();
const { offset } = useLoggingRouteContext();
const { log, data, pending, logExercise, deleteLoggedSet } =
    useWorkoutStore(offset);

const exerciseId = computed(() => parseExerciseIdParam(route));

const exerciseGroup = computed<ExerciseGroup | null>(() => {
    if (exerciseId.value == null || pending.value) return null;
    return findExerciseGroupByExerciseId(log.value, exerciseId.value);
});

const dayId = computed(() => data.value?.day.ID ?? 0);

const session = useExerciseLoggingSession({
    exerciseGroup,
    pending,
    dayId,
    offset,
    logExercise,
    deleteLoggedSet,
    router,
});

const cuesExercise = computed(
    () =>
        session.exerciseGroup?.planned ??
        session.exerciseGroup?.logged?.exercise ??
        null,
);

const pendingCuesFromSave = ref<string | null>(null);

const cuesTextFromSession = computed(() => {
    const raw = cuesExercise.value?.cues ?? "";
    return typeof raw === "string" ? raw.trim() : "";
});

const cuesText = computed(
    () => pendingCuesFromSave.value ?? cuesTextFromSession.value,
);

const repRolloverHintDismissed = ref(false);

watch(exerciseId, () => {
    pendingCuesFromSave.value = null;
    repRolloverHintDismissed.value = false;
});

watch(cuesTextFromSession, (fromSession) => {
    if (pendingCuesFromSave.value === fromSession) {
        pendingCuesFromSave.value = null;
    }
});

type PerformanceSummaryItem = {
    label: string;
    text: string;
    date?: string;
};

const formatLoggedSets = (
    sets: readonly { weight: number; reps: number }[] | undefined,
) => (sets ?? []).map((set) => `${set.weight}x${set.reps}`).join(", ");

const formatPerformanceDate = (iso: string | undefined) => {
    if (!iso) return undefined;
    return formatDate(iso.includes("T") ? iso : `${iso}T12:00:00`);
};

const performanceDatesDiffer = (
    first: string | undefined,
    second: string | undefined,
) => Boolean(first && second && !isSameDay(first, second));

const loggedSetsMatch = (
    first: ExerciseGroup["previous"],
    second: ExerciseGroup["max"],
) => {
    if (!first || !second) return false;
    if (first.ID > 0 && second.ID > 0 && first.ID === second.ID) return true;
    const firstSets = first.sets ?? [];
    const secondSets = second.sets ?? [];
    if (firstSets.length === 0 || firstSets.length !== secondSets.length) {
        return false;
    }
    return firstSets.every((set, index) => {
        const other = secondSets[index];
        return (
            other != null && other.weight === set.weight && other.reps === set.reps
        );
    });
};

const performanceSummary = computed<PerformanceSummaryItem[]>(() => {
    const previous = session.exerciseGroup?.previous;
    const max = session.exerciseGroup?.max;
    const previousText = formatLoggedSets(previous?.sets);
    const maxText = formatLoggedSets(max?.sets);
    if (!previousText && !maxText) return [];
    if (previousText && maxText && loggedSetsMatch(previous, max)) {
        return [{ label: "Last & Max", text: previousText }];
    }
    const showMaxDate = performanceDatesDiffer(previous?.log_date, max?.log_date);
    return [
        previousText ? { label: "Last", text: previousText } : null,
        maxText
            ? {
                  label: "Max",
                  text: maxText,
                  date: showMaxDate
                      ? formatPerformanceDate(max?.log_date)
                      : undefined,
              }
            : null,
    ].filter((item): item is PerformanceSummaryItem => item !== null);
});

const previousSets = computed(() =>
    (session.exerciseGroup?.previous?.sets ?? []).map((set) => ({
        weight: set.weight,
        reps: set.reps,
    })),
);

const exerciseName = computed(
    () =>
        session.exerciseGroup?.planned?.name ||
        session.exerciseGroup?.logged?.exercise?.name ||
        "",
);

const exerciseLoadType = computed<ExerciseLoadType>(
    () =>
        (session.exerciseGroup?.planned?.load_type ??
            session.exerciseGroup?.logged?.exercise?.load_type ??
            DEFAULT_EXERCISE_LOAD_TYPE) as ExerciseLoadType,
);

const tracksWeightSetup = computed(
    () => exerciseLoadType.value !== "weight_stack",
);

const { isActive: timerActive, displayText: timerText } = useGlobalRestTimer();

const headerText = computed(() =>
    timerActive.value ? timerText.value : exerciseName.value,
);

const repRolloverWeightHintBase = computed(() => {
    const g = session.exerciseGroup;
    const repRollover = g?.previous?.exercise?.rep_rollover;
    if (typeof repRollover !== "number") return "";
    const currentWeight = session.currentWeight;
    const setsAtCurrentWeight = (g?.previous?.sets ?? [])
        .map((set) => ({
            weight: Number(set.weight),
            reps: Number(set.reps),
        }))
        .filter(
            (set) => set.weight === currentWeight && Number.isFinite(set.reps),
        );
    if (setsAtCurrentWeight.length === 0) return "";
    if (!setsAtCurrentWeight.every((set) => set.reps >= repRollover)) return "";
    const minPreviousReps = Math.min(
        ...setsAtCurrentWeight.map((set) => set.reps),
    );
    return `Previously did ${minPreviousReps} reps, consider increase the weight`;
});

const maxPreviousWorkoutWeight = computed(() => {
    const sets = session.exerciseGroup?.previous?.sets ?? [];
    const best = pickBestLoggedSet(sets);
    return best ? best.weight : null;
});

const repRolloverWeightHint = computed(() =>
    repRolloverHintDismissed.value ? "" : repRolloverWeightHintBase.value,
);

const repRolloverValue = computed(() => {
    const g = session.exerciseGroup;
    const rollover =
        g?.planned?.rep_rollover ??
        g?.logged?.exercise?.rep_rollover ??
        g?.previous?.exercise?.rep_rollover;
    return typeof rollover === "number" && Number.isFinite(rollover)
        ? rollover
        : undefined;
});

const logSetWithRepRolloverHintDismiss = async () => {
    const weightLogged = session.currentWeight;
    const hadHint = repRolloverWeightHintBase.value.length > 0;
    const maxPrev = maxPreviousWorkoutWeight.value;
    const ok = await session.addNextSet();
    if (ok && hadHint && maxPrev != null && weightLogged > maxPrev) {
        repRolloverHintDismissed.value = true;
    }
};

const openProgressionDialog = () => {
    const id = exerciseId.value;
    if (id == null) return;
    void dialogManager.custom<void>({
        title: `${exerciseName.value} Progression`,
        component: ExerciseProgressionChartPanel,
        componentProps: {
            exerciseId: id,
            lastTimeSets: previousSets.value,
        },
        wide: true,
    });
};

const weightSetupSummary = computed(() => {
    if (!tracksWeightSetup.value) return "";
    const text = session.currentWeightSetup.trim();
    if (!text) return "";
    const parsed = parseWeightSetup(text);
    if (!parsed.parsed) return text;
    const normalizedText = formatExerciseWeightSetup(
        exerciseLoadType.value,
        parsed.platesPerSide,
    );
    const total = computeExerciseLoadLbs(
        exerciseLoadType.value,
        parsed.platesPerSide,
    );
    return `${normalizedText || "No plates"} (${total} lbs)`;
});

const openWeightSetupDialog = async () => {
    const result = await dialogManager.custom<WeightSetupDialogResult>({
        title: "Weight setup",
        component: WeightSetupDialog,
        componentProps: {
            currentSetup: session.currentWeightSetup,
            targetWeightLbs: session.currentWeight,
            loadType: exerciseLoadType.value,
        },
        wide: true,
    });
    if (result == null) return;
    session.updateWeightSetup(result.weightSetup);
    if (result.syncWeight) {
        session.commitWeightFromInput(result.totalLbs);
    }
};

const editCues = async () => {
    const exercise = cuesExercise.value;
    if (!exercise) {
        toast.push("Exercise not found", "error");
        return;
    }

    const saved = await dialogManager.custom<string>({
        title: `Edit ${exercise.name} Cues`,
        component: EditCuesDialog,
        componentProps: {
            exercise,
            currentCues: cuesText.value,
        },
    });
    if (saved === null) return;
    pendingCuesFromSave.value = saved;
};
</script>

<template>
    <div class="logging-view">
        <LoggingHeader @back="session.goBackToList()">
            <template #center>
                <h2 class="logging-title m-0 min-w-0 text-center font-medium">
                    {{ headerText }}
                </h2>
            </template>
            <template #right>
                <span class="set-indicator"
                    >Set {{ session.currentSetNumber }}</span
                >
            </template>
        </LoggingHeader>
        <div class="input-group">
            <div class="flex flex-col gap-2">
                <NumberSlider
                    label="Weight (lbs)"
                    :model-value="session.currentWeight"
                    :hint="repRolloverWeightHint"
                    :step="2.5"
                    :max="500"
                    @update:model-value="session.commitWeightFromInput"
                />
                <NumberSlider
                    label="Reps"
                    :model-value="session.currentReps"
                    :marker-value="repRolloverValue"
                    marker-label="Rollover"
                    integer-only
                    :step="1"
                    :max="50"
                    @update:model-value="session.commitRepsFromInput"
                />
            </div>
            <div v-if="tracksWeightSetup" class="input-container">
                <label>Weight setup</label>
                <button
                    type="button"
                    class="weight-setup-trigger"
                    @click="openWeightSetupDialog()"
                >
                    <span
                        class="min-w-0 truncate"
                        :class="{
                            'text-[rgb(120,120,120)]':
                                !session.currentWeightSetup.trim(),
                        }"
                        >{{
                            session.currentWeightSetup.trim()
                                ? weightSetupSummary ||
                                  session.currentWeightSetup
                                : "Set plates per side…"
                        }}</span
                    >
                </button>
            </div>
        </div>
        <div class="flex flex-row gap-2">
            <button
                class="bg-[#2c2c2c] hover:bg-[#525252] flex-1"
                type="button"
                @click="logSetWithRepRolloverHintDismiss()"
            >
                <span>Log Set</span>
            </button>
            <button
                class="bg-green-500 hover:bg-green-600 flex-1"
                type="button"
                @click="session.finishLogging()"
            >
                <span>Finish Exercise</span>
            </button>
        </div>
        <div v-if="cuesExercise" class="exercise-cues-wrap">
            <template v-if="cuesText">
                <span class="exercise-cues-label exercise-cues-label--tight"
                    >Cues</span
                >
                <button
                    type="button"
                    class="exercise-cues-edit exercise-cues-edit--overlay hover:bg-secondBg rounded-md"
                    @click="editCues()"
                >
                    <Pencil :size="16" class="my-0.5" />
                </button>
                <p class="exercise-cues">{{ cuesText }}</p>
            </template>
            <div v-else class="exercise-cues-heading">
                <span class="exercise-cues-label">Cues</span>
                <button
                    type="button"
                    class="exercise-cues-edit hover:bg-secondBg rounded-md"
                    @click="editCues()"
                >
                    <Pencil :size="16" class="my-0.5" />
                </button>
            </div>
        </div>
        <LoggedSetsList
            :logged-sets="session.loggedSets"
            @retry="session.retrySet"
            @delete="session.deleteSet"
            @edit="session.editSet"
        />
        <button
            v-if="performanceSummary.length > 0"
            type="button"
            class="previous-performance cursor-pointer text-left transition-colors hover:border-[rgb(80,80,80)] hover:bg-[rgb(32,32,32)]"
            @click="openProgressionDialog()"
        >
            <span
                v-for="item in performanceSummary"
                :key="item.label"
                class="flex flex-col gap-1"
            >
                <span class="previous-performance-label">
                    {{ item.label }}<span
                        v-if="item.date"
                        class="font-normal normal-case tracking-normal text-[rgb(100,100,100)]"
                    >
                        · {{ item.date }}</span
                    >
                </span>
                <span class="previous-performance-value">{{ item.text }}</span>
            </span>
        </button>
        <div class="input-container">
            <label>Notes</label>
            <textarea
                class="notes-input"
                :value="session.notes"
                placeholder="Add any notes about this set..."
                rows="3"
                @input="
                    session.updateNotes(
                        ($event.target as HTMLTextAreaElement).value,
                    )
                "
            ></textarea>
        </div>
    </div>
</template>

<style scoped>
.logging-view {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    width: 100%;
    max-width: 100%;
    padding: 0 0.75rem 0 0.375rem;
    box-sizing: border-box;
}

.logging-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: clamp(0.75rem, 0.55rem + 2.75cqi, 1.125rem);
}

.set-indicator {
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
    text-align: right;
    grid-column: 3;
    white-space: nowrap;
    flex-shrink: 0;
}

.previous-performance {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    width: 100%;
    padding: 0.875rem 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 0.5rem;
    background: rgb(24, 24, 24);
    color: inherit;
    font: inherit;
}

.previous-performance-label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgb(130, 130, 130);
}

.previous-performance-value {
    font-size: 1rem;
    color: rgb(220, 220, 220);
}

.exercise-cues-wrap {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.75rem 0.25rem 0;
    border-bottom: 1px solid rgb(56, 56, 56);
    margin-bottom: -0.25rem;
}

.exercise-cues-heading {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    padding-bottom: 0.5rem;
}

.exercise-cues-label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgb(130, 130, 130);
}

.exercise-cues-label--tight {
    padding-right: 2.75rem;
}

.exercise-cues-edit {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}

.exercise-cues-edit--overlay {
    position: absolute;
    top: 0.75rem;
    right: 0.25rem;
}

.exercise-cues {
    margin: 0;
    font-size: 0.95rem;
    line-height: 1.45;
    color: rgb(210, 210, 210);
    padding-bottom: 0.5rem;
    white-space: pre-wrap;
}

.input-group {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
}

.input-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.input-container label {
    font-weight: 500;
    font-size: 0.9rem;
    color: rgb(150, 150, 150);
}

.weight-setup-trigger {
    display: flex;
    align-items: center;
    width: 100%;
    min-height: 3rem;
    padding: 0 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    text-align: left;
    cursor: pointer;
    transition:
        border-color 0.2s,
        background-color 0.2s;
}

.weight-setup-trigger:hover {
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.notes-input {
    padding: 0.75rem 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
    transition:
        border-color 0.2s,
        background-color 0.2s;
}

.notes-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.notes-input::placeholder {
    color: rgb(100, 100, 100);
}

.button-group {
    display: flex;
    flex-direction: row;
    gap: 1rem;
}

.button-group button,
.finish-button {
    flex: 1;
    padding: 0.75rem 1.5rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.button-group button:hover,
.finish-button:hover {
    background: rgb(40, 40, 40);
}

.next-set-button {
    background: rgb(40, 80, 40) !important;
}

.next-set-button:hover {
    background: rgb(50, 100, 50) !important;
}

.finish-button {
    background: rgb(80, 80, 40) !important;
}

.finish-button:hover {
    background: rgb(100, 100, 50) !important;
}

@media (max-width: 767px) {
    .logging-view {
        padding: 0 1rem 0 0.5rem;
    }

    .logging-title {
        font-size: clamp(0.8125rem, 0.58rem + 3cqi, 1.25rem);
    }

    .set-indicator {
        font-size: 0.85rem;
    }

    .previous-performance {
        padding: 0.75rem 0.875rem;
    }

    .button-group {
        flex-direction: column;
    }
}
</style>
