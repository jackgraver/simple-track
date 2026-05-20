import type { Meal } from "~/types/diet";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useRouter } from "vue-router";
import { useDietLogsToday } from "~/pages/home/queries/useDietLogsToday";
import {
    useLogPlannedMeal,
    useDeleteLoggedMeal,
} from "~/pages/home/queries/useMealMutations";
import { LOG_TYPE, EDIT_TYPE, EDIT_LOGGED_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { withDietDayOffsetQuery } from "~/utils/dateUtil";
import { mealHasQuickEntryFood } from "~/utils/dietMealQuickLog";
import QuickLogDialog from "~/pages/diet/components/dialog/QuickLogDialog.vue";

export function useDietDayMealHandlers(getOffset: () => number) {
    const router = useRouter();
    const { data } = useDietLogsToday(getOffset);

    const logPlannedMealMutation = useLogPlannedMeal(getOffset);
    const deleteLoggedMealMutation = useDeleteLoggedMeal(getOffset);

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
            router.push({
                name: "diet-log",
                query: withDietDayOffsetQuery(getOffset(), { type: LOG_TYPE }),
            });
            return;
        }
        if (meal && type === EDIT_TYPE && mealHasQuickEntryFood(meal)) {
            await dialogManager.custom<boolean>({
                title: "Edit quick log",
                component: QuickLogDialog,
                componentProps: {
                    dateOffset: getOffset(),
                    editingMeal: meal,
                },
            });
            return;
        }
        const rowDayId = data.value?.day.ID;
        const query = withDietDayOffsetQuery(getOffset(), {
            type,
            id: String(meal?.ID ?? ""),
        });
        if (
            rowDayId &&
            meal &&
            (type === EDIT_TYPE || type === EDIT_LOGGED_TYPE)
        ) {
            query.dayId = String(rowDayId);
        }
        router.push({
            name: "diet-log",
            query,
        });
    };

    const deleteLoggedMeal = async (meal: Meal) => {
        const name = meal.name.trim();
        const quick = mealHasQuickEntryFood(meal);
        const confirmed = await dialogManager.confirm({
            title: quick ? "Remove quick log" : "Delete Logged Meal",
            message: quick
                ? `Remove “${name}” from today's log? This one-off entry will be discarded.`
                : `Are you sure you want to delete “${name}”?`,
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

    return {
        data,
        logPlannedMeal,
        logMeal,
        deleteLoggedMeal,
    };
}
