<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";
import { Check } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import TodayLogExerciseDialog from "./LogExerciseDialog.vue";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    exercise: LoggedExercise;
    planned: boolean;
}>();

const actualExercise = props.exercise.exercise;

const logExercise = async () => {
    dialogManager
        .custom({
            title: "Log " + actualExercise.name,
            component: TodayLogExerciseDialog,
            props: {
                exercise: props.exercise,
            },
        })
        .then((success) => {
            if (success) {
                toast.push("Log Exercise Successfully!", "success");
            } else {
                toast.push("Log Exercise Failed!", "error");
            }
        });
};
</script>

<template>
    <article class="workout-card">
        <header class="card-header">
            <h2>{{ actualExercise?.name }}</h2>
            <template v-if="!planned">
                {{
                    props.exercise.sets[props.exercise.sets.length - 1]
                        ?.weight ?? "X"
                }}
                x
                {{
                    props.exercise.sets[props.exercise.sets.length - 1]?.reps ??
                    "X"
                }}
            </template>
        </header>
        <template v-if="!planned">
            <span v-if="props.exercise.weight_setup">
                {{ props.exercise.weight_setup }}
            </span>
            <span
                v-if="
                    (props.exercise.sets[props.exercise.sets.length - 1]
                        ?.reps ?? 0) > props.exercise.exercise?.rep_rollover
                "
                class="info"
            >
                Up weight next session
            </span>
            <span v-if="props.exercise.percent_change" class="fact">
                {{ props.exercise.percent_change < 0 ? "Down" : "Up" }}
                {{ props.exercise.percent_change.toFixed(2) }}% since last
                workout
            </span>
        </template>
        <footer v-if="planned">
            <button @click="logExercise" class="check">
                <Check />
            </button>
        </footer>
    </article>
</template>

<style scoped>
.workout-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid #333;
    border-radius: 0.5rem;
    background: #1a1a1a;
    color: #fff;
    min-height: 120px;
}

.workout-card span {
    font-size: small;
}

.card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
}

.workout-card header h2 {
    margin: 0;
}

.workout-card footer {
    display: flex;
    justify-content: flex-end;
    margin-top: auto; /* pushes footer to bottom */
}

.card-header h3 {
    margin: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.workout-card .info {
    color: rgb(206, 206, 48);
}

.workout-card .fact {
    color: rgb(63, 197, 46);
}

.check {
    background-color: rgb(63, 197, 46);
    color: #fff;
    border-radius: 0.25rem;
    padding: 0.25rem 0.5rem;
}
</style>
