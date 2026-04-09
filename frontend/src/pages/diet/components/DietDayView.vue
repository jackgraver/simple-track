<script setup lang="ts">
import type { Meal } from "~/types/diet";
import LogEditedDialog from "~/pages/home/dialog/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { ChevronDown, ChevronUp } from "lucide-vue-next";
import { useRouter } from "vue-router";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import {
    useLogPlannedMeal,
    useDeleteLoggedMeal,
    useEditLoggedMeal,
} from "~/pages/home/queries/useMealMutations";
import { computed, ref } from "vue";
import MealCard from "./MealCard.vue";
import MacroBars from "~/pages/diet/components/MacroBars.vue";
import {
    EDIT_LOGGED_TYPE,
    EDIT_TYPE,
    LOG_TYPE,
} from "~/pages/diet/logmeal/logmealMode";

const router = useRouter();

const props = defineProps<{
    dateOffset: number;
}>();

const {
    data,
    isPending: pending,
    error,
} = useDietLogsToday(() => props.dateOffset);

const logPlannedMealMutation = useLogPlannedMeal(props.dateOffset);
const deleteLoggedMealMutation = useDeleteLoggedMeal(props.dateOffset);
const editLoggedMealMutation = useEditLoggedMeal(props.dateOffset);

const logPlannedMeal = async (meal: Meal) => {
    try {
        await logPlannedMealMutation.mutateAsync(meal.ID);
        toast.push("Planned Meal Log Successfully!", "success");
    } catch {
        toast.push("Planned Meal Log Failed!", "error");
    }
};

const logMeal = async (
    meal: Meal | null,
    type: typeof LOG_TYPE | typeof EDIT_TYPE | typeof EDIT_LOGGED_TYPE,
) => {
    if (type === LOG_TYPE) {
        router.push({ name: "diet-log", query: { type: LOG_TYPE } });
        return;
    }
    router.push({
        name: "diet-log",
        query: { type, id: String(meal?.ID ?? "") },
    });
};

const deleteLoggedMeal = async (meal: Meal) => {
    const confirmed = await dialogManager.confirm({
        title: "Delete Logged Meal",
        message: "Are you sure you want to delete this meal?",
    });

    if (!confirmed) return;

    if (!data.value?.day.ID) {
        toast.push("Cannot delete: day ID not found", "error");
        return;
    }

    try {
        await deleteLoggedMealMutation.mutateAsync({
            mealId: meal.ID,
            dayId: data.value.day.ID,
        });
        toast.push("Delete Successfully!", "success");
    } catch {
        toast.push("Delete Failed!", "error");
    }
};

const editLogMeal = (meal: Meal) => {
    const oldMealID = meal.ID;
    dialogManager
        .custom<Meal>({
            title: "Log Edited Meal",
            component: LogEditedDialog,
            props: { meal },
        })
        .then(async (editedMeal) => {
            if (!editedMeal) return;
            try {
                await editLoggedMealMutation.mutateAsync({
                    meal: editedMeal,
                    oldMealId: oldMealID,
                });
                toast.push("Meal Edited Successfully!", "success");
            } catch (error: unknown) {
                const msg =
                    error instanceof Error ? error.message : "Unknown error";
                toast.push("Log Edited Failed! " + msg, "error");
            }
        })
        .catch((err) => {
            console.error("Dialog error:", err);
            toast.push("Dialog Error", "error");
        });
};

const start = ref(0);

const visibleItems = computed(() =>
    data.value?.day.plannedMeals.slice(start.value, start.value + 2),
);

function next() {
    start.value = Math.min(
        start.value + 1,
        (data?.value?.day?.plannedMeals?.length ?? 0) - 2,
    );
}
function prev() {
    start.value = Math.max(start.value - 1, 0);
}
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="flex w-full flex-col gap-4">
        <div v-if="data" class="w-full">
            <MacroBars
                :totalCalories="data?.totalCalories ?? 0"
                :totalProtein="data?.totalProtein ?? 0"
                :totalFiber="data?.totalFiber ?? 0"
                :totalCarbs="data?.totalCarbs ?? 0"
                :plannedCalories="data?.day.plan.calories ?? 0"
                :plannedProtein="data?.day.plan.protein ?? 0"
                :plannedFiber="data?.day.plan.fiber ?? 0"
                :plannedCarbs="data?.day.plan.carbs ?? 0"
            />
            <div class="flex w-full flex-col gap-2 sm:flex-row sm:gap-2">
                <div class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
                    <h2 class="mb-0 text-lg font-semibold">Logged</h2>
                    <span
                        v-if="data.day.loggedMeals.length === 0"
                        class="text-zinc-500"
                        >Nothing logged yet.</span
                    >
                    <MealCard
                        v-for="log in data.day.loggedMeals"
                        :key="log.ID"
                        :meal="log.meal"
                        :on-log-planned="logPlannedMeal"
                        :on-log-edited="logMeal"
                        :on-delete="deleteLoggedMeal"
                        :on-edit="editLogMeal"
                        type="logged"
                    />
                </div>
                <div class="flex min-w-0 flex-1 flex-col gap-2 pt-2">
                    <div class="flex flex-row items-end gap-2">
                        <h2 class="mb-0 flex-1 text-lg font-semibold">
                            Planned
                        </h2>
                        <template v-if="data.day.plannedMeals.length > 2">
                            <button
                                type="button"
                                aria-label="Previous"
                                @click="prev"
                            >
                                <ChevronUp />
                            </button>
                            <button
                                type="button"
                                aria-label="Next"
                                @click="next"
                            >
                                <ChevronDown />
                            </button>
                        </template>
                    </div>
                    <span
                        v-if="visibleItems?.length === 0"
                        class="text-zinc-500"
                        >Nothing else planned.</span
                    >
                    <MealCard
                        v-for="log in visibleItems"
                        :key="log.ID"
                        :meal="log.meal"
                        :on-log-planned="logPlannedMeal"
                        :on-log-edited="logMeal"
                        :on-delete="deleteLoggedMeal"
                        :on-edit="editLogMeal"
                        type="planned"
                    />
                </div>
            </div>
        </div>
    </div>
</template>
