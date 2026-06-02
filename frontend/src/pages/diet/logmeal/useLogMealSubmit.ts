import type { ComputedRef, Ref } from "vue";
import type { Meal } from "~/types/diet";
import { toast } from "~/composables/toast/useToast";
import {
    useCreateMeal,
    useCreateSavedMeal,
    useLogEditedMeal,
    useUpdateLoggedMeal,
    useUpdateSavedMeal,
} from "./queries/useMealMutations";
import { EDIT_VARIANT } from "./logmealMode";
import type { EditMealVariant } from "./logmealMode";
import { emptyMeal } from "./useEditableMeal";

type DraftStorage = {
    clearAndPause: () => void;
};

export function useLogMealSubmit(options: {
    meal: Ref<Meal>;
    dayOffset: ComputedRef<number>;
    editVariant: ComputedRef<EditMealVariant | null>;
    routeMealId: ComputedRef<number>;
    mealLogDayId: ComputedRef<number | undefined>;
    draftStorage: DraftStorage;
}) {
    const createMealMutation = useCreateMeal();
    const createSavedMealMutation = useCreateSavedMeal();
    const logEditedMealMutation = useLogEditedMeal();
    const updateLoggedMealMutation = useUpdateLoggedMeal();
    const updateSavedMealMutation = useUpdateSavedMeal();

    function resetAfterDraftSubmit() {
        options.draftStorage.clearAndPause();
        options.meal.value = emptyMeal();
    }

    async function logMealToDay(saveToLibrary: boolean) {
        const mealToCreate = { ...options.meal.value, ID: 0 };
        try {
            await createMealMutation.mutateAsync({
                meal: mealToCreate,
                log: true,
                saveToLibrary,
                offset: options.dayOffset.value,
            });
            toast.push(
                saveToLibrary
                    ? "Meal logged and saved for later!"
                    : "Meal logged!",
                "success",
            );
            resetAfterDraftSubmit();
        } catch (error: unknown) {
            const message =
                error instanceof Error ? error.message : String(error);
            toast.push("Log meal failed! " + message, "error");
        }
    }

    async function saveSavedMealTemplate() {
        const name = options.meal.value.name.trim();
        if (!name || options.meal.value.items.length === 0) {
            toast.push("Add a name and at least one food.", "error");
            return;
        }
        try {
            await createSavedMealMutation.mutateAsync({
                name,
                items: options.meal.value.items.map((i) => ({
                    food_id: i.food_id,
                    amount: i.amount,
                    group_id: i.group_id ?? "",
                    group_label: i.group_label ?? "",
                    composite_food_id: i.composite_food_id ?? null,
                })),
            });
            toast.push("Saved meal created!", "success");
            resetAfterDraftSubmit();
        } catch (error: unknown) {
            const message =
                error instanceof Error ? error.message : String(error);
            toast.push("Could not save meal. " + message, "error");
        }
    }

    async function logEditedMeal() {
        const mealToLog = { ...options.meal.value, ID: 0 };
        const plannedSourceMealId =
            options.editVariant.value === EDIT_VARIANT.planned &&
            options.routeMealId.value !== 0
                ? options.routeMealId.value
                : undefined;
        try {
            await logEditedMealMutation.mutateAsync({
                meal: mealToLog,
                plannedSourceMealId,
                dayId: options.mealLogDayId.value,
            });
            toast.push("Meal Logged Successfully!", "success");
        } catch (error: unknown) {
            const message =
                error instanceof Error ? error.message : String(error);
            toast.push("Log Edited Failed! " + message, "error");
        }
    }

    async function updateLoggedMeal() {
        const oldMealID = options.meal.value.ID;
        const mealToUpdate = { ...options.meal.value, ID: 0 };
        try {
            await updateLoggedMealMutation.mutateAsync({
                meal: mealToUpdate,
                oldMealId: oldMealID,
                dayId: options.mealLogDayId.value,
            });
            toast.push("Meal Updated Successfully!", "success");
        } catch (error: unknown) {
            const message =
                error instanceof Error ? error.message : String(error);
            toast.push("Update Failed! " + message, "error");
        }
    }

    async function saveEditedSavedMeal() {
        const name = options.meal.value.name.trim();
        if (!name || options.meal.value.items.length === 0) {
            toast.push("Add a name and at least one food.", "error");
            return;
        }
        if (options.routeMealId.value === 0) return;
        try {
            await updateSavedMealMutation.mutateAsync({
                savedMealId: options.routeMealId.value,
                name,
                items: options.meal.value.items.map((i) => ({
                    food_id: i.food_id,
                    amount: i.amount,
                    group_id: i.group_id ?? "",
                    group_label: i.group_label ?? "",
                    composite_food_id: i.composite_food_id ?? null,
                })),
            });
            toast.push("Saved meal updated!", "success");
        } catch (error: unknown) {
            const message =
                error instanceof Error ? error.message : String(error);
            toast.push("Could not update saved meal. " + message, "error");
        }
    }

    return {
        logMealToDay,
        saveSavedMealTemplate,
        logEditedMeal,
        updateLoggedMeal,
        saveEditedSavedMeal,
    };
}
