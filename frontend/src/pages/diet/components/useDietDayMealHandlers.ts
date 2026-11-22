import type { Meal } from "~/types/diet";
import { toast } from "~/composables/toast/useToast";
import { dialogManager } from "~/composables/dialog/useDialog";
import { useRouter } from "vue-router";
import {
    useDietLogsToday,
    useLogPlannedMeal,
    useDeleteLoggedMeal,
    dayRef,
} from "~/api/diet/queries";
import { isTempId } from "~/api/diet/dayPatches";
import { newClientId } from "~/api/ids";
import { LOG_TYPE, EDIT_TYPE, EDIT_LOGGED_TYPE } from "~/pages/diet/logmeal/logmealMode";
import { withDietDayOffsetQuery } from "~/utils/dateUtil";
import { mealHasQuickEntryFood } from "~/utils/dietMealQuickLog";
import QuickLogDialog from "~/pages/diet/components/dialog/QuickLogDialog.vue";

export function useDietDayMealHandlers(getOffset: () => number) {
    const router = useRouter();
    const { data } = useDietLogsToday(getOffset);

    const logPlannedMealMutation = useLogPlannedMeal();
    const deleteLoggedMealMutation = useDeleteLoggedMeal();

    // Optimistic: the meal moves to the logged list (pending) and the totals move on tap. A failure
    // rolls back with a Retry toast.
    const logPlannedMeal = (meal: Meal) => {
        const dayId = data.value?.day.ID;
        if (!dayId || isTempId(meal.ID)) return;
        logPlannedMealMutation.mutate({
            mealId: meal.ID,
            dayId,
            ...dayRef(getOffset()),
            client_id: newClientId(),
        });
    };

    const logMeal = async (
        meal: Meal | null,
        type: typeof LOG_TYPE | typeof EDIT_TYPE | typeof EDIT_LOGGED_TYPE,
    ) => {
        // A row still saving has a temp id: editing it would send a fresh entry without
        // `replace_meal_id` and the server would keep both. Wait for the real id.
        if (meal && type !== LOG_TYPE && isTempId(meal.ID)) {
            toast.push("Still saving, try again in a moment");
            return;
        }
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
            tone: "danger",
            message: quick
                ? `Remove “${name}” from today's log? This one-off entry will be discarded.`
                : `Are you sure you want to delete “${name}”?`,
        });
        if (!confirmed) return;

        if (!data.value?.day.ID) {
            toast.push("Cannot delete: day ID not found", "error");
            return;
        }

        // A row still saving is removed now too: the delete queues behind its log (same day scope).
        deleteLoggedMealMutation.mutate({
            mealId: meal.ID,
            mealClientId: meal.client_id,
            dayId: data.value.day.ID,
            ...dayRef(getOffset()),
            client_id: newClientId(),
        });
    };

    return {
        data,
        logPlannedMeal,
        logMeal,
        deleteLoggedMeal,
    };
}
