import { apiClient } from "~/api/client";
import type { CompositeFood, DietDay, Meal, Plan, SavedMeal } from "~/types/diet";

export type DietLogsTodayResponse = {
    day: DietDay;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
    totalFat: number;
};

export type MealResponse = {
    meal: Meal;
};

export type CreateMealResponse = {
    meal_id: number;
};

export type CreateSavedMealResponse = {
    saved_meal_id: number;
};

export type CreateCompositeFoodResponse = {
    composite_food: CompositeFood;
};

export type LogMealResponse = {
    day: DietDay;
    totalCalories: number;
    totalProtein: number;
    totalFiber: number;
    totalCarbs: number;
    totalFat: number;
};

export type UpdatePlanMacrosPayload = {
    calories: number;
    protein: number;
    fiber: number;
    carbs: number;
    fat: number;
};

export type QuickLogPayload = {
    name: string;
    calories: number;
    protein: number;
    carbs: number;
    fat: number;
    fiber: number;
    offset?: number;
    replace_meal_id?: number;
    /** Retry-safe: the server creates nothing for a known id and returns the current day. */
    client_id?: string;
};

export async function quickLog(payload: QuickLogPayload): Promise<DietLogsTodayResponse> {
    const body: Record<string, unknown> = {
        name: payload.name,
        calories: payload.calories,
        protein: payload.protein,
        carbs: payload.carbs,
        fat: payload.fat,
        fiber: payload.fiber,
        offset: payload.offset ?? 0,
    };
    if (payload.client_id) body.client_id = payload.client_id;
    if (
        payload.replace_meal_id != null &&
        payload.replace_meal_id > 0
    ) {
        body.replace_meal_id = payload.replace_meal_id;
    }
    const response = await apiClient.post<DietLogsTodayResponse>(
        "/diet/meals/quick-log",
        body,
    );
    return response.data;
}

export async function getDietLogsToday(
    offset: number = 0,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.get<DietLogsTodayResponse>(
        "/diet/logs/today",
        {
            params: { offset },
        },
    );
    return response.data;
}

export async function getMonthPlannedSummary(
    monthOffset: number,
): Promise<number[]> {
    const response = await apiClient.get<{ planned_counts: number[] }>(
        "/diet/logs/month-planned-summary",
        { params: { monthoffset: monthOffset } },
    );
    return response.data.planned_counts;
}

/** `dayId` is the diet day row id (`day.ID`); 0 or omitted means today. */
export async function logPlannedMeal(
    mealId: number,
    dayId = 0,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.post<DietLogsTodayResponse>(
        "/diet/meals/meal/log-planned",
        dayId > 0 ? { meal_id: mealId, day_id: dayId } : { meal_id: mealId },
    );
    return response.data;
}

export async function deleteLoggedMeal(
    mealId: number,
    dayId: number,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.delete<DietLogsTodayResponse>(
        "/diet/meals/meal/logged",
        {
            data: {
                meal_id: mealId,
                day_id: dayId,
            },
        },
    );
    return response.data;
}

export async function deletePlannedMeal(
    plannedMealId: number,
    offset = 0,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.delete<DietLogsTodayResponse>(
        "/diet/meals/planned",
        {
            data: {
                planned_meal_id: plannedMealId,
                offset,
            },
        },
    );
    return response.data;
}

export async function addPlannedMealFromSaved(
    savedMealId: number,
    offset = 0,
    clientId?: string,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.post<DietLogsTodayResponse>(
        "/diet/meals/planned/from-saved",
        {
            saved_meal_id: savedMealId,
            offset,
            ...(clientId && { client_id: clientId }),
        },
    );
    return response.data;
}

/** Appends every planned meal of `fromOffset` to each day in `toOffsets` (after what's there). */
export async function copyPlannedDay(
    fromOffset: number,
    toOffsets: number[],
): Promise<{ copied_meals: number; days: number }> {
    const response = await apiClient.post<{ copied_meals: number; days: number }>(
        "/diet/meals/planned/copy-day",
        { from_offset: fromOffset, to_offsets: toOffsets },
    );
    return response.data;
}

export async function reorderPlannedMeals(
    plannedMealIds: number[],
    offset = 0,
): Promise<DietLogsTodayResponse> {
    const response = await apiClient.post<DietLogsTodayResponse>(
        "/diet/meals/planned/reorder",
        {
            planned_meal_ids: plannedMealIds,
            offset,
        },
    );
    return response.data;
}

