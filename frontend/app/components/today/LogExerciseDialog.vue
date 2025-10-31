<script setup lang="ts">
import type { LoggedExercise, LoggedSet } from "~/types/workout";
import { Check, Plus } from "lucide-vue-next";
import { toast } from "~/composables/toast/useToast";

const props = defineProps<{
    exercise: LoggedExercise;
    onResolve?: (loggedExercise: LoggedExercise | null) => void;
}>();

props.exercise.ID = 0;

const log = ref<LoggedSet[]>(
    props.exercise.sets || [
        {
            weight: 0,
            reps: 0,
            logged_exercise_id: 0,
            ID: 0,
            created_at: "",
            updated_at: "",
        },
    ],
);

const weight = computed(() => {
    if (!log.value.length) return 0;
    return log.value.reduce(
        (max, set) => (set.weight > max ? set.weight : max),
        0,
    );
});

const addSet = () => {
    log.value.push({
        weight: weight.value,
        reps: 0,
        logged_exercise_id: 0,
        ID: 0,
        created_at: "",
        updated_at: "",
    });
};

const logExercise = async () => {
    console.log("log", props.exercise);
    const { response, error } = await useAPIPost<{
        exercise: LoggedExercise;
    }>(`workout/exercise/log`, "POST", {
        exercise: props.exercise,
    });

    if (error) {
        props.onResolve?.(null);
    } else if (response) {
        props.onResolve?.(response.exercise);
    }
};
</script>

<template>
    <div class="container">
        <div v-for="(set, index) in log" :key="index" class="input-set">
            <div class="item">
                <label for="weight">Weight</label>
                <input type="number" v-model.number="set.weight" />
            </div>
            <div class="item">
                <label for="reps">Reps</label>
                <input type="number" v-model.number="set.reps" />
            </div>
        </div>
        <button @click="addSet">
            <Plus />
        </button>
        <label for="weight_setup">Weight Setup</label>
        <input type="text" v-model="exercise.weight_setup" />
        <button @click="logExercise" class="check">
            <Check />
        </button>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem 0 1rem 0;
}

.input-set {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
}

.item {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 0.5rem;
}
</style>
