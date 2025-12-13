<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import AddExerciseDialog from "~/components/AddExerciseDialog.vue";
import { X, Plus } from "lucide-vue-next";

const { data, refresh } = useAPIGet<{ plans: WorkoutPlan[] }>("workout/plans/all");

const plans = computed(() => data.value?.plans || []);

const removeExerciseFromPlan = async (plan: WorkoutPlan, exercise: Exercise) => {
    const { response, error } = await useAPIPost(
        `workout/plans/${plan.ID}/exercises/remove`,
        "DELETE",
        { exercise_id: exercise.ID },
    );

    if (error) {
        toast.push("Failed to remove exercise: " + error.message, "error");
        return;
    }

    toast.push(`Removed ${exercise.name} from ${plan.name}`, "success");
    await refresh();
};

const openAddDialog = (plan: WorkoutPlan) => {
    dialogManager
        .custom({
            title: `Add Exercise to ${plan.name}`,
            component: AddExerciseDialog,
            props: {
                plan: plan,
            },
        })
        .then(() => {
            refresh();
        });
};
</script>

<template>
    <div class="manage-plans-container">
        <h1>Manage Workout Plans</h1>
        <div v-if="!data" class="loading">Loading...</div>
        <div v-else class="plans-list">
            <div v-for="plan in plans" :key="plan.ID" class="plan-card">
                <div class="plan-header">
                    <h2>{{ plan.name }}</h2>
                    <button @click="openAddDialog(plan)" class="add-button">
                        <Plus :size="18" />
                        Add Exercise
                    </button>
                </div>
                <div class="exercises-list">
                    <div
                        v-for="exercise in plan.exercises"
                        :key="exercise.ID"
                        class="exercise-item"
                    >
                        <span>{{ exercise.name }}</span>
                        <button
                            @click="removeExerciseFromPlan(plan, exercise)"
                            class="remove-button"
                        >
                            <X :size="16" />
                        </button>
                    </div>
                    <div v-if="plan.exercises.length === 0" class="empty-state">
                        No exercises in this plan
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.manage-plans-container {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
}

h1 {
    margin-bottom: 2rem;
    font-size: 2rem;
}

.loading {
    text-align: center;
    padding: 2rem;
}

.plans-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
}

.plan-card {
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 0.5rem;
    padding: 1.5rem;
}

.plan-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #333;
}

.plan-header h2 {
    margin: 0;
    font-size: 1.5rem;
}

.add-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: #4a9eff;
    color: white;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.2s;
}

.add-button:hover {
    background: #3a8eef;
}

.exercises-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.exercise-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background: #2a2a2a;
    border-radius: 0.25rem;
    border: 1px solid #444;
}

.remove-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    background: transparent;
    border: none;
    color: #ff6b6b;
    cursor: pointer;
    border-radius: 0.25rem;
    transition: background 0.2s;
}

.remove-button:hover {
    background: rgba(255, 107, 107, 0.1);
}

.empty-state {
    padding: 1rem;
    text-align: center;
    color: #888;
    font-style: italic;
}
</style>