export async function getMealById(id: number): Promise<MealResponse> {
    const response = await apiClient.get<Meal>(`/diet/meals/meal/${id}`);
    return { meal: response.data };
}

export type SavedMealItemPayload = {
    food_id: number;
    amount: number;
    group_id?: string;
    group_label?: string;
    composite_food_id?: number | null;
};

export async function createSavedMeal(payload: {
    name: string;
    items: SavedMealItemPayload[];
}): Promise<CreateSavedMealResponse> {
    const response = await apiClient.post<CreateSavedMealResponse>(
        "/diet/meals/saved-meal/new",
        payload,
    );
    return response.data;
}

export async function getSavedMeals(): Promise<SavedMeal[]> {
    const response = await apiClient.get<{ saved_meals: SavedMeal[] }>(
        "/diet/meals/saved-meal/all",
    );
    return response.data.saved_meals;
}

export async function getSavedMealById(id: number): Promise<SavedMeal> {
    const response = await apiClient.get<{ saved_meal: SavedMeal }>(
        `/diet/meals/saved-meal/${id}`,
    );
    return response.data.saved_meal;
}

export async function updateSavedMeal(
    savedMealId: number,
    payload: { name: string; items: SavedMealItemPayload[] },
): Promise<void> {
    await apiClient.put(`/diet/meals/saved-meal/${savedMealId}`, payload);
}

export async function deleteSavedMeal(
    savedMealId: number,
    options?: { force?: boolean },
): Promise<void> {
    const config =
        options?.force === true
            ? { params: { force: "true" as const } }
            : undefined;
    await apiClient.delete(`/diet/meals/saved-meal/${savedMealId}`, config);
}

export type SavedMealDeleteDependentsInfo = {
    reference_count: number;
};

export async function previewSavedMealDelete(
    savedMealId: number,
): Promise<SavedMealDeleteDependentsInfo> {
    const response = await apiClient.delete<SavedMealDeleteDependentsInfo>(
        `/diet/meals/saved-meal/${savedMealId}`,
    );
    return response.data;
}

export async function createCompositeFood(body: {
    name: string;
    items: { food_id: number; amount: number }[];
}): Promise<CreateCompositeFoodResponse> {
    const response = await apiClient.post<CreateCompositeFoodResponse>(
        "/diet/meals/composite-food/new",
        body,
    );
    return response.data;
}

export async function createMeal(
    meal: Meal,
    log: boolean,
    options?: { saveToLibrary?: boolean; offset?: number },
): Promise<CreateMealResponse> {
    const body: {
        meal: Meal;
        log: boolean;
        save_to_library?: boolean;
        offset?: number;
    } = {
        meal,
        log,
    };
    if (log) {
        body.save_to_library = options?.saveToLibrary === true;
        body.offset = options?.offset ?? 0;
    }
    const response = await apiClient.post<CreateMealResponse>(
        "/diet/meals/meal/new",
        body,
    );
    return response.data;
}

export async function logEditedMeal(
    meal: Meal,
    options?: { plannedSourceMealId?: number; dayId?: number },
): Promise<LogMealResponse> {
    const body: {
        meal: Meal;
        planned_source_meal_id?: number;
        day_id?: number;
    } = { meal };
    const pid = options?.plannedSourceMealId;
    if (pid != null && pid > 0) {
        body.planned_source_meal_id = pid;
    }
    const did = options?.dayId;
    if (did != null && did > 0) {
        body.day_id = did;
    }
    const response = await apiClient.post<LogMealResponse>(
        "/diet/meals/meal/logedited",
        body,
    );
    return response.data;
}

export async function updateLoggedMeal(
    meal: Meal,
    oldMealId: number,
    options?: { dayId?: number },
): Promise<LogMealResponse> {
    const body: { meal: Meal; oldMealID: number; day_id?: number } = {
        meal,
        oldMealID: oldMealId,
    };
    const did = options?.dayId;
    if (did != null && did > 0) {
        body.day_id = did;
    }
    const response = await apiClient.post<LogMealResponse>(
        "/diet/meals/meal/editlogged",
        body,
    );
    return response.data;
}

export async function updatePlanMacros(
    planId: number,
    payload: UpdatePlanMacrosPayload,
): Promise<{ plan: Plan }> {
    const response = await apiClient.put<{ plan: Plan }>(
        `/diet/plans/plan/${planId}`,
        payload,
    );
    return response.data;
}

export async function getAllPlans(): Promise<Plan[]> {
    const response = await apiClient.get<{ plans: Plan[] | null }>("/diet/plans/plan/all");
    return response.data.plans ?? [];
}
