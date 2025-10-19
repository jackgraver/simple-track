<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
import TodayLogEditedDialog from "~/components/today/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import LogOtherMealDialog from "../LogOtherMealDialog.vue";

function formatNum(n: number): string {
    const s = n.toFixed(2); // always 2 decimals
    return s.replace(/\.?0+$/, ""); // drop trailing zeros and optional dot
}

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
</script>

<template>
    <div v-if="pending">Loading...</div>
    <div v-else-if="error">Error: {{ error.message }}</div>
    <div v-else class="container">
        <div v-if="data">
            <div class="title-row">
                <h1 style="flex: 1">
                    {{
                        formatDate(data.day.date) +
                        ", " +
                        dayOfWeek(data.day.date)
                    }}
                </h1>
                <NuxtLink class="link" to="/mealplan"
                    >Manage Meal Plan</NuxtLink
                >
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
                    <div class="title-row">
                        <h2>Logged</h2>
                        <button @click="logOtherMeal">Log Other</button>
                    </div>
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
                    <div class="title-row">
                        <h2>Planned</h2>
                    </div>
                    <TodayMealCard
                        v-for="log in data.day.plannedMeals"
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
    flex-direction: row;
    gap: 1rem;
    width: 75%;
}

.title-row {
    display: flex;
    flex-direction: row;
    gap: 1rem;
    align-items: center;
}

.link {
    color: rgb(177, 177, 177);
    text-decoration: none;
    background-color: rgb(56, 56, 56);
    padding: 6px 12px;
    margin-top: 6px;
    border-radius: 0.25rem;
    font-weight: bold;
}

.link:hover {
    color: rgb(199, 199, 199);
    transition: color 0.3s;
    background-color: rgb(82, 82, 82);
}

.meals-section {
    display: flex;
    flex-direction: row;
    gap: 0.5rem;
    padding-top: 1rem;
}

.meals-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-top: 1rem;
}

.expected-header {
    display: flex;
    flex-direction: row;
}
</style>
