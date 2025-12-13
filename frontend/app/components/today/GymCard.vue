<script setup lang="ts">
import type { Exercise, LoggedExercise } from "~/types/workout";
import { Check } from "lucide-vue-next";
import { dialogManager } from "~/composables/dialog/useDialog";
import TodayLogExerciseDialog from "./LogExerciseDialog.vue";
import { toast } from "~/composables/toast/useToast";

const weightString = (log: LoggedExercise): string => {
    if (log.sets[0]?.weight) {
        return log.sets[0]?.weight + " lbs";
    }
    return "";
};

const props = defineProps<{
    exercise: LoggedExercise;
    previous: LoggedExercise | null;
    planned: boolean;
}>();

const localExercise = ref<LoggedExercise>(props.exercise);
const localPlanned = ref<boolean>(props.planned);

const actualExercise = computed<Exercise>(() => localExercise.value.exercise);

const logExercise = async () => {
    dialogManager
        .custom<LoggedExercise>({
            title: "Log " + actualExercise.value.name,
            component: TodayLogExerciseDialog,
            props: {
                exercise: localExercise,
                previousWeight: props.previous?.sets[0]?.weight,
            },
        })
        .then((loggedExercise) => {
            if (loggedExercise === null) {
                return;
            }
            if (loggedExercise) {
                localExercise.value = loggedExercise;
                localPlanned.value = false;
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
            <template v-if="previous">
                <h3>{{ weightString(previous) }}</h3>
            </template>
        </header>
        <div class="sets">
            <template v-if="localExercise.sets.length > 0">
                <span v-for="set in localExercise.sets" :key="set.ID">
                    {{ set.reps }} x {{ set.weight }}
                </span>
            </template>
            <template v-else-if="previous">
                <span v-for="set in previous.sets" :key="set.ID">
                    {{ set.reps }} x {{ set.weight }}
                </span>
            </template>
        </div>
        <template v-if="!localPlanned">
            <span v-if="localExercise.weight_setup">
                {{ localExercise.weight_setup }}
            </span>
            <span
                v-if="
                    (localExercise.sets[localExercise.sets.length - 1]?.reps ??
                        0) > localExercise.exercise?.rep_rollover
                "
                class="info"
            >
                Up weight next session
            </span>
            <span v-if="localExercise.percent_change" class="fact">
                {{ localExercise.percent_change < 0 ? "Down" : "Up" }}
                {{ localExercise.percent_change.toFixed(2) }}% since last
                workout
            </span>
        </template>
        <footer v-if="localPlanned">
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

.sets {
    display: flex;
    flex-direction: row;
    gap: 0.3rem;
}

.sets span:not(:last-child)::after {
    content: ",";
    margin-left: 0.05rem;
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
