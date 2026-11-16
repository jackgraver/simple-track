<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from "vue";
import { Minus, Plus } from "lucide-vue-next";

export type StepDirection = "plus" | "minus";

const props = withDefaults(
    defineProps<{
        label: string;
        hint?: string;
        inputStep?: string;
        /** Reps-style: only whole numbers commit. */
        integerOnly?: boolean;
        /** Cardio-style: finite values are rounded on commit. */
        roundToInteger?: boolean;
        /** Shown with commit errors (e.g. save validation). */
        error?: string;
        /**
         * Override +/- behavior; parent is responsible for updating the bound value.
         * If omitted, +/- adjust by 1 (minus clamps at 0).
         */
        stepWith?: (direction: StepDirection) => void;
    }>(),
    { integerOnly: false, roundToInteger: false },
);

const model = defineModel<number>({ required: true });

const editMode = ref(false);
const draft = ref("");
const commitError = ref("");

const inputId = useId();

const displayError = computed(() => props.error || commitError.value);

watch(
    () => model.value,
    () => {
        editMode.value = false;
    },
);

const clearCommitError = () => {
    commitError.value = "";
};

const enterEdit = () => {
    clearCommitError();
    draft.value = model.value.toString();
    editMode.value = true;
    nextTick(() => {
        const input = document.getElementById(
            inputId,
        ) as HTMLInputElement | null;
        if (input) {
            input.focus();
            input.select();
        }
    });
};

const exitEdit = () => {
    const trimmed = draft.value.trim();
    clearCommitError();
    if (trimmed === "") {
        editMode.value = false;
        return;
    }
    const n = Number(trimmed);
    if (!Number.isFinite(n) || n < 0) {
        if (props.roundToInteger) {
            commitError.value = "Enter a valid number of minutes.";
        }
        editMode.value = false;
        return;
    }
    if (props.integerOnly && !Number.isInteger(n)) {
        editMode.value = false;
        return;
    }
    const next = props.roundToInteger ? Math.round(n) : n;
    model.value = next;
    editMode.value = false;
};

const onInput = (e: Event) => {
    draft.value = (e.target as HTMLInputElement).value;
};

const applyDefaultStep = (direction: StepDirection) => {
    const v = model.value || 0;
    if (direction === "plus") {
        model.value = v + 1;
    } else {
        model.value = Math.max(0, v - 1);
    }
};

const onIncrement = () => {
    clearCommitError();
    if (props.stepWith) props.stepWith("plus");
    else applyDefaultStep("plus");
};

const onDecrement = () => {
    clearCommitError();
    if (props.stepWith) props.stepWith("minus");
    else applyDefaultStep("minus");
};
</script>

<template>
    <div class="stepper-container">
        <label :for="inputId">{{ label }}</label>
        <div class="stepper">
            <button
                v-if="!editMode"
                class="stepper-button"
                type="button"
                @click="onDecrement"
            >
                <Minus :size="20" />
            </button>
            <div v-if="!editMode" class="stepper-display" @click="enterEdit">
                {{ model || 0 }}
            </div>
            <input
                v-else
                :id="inputId"
                type="number"
                class="stepper-input"
                :value="draft"
                min="0"
                :step="inputStep ?? '1'"
                @input="onInput"
                @blur="exitEdit"
                @keyup.enter="exitEdit"
                @keyup.escape="exitEdit"
            />
            <button
                v-if="!editMode"
                class="stepper-button"
                type="button"
                @click="onIncrement"
            >
                <Plus :size="20" />
            </button>
        </div>
        <p
            v-if="hint"
            class="mt-1 mb-0 text-[0.85rem] leading-snug text-amber-200/90"
        >
            {{ hint }}
        </p>
        <p
            v-if="displayError"
            class="mt-1 mb-0 text-[0.85rem] leading-snug text-red-400"
        >
            {{ displayError }}
        </p>
    </div>
</template>

<style scoped>
.stepper-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.stepper-container label {
    font-weight: 500;
    font-size: 0.9rem;
    color: rgb(150, 150, 150);
}

.stepper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    touch-action: manipulation;
}

.stepper-button {
    width: 3rem;
    height: 3rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 1.5rem;
    font-weight: 300;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
        background-color 0.2s,
        border-color 0.2s;
    user-select: none;
}

.stepper-button:hover {
    background: rgb(40, 40, 40);
    border-color: rgb(100, 100, 100);
}

.stepper-button:active {
    background: rgb(50, 50, 50);
}

.stepper-display {
    min-width: 6rem;
    height: 3rem;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
    font-weight: 500;
    cursor: pointer;
    user-select: none;
    padding: 0 1rem;
    border-radius: 5px;
    transition: background-color 0.2s;
}

.stepper-display:hover {
    background: rgb(30, 30, 30);
}

.stepper-input {
    min-width: 6rem;
    height: 3rem;
    padding: 0 1rem;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    color: inherit;
    font-size: 2rem;
    font-weight: 500;
    text-align: center;
}

.stepper-input:focus {
    outline: none;
    border-color: rgb(100, 100, 100);
    background: rgb(35, 35, 35);
}
</style>
