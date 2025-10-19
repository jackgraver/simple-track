<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
import TodayLogEditedDialog from "~/components/today/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import LogOtherMealDialog from "../LogOtherMealDialog.vue";

const { data, pending, error } = useAPIGet<{
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
}>(`mealplan/today`);

const logPlannedMeal = async (meal: Meal) => {
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
    }>(`mealplan/meal/log-planned`, "POST", {
        meal_id: meal.ID,
    });

    if (error) {
        toast.push("Planned Meal Log Failed!", "error");
    } else if (response) {
        toast.push("Planned Meal Log Successfully!", "success");
        if (data.value) {
            data.value = {
                day: response.day,
                totalCalories: response.totalCalories,
                totalProtein: response.totalProtein,
                totalFiber: response.totalFiber,
            };
        }
    }
};

const logEditedMeal = async (meal: Meal) => {
    dialogManager
        .custom<Meal>({
            title: "Log Edited Meal",
            component: TodayLogEditedDialog,
            props: { meal },
        })
        .then(async (editedMeal) => {
            if (editedMeal) {
                const { response, error } = await useAPIPost<{
                    day: Day;
                    totalCalories: number;
                    totalProtein: number;
                    totalFiber: number;
                }>("mealplan/meal/logedited", "POST", { meal: editedMeal });

                if (error)
                    toast.push("Log Edited Failed!" + error.message, "error");
                else if (response) {
                    toast.push("Planned Meal Log Successfully!", "success");
                    if (data.value) {
                        data.value = {
                            day: response.day,
                            totalCalories: response.totalCalories,
                            totalProtein: response.totalProtein,
                            totalFiber: response.totalFiber,
                        };
                    }
                }
            }
        })
        .catch((err) => {
            console.error("Dialog error:", err);
            toast.push("Dialog Error", "error");
        });
};

const deleteLoggedMeal = async (meal: Meal) => {
    console.log(meal);
    dialogManager
        .confirm({
            title: "Delete Logged Meal",
            message: "Are you sure you want to delete this meal?",
        })
        .then(async (confirmed) => {
            if (!confirmed) return;

            const { response, error } = await useAPIPost<{
                day: Day;
                totalCalories: number;
                totalProtein: number;
                totalFiber: number;
            }>(`mealplan/meal/logged`, "DELETE", {
                meal_id: meal.ID,
                day_id: data.value?.day.ID,
            });

            if (error) {
                toast.push("Delete Failed!", "error");
            } else if (response) {
                toast.push("Delete Successfully!", "success");
                if (data.value) {
                    data.value = {
                        day: response.day,
                        totalCalories: response.totalCalories,
                        totalProtein: response.totalProtein,
                        totalFiber: response.totalFiber,
                    };
                }
            }
        });
};

const editLogMeal = (meal: Meal) => {
    const oldMealID = meal.ID;
    dialogManager
        .custom<Meal>({
            title: "Log Edited Meal",
            component: TodayLogEditedDialog,
            props: { meal },
        })
        .then(async (editedMeal) => {
            console.log(editedMeal);
            const { response, error } = await useAPIPost<{
                day: Day;
                totalCalories: number;
                totalProtein: number;
                totalFiber: number;
            }>("mealplan/meal/editlogged", "POST", {
                meal: editedMeal,
                oldMealID: oldMealID,
            });

            if (error)
                toast.push("Log Edited Failed!" + error.message, "error");
            else if (response) {
                toast.push("Planned Meal Log Successfully!", "success");
                if (data.value) {
                    data.value = {
                        day: response.day,
                        totalCalories: response.totalCalories,
                        totalProtein: response.totalProtein,
                        totalFiber: response.totalFiber,
                    };
                }
            }
        })
        .catch((err) => {
            console.error("Dialog error:", err);
            toast.push("Dialog Error", "error");
        });
};

const logOtherMeal = async () => {
    dialogManager.custom<any>({
        title: "Log Other Meal",
        component: LogOtherMealDialog,
        props: { meal: null },
    });
};

const start = ref(0);

const visibleItems = computed(() =>
    data.value?.day.plannedMeals.slice(start.value, start.value + 2),
);

console.log(visibleItems);

function next() {
    console.log("next", start.value);
    start.value = Math.min(
        start.value + 1,
        (data?.value?.day?.plannedMeals?.length ?? 0) - 2,
    );
    console.log(start.value);
}
function prev() {
    start.value = Math.max(start.value - 1, 0);
}
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div v-if="data" style="width: 100%">
            <div class="title-row">
                <h1 style="flex: 1">
                    {{
                        formatDate(data.day.date) +
                        ", " +
                        dayOfWeek(data.day.date)
                    }}
                </h1>
                <!-- <NuxtLink class="link" to="/mealplan"
                    >Manage Meal Plan</NuxtLink
                > -->
                <button @click="logOtherMeal">Log Meal</button>
            </div>
            <TodayBars
                :totalCalories="data?.totalCalories ?? 0"
                :totalProtein="data?.totalProtein ?? 0"
                :totalFiber="data?.totalFiber ?? 0"
                :plannedCalories="data?.day.plan.calories ?? 0"
                :plannedProtein="data?.day.plan.protein ?? 0"
                :plannedFiber="data?.day.plan.fiber ?? 0"
            />
            <div class="meals-section">
                <div class="meals-container">
                    <h2>Logged</h2>
                    <TodayMealCard
                        v-for="log in data.day.loggedMeals"
                        :key="log.ID"
                        :meal="log.meal"
                        :on-log-planned="logPlannedMeal"
                        :on-log-edited="logEditedMeal"
                        :on-delete="deleteLoggedMeal"
                        :on-edit="editLogMeal"
                        type="logged"
                    />
                </div>
                <div class="meals-container">
                    <div class="small-title-row">
                        <h2>Planned</h2>
                        <button @click="prev">^</button>
                        <button @click="next">v</button>
                    </div>
                    <TodayMealCard
                        v-for="log in visibleItems"
                        :key="log.ID"
                        :meal="log.meal"
                        :on-log-planned="logPlannedMeal"
                        :on-log-edited="logEditedMeal"
                        :on-delete="deleteLoggedMeal"
                        :on-edit="editLogMeal"
                        type="planned"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    /* Fill the width provided by the page wrapper */
    width: 100%;
}

.title-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
}

.title-row button {
    margin-top: 6px;
    border-radius: 4px;
    font-size: large;
    padding: 6px 12px;
    font-weight: bold;
    text-decoration: none;
    font-size: large;
    padding: 6px 16px;
}

.small-title-row {
    display: flex;
    flex-direction: row;
    align-items: end;
}

.small-title-row h2 {
    flex: 1;
}

.meals-section {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    width: 100%;
}

.meals-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-top: 0.5rem;
    /* Split the meals section evenly (50/50) */
    flex: 1 1 0%;
}

.meals-container h2 {
    margin-bottom: 0;
}

.expected-header {
    display: flex;
    flex-direction: row;
}
</style>
