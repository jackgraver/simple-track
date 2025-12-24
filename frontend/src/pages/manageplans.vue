<script setup lang="ts">
import type { WorkoutPlan, Exercise } from "~/types/workout";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import AddExerciseDialog from "~/shared/AddExerciseDialog.vue";
import { X, Plus } from "lucide-vue-next";
import { useAPIGet, useAPIPost } from "~/composables/useApiFetch";
import { computed } from "vue";

const { data, refresh } = useAPIGet<{ plans: WorkoutPlan[] }>("workout/plans/all");

const plans = computed(() => data.value?.plans || []);

const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

const getDayName = (dayOfWeek: number | null): string | undefined => {
    if (dayOfWeek === null) return "Unassigned";
    return dayNames[dayOfWeek];
};

const getAssignedDays = computed(() => {
    const assigned: number[] = [];
    plans.value.forEach(plan => {
        if (plan.day_of_week !== null) {
            assigned.push(plan.day_of_week);
        }
    });
    return assigned;
});

const isDayAssigned = (dayOfWeek: number, currentPlan: WorkoutPlan): boolean => {
    return getAssignedDays.value.includes(dayOfWeek) && 
           plans.value.find(p => p.day_of_week === dayOfWeek)?.ID !== currentPlan.ID;
};

const getAssignedPlanName = (dayOfWeek: number, currentPlan: WorkoutPlan): string => {
    const assignedPlan = plans.value.find(p => p.day_of_week === dayOfWeek && p.ID !== currentPlan.ID);
    return assignedPlan?.name || 'Assigned';
};

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
        .custom<boolean>({
            title: `Add Exercise to ${plan.name}`,
            component: AddExerciseDialog,
            props: {
                plan: plan,
            },
        })
        .then((result) => {
            if (result !== null) {
                refresh();
            }
        });
};

const handleDayChange = async (plan: WorkoutPlan, event: Event) => {
    const select = event.target as HTMLSelectElement;
    const day = parseInt(select.value);
    const previousValue = plan.day_of_week !== null ? plan.day_of_week.toString() : '';
    
    if (isNaN(day)) {
        select.value = previousValue;
        return;
    }

    const currentlyAssignedPlan = plans.value.find(p => p.day_of_week === day && p.ID !== plan.ID);
    
    if (currentlyAssignedPlan) {
        const confirmed = await dialogManager.confirm({
            title: 'Day Already Assigned',
            message: dayNames[day] + ' is currently assigned to "' + currentlyAssignedPlan.name + '". Assigning "' + plan.name + '" will unassign "' + currentlyAssignedPlan.name + '". Continue?',
            confirmText: 'Yes, Reassign',
            cancelText: 'Cancel',
        });
        
        if (!confirmed) {
            select.value = previousValue;
            return;
        }
    }
    
    await assignPlanToDay(plan, day);
};

const assignPlanToDay = async (plan: WorkoutPlan, dayOfWeek: number) => {
    const { response, error } = await useAPIPost<{ plan: WorkoutPlan }>(
        `workout/plans/${plan.ID}/assign-day`,
        "POST",
        { day_of_week: dayOfWeek },
    );

    if (error) {
        toast.push("Failed to assign day: " + error.message, "error");
        return;
    }

    toast.push(`Assigned ${plan.name} to ${dayNames[dayOfWeek]}`, "success");
    await refresh();
};

const unassignPlanFromDay = async (plan: WorkoutPlan) => {
    const { response, error } = await useAPIPost<{ plan: WorkoutPlan }>(
        `workout/plans/${plan.ID}/assign-day`,
        "DELETE",
        {},
    );

    if (error) {
        toast.push("Failed to unassign day: " + error.message, "error");
        return;
    }

    toast.push(`Unassigned ${plan.name} from ${getDayName(plan.day_of_week)}`, "success");
    await refresh();
};
</script>

<template>
    <div class="manage-plans-container">
        <h1>Manage Workout Plans</h1>
        <div v-if="!data" class="loading">Loading...</div>
        <div v-else class="plans-list">
            <div v-for="plan in plans" :key="plan.ID" class="plan-card">
                <div class="plan-header">
                    <div class="plan-title-section">
                        <h2>{{ plan.name }}</h2>
                        <div class="day-assignment">
                            <span v-if="plan.day_of_week !== null" class="day-badge">
                                {{ getDayName(plan.day_of_week) }}
                            </span>
                            <span v-else class="day-badge unassigned">Unassigned</span>
                        </div>
                    </div>
                    <button @click="openAddDialog(plan)" class="add-button">
                        <Plus :size="18" />
                        Add Exercise
                    </button>
                </div>
                <div class="day-selector-section">
                    <label>Assign to Day:</label>
                    <div class="day-selector">
                        <select 
                            :value="plan.day_of_week !== null ? plan.day_of_week : ''"
                            @change="handleDayChange(plan, $event)"
                            class="day-select"
                        >
                            <option value="">-- Select Day --</option>
                            <option 
                                v-for="(dayName, index) in dayNames" 
                                :key="index"
                                :value="index"
                            >
                                {{ dayName }}{{ isDayAssigned(index, plan) ? ` (${getAssignedPlanName(index, plan)})` : '' }}
                            </option>
                        </select>
                        <button 
                            v-if="plan.day_of_week !== null"
                            @click="unassignPlanFromDay(plan)"
                            class="unassign-button"
                        >
                            Clear
                        </button>
                    </div>
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

.plan-title-section {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.plan-header h2 {
    margin: 0;
    font-size: 1.5rem;
}

.day-assignment {
    display: flex;
    align-items: center;
}

.day-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    background: #4a9eff;
    color: white;
    border-radius: 0.25rem;
    font-size: 0.85rem;
    font-weight: 500;
}

.day-badge.unassigned {
    background: #666;
}

.day-selector-section {
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #333;
}

.day-selector-section label {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: #aaa;
}

.day-selector {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.day-select {
    flex: 1;
    padding: 0.5rem;
    background: #2a2a2a;
    color: white;
    border: 1px solid #444;
    border-radius: 0.25rem;
    font-size: 0.9rem;
    cursor: pointer;
}

.day-select:focus {
    outline: none;
    border-color: #4a9eff;
}

.day-select option:disabled {
    color: #666;
    font-style: italic;
}

.unassign-button {
    padding: 0.5rem 1rem;
    background: #666;
    color: white;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.2s;
}

.unassign-button:hover {
    background: #777;
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

