import type { Meal } from "~/types/diet";
import LogEditedDialog from "~/pages/home/dialog/LogEditedDialog.vue";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useRouter } from "vue-router";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import {
    useLogPlannedMeal,
    useDeleteLoggedMeal,
    useEditLoggedMeal,
} from "~/pages/home/queries/useMealMutations";
import {
    EDIT_LOGGED_TYPE,
    EDIT_TYPE,
    LOG_TYPE,
} from "~/pages/diet/logmeal/logmealMode";

export function useDietDayMealHandlers(getOffset: () => number) {
    const router = useRouter();
    const { data } = useDietLogsToday(getOffset);

    const logPlannedMealMutation = useLogPlannedMeal(getOffset());
    const deleteLoggedMealMutation = useDeleteLoggedMeal(getOffset());
    const editLoggedMealMutation = useEditLoggedMeal(getOffset());

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
                componentProps: { meal },
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
                        error instanceof Error
                            ? error.message
                            : "Unknown error";
                    toast.push("Log Edited Failed! " + msg, "error");
                }
            })
            .catch((err) => {
                console.error("Dialog error:", err);
                toast.push("Dialog Error", "error");
            });
    };

    return {
        data,
        logPlannedMeal,
        logMeal,
        deleteLoggedMeal,
        editLogMeal,
    };
}
