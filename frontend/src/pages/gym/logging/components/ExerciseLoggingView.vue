<script setup lang="ts">
import type { ExerciseLoggingSessionViewModel } from "../composables/useExerciseLoggingSession";
import { ArrowLeft } from "lucide-vue-next";
import LoggedSetsList from "./LoggedSetsList.vue";
import NumericStepper from "./NumericStepper.vue";

defineProps<{
    session: ExerciseLoggingSessionViewModel;
}>();
</script>

<template>
    <div class="logging-view">
        <div class="logging-header">
            <button
                class="back-button"
                type="button"
                @click="session.goBackToList()"
            >
                <ArrowLeft :size="20" />
            </button>
            <h2>
                {{
                    session.exerciseGroup?.planned?.name ||
                    session.exerciseGroup?.logged?.exercise?.name
                }}
            </h2>
            <span class="set-indicator">Set {{ session.currentSetNumber }}</span>
        </div>
        <LoggedSetsList
            :logged-sets="session.loggedSets"
            @retry="session.retrySet"
            @delete="session.deleteSet"
            @edit="session.editSet"
        />
        <div class="input-group">
            <NumericStepper
                label="Weight (lbs)"
                :model-value="session.currentWeight"
                :edit-mode="session.weightEditMode"
                :input-value="session.weightInputValue"
                input-step="0.5"
                @increment="session.incrementWeight"
                @decrement="session.decrementWeight"
                @update:input-value="session.updateWeightInputValue"
                @enter-edit="session.enterWeightEditMode"
                @exit-edit="session.exitWeightEditMode"
            />
            <NumericStepper
                label="Reps"
                :model-value="session.currentReps"
                :edit-mode="session.repsEditMode"
                :input-value="session.repsInputValue"
                @increment="session.incrementReps"
                @decrement="session.decrementReps"
                @update:input-value="session.updateRepsInputValue"
                @enter-edit="session.enterRepsEditMode"
                @exit-edit="session.exitRepsEditMode"
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
        <div class="button-group">
            <button class="next-set-button" type="button" @click="session.addNextSet()">
                <span>Next Set</span>
            </button>
            <button class="finish-button" type="button" @click="session.finishLogging()">
                <span>Finish</span>
            </button>
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

.logging-header {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 0.5rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgb(56, 56, 56);
    width: 100%;
    box-sizing: border-box;
}

.back-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    height: 2.5rem;
    background: transparent;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 0.25rem;
    color: inherit;
    cursor: pointer;
    transition:
        background-color 0.2s,
        border-color 0.2s;
    padding: 0;
}

.back-button:hover {
    background: rgb(40, 40, 40);
    border-color: rgb(100, 100, 100);
}

.logging-header h2 {
    margin: 0;
    text-align: center;
    grid-column: 2;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
}

.set-indicator {
    color: rgb(150, 150, 150);
    font-size: 0.9rem;
    text-align: right;
    grid-column: 3;
    white-space: nowrap;
    flex-shrink: 0;
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

    .logging-header {
        gap: 0.75rem;
    }

    .logging-header h2 {
        font-size: 1.25rem;
    }

    .set-indicator {
        font-size: 0.85rem;
    }

    .back-button {
        width: 2.25rem;
        height: 2.25rem;
    }

    .button-group {
        flex-direction: column;
    }
}
</style>
