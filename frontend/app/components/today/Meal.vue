<script setup lang="ts">
import type { Day, Meal } from "~/types/diet";
import TodayLogEditedDialog from "~/components/today/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import {
    ChevronDown,
    ChevronUp,
    ChevronLeft,
    ChevronRight,
} from "lucide-vue-next";

const router = useRouter();

const emit = defineEmits<{
    (e: "date-change", direction: "next" | "prev"): void;
}>();

const props = defineProps<{
    dateOffset: number;
}>();

const { data, pending, error } = useAPIGet<{
    day: Day;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
}>(`mealplan/today?offset=${props.dateOffset}`);

const logPlannedMeal = async (meal: Meal) => {
    const { response, error } = await useAPIPost<{
        day: Day;
        totalCalories: number;
        totalProtein: number;
        totalFiber: number;
        totalCarbs: number;
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
                totalCarbs: response.totalCarbs,
            };
        }
    }
};

const logMeal = async (
    meal: Meal | null,
    type: "edit" | "editlogged" | "create",
) => {
    if (type === "create") {
        router.push(`/logmeal?type=${type}`);
        return;
    }
    router.push(`/logmeal?type=${type}&id=${meal?.ID}`);
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
                totalCarbs: number;
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
                        totalCarbs: response.totalCarbs,
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
                totalCarbs: number;
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
                        totalCarbs: response.totalCarbs,
                    };
                }
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
                <button @click="emit('date-change', 'next')">
                    <ChevronLeft />
                </button>
                <h1 style="flex: 1">
                    {{
                        formatDate(data.day.date) +
                        ", " +
                        dayOfWeek(data.day.date)
                    }}
                </h1>
                <button @click="emit('date-change', 'prev')">
                    <ChevronRight />
                </button>
                <button @click="logMeal(null, 'create')">Log Meal</button>
            </div>
            <TodayBars
                :totalCalories="data?.totalCalories ?? 0"
                :totalProtein="data?.totalProtein ?? 0"
                :totalFiber="data?.totalFiber ?? 0"
                :totalCarbs="data?.totalCarbs ?? 0"
                :plannedCalories="data?.day.plan.calories ?? 0"
                :plannedProtein="data?.day.plan.protein ?? 0"
                :plannedFiber="data?.day.plan.fiber ?? 0"
                :plannedCarbs="data?.day.plan.carbs ?? 0"
            />
            <div class="meals-section">
                <div class="meals-container">
                    <h2>Logged</h2>
                    <span v-if="data.day.loggedMeals.length === 0">
                        Nothing logged yet.
                    </span>
                    <TodayMealCard
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
                <div class="meals-container">
                    <div class="small-title-row">
                        <h2>Planned</h2>
                        <button @click="prev"><ChevronUp /></button>
                        <button @click="next"><ChevronDown /></button>
                    </div>
                    <span v-if="visibleItems?.length === 0">
                        Nothing else planned.
                    </span>
                    <TodayMealCard
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

<style scoped>
.container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
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

.small-title-row button {
    margin-bottom: 0;
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
