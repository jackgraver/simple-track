<script setup lang="ts">
import type { ExerciseGroup } from "../../store/useWorkoutStore";
import { useWorkoutStore } from "../../store/useWorkoutStore";
import { useExerciseLoggingSession } from "../composables/useExerciseLoggingSession";
import { useLoggingRouteContext } from "../../composables/useLoggingRouteContext";
import {
    findExerciseGroupByExerciseId,
    parseExerciseIdParam,
} from "../domain/exerciseRouteGroup";
import LoggingHeader from "./LoggingHeader.vue";
import LoggedSetsList from "./LoggedSetsList.vue";
import NumericStepper from "./NumericStepper.vue";
import { useGlobalRestTimer } from "~/composables/useGlobalRestTimer";
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

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

const cuesText = computed(() => {
    const g = session.exerciseGroup;
    if (!g) return "";
    const raw = g.planned?.cues ?? g.logged?.exercise?.cues ?? "";
    return typeof raw === "string" ? raw.trim() : "";
});

const previousPerformanceText = computed(() => {
    const sets = session.exerciseGroup?.previous?.sets ?? [];
    if (sets.length === 0) return "";

    return sets.map((set) => `${set.weight}x${set.reps}`).join(", ");
});

const exerciseName = computed(
    () =>
        session.exerciseGroup?.planned?.name ||
        session.exerciseGroup?.logged?.exercise?.name ||
        "",
);

const { isActive: timerActive, displayText: timerText } = useGlobalRestTimer();

const headerText = computed(() =>
    timerActive.value ? timerText.value : exerciseName.value,
);

const repRolloverWeightHint = computed(() => {
    const g = session.exerciseGroup;
    const repRollover = g?.previous?.exercise?.rep_rollover;
    if (typeof repRollover !== "number") return "";
    const previousReps = (g?.previous?.sets ?? [])
        .map((set) => Number(set.reps))
        .filter((reps) => Number.isFinite(reps));
    if (previousReps.length === 0) return "";
    const minPreviousReps = Math.min(...previousReps);
    if (minPreviousReps < repRollover) return "";
    return `Previously did ${minPreviousReps} reps, consider increase the weight`;
});
</script>

<template>
    <div class="logging-view">
        <LoggingHeader @back="session.goBackToList()">
            <template #center>
                <h2
                    class="logging-title m-0 min-w-0 truncate text-center text-lg font-medium"
                >
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
            <NumericStepper
                label="Weight (lbs)"
                :model-value="session.currentWeight"
                :hint="repRolloverWeightHint"
                input-step="0.5"
                :step-with="session.stepWeight"
                @update:model-value="session.commitWeightFromInput"
            />
            <NumericStepper
                label="Reps"
                :model-value="session.currentReps"
                integer-only
                :step-with="session.stepReps"
                @update:model-value="session.commitRepsFromInput"
            />
            <div class="input-container">
                <label>Weight Setup</label>
                <input
                    type="text"
                    class="weight-setup-input"
                    :value="session.currentWeightSetup"
                    placeholder="2x45 + 10"
                    @input="
                        session.updateWeightSetup(
                            ($event.target as HTMLInputElement).value,
                        )
                    "
                />
            </div>
        </div>
        <div class="flex flex-row gap-2">
            <button
                class="bg-[#2c2c2c] hover:bg-[#525252] flex-1"
                type="button"
                @click="session.addNextSet()"
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
        <div v-if="cuesText" class="exercise-cues-wrap">
            <span class="exercise-cues-label">Cues</span>
            <p class="exercise-cues">{{ cuesText }}</p>
        </div>
        <LoggedSetsList
            :logged-sets="session.loggedSets"
            @retry="session.retrySet"
            @delete="session.deleteSet"
            @edit="session.editSet"
        />
        <div v-if="previousPerformanceText" class="previous-performance">
            <span class="previous-performance-label">Last time</span>
            <span class="previous-performance-value">{{
                previousPerformanceText
            }}</span>
        </div>
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
    gap: 0.25rem;
    padding: 0.875rem 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 0.5rem;
    background: rgb(24, 24, 24);
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
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.75rem 0.25rem 0;
    border-bottom: 1px solid rgb(56, 56, 56);
    margin-bottom: -0.25rem;
}

.exercise-cues-label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgb(130, 130, 130);
}

.exercise-cues {
    margin: 0;
    font-size: 0.95rem;
    line-height: 1.45;
    color: rgb(210, 210, 210);
    white-space: pre-wrap;
}

.input-group {
    display: flex;
    flex-direction: column;
    gap: 2rem;
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

.weight-setup-input {
    height: 3rem;
    padding: 0 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1rem;
    transition:
        border-color 0.2s,
        background-color 0.2s;
}

.weight-setup-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}

.weight-setup-input::placeholder {
    color: rgb(100, 100, 100);
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
        font-size: 1.25rem;
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
