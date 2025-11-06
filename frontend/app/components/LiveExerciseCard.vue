<script setup lang="ts">
import type { Exercise, LoggedExercise, LoggedSet } from "~/types/workout";
import { Plus, NotebookPen, Loader, Check, Logs } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import ConfirmWeightSetupDialog from "./ConfirmWeightSetupDialog.vue";

type ExerciseGroup = {
    planned: Exercise;
    logged: LoggedExercise;
    previous: LoggedExercise;
};

const props = defineProps<{
    data: ExerciseGroup;
    addSet: (exercise: LoggedExercise) => void;
    logExercise: (
        exercise: LoggedExercise,
        type: "logged" | "previous",
    ) => Promise<boolean>;
    removeSet: (exercise: LoggedExercise, set: LoggedSet) => void;
}>();

const exercise = ref<LoggedExercise>(props.data.logged || props.data.previous);
const type = computed<"logged" | "previous">(() =>
    props.data.logged ? "logged" : "previous",
);

const logStatus = ref<"pending" | "logged" | "not-logged">(
    type.value === "logged" ? "logged" : "not-logged",
);

const initialWeight = computed(() => {
    if (exercise.value?.sets[0]?.weight) {
        return exercise.value?.sets[0]?.weight;
    }
    return 0;
});

watch(exercise.value?.sets, () => {
    logStatus.value = "not-logged";
});

const innerAddSet = (exercise: LoggedExercise) => {
    logStatus.value = "not-logged";
    props.addSet(exercise);
};

const innerLogExercise = async (exercise: LoggedExercise) => {
    logStatus.value = "pending";

    for (const set of exercise.sets) {
        if (set.weight !== initialWeight.value) {
            const res = await dialogManager.custom<string>({
                title: "Update Weight Setup",
                component: ConfirmWeightSetupDialog,
                props: {
                    weightSetup: exercise.weight_setup,
                },
            });

            if (res === "cancel") {
                logStatus.value = "not-logged";
                return;
            }

            exercise.weight_setup = res;
            break;
        }
    }

    props.logExercise(exercise, type.value).then((res) => {
        logStatus.value = res ? "logged" : "not-logged";
    });
};

const checkIfZerod = (set: LoggedSet) => {
    if (set.reps === 0 && set.weight === 0)
        props.removeSet(exercise.value, set);
};
</script>
<template>
    <article>
        <header>
            <h1>{{ exercise.exercise.name }}</h1>
            <span>{{ exercise.weight_setup }}</span>
        </header>
        <template v-for="(set, i) in exercise.sets">
            <div class="set">
                <div class="set-input">
                    <label v-if="i === 0">Weight</label>
                    <input
                        type="number"
                        v-model="set.weight"
                        @change="() => checkIfZerod(set)"
                    />
                </div>
                <div class="set-input">
                    <label v-if="i === 0">Reps</label>
                    <input
                        type="number"
                        v-model="set.reps"
                        @change="() => checkIfZerod(set)"
                    />
                </div>
            </div>
        </template>
        <footer>
            <button @click="innerAddSet(exercise)"><Plus /></button>
            <button
                class="confirm-button"
                @click="innerLogExercise(exercise)"
                :disabled="logStatus === 'logged'"
            >
                <NotebookPen v-if="logStatus === 'not-logged'" />
                <Loader class="spinner" v-if="logStatus === 'pending'" />
                <Check v-if="logStatus === 'logged'" />
            </button>
        </footer>
    </article>
</template>

<style scoped>
article {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 95%;
    border: 1px solid rgb(56, 56, 56);
    border-radius: 5px;
    background: rgb(27, 27, 27);
    padding: 1rem;
}

article header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

article footer {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    align-items: center;
}

article footer button {
    width: 100%;
}

article h1 {
    margin: 0;
}

.set {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
}

.set-input {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    justify-content: center;
}

.set-input input {
    width: 100%;
    min-width: 0;
    box-sizing: border-box;
}

.spinner {
    animation-name: spin;
    animation-duration: 3000ms;
    animation-iteration-count: infinite;
    animation-timing-function: linear;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}
</style>
